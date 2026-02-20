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

// NB67EventProcessor 是 Benthos/Connect 自定义处理器，
// 实现 HVAC 预警、原生故障位告警、部件寿命三类事件的构建逻辑。
type NB67EventProcessor struct{}

// init 在程序启动时自动注册处理器。
func init() {
	err := service.RegisterProcessor(
		"nb67_event_builder",
		service.NewConfigSpec().
			Summary("NB67 空调事件构建处理器").
			Description(
				"读取 nb67_parser 解码后的 signal-parsed 消息，\n"+
					"按照 NB67 空调预警码表（20240802）和部件码表（20240802）生成三类事件：\n"+
					"  predict_event - HVAC 算法预警（26 种预警码）\n"+
					"  alarm_event   - 原生故障位告警（直接映射 binary 故障位）\n"+
					"  life_event    - 部件寿命预警（风机/压缩机/阀门）\n",
			),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return &NB67EventProcessor{}, nil
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
		return nil, fmt.Errorf("读取消息字节失败: %w", err)
	}

	var input parsedInput
	if err := json.Unmarshal(rawBytes, &input); err != nil {
		return nil, fmt.Errorf("解析输入 JSON 失败: %w", err)
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

	raw := input.Raw

	// 构建三类事件命中列表
	predictHits := buildPredictHits(raw, input.CarriageID)
	alarmHits := buildAlarmHits(raw)
	lifeHits := buildLifeHits(raw, input.CarriageID)

	// 三类命中均为空时丢弃消息，不向下游输出任何内容
	if len(predictHits) == 0 && len(alarmHits) == 0 && len(lifeHits) == 0 {
		return service.MessageBatch{}, nil
	}

	// 聚合输出
	output := EventOutput{
		PredictEvent: SubEvent{
			EventMeta: meta,
			Hits:      predictHits,
			Source:    "connect-rule-v2",
		},
		AlarmEvent: SubEvent{
			EventMeta: meta,
			Hits:      alarmHits,
			Source:    "raw-fault-bit",
		},
		LifeEvent: SubEvent{
			EventMeta: meta,
			Hits:      lifeHits,
			Source:    "part-life-v2",
		},
	}

	outBytes, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("序列化输出 JSON 失败: %w", err)
	}

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

func buildPredictHits(raw map[string]any, carriageID int) []PredictHit {
	// 初始化为空 slice（非 nil），序列化时输出 [] 而非 null
	hits := make([]PredictHit, 0)
	// raw 为空时所有 rawInt 返回 0，会误触发 FasU1(0)<350 等规则，必须提前退出
	if len(raw) == 0 {
		return hits
	}
	base := carriageID * 100

	// ---------- 1~4. 冷媒泄漏预警（ref_leak）----------
	// HVAC_01: 机组1系统1(u11)  HVAC_02: 机组1系统2(u12)
	// HVAC_03: 机组2系统1(u21)  HVAC_04: 机组2系统2(u22)
	// 制冷或热泵(mode=2|3): 频率>30Hz 且 吸气压力<2.0bar(raw<20)
	// 通风(mode=1): 高压压力<5bar(raw<50)
	wModeU1 := rawInt(raw, "WmodeU1")
	wModeU2 := rawInt(raw, "WmodeU2")

	// u11
	if (wModeU1 == 2 || wModeU1 == 3) && rawInt(raw, "FCpU11") > 300 && rawInt(raw, "SuckpU11") < 20 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 1), Name: "机组1系统1冷媒泄露预警", Severity: 3})
	}
	if wModeU1 == 1 && rawInt(raw, "HighpressU11") < 50 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 1), Name: "机组1系统1冷媒泄露预警", Severity: 3})
	}
	// u12
	if (wModeU1 == 2 || wModeU1 == 3) && rawInt(raw, "FCpU12") > 300 && rawInt(raw, "SuckpU12") < 20 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 2), Name: "机组1系统2冷媒泄露预警", Severity: 3})
	}
	if wModeU1 == 1 && rawInt(raw, "HighpressU12") < 50 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 2), Name: "机组1系统2冷媒泄露预警", Severity: 3})
	}
	// u21
	if (wModeU2 == 2 || wModeU2 == 3) && rawInt(raw, "FCpU21") > 300 && rawInt(raw, "SuckpU21") < 20 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 3), Name: "机组2系统1冷媒泄露预警", Severity: 3})
	}
	if wModeU2 == 1 && rawInt(raw, "HighpressU21") < 50 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 3), Name: "机组2系统1冷媒泄露预警", Severity: 3})
	}
	// u22
	if (wModeU2 == 2 || wModeU2 == 3) && rawInt(raw, "FCpU22") > 300 && rawInt(raw, "SuckpU22") < 20 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 4), Name: "机组2系统2冷媒泄露预警", Severity: 3})
	}
	if wModeU2 == 1 && rawInt(raw, "HighpressU22") < 50 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 4), Name: "机组2系统2冷媒泄露预警", Severity: 3})
	}

	// ---------- 5~6. 制冷系统预警（f_cp）----------
	// HVAC_05: 机组1  HVAC_06: 机组2
	// 同频压缩机电流差>2A(raw>20) 或 过热度>20℃(raw>200) 或 <-8℃(raw<-80)
	cpU1Diff := rawInt(raw, "ICpU11") - rawInt(raw, "ICpU12")
	if rawInt(raw, "FCpU11") == rawInt(raw, "FCpU12") && (cpU1Diff > 20 || cpU1Diff < -20) {
		hits = append(hits, PredictHit{Code: hvacCode(base, 5), Name: "机组1制冷系统预警", Severity: 3})
	}
	spU11 := rawInt(raw, "SpU11")
	spU12 := rawInt(raw, "SpU12")
	if spU11 > 200 || spU11 < -80 || spU12 > 200 || spU12 < -80 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 5), Name: "机组1制冷系统预警", Severity: 3})
	}

	cpU2Diff := rawInt(raw, "ICpU21") - rawInt(raw, "ICpU22")
	if rawInt(raw, "FCpU21") == rawInt(raw, "FCpU22") && (cpU2Diff > 20 || cpU2Diff < -20) {
		hits = append(hits, PredictHit{Code: hvacCode(base, 6), Name: "机组2制冷系统预警", Severity: 3})
	}
	spU21 := rawInt(raw, "SpU21")
	spU22 := rawInt(raw, "SpU22")
	if spU21 > 200 || spU21 < -80 || spU22 > 200 || spU22 < -80 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 6), Name: "机组2制冷系统预警", Severity: 3})
	}

	// ---------- 7~8. 温度传感器预警（f_fas/f_ras）----------
	// HVAC_07: 新风温度传感器  HVAC_08: 回风温度传感器
	// U1 与 U2 温差>8℃(raw>80)
	fasDiff := rawInt(raw, "FasU1") - rawInt(raw, "FasU2")
	if fasDiff > 80 || fasDiff < -80 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 7), Name: "新风温度传感器预警", Severity: 3})
	}
	rasDiff := rawInt(raw, "RasU1") - rawInt(raw, "RasU2")
	if rasDiff > 80 || rasDiff < -80 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 8), Name: "回风温度传感器预警", Severity: 3})
	}

	// ---------- 9. 车厢温度超温预警（cabin_overtemp）----------
	// HVAC_09: BfltTempover 故障位置位
	if rawBool(raw, "BfltTempover") {
		hits = append(hits, PredictHit{Code: hvacCode(base, 9), Name: "车厢温度超温预警", Severity: 3})
	}

	// ---------- 10~11. 滤网脏堵预警（f_presdiff）----------
	// HVAC_10: 机组1  HVAC_11: 机组2
	// 通风机运行 且 压差>300Pa(raw>3000)
	if rawBool(raw, "CfbkEfU11") && rawInt(raw, "PresdiffU1") > 3000 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 10), Name: "机组1滤网脏堵预警", Severity: 2})
	}
	if rawBool(raw, "CfbkEfU21") && rawInt(raw, "PresdiffU2") > 3000 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 11), Name: "机组2滤网脏堵预警", Severity: 2})
	}

	// ---------- 12~15. 通风机电流预警（f_ef）----------
	// HVAC_12: 机组1通风机1  HVAC_13: 机组1通风机2
	// HVAC_14: 机组2通风机1  HVAC_15: 机组2通风机2
	// 运行时电流>2A(raw>20)
	if rawBool(raw, "CfbkEfU11") && rawInt(raw, "IEfU11") > 20 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 12), Name: "机组1通风机1电流预警", Severity: 3})
	}
	if rawBool(raw, "CfbkEfU11") && rawInt(raw, "IEfU12") > 20 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 13), Name: "机组1通风机2电流预警", Severity: 3})
	}
	if rawBool(raw, "CfbkEfU21") && rawInt(raw, "IEfU21") > 20 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 14), Name: "机组2通风机1电流预警", Severity: 3})
	}
	if rawBool(raw, "CfbkEfU21") && rawInt(raw, "IEfU22") > 20 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 15), Name: "机组2通风机2电流预警", Severity: 3})
	}

	// ---------- 16~19. 冷凝风机电流预警（f_cf）----------
	// HVAC_16: 机组1冷凝风机1  HVAC_17: 机组1冷凝风机2
	// HVAC_18: 机组2冷凝风机1  HVAC_19: 机组2冷凝风机2
	// 运行时电流>2.9A(raw>29)
	if rawBool(raw, "CfbkCfU11") && rawInt(raw, "ICfU11") > 29 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 16), Name: "机组1冷凝风机1电流预警", Severity: 3})
	}
	if rawBool(raw, "CfbkCfU11") && rawInt(raw, "ICfU12") > 29 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 17), Name: "机组1冷凝风机2电流预警", Severity: 3})
	}
	if rawBool(raw, "CfbkCfU21") && rawInt(raw, "ICfU21") > 29 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 18), Name: "机组2冷凝风机1电流预警", Severity: 3})
	}
	if rawBool(raw, "CfbkCfU21") && rawInt(raw, "ICfU22") > 29 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 19), Name: "机组2冷凝风机2电流预警", Severity: 3})
	}

	// ---------- 20. 废排风机电流预警（f_exufan）----------
	// HVAC_20: 运行时电流>4.0A(raw>40)
	if rawBool(raw, "CfbkExufan") && rawInt(raw, "IExufan") > 40 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 20), Name: "废排风机电流预警", Severity: 3})
	}

	// ---------- 21~24. 压缩机电流预警（f_fas_u）----------
	// HVAC_21: 机组1压缩机1  HVAC_22: 机组1压缩机2
	// HVAC_23: 机组2压缩机1  HVAC_24: 机组2压缩机2
	// 新风温度<35℃(raw<350) 且 压缩机电流>18A(raw>180)
	if rawInt(raw, "FasU1") < 350 && rawInt(raw, "ICpU11") > 180 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 21), Name: "机组1压缩机1电流预警", Severity: 3})
	}
	if rawInt(raw, "FasU1") < 350 && rawInt(raw, "ICpU12") > 180 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 22), Name: "机组1压缩机2电流预警", Severity: 3})
	}
	if rawInt(raw, "FasU2") < 350 && rawInt(raw, "ICpU21") > 180 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 23), Name: "机组2压缩机1电流预警", Severity: 3})
	}
	if rawInt(raw, "FasU2") < 350 && rawInt(raw, "ICpU22") > 180 {
		hits = append(hits, PredictHit{Code: hvacCode(base, 24), Name: "机组2压缩机2电流预警", Severity: 3})
	}

	// ---------- 25~26. 空气质量预警（f_aq）----------
	// HVAC_25: 机组1  HVAC_26: 机组2
	// 通风机运行 且 CO2>1200ppm 或 PM2.5>75 或 PM10>150 或 TVOC>600
	if rawBool(raw, "CfbkEfU11") &&
		(rawInt(raw, "AqCo2U1") > 1200 || rawInt(raw, "AqPm25U1") > 75 ||
			rawInt(raw, "AqPm10U1") > 150 || rawInt(raw, "AqTvocU1") > 600) {
		hits = append(hits, PredictHit{Code: hvacCode(base, 25), Name: "机组1空气质量预警", Severity: 3})
	}
	if rawBool(raw, "CfbkEfU21") &&
		(rawInt(raw, "AqCo2U2") > 1200 || rawInt(raw, "AqPm25U2") > 75 ||
			rawInt(raw, "AqPm10U2") > 150 || rawInt(raw, "AqTvocU2") > 600) {
		hits = append(hits, PredictHit{Code: hvacCode(base, 26), Name: "机组2空气质量预警", Severity: 3})
	}

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

	// 低压 / 高压故障
	check("BlpfltCompU11", "blpflt_comp_u11", "低压故障U1-1", 2)
	check("BlpfltCompU12", "blpflt_comp_u12", "低压故障U1-2", 2)
	check("BscfltCompU11", "bscflt_comp_u11", "高压故障U1-1", 2)
	check("BscfltCompU12", "bscflt_comp_u12", "高压故障U1-2", 2)
	check("BlpfltCompU21", "blpflt_comp_u21", "低压故障U2-1", 2)
	check("BlpfltCompU22", "blpflt_comp_u22", "低压故障U2-2", 2)
	check("BscfltCompU21", "bscflt_comp_u21", "高压故障U2-1", 2)
	check("BscfltCompU22", "bscflt_comp_u22", "高压故障U2-2", 2)
	// 车厢温度超温
	check("BfltTempover", "bflt_tempover", "车厢温度超温", 1)
	// 通风机过流故障
	check("BocfltEfU11", "bocflt_ef_u11", "通风机过流故障U1-1", 2)
	check("BocfltEfU12", "bocflt_ef_u12", "通风机过流故障U1-2", 2)
	check("BocfltCfU11", "bocflt_cf_u11", "冷凝风机过流故障U1-1", 2)
	check("BocfltCfU12", "bocflt_cf_u12", "冷凝风机过流故障U1-2", 2)
	check("BocfltEfU21", "bocflt_ef_u21", "通风机过流故障U2-1", 2)
	check("BocfltEfU22", "bocflt_ef_u22", "通风机过流故障U2-2", 2)
	check("BocfltCfU21", "bocflt_cf_u21", "冷凝风机过流故障U2-1", 2)
	check("BocfltCfU22", "bocflt_cf_u22", "冷凝风机过流故障U2-2", 2)
	// 变频器保护故障
	check("BfltVfdU11", "bflt_vfd_u11", "变频器保护故障U1-1", 2)
	check("BfltVfdU12", "bflt_vfd_u12", "变频器保护故障U1-2", 2)
	check("BfltVfdU21", "bflt_vfd_u21", "变频器保护故障U2-1", 2)
	check("BfltVfdU22", "bflt_vfd_u22", "变频器保护故障U2-2", 2)
	// 供电故障
	check("BfltPowersupplyU1", "bflt_powersupply_u1", "供电故障U1", 1)
	check("BfltPowersupplyU2", "bflt_powersupply_u2", "供电故障U2", 1)
	// 废排风机故障
	check("BfltExhaustfan", "bflt_exhaustfan", "废排风机故障", 2)

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
