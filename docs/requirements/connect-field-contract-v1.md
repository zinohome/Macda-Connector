# Connect 字段契约清单（V1）

日期：2026-02-20

---

## 1. 契约目的

本清单用于约束从 `signal-in` 到 API 展示的字段语义，确保：

1. 解析器、存储、API、前端一致
2. 字段变更可追踪
3. 预测与告警可解释

---

## 2. 核心字段（强制）

## 2.1 标识字段

- `line_id`：线路 ID，整型
- `train_id`：列车号，整型
- `carriage_id`：车厢号，整型
- `device_id`：设备唯一标识，格式 `HVAC-{line}-{train}-{carriage}`
- `frame_no`：帧序号，可空

## 2.2 时间字段

- `event_time`：设备业务时间（主时间）
- `ingest_time`：接入时间（系统时间）
- `process_time`：规则/预测计算时间（可空）
- `event_time_valid`：事件时间是否可信
- `query_time`：展示时间轴（由环境策略计算）

## 2.3 元数据字段

- `schema_version`：例如 `nb67.parsed.v1`
- `parser_version`：例如 `nb67-go-v1`
- `quality_code`：0 正常，>0 表示解析/质量异常

---

## 3. 热点测点字段（第一批）

## 3.1 运行模式

- `wmode_u1`, `wmode_u2`

## 3.2 压缩机关键量

- `f_cp_u11`, `f_cp_u12`, `f_cp_u21`, `f_cp_u22`
- `i_cp_u11`, `i_cp_u12`, `i_cp_u21`, `i_cp_u22`
- `suckp_u11`, `suckp_u12`, `suckp_u21`, `suckp_u22`
- `highpress_u11`, `highpress_u12`, `highpress_u21`, `highpress_u22`

## 3.3 环境与通风关键量

- `fas_u1`, `fas_u2`, `ras_u1`, `ras_u2`
- `presdiff_u1`, `presdiff_u2`
- `aq_co2_u1`, `aq_co2_u2`
- `aq_tvoc_u1`, `aq_tvoc_u2`
- `aq_pm2_5_u1`, `aq_pm2_5_u2`
- `aq_pm10_u1`, `aq_pm10_u2`

## 3.4 故障位（布尔）

- `bocflt_ef_u11`, `bocflt_ef_u12`, `bocflt_ef_u21`, `bocflt_ef_u22`
- `bocflt_cf_u11`, `bocflt_cf_u12`, `bocflt_cf_u21`, `bocflt_cf_u22`
- `blpflt_comp_u11`, `blpflt_comp_u12`, `blpflt_comp_u21`, `blpflt_comp_u22`
- `bscflt_comp_u11`, `bscflt_comp_u12`, `bscflt_comp_u21`, `bscflt_comp_u22`
- `bflt_tempover`, `bflt_diffpres_u1`, `bflt_diffpres_u2`

---

## 4. 事件契约

## 4.1 告警事件 `alarm_event`

- `event_id`
- `event_key`（`device_id + code`）
- `fault_code`, `fault_name`, `level`
- `start_time`, `end_time`, `status`（open/closed）
- `signal_snapshot`（触发时信号快照）

## 4.2 预测事件 `predict_event`

- `event_id`
- `event_key`
- `predict_code`, `predict_name`, `severity`, `score`
- `start_time`, `end_time`, `status`
- `feature_snapshot`

---

## 5. API 契约（第一版）

## 5.1 首页总览

请求参数：`line_id`, `start_time`, `end_time`

返回字段：

- `total_trains`
- `alarm_train_count`
- `warning_train_count`
- `normal_train_count`
- `last_time`

## 5.2 车厢状态

请求参数：`line_id`, `train_id`

返回字段：

- `carriage_id`
- `alarm_count`
- `warning_count`
- `health_status`
- `last_time`

## 5.3 告警列表

请求参数：`line_id`, `train_id`, `carriage_id?`, `type`, `start_time`, `end_time`, `page`, `limit`

返回字段：

- `event_id`
- `type`（alarm/predict）
- `code`
- `name`
- `level_or_severity`
- `start_time`
- `end_time`
- `duration_sec`
- `advice`

## 5.4 温度趋势

请求参数：`device_id`, `metric`, `start_time`, `end_time`, `bucket`

返回字段：

- `query_time`
- `value`

---

## 6. 时间轴策略契约（开发/生产）

- `prod`：`query_time = event_time`
- `dev`：`query_time = ingest_time`
- `smart`：`query_time = CASE WHEN event_time_valid THEN event_time ELSE ingest_time END`

开发联调只有单时间测试样本时，统一切 `dev` 或 `smart`。

---

## 7. 字段变更流程（强制）

任一字段新增/删除/重命名，必须同时更新：

1. 解析器映射
2. DDL 与视图
3. 事件规则
4. API 返回契约
5. 本契约文档

---

## 8. 首批验收清单

1. `fact_raw` 连续写入并可按 `device_id + query_time` 查询
2. 首页总览接口可返回 train 级计数
3. 车厢状态接口可返回 alarm/warning 数
4. 告警列表可展示 start/end/duration
5. 单时间测试数据在 dev 模式下可见趋势曲线
