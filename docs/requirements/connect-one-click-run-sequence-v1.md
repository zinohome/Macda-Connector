# Connect 一键联调顺序（V1）

日期：2026-02-20

---

## 1. 目的

把 parser / storage-writer / storage-adapter / event-builder 四段链路最小化跑通。

---

## 2. 前置条件

1. Redpanda 集群可达：`redpanda-1:9092, redpanda-2:9092, redpanda-3:9092`
2. topic 已创建或允许自动创建
3. 可执行文件：`connect-nb67`（包含 `nb67_parser` 自定义处理器）

---

## 3. 配置文件

- [connect/config/nb67-parser.yaml](connect/config/nb67-parser.yaml)
- [connect/config/nb67-storage-writer.yaml](connect/config/nb67-storage-writer.yaml)
- [connect/config/nb67-event-builder.yaml](connect/config/nb67-event-builder.yaml)

---

## 4. 启动顺序

在 4 个终端分别执行：

### 终端 A：Parser

```bash
cd connect/cmd/connect-nb67
./connect-nb67 -c ../../config/nb67-parser.yaml
```

### 终端 B：Storage Writer（当前输出到中间 topic）

```bash
cd connect/cmd/connect-nb67
./connect-nb67 -c ../../config/nb67-storage-writer.yaml
```

### 终端 C：Storage Adapter（消费中间 topic 写 TimescaleDB）

```bash
cd connect/cmd/storage-adapter
cp ../../config/storage-adapter.env.example .env
set -a && source .env && set +a
go run .
```

### 终端 D：Event Builder

```bash
cd connect/cmd/connect-nb67
./connect-nb67 -c ../../config/nb67-event-builder.yaml
```

---

## 5. 结果检查

至少检查以下 topic 有持续消息：

- `signal.parsed.v1`
- `signal.storage.fact_raw.v1`
- `signal.event.predict.v1`
- `signal.event.alarm.v1`

如果你有 `rpk`：

```bash
rpk topic consume signal.parsed.v1 -n 3
rpk topic consume signal.storage.fact_raw.v1 -n 3
rpk topic consume signal.event.predict.v1 -n 3
rpk topic consume signal.event.alarm.v1 -n 3
```

---

## 6. 常见问题

1. `failed to create ... consumer group`
   - 检查 broker 地址是否可达，topic 权限是否正常。

2. `unknown processor type: nb67_parser`
   - 当前执行的二进制不是 `connect-nb67`，或未包含 `main.go` 注册逻辑。

3. 下游 topic 没数据
   - 先确认 `signal-in` 是否有消息，再查 parser 日志是否持续处理。

---

## 7. 数据库落地检查

`storage-writer` 会输出到 `signal.storage.fact_raw.v1`，由 `storage-adapter` 入库到 TimescaleDB。

可在 TimescaleDB 执行：

```sql
SELECT device_id, event_time, ingest_time, quality_code
FROM hvac.fact_raw
ORDER BY ingest_time DESC
LIMIT 10;
```

数据库 DDL 参考：
- [docs/requirements/sql/connect-timescaledb-ddl-v1.sql](docs/requirements/sql/connect-timescaledb-ddl-v1.sql)
