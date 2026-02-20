# Connect 配置拆分与拓扑（V1）

日期：2026-02-20

---

## 1. 设计目标

把当前单体配置拆为 3 条职责清晰的流水线：

1. parser：二进制解析与标准化
2. storage-writer：批量入库事实层
3. event-builder：规则判定与事件输出

这样可以独立扩缩容、独立排障、独立回放。

---

## 2. Topic 拓扑

- 输入：`signal-in`
- 解析输出：`signal.parsed.v1`
- 事件输出：
  - `signal.event.alarm.v1`
  - `signal.event.predict.v1`
- 可选（后续）：`signal.feature.5m.v1`

Consumer Group 建议：

- `macda-parser-v1`
- `macda-storage-v1`
- `macda-event-v1`

---

## 3. 文件落位建议

放在目录：`connect/config/`

- `nb67-parser.yaml`
- `nb67-storage-writer.yaml`
- `nb67-event-builder.yaml`

---

## 4. parser 配置模板

```yaml
input:
  kafka:
    addresses: ["redpanda-1:9092", "redpanda-2:9092", "redpanda-3:9092"]
    topics: ["signal-in"]
    consumer_group: "macda-parser"
    start_from_oldest: false

pipeline:
  threads: 8
  processors:
    - nb67_parser:
        log_sample_every: 100

    - mapping: |
        root = this

        root.meta = {
          "schema_version": "nb67.parsed",
          "parser_version": this.parser_version.or("nb67-go-v1"),
          "ingest_time": now(),
          "process_time": now()
        }

        # event_time from device fields
        root.event_time = (
          "%04d-%02.0f-%02.0f %02.0f:%02.0f:%02.0f".format(
            this.src_year + 2000,
            this.src_month.number(),
            this.src_day.number(),
            this.src_hour.number(),
            this.src_minute.number(),
            this.src_second.number()
          )
        ).parse_timestamp("2006-01-02 15:04:05")

        root.ids = {
          "line_id": this.line_no,
          "train_id": this.train_no,
          "carriage_id": this.carriage_no,
          "device_id": "HVAC-%02d-%05d-%02d".format(this.line_no, this.train_no, this.carriage_no)
        }

        # quality check
        root.meta.event_time_valid = if root.event_time > now() + "24h".parse_duration() || root.event_time < now() - "365d".parse_duration() {
          false
        } else {
          true
        }

output:
  kafka:
    addresses: ["redpanda-1:9092", "redpanda-2:9092", "redpanda-3:9092"]
    topic: "signal.parsed.v1"
    key: ${! this.ids.device_id }
    partitioner: round_robin
    compression: snappy
    batching:
      count: 200
      period: 100ms
```

---

## 5. storage-writer 配置模板

说明：Redpanda Connect 推荐通过 SQL output 插件或 HTTP bridge 入库；如果当前环境暂不使用 SQL output，可先发到 `storage-adapter` 服务。

```yaml
input:
  kafka:
    addresses: ["redpanda-1:9092", "redpanda-2:9092", "redpanda-3:9092"]
    topics: ["signal.parsed.v1"]
    consumer_group: "macda-storage"
    start_from_oldest: false

pipeline:
  threads: 4
  processors:
    - mapping: |
        root = {
          "event_time": this.event_time,
          "ingest_time": this.meta.ingest_time,
          "process_time": this.meta.process_time,

          "line_id": this.ids.line_id,
          "train_id": this.ids.train_id,
          "carriage_id": this.ids.carriage_id,
          "device_id": this.ids.device_id,

          "parser_version": this.meta.parser_version,
          "event_time_valid": this.meta.event_time_valid,

          # hot fields
          "wmode_u1": this.wmode_u1,
          "wmode_u2": this.wmode_u2,
          "f_cp_u11": this.f_cp_u11,
          "f_cp_u12": this.f_cp_u12,
          "f_cp_u21": this.f_cp_u21,
          "f_cp_u22": this.f_cp_u22,
          "suckp_u11": this.suckp_u11,
          "suckp_u12": this.suckp_u12,
          "suckp_u21": this.suckp_u21,
          "suckp_u22": this.suckp_u22,
          "highpress_u11": this.highpress_u11,
          "highpress_u12": this.highpress_u12,
          "highpress_u21": this.highpress_u21,
          "highpress_u22": this.highpress_u22,
          "fas_u1": this.fas_u1,
          "fas_u2": this.fas_u2,
          "ras_u1": this.ras_u1,
          "ras_u2": this.ras_u2,
          "presdiff_u1": this.presdiff_u1,
          "presdiff_u2": this.presdiff_u2,

          # fault bits
          "bocflt_ef_u11": this.bocflt_ef_u11,
          "bocflt_ef_u12": this.bocflt_ef_u12,
          "bocflt_ef_u21": this.bocflt_ef_u21,
          "bocflt_ef_u22": this.bocflt_ef_u22,
          "blpflt_comp_u11": this.blpflt_comp_u11,
          "bscflt_comp_u11": this.bscflt_comp_u11,
          "bflt_tempover": this.bflt_tempover,

          "payload_json": this
        }

# 方式A：直接写 PostgreSQL（需要对应 output 插件）
# output:
#   sql_raw:
#     driver: postgres
#     dsn: postgres://postgres:passw0rd@timescaledb:5432/postgres?sslmode=disable
#     table: hvac.fact_raw
#
# 方式B：HTTP 交给 storage-adapter
output:
  http_client:
    url: http://storage-adapter:8080/api/v1/raw/batch
    verb: POST
    headers:
      Content-Type: application/json
    batching:
      count: 300
      period: 200ms
```

---

## 6. event-builder 配置模板

```yaml
input:
  kafka:
    addresses: ["redpanda-1:9092", "redpanda-2:9092", "redpanda-3:9092"]
    topics: ["signal.parsed.v1"]
    consumer_group: "macda-event"
    start_from_oldest: false

pipeline:
  threads: 2
  processors:
    - mapping: |
        root = this

        # 这里只保留轻量实时规则，复杂窗口在DB feature层处理
        root.predict_hits = []

        if this.wmode_u1 == 2 && this.f_cp_u11 > 30 && this.suckp_u11 < 2 {
          root.predict_hits = root.predict_hits.append({"code":"ref_leak_u11", "severity":2})
        }

        if this.bflt_tempover == true {
          root.alarm_hits = [{"code":"bflt_tempover", "level":1}]
        } else {
          root.alarm_hits = []
        }

        root.event_meta = {
          "device_id": this.ids.device_id,
          "line_id": this.ids.line_id,
          "train_id": this.ids.train_id,
          "carriage_id": this.ids.carriage_id,
          "event_time": this.event_time,
          "ingest_time": this.meta.ingest_time
        }

output:
  broker:
    pattern: fan_out
    outputs:
      - kafka:
          addresses: ["redpanda-1:9092", "redpanda-2:9092", "redpanda-3:9092"]
          topic: "signal.event.predict.v1"
          key: ${! this.event_meta.device_id }
      - kafka:
          addresses: ["redpanda-1:9092", "redpanda-2:9092", "redpanda-3:9092"]
          topic: "signal.event.alarm.v1"
          key: ${! this.event_meta.device_id }
```

---

## 7. 扩缩容建议

- parser：实例数 <= `signal-in` 分区数
- storage-writer：实例数 <= `signal.parsed.v1` 分区数
- event-builder：按规则复杂度扩缩容

---

## 8. 运维检查项

1. 每个 consumer group lag
2. parser 解析失败率
3. storage 写入成功率
4. event 生成速率与抖动率
5. end-to-end 延迟（event_time → API 可见）
