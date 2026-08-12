package main

// lifecycle_writer.go
//
// 预警生命周期写入器（Phase 2+3 合并实现）：
// 在 storage-adapter 进程内维护一份 per-device active fault_code 集合，
// 每收到一帧 predict 事件做 set-diff，只在状态变化时写入 hvac.warning_lifecycle：
//   - 新出现的 code → INSERT (open)
//   - 不再出现的 code → UPDATE end_time (close)
//   - 仍出现的 code → UPDATE last_seen_time（轻量心跳）
//
// 数据库层有部分唯一索引 UNIQUE(device_id, fault_code) WHERE end_time IS NULL 兜底，
// 即使应用层判定失误，也无法插入重复活跃行（migration 08-20260603）。
//
// 启动恢复：从 warning_lifecycle WHERE end_time IS NULL 拉取已有活跃集合，
// 进程重启不会导致整批重发或丢状态。

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	lifecycleInsertSQL = `
INSERT INTO hvac.warning_lifecycle (
    device_id, line_id, train_id, carriage_id, unit_id,
    fault_code, fault_name, warn_code, severity,
    start_time, last_seen_time, trigger_snapshot, source
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
ON CONFLICT (device_id, fault_code) WHERE end_time IS NULL DO UPDATE
SET last_seen_time = EXCLUDED.last_seen_time;
`

	lifecycleCloseSQL = `
UPDATE hvac.warning_lifecycle
   SET end_time = $1,
       last_seen_time = $1
 WHERE device_id = $2
   AND fault_code = $3
   AND end_time IS NULL;
`

	lifecycleTouchSQL = `
UPDATE hvac.warning_lifecycle
   SET last_seen_time = $1
 WHERE device_id = $2
   AND fault_code = $3
   AND end_time IS NULL;
`

	lifecycleRecoverSQL = `
SELECT device_id, fault_code
  FROM hvac.warning_lifecycle
 WHERE end_time IS NULL;
`

	// lifecycleSweepSQL 由后台 sweeper 周期性执行，把 last_seen_time 停摆超过 grace 的活跃行
	// 主动 close，end_time := last_seen_time（符合 DDL 08-20260603 中 last_seen_time 列的
	// COMMENT 语义："若连续 N 秒未刷新，可由清扫器兜底将 end_time = last_seen_time"）。
	// RETURNING 用于回收被关闭的 (device_id, fault_code) 以同步内存 active map，
	// 避免下次同 code 帧到来时 diff 判成 touch 却在 DB 侧无匹配行。
	lifecycleSweepSQL = `
UPDATE hvac.warning_lifecycle
   SET end_time = last_seen_time
 WHERE end_time IS NULL
   AND last_seen_time < now() - $1::interval
RETURNING device_id, fault_code;
`
)

// PredictHitMeta 是 lifecycle 关心的最小 hit 信息。
type PredictHitMeta struct {
	FaultCode string
	FaultName string
	Severity  int16
	Payload   json.RawMessage // 触发时刻 hit 原文（含 trigger_condition_snapshot）
}

// PredictFrame 是 nb67 单条 predict 消息的设备级摘要。
// 关键：空帧（Hits=nil/empty）也合法 —— 表示 "该 device 在 EventTimeText 时刻没有任何活跃预警"，
// 是 lifecycle 状态机判定 close 的唯一信号源。
type PredictFrame struct {
	DeviceID      string
	LineID        int32
	TrainID       int32
	CarriageID    int32
	EventTimeText string
	Hits          []PredictHitMeta
}

// LifecycleWriter 维护 per-device 的活跃 fault_code 集合，
// 并把状态变化落 hvac.warning_lifecycle。
type LifecycleWriter struct {
	pool *pgxpool.Pool
	mu   sync.Mutex
	// active[deviceID] = set(fault_code)
	active map[string]map[string]struct{}
}

func newLifecycleWriter(pool *pgxpool.Pool) *LifecycleWriter {
	return &LifecycleWriter{
		pool:   pool,
		active: make(map[string]map[string]struct{}),
	}
}

// Recover 从 DB 拉取所有 end_time IS NULL 的活跃行，重建内存集合。
// 应在 Kafka 消费开始前调用一次。
func (w *LifecycleWriter) Recover(ctx context.Context) error {
	rows, err := w.pool.Query(ctx, lifecycleRecoverSQL)
	if err != nil {
		return fmt.Errorf("lifecycle recover query: %w", err)
	}
	defer rows.Close()

	w.mu.Lock()
	defer w.mu.Unlock()
	w.active = make(map[string]map[string]struct{})

	total := 0
	for rows.Next() {
		var deviceID, faultCode string
		if err := rows.Scan(&deviceID, &faultCode); err != nil {
			return fmt.Errorf("lifecycle recover scan: %w", err)
		}
		set, ok := w.active[deviceID]
		if !ok {
			set = make(map[string]struct{})
			w.active[deviceID] = set
		}
		set[faultCode] = struct{}{}
		total++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("lifecycle recover iter: %w", err)
	}
	log.Printf("[INFO] lifecycle recover: %d active warnings across %d devices", total, len(w.active))
	return nil
}

// ProcessFrames 对一批 PredictFrame 做 per-device, per-frame 的 diff，并落库。
// 同一 device 在同一 batch 内可能多帧，按 EventTimeText 升序逐帧推进状态。
func (w *LifecycleWriter) ProcessFrames(ctx context.Context, frames []PredictFrame) error {
	if len(frames) == 0 {
		return nil
	}
	// 分组：device → []frame
	byDevice := make(map[string][]PredictFrame)
	for _, f := range frames {
		if f.DeviceID == "" {
			continue
		}
		byDevice[f.DeviceID] = append(byDevice[f.DeviceID], f)
	}
	for device, list := range byDevice {
		// 按 EventTimeText 升序（RFC3339-like 字典序 == 时间序）
		sortFramesByTime(list)
		for _, frame := range list {
			eventTime, ok := parseEventTime(frame.EventTimeText, true, time.Now().UTC())
			_ = ok
			if err := w.applyFrame(ctx, device, frame, eventTime); err != nil {
				return err
			}
		}
	}
	return nil
}

// applyFrame 对单个 (device, event_time) 帧做 diff 与写库。
func (w *LifecycleWriter) applyFrame(ctx context.Context, deviceID string, frame PredictFrame, eventTime time.Time) error {
	w.mu.Lock()
	prev, ok := w.active[deviceID]
	if !ok {
		prev = make(map[string]struct{})
		w.active[deviceID] = prev
	}

	currCodes := make(map[string]PredictHitMeta, len(frame.Hits))
	for _, h := range frame.Hits {
		if h.FaultCode == "" {
			continue
		}
		currCodes[h.FaultCode] = h
	}

	var toOpen []PredictHitMeta
	var toTouch []PredictHitMeta
	for _, h := range frame.Hits {
		if h.FaultCode == "" {
			continue
		}
		if _, exists := prev[h.FaultCode]; exists {
			toTouch = append(toTouch, h)
		} else {
			toOpen = append(toOpen, h)
		}
	}

	var toClose []string
	for code := range prev {
		if _, stillActive := currCodes[code]; !stillActive {
			toClose = append(toClose, code)
		}
	}

	// 推进内存状态（即便 DB 写失败也保持单调推进，避免重启风暴）
	for _, h := range toOpen {
		prev[h.FaultCode] = struct{}{}
	}
	for _, c := range toClose {
		delete(prev, c)
	}
	w.mu.Unlock()

	if len(toOpen) == 0 && len(toClose) == 0 && len(toTouch) == 0 {
		return nil
	}

	tx, err := w.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("lifecycle tx begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, h := range toOpen {
		warnCode := inferWarnCode(h.FaultCode)
		unitID := inferUnitID(h.FaultCode)
		if _, err := tx.Exec(ctx, lifecycleInsertSQL,
			deviceID, frame.LineID, frame.TrainID, frame.CarriageID, unitID,
			h.FaultCode, h.FaultName, warnCode, h.Severity,
			eventTime, eventTime, h.Payload, "connect-rule-v2",
		); err != nil {
			return fmt.Errorf("lifecycle insert %s/%s: %w", deviceID, h.FaultCode, err)
		}
	}
	for _, h := range toTouch {
		if _, err := tx.Exec(ctx, lifecycleTouchSQL, eventTime, deviceID, h.FaultCode); err != nil {
			return fmt.Errorf("lifecycle touch %s/%s: %w", deviceID, h.FaultCode, err)
		}
	}
	for _, code := range toClose {
		if _, err := tx.Exec(ctx, lifecycleCloseSQL, eventTime, deviceID, code); err != nil {
			return fmt.Errorf("lifecycle close %s/%s: %w", deviceID, code, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("lifecycle tx commit: %w", err)
	}
	if len(toOpen) > 0 || len(toClose) > 0 {
		log.Printf("[INFO] lifecycle %s: +%d -%d ~%d", deviceID, len(toOpen), len(toClose), len(toTouch))
	}
	return nil
}

// inferWarnCode 用 HVAC 编码序号推 warning_config.warn_code，对齐 BFF 的 hvacSeqToWarnCode。
// 输入：HVACxyz，xyz % 100 = seq；带 _v/_c 后缀的特殊处理冷媒泄漏。
func inferWarnCode(faultCode string) string {
	if !strings.HasPrefix(faultCode, "HVAC") {
		return ""
	}
	numStr := strings.TrimPrefix(faultCode, "HVAC")
	if idx := strings.Index(numStr, "_"); idx >= 0 {
		numStr = numStr[:idx]
	}
	var n int
	fmt.Sscanf(numStr, "%d", &n)
	seq := n % 100
	switch {
	case seq >= 1 && seq <= 4:
		if strings.HasSuffix(faultCode, "_v") {
			return "WARN_REFRIGERANT_LEAK_VENT"
		}
		return "WARN_REFRIGERANT_LEAK_COOLING"
	case seq >= 5 && seq <= 6:
		return "WARN_COOLING_SYSTEM"
	case seq >= 7 && seq <= 8:
		return "WARN_TEMP_SENSOR"
	case seq == 9:
		return "WARN_CABIN_OVERHEAT"
	case seq >= 10 && seq <= 11:
		return "WARN_FILTER_CLOG"
	case seq >= 12 && seq <= 15:
		return "WARN_EF_CURRENT"
	case seq >= 16 && seq <= 19:
		return "WARN_CF_CURRENT"
	case seq == 20:
		return "WARN_EXUF_CURRENT"
	case seq >= 21 && seq <= 24:
		return "WARN_CP_CURRENT"
	case seq == 25, seq == 26:
		return "WARN_AQ_CO2"
	}
	return ""
}

// inferUnitID 用 fault_code 序号推 unit_id（机组号 1/2），单数侧→机组1 / 双数侧→机组2。
func inferUnitID(faultCode string) *int16 {
	if !strings.HasPrefix(faultCode, "HVAC") {
		return nil
	}
	numStr := strings.TrimPrefix(faultCode, "HVAC")
	if idx := strings.Index(numStr, "_"); idx >= 0 {
		numStr = numStr[:idx]
	}
	var n int
	fmt.Sscanf(numStr, "%d", &n)
	seq := n % 100
	var u int16
	switch seq {
	case 1, 2, 5, 10, 12, 13, 16, 17, 21, 22, 25:
		u = 1
	case 3, 4, 6, 11, 14, 15, 18, 19, 23, 24, 26:
		u = 2
	default:
		return nil
	}
	return &u
}

// SweepClosed 是 RunSweeper 一次扫描返回的被关闭条目，
// 便于内存 active map 同步（也供单测断言 sweep 效果）。
type SweepClosed struct {
	DeviceID  string
	FaultCode string
}

// RunSweeper 是一个阻塞式后台循环，按 interval 周期性调用 sweepOnce，
// 直到 ctx 取消。用法：main.go 里 `go lifecycle.RunSweeper(ctx, ...)`。
//
// grace 是 "活跃行 last_seen_time 落后 wall clock 多久后视为需要兜底 close" 的阈值，
// 必须显著大于生产车厢正常帧间隔的 P99，避免误关正在活跃的预警。
//
// 该 sweeper 是对 issue #26 根因的兜底修复：
//   现有 close 语义完全依赖 nb67_event_processor 的 prevPredictHadHits 状态机
//   在 "上一帧有命中、本帧清空" 的边沿主动发一次 hits=[] 过渡帧。生产上真实车厢
//   停发 / 断线 / 车厢下线时，这条过渡帧永远不产生 → warning_lifecycle 里
//   end_time 永久为 NULL → 前端一直显示 "活跃预警" 直到同车厢下一次触发。
//   sweeper 打破这一硬依赖：只要 last_seen_time 停摆足够久，即使 close 帧从未到达，
//   也在 grace 秒后把 end_time 补齐。
func (w *LifecycleWriter) RunSweeper(ctx context.Context, interval, grace time.Duration) {
	if interval <= 0 || grace <= 0 {
		log.Printf("[WARN] lifecycle sweeper disabled (interval=%v grace=%v)", interval, grace)
		return
	}
	log.Printf("[INFO] lifecycle sweeper started: interval=%v grace=%v", interval, grace)
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] lifecycle sweeper stopped")
			return
		case <-tick.C:
			closed, err := w.sweepOnce(ctx, grace)
			if err != nil {
				log.Printf("[WARN] lifecycle sweep: %v", err)
				continue
			}
			if len(closed) > 0 {
				w.applySweepClosed(closed)
				log.Printf("[INFO] lifecycle sweep: closed %d stale warnings (grace=%v)", len(closed), grace)
			}
		}
	}
}

// sweepOnce 执行一次 UPDATE ... RETURNING 并返回被 close 的条目。
// 分离出来便于单测（RunSweeper 循环侧不便直接测）。
func (w *LifecycleWriter) sweepOnce(ctx context.Context, grace time.Duration) ([]SweepClosed, error) {
	// pgx 的 $N::interval 需要字符串形式；用秒粒度避免 duration.String() 的 "1h30m0s"
	// 之类 postgres 不认识的写法。
	graceStr := fmt.Sprintf("%d seconds", int(grace.Seconds()))
	rows, err := w.pool.Query(ctx, lifecycleSweepSQL, graceStr)
	if err != nil {
		return nil, fmt.Errorf("sweep query: %w", err)
	}
	defer rows.Close()
	var closed []SweepClosed
	for rows.Next() {
		var c SweepClosed
		if err := rows.Scan(&c.DeviceID, &c.FaultCode); err != nil {
			return nil, fmt.Errorf("sweep scan: %w", err)
		}
		closed = append(closed, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sweep iter: %w", err)
	}
	return closed, nil
}

// applySweepClosed 把 sweeper 关闭的 (device, code) 从内存 active map 里删掉，
// 与 DB 保持一致；否则同一 code 下一帧到来时会被判成 touch，UPDATE 无匹配行、
// 无害但会持续 log noise。
func (w *LifecycleWriter) applySweepClosed(closed []SweepClosed) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, c := range closed {
		if set, ok := w.active[c.DeviceID]; ok {
			delete(set, c.FaultCode)
			if len(set) == 0 {
				delete(w.active, c.DeviceID)
			}
		}
	}
}

// sortFramesByTime 按 EventTimeText 字符串升序排序（RFC3339 字典序 == 时间序）。
func sortFramesByTime(frames []PredictFrame) {
	for i := 1; i < len(frames); i++ {
		for j := i; j > 0 && frames[j-1].EventTimeText > frames[j].EventTimeText; j-- {
			frames[j-1], frames[j] = frames[j], frames[j-1]
		}
	}
}
