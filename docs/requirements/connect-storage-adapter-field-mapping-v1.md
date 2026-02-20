# Storage Adapter 全字段映射清单（V1）

日期：2026-02-20

## 1. 映射目标

- 保证 NB67 二进制字段 **全量入库不遗漏**。
- 保证中间字段到 `hvac.fact_raw` 的展开列可追溯。
- 保证字段新增时有固定扩展路径。

## 2. 三层映射规则

### 2.1 Binary → Intermediate

- 来源：`connect/cmd/connect-nb67/nb67.go` 的 `Nb67` 结构（共 318 字段）。
- 全量字段清单：
  - [docs/reference/connect-storage-adapter-binary-field-list-v1.txt](docs/reference/connect-storage-adapter-binary-field-list-v1.txt)
- 映射规则（统一）：
  - `binary_field = X`
  - `intermediate_path = signal.parsed.v1.raw.X`

### 2.2 Intermediate → Table(JSONB)

- 映射规则（统一）：
  - `table_path = hvac.fact_raw.payload_json.raw.X`
- 说明：318 个二进制字段全部通过 `payload_json` 原样保留。

### 2.3 Intermediate → Table(热点列)

以下字段会同时展开为独立列（用于检索/聚合加速）：

| intermediate_field | fact_raw column | source path |
|---|---|---|
| event_time_text（解析） | event_time | `this.event_time_text` |
| ingest_time | ingest_time | `this.ingest_time` |
| process_time | process_time | `this.process_time` |
| line_id | line_id | `this.line_id` |
| train_id | train_id | `this.train_id` |
| carriage_id | carriage_id | `this.carriage_id` |
| device_id | device_id | `this.device_id` |
| frame_no | frame_no | `this.frame_no` |
| parser_version | parser_version | `this.parser_version` |
| quality_code | quality_code | `this.quality_code` |
| event_time_valid | event_time_valid | `this.event_time_valid` |
| wmode_u1 | wmode_u1 | `this.raw.WmodeU1` |
| wmode_u2 | wmode_u2 | `this.raw.WmodeU2` |
| f_cp_u11 | f_cp_u11 | `this.raw.FCpU11` |
| f_cp_u12 | f_cp_u12 | `this.raw.FCpU12` |
| f_cp_u21 | f_cp_u21 | `this.raw.FCpU21` |
| f_cp_u22 | f_cp_u22 | `this.raw.FCpU22` |
| i_cp_u11 | i_cp_u11 | `this.raw.ICpU11` |
| i_cp_u12 | i_cp_u12 | `this.raw.ICpU12` |
| i_cp_u21 | i_cp_u21 | `this.raw.ICpU21` |
| i_cp_u22 | i_cp_u22 | `this.raw.ICpU22` |
| suckp_u11 | suckp_u11 | `this.raw.SuckpU11` |
| suckp_u12 | suckp_u12 | `this.raw.SuckpU12` |
| suckp_u21 | suckp_u21 | `this.raw.SuckpU21` |
| suckp_u22 | suckp_u22 | `this.raw.SuckpU22` |
| highpress_u11 | highpress_u11 | `this.raw.HighpressU11` |
| highpress_u12 | highpress_u12 | `this.raw.HighpressU12` |
| highpress_u21 | highpress_u21 | `this.raw.HighpressU21` |
| highpress_u22 | highpress_u22 | `this.raw.HighpressU22` |
| fas_u1 | fas_u1 | `this.raw.FasU1` |
| fas_u2 | fas_u2 | `this.raw.FasU2` |
| ras_u1 | ras_u1 | `this.raw.RasU1` |
| ras_u2 | ras_u2 | `this.raw.RasU2` |
| presdiff_u1 | presdiff_u1 | `this.raw.PresdiffU1` |
| presdiff_u2 | presdiff_u2 | `this.raw.PresdiffU2` |
| aq_co2_u1 | aq_co2_u1 | `this.raw.AqCo2U1` |
| aq_co2_u2 | aq_co2_u2 | `this.raw.AqCo2U2` |
| aq_tvoc_u1 | aq_tvoc_u1 | `this.raw.AqTvocU1` |
| aq_tvoc_u2 | aq_tvoc_u2 | `this.raw.AqTvocU2` |
| aq_pm2_5_u1 | aq_pm2_5_u1 | `this.raw.AqPm25U1` |
| aq_pm2_5_u2 | aq_pm2_5_u2 | `this.raw.AqPm25U2` |
| aq_pm10_u1 | aq_pm10_u1 | `this.raw.AqPm10U1` |
| aq_pm10_u2 | aq_pm10_u2 | `this.raw.AqPm10U2` |
| bocflt_ef_u11 | bocflt_ef_u11 | `this.raw.BocfltEfU11` |
| bocflt_ef_u12 | bocflt_ef_u12 | `this.raw.BocfltEfU12` |
| bocflt_ef_u21 | bocflt_ef_u21 | `this.raw.BocfltEfU21` |
| bocflt_ef_u22 | bocflt_ef_u22 | `this.raw.BocfltEfU22` |
| bocflt_cf_u11 | bocflt_cf_u11 | `this.raw.BocfltCfU11` |
| bocflt_cf_u12 | bocflt_cf_u12 | `this.raw.BocfltCfU12` |
| bocflt_cf_u21 | bocflt_cf_u21 | `this.raw.BocfltCfU21` |
| bocflt_cf_u22 | bocflt_cf_u22 | `this.raw.BocfltCfU22` |
| blpflt_comp_u11 | blpflt_comp_u11 | `this.raw.BlpfltCompU11` |
| blpflt_comp_u12 | blpflt_comp_u12 | `this.raw.BlpfltCompU12` |
| blpflt_comp_u21 | blpflt_comp_u21 | `this.raw.BlpfltCompU21` |
| blpflt_comp_u22 | blpflt_comp_u22 | `this.raw.BlpfltCompU22` |
| bscflt_comp_u11 | bscflt_comp_u11 | `this.raw.BscfltCompU11` |
| bscflt_comp_u12 | bscflt_comp_u12 | `this.raw.BscfltCompU12` |
| bscflt_comp_u21 | bscflt_comp_u21 | `this.raw.BscfltCompU21` |
| bscflt_comp_u22 | bscflt_comp_u22 | `this.raw.BscfltCompU22` |
| bflt_tempover | bflt_tempover | `this.raw.BfltTempover` |
| bflt_diffpres_u1 | bflt_diffpres_u1 | `this.raw.BfltDiffpresU1` |
| bflt_diffpres_u2 | bflt_diffpres_u2 | `this.raw.BfltDiffpresU2` |
| payload_json | payload_json | `this` |

## 3. 存储适配器实现对应文件

- 入库服务：
  - [connect/cmd/storage-adapter/main.go](connect/cmd/storage-adapter/main.go)
  - [connect/cmd/storage-adapter/consumer.go](connect/cmd/storage-adapter/consumer.go)
  - [connect/cmd/storage-adapter/types.go](connect/cmd/storage-adapter/types.go)
  - [connect/cmd/storage-adapter/transform.go](connect/cmd/storage-adapter/transform.go)
- 上游映射：
  - [connect/config/nb67-storage-writer.yaml](connect/config/nb67-storage-writer.yaml)
- 目标 DDL：
  - [docs/requirements/sql/connect-timescaledb-ddl-v1.sql](docs/requirements/sql/connect-timescaledb-ddl-v1.sql)

## 4. 字段新增流程（强制）

1. 更新 `NB67.ksy` 并重新生成 `nb67.go`。
2. 解析后字段自动进入 `raw`，无损进 `payload_json`。
3. 若需要热点查询，新增 `fact_raw` 列 + storage-writer 映射 + adapter 结构体字段。
4. 同步更新本文档与字段清单。
