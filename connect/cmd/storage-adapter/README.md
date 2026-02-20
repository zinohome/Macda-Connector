# storage-adapter

`storage-adapter` 负责消费 `signal.storage.fact_raw.v1`，并批量写入 TimescaleDB `hvac.fact_raw`。

## 能力

- Kafka Consumer Group 消费（可横向扩展）
- 批量写入 PostgreSQL/TimescaleDB（`pgx.Batch`）
- 幂等去重（`ON CONFLICT (device_id, event_time, ingest_time) DO NOTHING`）
- `event_time` 解析失败自动回退并标记 `event_time_valid=false`
- `payload_json` 全量保留，保证字段不丢失

## 环境变量

- `KAFKA_BROKERS`：默认 `redpanda-1:9092,redpanda-2:9092,redpanda-3:9092`
- `KAFKA_TOPIC`：默认 `signal.storage.fact_raw.v1`
- `KAFKA_GROUP`：默认 `macda-storage-adapter-v1`
- `PG_DSN`：默认 `postgres://postgres:passw0rd@timescaledb:5432/postgres?sslmode=disable`
- `BATCH_SIZE`：默认 `300`
- `FLUSH_INTERVAL_MS`：默认 `300`

## 本地运行

```bash
cd connect/cmd/storage-adapter
go mod tidy
go run .
```

## 构建

```bash
cd connect/cmd/storage-adapter
go build -o storage-adapter .
```
