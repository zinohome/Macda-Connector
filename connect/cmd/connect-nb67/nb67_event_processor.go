package main

// nb67_event_processor.go
//
// NB67 空调事件构建处理器（Go 原生实现）
//
// 将原先写在 YAML mapping: | 块中的 Bloblang 脚本全部迁移到 Go 代码，
// 解决 Bloblang 解析器对注释中多字节（中文）字符的误判问题，
// 同时获得完整的类型检查、单元测试和 IDE 支持。
//
// 注册处理器名称：nb67_event_builder
// 输入消息：nb67_parser 输出的 signal-parsed JSON（含 raw 字段）
// 输出消息：三个子事件聚合体，YAML 通过 fan_out + mapping 分拣到三个 topic
//
// 事件码规范：
//   HVAC 预警码 = "HVAC" + string(carriage_id*100 + seq)  （来源：NB67 空调预警码表 20240802）
//   部件寿命码  = string(carriage_id*1000 + 50000 + offset)（来源：NB67 空调部件码表 20240802）
//
// Author: Macda Connect Team
// Google Go Style Guide: https://google.github.io/styleguide/go/

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/benthosdev/benthos/v4/public/service"
)

// ============================================================
// 寿命阈值常量（只需在此处修改）
// ============================================================

const (
	// 风机类（通风机 / 冷凝风机 / 废排风机）额定寿命 25000h = 90000000s
	fanLifeS = 90_000_000 // 额定寿命（秒）
	fanWarnS = 67_500_000 // 75% 预警
	fanCritS = 81_000_000 // 90% 严重

	// 压缩机额定寿命 50000h = 180000000s
	cpLifeS = 180_000_000
	cpWarnS = 135_000_000
	cpCritS = 162_000_000

	// 阀门类额定寿命 1000000 次
	valveLifeN = 1_000_000
	valveWarnN = 750_000
	valveCritN = 900_000
)

// ============================================================
// 数据结构定义
// ============================================================

// EventMeta 事件元数据，对应 signal-parsed 消息头。
type EventMeta struct {
	SchemaVersion string `json:"schema_version"`
	LineID        string `json:"line_id"`
	TrainID       string `json:"train_id"`
	CarriageID    int    `json:"carriage_id"`
	DeviceID      string `json:"device_id"`
	EventTimeText string `json:"event_time_text"`
	IngestTime    string `json:"ingest_time"`
	ProcessTime   string `json:"process_time"`
}

// PredictHit 预警命中条目（基于算法规则）。
type PredictHit struct {
	Code     string `json:"code"`     // e.g. "HVAC301"
	Name     string `json:"name"`     // 中文名称
	Severity int    `json:"severity"` // 3=高 2=中 1=低
}

// AlarmHit 原生故障位命中条目（直接映射 binary 故障位）。
type AlarmHit struct {
	Code  string `json:"code"`  // e.g. "blpflt_comp_u11"
	Name  string `json:"name"`  // 中文名称
	Level int    `json:"level"` // 1=严重 2=一般
}

// LifeHit 部件寿命预警条目。
type LifeHit struct {
	Code     string `json:"code"`     // 部件码
	Name     string `json:"name"`     // 中文名称
	Severity int    `json:"severity"` // 3=高 2=中
	Value    int64  `json:"value"`    // 当前累计值（秒或次）
	Limit    int64  `json:"limit"`    // 额定寿命（秒或次）
}

// SubEvent 单个子事件，用于输出到对应 topic。
type SubEvent struct {
	EventMeta EventMeta   `json:"event_meta"`
	Hits      interface{} `json:"hits"`   // []PredictHit | []AlarmHit | []LifeHit
	Source    string      `json:"source"` // 来源标识
}

// EventOutput 处理器输出的聚合事件包，YAML fan_out 分拣用。
type EventOutput struct {
	PredictEvent SubEvent `json:"predict_event"`
	AlarmEvent   SubEvent `json:"alarm_event"`
	LifeEvent    SubEvent `json:"life_event"`
}

// parsedInput 是从上游 signal-parsed 消息解析的输入结构。
// 仅保留事件构建所需字段，Raw 字段来自 nb67_parser 输出的 raw 对象。
type parsedInput struct {
	LineID        string         `json:"line_id"`
	TrainID       string         `json:"train_id"`
	CarriageID    int            `json:"carriage_id"`
	DeviceID      string         `json:"device_id"`
	EventTimeText string         `json:"event_time_text"`
	IngestTime    string         `json:"ingest_time"`
	Raw           map[string]any `json:"raw"`
}

// ============================================================
// 处理器注册与实现
// ============================================================

// ============================================================
// 状态管理：规则计时器
// ============================================================

type ruleState struct {
	firstSeen time.Time
}

// NB67EventProcessor 增加了状态表，用于判断规则持续时间。
type NB67EventProcessor struct {
	// key: DeviceID + RuleCode
	// value: *ruleState
	states  sync.Map
	logger  *service.Logger
	runtime string // ENV "RUNTIME": "DEV" | "PRD"
}

// checkRule 判定规则是否满足持续时间要求，使用消息中的 currentTime。
func (p *NB67EventProcessor) checkRule(condition bool, duration time.Duration, deviceID string, ruleCode string, currentTime time.Time) bool {
	key := deviceID + ":" + ruleCode
	if !condition {
		p.states.Delete(key)
		return false
	}

	val, loaded := p.states.LoadOrStore(key, &ruleState{firstSeen: currentTime})
	state := val.(*ruleState)

	if !loaded {
		return duration <= 0
	}

	// 计算消息间的时间差，而不是系统运行时间差
	return currentTime.Sub(state.firstSeen) >= duration
}

// init 在程序启动时自动注册处理器。
func init() {
	err := service.RegisterProcessor(
		"nb67_event_builder",
		service.NewConfigSpec().
			Summary("NB67 空调事件构建处理器").
			Description("支持状态化持续时间判定的事件构建器"),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			rt := os.Getenv("RUNTIME")
			if rt == "" {
				rt = "PRD"
			}
			return &NB67EventProcessor{
				logger:  mgr.Logger(),
				runtime: rt,
			}, nil
		},
	)
	if err != nil {
		panic(fmt.Sprintf("注册 nb67_event_builder 处理器失败: %v", err))
	}
}

// Process 实现 service.Processor 接口，处理每条输入消息。
func (p *NB67EventProcessor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error) {
	// 解析输入 JSON
	rawBytes, err := msg.AsBytes()
	if err != nil {
		// 读取字节失败，通常是内核或内存极端情况，直接丢弃
		return service.MessageBatch{}, nil
	}

	var input parsedInput
	if err := json.Unmarshal(rawBytes, &input); err != nil {
		// 【关键修复】：如果此 Processor 报错返回 error，Benthos 会透传原始巨大消息。
		// 为了保护下游 Topic，我们此处拦截错误并返回空 Batch 丢弃它。
		p.logger.Errorf("NB67处理器解析 JSON 失败（可能是非标准数据），已拦截丢弃，防止污染 Topic: %v", err)
		return service.MessageBatch{}, nil
	}

	// 极其重要：如果 raw 为空，说明不是合法的 signal-parsed 数据，丢弃
	if len(input.Raw) == 0 {
		return service.MessageBatch{}, nil
	}

	// 构建事件元数据
	meta := EventMeta{
		SchemaVersion: "nb67.event",
		LineID:        input.LineID,
		TrainID:       input.TrainID,
		CarriageID:    input.CarriageID,
		DeviceID:      input.DeviceID,
		EventTimeText: input.EventTimeText,
		IngestTime:    input.IngestTime,
		ProcessTime:   time.Now().UTC().Format(time.RFC3339Nano),
	}

	// 【核心修复】：根据 RUNTIME 环境选择时间源
	var currentTime time.Time
	var parseErr error
	if p.runtime == "DEV" {
		// DEV 模式使用入库时间（RFC3339 格式）
		currentTime, parseErr = time.Parse(time.RFC3339, input.IngestTime)
	} else {
		// PRD 模式使用原始物理时间（自定义文本格式）
		currentTime, parseErr = time.Parse("2006-01-02 15:04:05", input.EventTimeText)
	}

	if parseErr != nil {
		p.logger.Warnf("解析时间源失败 [Mode:%s, Ingest:%s, Event:%s]: %v", p.runtime, input.IngestTime, input.EventTimeText, parseErr)
		currentTime = time.Now()
	}

	// 构建三类事件命中列表
	predictHits := p.buildPredictHits(input.Raw, input.CarriageID, input.DeviceID, currentTime)
	alarmHits := buildAlarmHits(input.Raw)
	lifeHits := buildLifeHits(input.Raw, input.CarriageID)

	// 如果三类命中均为空，直接拦截，不向下游输出任何内容
	if len(predictHits) == 0 && len(alarmHits) == 0 && len(lifeHits) == 0 {
		return service.MessageBatch{}, nil
	}

	// 聚合输出
	output := EventOutput{
		PredictEvent: SubEvent{EventMeta: meta, Hits: predictHits, Source: "connect-rule-v2"},
		AlarmEvent:   SubEvent{EventMeta: meta, Hits: alarmHits, Source: "raw-fault-bit"},
		LifeEvent:    SubEvent{EventMeta: meta, Hits: lifeHits, Source: "part-life-v2"},
	}

	outBytes, err := json.Marshal(output)
	if err != nil {
		p.logger.Errorf("NB67处理器序列化 JSON 失败: %v", err)
		return service.MessageBatch{}, nil
	}

	// 替换原始消息内容
	outMsg := msg.Copy()
	outMsg.SetBytes(outBytes)
	return service.MessageBatch{outMsg}, nil
}

// Close 实现 service.Processor 接口。
func (p *NB67EventProcessor) Close(ctx context.Context) error {
	return nil
}

// ============================================================
// 辅助函数：从 raw map 安全读取数值
// ============================================================

// rawInt 从 raw map 读取整数字段，找不到或类型不匹配时返回 0。
func rawInt(raw map[string]any, key string) int64 {
	v, ok := raw[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	}
	return 0
}

// rawBool 从 raw map 读取布尔字段，找不到或类型不匹配时返回 false。
func rawBool(raw map[string]any, key string) bool {
	v, ok := raw[key]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

// hvacCode 生成 HVAC 预警码字符串，格式：HVAC{carriageID*100+seq}
func hvacCode(hvacBase int, seq int) string {
	return fmt.Sprintf("HVAC%d", hvacBase+seq)
}

// ============================================================
// buildPredictHits：HVAC 算法预警（HVAC_01 ~ HVAC_26）
// 来源：NB67 空调预警码表 20240802
// 官方码格式：HVAC{carriage_id * 100 + seq}
//   carriage_id 1~6 对应 Tc1/Mp1/M1/M2/Mp2/Tc2
//   seq 01~26 对应 26 种预警类型
// ============================================================

// buildPredictHits 全量实现 HVAC101 ~ HVAC126 业务逻辑
func (p *NB67EventProcessor) buildPredictHits(raw map[string]any, carriageID int, deviceID string, currentTime time.Time) []PredictHit {
	hits := make([]PredictHit, 0)
	if len(raw) == 0 {
		return hits
	}
	base := carriageID * 100

	// 辅助变量
	wModeU1 := rawInt(raw, "WmodeU1")
	wModeU2 := rawInt(raw, "WmodeU2")

	// ================================================================
	// 1. 冷媒泄漏预警 (HVAC_01 ~ HVAC_04)
	// ================================================================
	checkRefLeak := func(mode int64, hvacSeq int, name string) {
		code := hvacCode(base, hvacSeq)
		uIdx := (hvacSeq + 1) / 2
		sIdx := (hvacSeq+1)%2 + 1
		fcp := rawInt(raw, fmt.Sprintf("FCpU%d%d", uIdx, sIdx))
		suckp := rawInt(raw, fmt.Sprintf("SuckpU%d%d", uIdx, sIdx))
		highp := rawInt(raw, fmt.Sprintf("HighpressU%d%d", uIdx, sIdx))

		// 条件1：制冷模式 + 频率>30Hz + 吸气<2.0bar -> 持续5分钟
		isCoolingLeak := (mode == 2 || mode == 3) && fcp > 300 && suckp < 20
		if p.checkRule(isCoolingLeak, 5*time.Minute, deviceID, code+"_c", currentTime) {
			hits = append(hits, PredictHit{Code: code, Name: name, Severity: 3})
			return
		}
		// 条件2：通风模式 + 高压<5bar -> 持续15分钟
		isVentLeak := mode == 1 && highp < 50
		if p.checkRule(isVentLeak, 15*time.Minute, deviceID, code+"_v", currentTime) {
			hits = append(hits, PredictHit{Code: code, Name: name, Severity: 3})
		}
	}
	checkRefLeak(wModeU1, 1, "机组1系统1冷媒泄露预警")
	checkRefLeak(wModeU1, 2, "机组1系统2冷媒泄露预警")
	checkRefLeak(wModeU2, 3, "机组2系统1冷媒泄露预警")
	checkRefLeak(wModeU2, 4, "机组2系统2冷媒泄露预警")

	// ================================================================
	// 2. 制冷系统预警 (HVAC_05 ~ HVAC_06)
	// ================================================================
	checkCpSys := func(uIdx int, name string) {
		code := hvacCode(base, uIdx+4)
		f1 := rawInt(raw, fmt.Sprintf("FCpU%d1", uIdx))
		f2 := rawInt(raw, fmt.Sprintf("FCpU%d2", uIdx))
		i1 := rawInt(raw, fmt.Sprintf("ICpU%d1", uIdx))
		i2 := rawInt(raw, fmt.Sprintf("ICpU%d2", uIdx))
		sp1 := rawInt(raw, fmt.Sprintf("SpU%d1", uIdx))
		sp2 := rawInt(raw, fmt.Sprintf("SpU%d2", uIdx))

		// 条件1：同频电流差 > 2A -> 持续3分钟
		isCurrentDiff := f1 == f2 && f1 > 0 && (i1-i2 > 20 || i1-i2 < -20)
		if p.checkRule(isCurrentDiff, 3*time.Minute, deviceID, code+"_i", currentTime) {
			hits = append(hits, PredictHit{Code: code, Name: name, Severity: 3})
			return
		}
		// 条件2：运行 > 5min 后，过热度异常 -> 持续10分钟
		isRunning := f1 > 0 || f2 > 0
		hasBeenRunning := p.checkRule(isRunning, 5*time.Minute, deviceID, code+"_run", currentTime)
		isSpErr := hasBeenRunning && (sp1 > 200 || sp1 < -80 || sp2 > 200 || sp2 < -80)
		if p.checkRule(isSpErr, 10*time.Minute, deviceID, code+"_sp", currentTime) {
			hits = append(hits, PredictHit{Code: code, Name: name, Severity: 3})
		}
	}
	checkCpSys(1, "机组1制冷系统预警")
	checkCpSys(2, "机组2制冷系统预警")

	// ================================================================
	// 3. 传感器预警 (HVAC_07 ~ HVAC_11)
	// ================================================================
	// HVAC_07/08: 温差 > 8℃ -> 持续 5 分钟
	fasCondition := rawInt(raw, "FasU1")-rawInt(raw, "FasU2") > 80 || rawInt(raw, "FasU1")-rawInt(raw, "FasU2") < -80
	if p.checkRule(fasCondition, 5*time.Minute, deviceID, hvacCode(base, 7), currentTime) {
		hits = append(hits, PredictHit{Code: hvacCode(base, 7), Name: "新风温度传感器预警", Severity: 3})
	}
	rasCondition := rawInt(raw, "RasU1")-rawInt(raw, "RasU2") > 80 || rawInt(raw, "RasU1")-rawInt(raw, "RasU2") < -80
	if p.checkRule(rasCondition, 5*time.Minute, deviceID, hvacCode(base, 8), currentTime) {
		hits = append(hits, PredictHit{Code: hvacCode(base, 8), Name: "回风温度传感器预警", Severity: 3})
	}

	// HVAC_09: 车厢超温预警 -> 系统正常运行 > 20min，且车温 > 4℃ -> 持续 2 分钟
	coolingNormal := len(buildAlarmHits(raw)) == 0 && (wModeU1 == 2 || wModeU2 == 2)
	sysRunningLong := p.checkRule(coolingNormal, 20*time.Minute, deviceID, "cooling_normal_20", currentTime)
	isOvertemp := sysRunningLong && (rawInt(raw, "Tveh1") > 40 || rawInt(raw, "Tveh2") > 40)
	if p.checkRule(isOvertemp, 2*time.Minute, deviceID, hvacCode(base, 9), currentTime) {
		hits = append(hits, PredictHit{Code: hvacCode(base, 9), Name: "车厢温度超温预警", Severity: 3})
	}

	// HVAC_10/11: 压差 > 300Pa -> 持续 30 分钟
	if p.checkRule(rawBool(raw, "CfbkEfU11") && rawInt(raw, "PresdiffU1") > 3000 && rawInt(raw, "PresdiffU1") < 32767, 30*time.Minute, deviceID, hvacCode(base, 10), currentTime) {
		hits = append(hits, PredictHit{Code: hvacCode(base, 10), Name: "机组1滤网脏堵预警", Severity: 2})
	}
	if p.checkRule(rawBool(raw, "CfbkEfU21") && rawInt(raw, "PresdiffU2") > 3000 && rawInt(raw, "PresdiffU2") < 32767, 30*time.Minute, deviceID, hvacCode(base, 11), currentTime) {
		hits = append(hits, PredictHit{Code: hvacCode(base, 11), Name: "机组2滤网脏堵预警", Severity: 2})
	}

	// ================================================================
	// 4. 风机电流预警 (HVAC_12 ~ HVAC_20) -> 持续 10 分钟
	// ================================================================
	checkFanI := func(cfbkField, iField string, threshold int64, seq int, name string) {
		code := hvacCode(base, seq)
		isOverI := rawBool(raw, cfbkField) && rawInt(raw, iField) > threshold
		if p.checkRule(isOverI, 10*time.Minute, deviceID, code, currentTime) {
			hits = append(hits, PredictHit{Code: code, Name: name, Severity: 3})
		}
	}
	checkFanI("CfbkEfU11", "IEfU11", 20, 12, "机组1通风机1电流预警")
	checkFanI("CfbkEfU11", "IEfU12", 20, 13, "机组1通风机2电流预警")
	checkFanI("CfbkEfU21", "IEfU21", 20, 14, "机组2通风机1电流预警")
	checkFanI("CfbkEfU21", "IEfU22", 20, 15, "机组2通风机2电流预警")
	checkFanI("CfbkCfU11", "ICfU11", 29, 16, "机组1冷凝风机1电流预警")
	checkFanI("CfbkCfU11", "ICfU12", 29, 17, "机组1冷凝风机2电流预警")
	checkFanI("CfbkCfU21", "ICfU21", 29, 18, "机组2冷凝风机1电流预警")
	checkFanI("CfbkCfU21", "ICfU22", 29, 19, "机组2冷凝风机2电流预警")
	checkFanI("CfbkExufan", "IExufan", 40, 20, "废排风机电流预警")

	// ================================================================
	// 5. 压缩机电流预警 (HVAC_21 ~ HVAC_24) -> 新风 < 35℃ 且 I > 18A -> 持续 10 分钟
	// ================================================================
	checkCpI := func(fasField, iField string, seq int, name string) {
		code := hvacCode(base, seq)
		isOverI := rawInt(raw, fasField) < 350 && rawInt(raw, iField) > 180
		if p.checkRule(isOverI, 10*time.Minute, deviceID, code, currentTime) {
			hits = append(hits, PredictHit{Code: code, Name: name, Severity: 3})
		}
	}
	checkCpI("FasU1", "ICpU11", 21, "机组1压缩机1电流预警")
	checkCpI("FasU1", "ICpU12", 22, "机组1压缩机2电流预警")
	checkCpI("FasU2", "ICpU21", 23, "机组2压缩机1电流预警")
	checkCpI("FasU2", "ICpU22", 24, "机组2压缩机2电流预警")

	// ================================================================
	// 6. 空气质量预警 (HVAC_125 ~ HVAC_126)
	// ================================================================
	checkAQ := func(uIdx int, name string) {
		code := hvacCode(base, uIdx+24) // 101+24=125, 126
		fanRunning := rawBool(raw, fmt.Sprintf("CfbkEfU%d1", uIdx))
		hasBeenRunning := p.checkRule(fanRunning, 20*time.Minute, deviceID, code+"_fanrun", currentTime)

		aqErr := hasBeenRunning && (rawInt(raw, fmt.Sprintf("AqCo2U%d", uIdx)) > 1200 ||
			rawInt(raw, fmt.Sprintf("AqPm25U%d", uIdx)) > 75 ||
			rawInt(raw, fmt.Sprintf("AqPm10U%d", uIdx)) > 150 ||
			rawInt(raw, fmt.Sprintf("AqTvocU%d", uIdx)) > 600)

		if p.checkRule(aqErr, 20*time.Minute, deviceID, code+"_pollute", currentTime) {
			hits = append(hits, PredictHit{Code: code, Name: name, Severity: 3})
		}
	}
	checkAQ(1, "机组1空气质量预警")
	checkAQ(2, "机组2空气质量预警")

	return hits
}

// ============================================================
// buildAlarmHits：原生故障位告警（直接映射 binary 故障位字段）
// ============================================================

func buildAlarmHits(raw map[string]any) []AlarmHit {
	// 初始化为空 slice（非 nil），序列化时输出 [] 而非 null
	hits := make([]AlarmHit, 0)
	// raw 为空时直接返回，避免误判断
	if len(raw) == 0 {
		return hits
	}

	// 辅助函数：批量检查故障位
	check := func(field, code, name string, level int) {
		if rawBool(raw, field) {
			hits = append(hits, AlarmHit{Code: code, Name: name, Level: level})
		}
	}

	// ================================================================
	// 1. 核心系统与环境告警 (Level 1)
	// ================================================================
	check("BfltPowersupplyU1", "bflt_powersupply_u1", "机组1供电故障", 1)
	check("BfltPowersupplyU2", "bflt_powersupply_u2", "机组2供电故障", 1)
	check("BfltTempover", "bflt_tempover", "车厢温度超温", 1)
	check("BfltEmergivt", "bflt_emergivt", "紧急通风故障", 1)

	// ================================================================
	// 2. 压缩机与压力系统 (Level 2)
	// ================================================================
	// 高低压故障 (Blpflt/Bscflt)
	check("BlpfltCompU11", "blpflt_comp_u11", "低压故障U1-1", 2)
	check("BlpfltCompU12", "blpflt_comp_u12", "低压故障U1-2", 2)
	check("BlpfltCompU21", "blpflt_comp_u21", "低压故障U2-1", 2)
	check("BlpfltCompU22", "blpflt_comp_u22", "低压故障U2-2", 2)
	check("BscfltCompU11", "bscflt_comp_u11", "高压故障U1-1", 2)
	check("BscfltCompU12", "bscflt_comp_u12", "高压故障U1-2", 2)
	check("BscfltCompU21", "bscflt_comp_u21", "高压故障U2-1", 2)
	check("BscfltCompU22", "bscflt_comp_u22", "高压故障U2-2", 2)

	// 单系统保护位 (BfltHigh/Lowpres)
	check("BfltHighpresU11", "bflt_highpres_u11", "高压保护U1-1", 2)
	check("BfltHighpresU12", "bflt_highpres_u12", "高压保护U1-2", 2)
	check("BfltHighpresU21", "bflt_highpres_u21", "高压保护U2-1", 2)
	check("BfltHighpresU22", "bflt_highpres_u22", "高压保护U2-2", 2)
	check("BfltLowpresU11", "bflt_lowpres_u11", "低压保护U1-1", 2)
	check("BfltLowpresU12", "bflt_lowpres_u12", "低压保护U1-2", 2)
	check("BfltLowpresU21", "bflt_lowpres_u21", "低压保护U2-1", 2)
	check("BfltLowpresU22", "bflt_lowpres_u22", "低压保护U2-2", 2)

	// ================================================================
	// 3. 变频器与风机系统 (Level 2)
	// ================================================================
	// 变频器故障
	check("BfltVfdU11", "bflt_vfd_u11", "变频器故障U1-1", 2)
	check("BfltVfdU12", "bflt_vfd_u12", "变频器故障U1-2", 2)
	check("BfltVfdU21", "bflt_vfd_u21", "变频器故障U2-1", 2)
	check("BfltVfdU22", "bflt_vfd_u22", "变频器故障U2-2", 2)
	check("BfltVfdComU11", "bflt_vfd_com_u11", "变频器通信故障U1-1", 2)
	check("BfltVfdComU12", "bflt_vfd_com_u12", "变频器通信故障U1-2", 2)
	check("BfltVfdComU21", "bflt_vfd_com_u21", "变频器通信故障U2-1", 2)
	check("BfltVfdComU22", "bflt_vfd_com_u22", "变频器通信故障U2-2", 2)

	// 通风机过流
	check("BocfltEfU11", "bocflt_ef_u11", "通风机过流U1-1", 2)
	check("BocfltEfU12", "bocflt_ef_u12", "通风机过流U1-2", 2)
	check("BocfltEfU21", "bocflt_ef_u21", "通风机过流U2-1", 2)
	check("BocfltEfU22", "bocflt_ef_u22", "通风机过流U2-2", 2)

	// 冷凝风机过流与通风故障
	check("BocfltCfU11", "bocflt_cf_u11", "冷凝风机过流U1-1", 2)
	check("BocfltCfU12", "bocflt_cf_u12", "冷凝风机过流U1-2", 2)
	check("BocfltCfU21", "bocflt_cf_u21", "冷凝风机过流U2-1", 2)
	check("BocfltCfU22", "bocflt_cf_u22", "冷凝风机过流U2-2", 2)
	check("BscfltVentU11", "bscflt_vent_u11", "通风故障U1-1", 2)
	check("BscfltVentU12", "bscflt_vent_u12", "通风故障U1-2", 2)
	check("BscfltVentU21", "bscflt_vent_u21", "通风故障U2-1", 2)
	check("BscfltVentU22", "bscflt_vent_u22", "通风故障U2-2", 2)

	// 废排风机
	check("BfltExhaustfan", "bflt_exhaustfan", "废排风机故障", 2)

	// ================================================================
	// 4. 阀门与执行器 (Level 2)
	// ================================================================
	check("BfltFadU11", "bflt_fad_u11", "新风阀故障U1-1", 2)
	check("BfltFadU12", "bflt_fad_u12", "新风阀故障U1-2", 2)
	check("BfltFadU21", "bflt_fad_u21", "新风阀故障U2-1", 2)
	check("BfltFadU22", "bflt_fad_u22", "新风阀故障U2-2", 2)
	check("BfltRadU11", "bflt_rad_u11", "回风阀故障U1-1", 2)
	check("BfltRadU12", "bflt_rad_u12", "回风阀故障U1-2", 2)
	check("BfltRadU21", "bflt_rad_u21", "回风阀故障U2-1", 2)
	check("BfltRadU22", "bflt_rad_u22", "回风阀故障U2-2", 2)
	check("BfltExhaustval", "bflt_exhaustval", "废排风阀故障", 2)

	// ================================================================
	// 5. 传感器系统 (Level 2)
	// ================================================================
	check("BfltDiffpresU1", "bflt_diffpres_u1", "压差传感器故障U1", 2)
	check("BfltDiffpresU2", "bflt_diffpres_u2", "压差传感器故障U2", 2)
	check("BfltAirmonU1", "bflt_airmon_u1", "空气质量传感器故障U1", 2)
	check("BfltAirmonU2", "bflt_airmon_u2", "空气质量传感器故障U2", 2)
	check("BfltCurrentmon", "bflt_currentmon", "电流监测故障", 2)

	// 温度/盘管传感器
	check("BfltVehtempU1", "bflt_vehtemp_u1", "车厢温度传感器故障U1", 2)
	check("BfltVehtempU2", "bflt_vehtemp_u2", "车厢温度传感器故障U2", 2)
	check("BfltRnttempU1", "bflt_rnttemp_u1", "回风温度传感器故障U1", 2)
	check("BfltRnttempU2", "bflt_rnttemp_u2", "回风温度传感器故障U2", 2)
	check("BfltFrstempU1", "bflt_frstemp_u1", "冰霜温度传感器故障U1", 2)
	check("BfltFrstempU2", "bflt_frstemp_u2", "冰霜温度传感器故障U2", 2)
	check("BfltCoiltempU11", "bflt_coiltemp_u11", "盘管温度传感器故障U1-1", 2)
	check("BfltCoiltempU12", "bflt_coiltemp_u12", "盘管温度传感器故障U1-2", 2)
	check("BfltCoiltempU21", "bflt_coiltemp_u21", "盘管温度传感器故障U2-1", 2)
	check("BfltCoiltempU22", "bflt_coiltemp_u22", "盘管温度传感器故障U2-2", 2)

	// 送风/吸气传感器
	check("BfltSplytempU11", "bflt_splytemp_u11", "送风温度传感器故障U1-1", 2)
	check("BfltSplytempU12", "bflt_splytemp_u12", "送风温度传感器故障U1-2", 2)
	check("BfltSplytempU21", "bflt_splytemp_u21", "送风温度传感器故障U2-1", 2)
	check("BfltSplytempU22", "bflt_splytemp_u22", "送风温度传感器故障U2-2", 2)
	check("BfltInsptempU11", "bflt_insptemp_u11", "吸气温度传感器故障U1-1", 2)
	check("BfltInsptempU12", "bflt_insptemp_u12", "吸气温度传感器故障U1-2", 2)
	check("BfltInsptempU21", "bflt_insptemp_u21", "吸气温度传感器故障U2-1", 2)
	check("BfltInsptempU22", "bflt_insptemp_u22", "吸气温度传感器故障U2-2", 2)

	// ================================================================
	// 6. 通信与其他组件 (Level 2)
	// ================================================================
	check("BfltTcms", "bflt_tcms", "TCMS通信故障", 2)
	check("BfltExpboardU1", "bflt_expboard_u1", "扩展板通信故障U1", 2)
	check("BfltExpboardU2", "bflt_expboard_u2", "扩展板通信故障U2", 2)
	check("BfltApU11", "bflt_ap_u11", "空气净化器故障U1-1", 2)
	check("BfltApU21", "bflt_ap_u21", "空气净化器故障U2-1", 2)

	return hits
}

// ============================================================
// buildLifeHits：部件寿命预警
// part_code = carriage_id * 1000 + 50000 + offset
// 风机时间单位：秒（raw）  阀门单位：次（raw）
// ============================================================

func buildLifeHits(raw map[string]any, carriageID int) []LifeHit {
	// 初始化为空 slice（非 nil），序列化时输出 [] 而非 null
	hits := make([]LifeHit, 0)
	// raw 为空时直接返回，避免误判断
	if len(raw) == 0 {
		return hits
	}
	lifeBase := int64(carriageID*1000 + 50_000)

	// checkFan 检查风机类寿命（单位：秒）
	checkFan := func(field string, offset int, name string) {
		val := rawInt(raw, field)
		code := fmt.Sprintf("%d", lifeBase+int64(offset))
		if val >= fanCritS {
			hits = append(hits, LifeHit{Code: code, Name: name, Severity: 3, Value: val, Limit: fanLifeS})
		} else if val >= fanWarnS {
			hits = append(hits, LifeHit{Code: code, Name: name, Severity: 2, Value: val, Limit: fanLifeS})
		}
	}

	// checkComp 检查压缩机类寿命（单位：秒）
	checkComp := func(field string, offset int, name string) {
		val := rawInt(raw, field)
		code := fmt.Sprintf("%d", lifeBase+int64(offset))
		if val >= cpCritS {
			hits = append(hits, LifeHit{Code: code, Name: name, Severity: 3, Value: val, Limit: cpLifeS})
		} else if val >= cpWarnS {
			hits = append(hits, LifeHit{Code: code, Name: name, Severity: 2, Value: val, Limit: cpLifeS})
		}
	}

	// checkValve 检查阀门类寿命（单位：次）
	checkValve := func(field string, offset int, name string) {
		val := rawInt(raw, field)
		code := fmt.Sprintf("%d", lifeBase+int64(offset))
		if val >= valveCritN {
			hits = append(hits, LifeHit{Code: code, Name: name, Severity: 3, Value: val, Limit: valveLifeN})
		} else if val >= valveWarnN {
			hits = append(hits, LifeHit{Code: code, Name: name, Severity: 2, Value: val, Limit: valveLifeN})
		}
	}

	// 机组1（offset +001~+006）
	checkFan("DwefOpTmU11", 1, "机组1通风机累计运行时间")
	checkFan("DwcfOpTmU11", 2, "机组1冷凝风机累计运行时间")
	checkComp("DwcpOpTmU11", 3, "机组1压缩机1累计运行时间")
	checkComp("DwcpOpTmU12", 4, "机组1压缩机2累计运行时间")
	checkValve("DwfadOpCntU1", 5, "机组1新风阀开关总次数")
	checkValve("DwradOpCntU1", 6, "机组1回风阀开关总次数")

	// 机组2（offset +011~+016）
	checkFan("DwefOpTmU21", 11, "机组2通风机累计运行时间")
	checkFan("DwcfOpTmU21", 12, "机组2冷凝风机累计运行时间")
	checkComp("DwcpOpTmU21", 13, "机组2压缩机1累计运行时间")
	checkComp("DwcpOpTmU22", 14, "机组2压缩机2累计运行时间")
	checkValve("DwfadOpCntU2", 15, "机组2新风阀开关总次数")
	checkValve("DwradOpCntU2", 16, "机组2回风阀开关总次数")

	// 废排（offset +021~+022）
	checkFan("DwexufanOpTm", 21, "废排风机累计运行时间")
	checkValve("DwdmpexuOpCnt", 22, "废排风阀开关总次数")

	return hits
}
