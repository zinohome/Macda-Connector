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

### 静态二进制打包
```bash
cd connect/cmd/storage-adapter
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o storage-adapter -ldflags="-s -w" .
```

### 制作镜像并推送部署 (Docker)
这符合 `baseEnv/docker-compose-Dev.yml` 中定义的 `harbor.naivehero.top:8443/macda2/storage-adapter:v1.0` 的标准格式，执行以下命令即可将其编译出镜像并供容器编排使用：

```bash
cd connect/cmd/storage-adapter

# 1. 编译 Docker 镜像 (依赖外层 go.mod 需切回根目录周边或利用自带 Dockerfile 的多层构建)
# 当前 Dockerfile 为独立打包。执行此命令:
docker build -t harbor.naivehero.top:8443/macda2/storage-adapter:v1.0 .

# 2. 如果你需要上传到公司自有私有库
docker push harbor.naivehero.top:8443/macda2/storage-adapter:v1.0
```

> **注意：** 在最新的项目中已将该组件配置为了 `docker-compose-Dev.yml` 里的 `profiles: ["backup_plan_a"]`。若你想开启它，可使用 `docker compose --profile backup_plan_a up -d` 启动。否则，系统将默认采用 Redpanda Connect 的 SQL Output (Plan B) 进行平滑写入。
