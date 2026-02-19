# 阶段1部署指南：Redpanda Connect NB67解析服务

## 📦 交付物清单

本阶段1交付以下文件，用于构建和部署NB67解析服务：

```
connect/
├── cmd/connect-nb67/          # 自定义Connect应用源代码
│   ├── main.go                # 启动器，注册nb67处理器
│   ├── nb67_processor.go       # NB67处理器实现（标准接口）
│   ├── nb67.go                # Kaitai生成的二进制解析器
│   └── go.mod                 # Go模块定义
├── config/
│   └── nb67-connect.yaml      # Redpanda Connect完整配置
├── codec/
│   ├── NB67.ksy               # Kaitai Struct定义（已更新新字段）
│   └── nb67.go                # Kaitai生成的Go代码
└── Dockerfile.connect         # 自定义Connect镜像构建文件

deploy/
└── docker-compose.stage1.yml  # 完整部署编排配置
```

## 🚀 快速开始（在192.168.32.17服务器上）

### 步骤1：准备源代码
```bash
# 将整个Macda-Connector目录上传到服务器
scp -r /path/to/Macda-Connector user@192.168.32.17:/opt/
cd /opt/Macda-Connector
```

### 步骤2：构建镜像
```bash
# 构建自定义Connect镜像
docker-compose -f deploy/docker-compose.stage1.yml build connect-nb67

# 验证镜像
docker images | grep macda-connect
```

### 步骤3：启动服务
```bash
# 启动Connect服务和网络检查
docker-compose -f deploy/docker-compose.stage1.yml up -d

# 等待服务就绪（约10-30秒）
docker-compose -f deploy/docker-compose.stage1.yml logs -f connect-nb67
```

### 步骤4：验证运行状态
```bash
# 查看容器状态
docker-compose -f deploy/docker-compose.stage1.yml ps

# 查看实时日志
docker-compose -f deploy/docker-compose.stage1.yml logs -f connect-nb67

# 应该看到类似输出：
# [NB67] Processed 100 frames: TrainNo=100 Carriage=3 CurStation=45
# [NB67] Processed 200 frames: TrainNo=105 Carriage=3 CurStation=60
```

## 📊 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      阶段1实时数据处理流程                        │
└─────────────────────────────────────────────────────────────────┘

[Redpanda Broker]                    [Docker容器：Connect-NB67]
   192.168.32.17:19092
        │                                            │
        │ topic: signal-in                           │
        │ (原始二进制NB67帧)                          │
        ├────────────────────────────────────────────┤
        │                                            │
        │   ┌──────────────────────────────────┐    │
        │   │ Input: Kafka Consumer            │    │
        │   │ Topic: signal-in                 │    │
        │   │ Group: macda-phase1-parser       │    │
        │   └──────────────────────────────────┘    │
        │              ↓                             │
        │   ┌──────────────────────────────────┐    │
        │   │ Processor 1: nb67_parser          │    │
        │   │ - Kaitai解析 (nb67.go)            │    │
        │   │ - 二进制→180+字段JSON             │    │
        │   │ - 新增字段：车站信息(452-460)     │    │
        │   └──────────────────────────────────┘    │
        │              ↓                             │
        │   ┌──────────────────────────────────┐    │
        │   │ Processor 2: Bloblang Mapping    │    │
        │   │ - 单位转换 (÷10)                  │    │
        │   │ - 特征提取 (设备ID、时间)        │    │
        │   │ - 故障检测 (压力/温度告警)       │    │
        │   └──────────────────────────────────┘    │
        │              ↓                             │
        │   ┌──────────────────────────────────┐    │
        │   │ Output: Kafka Producer            │    │
        │   │ Topic: signal-parsed             │    │
        │   │ Format: JSON (完全标准化)        │    │
        │   └──────────────────────────────────┘    │
        │                                            │
        ├────────────────────────────────────────────┤
        │                  ↓
     topic: signal-parsed
     (标准化JSON数据)
```

## 🔍 监控和验证

### 1. 查看Connect管理界面
```bash
# 启动Redpanda Console
docker-compose -f deploy/docker-compose.stage1.yml --profile monitor up -d redpanda-console

# 访问：http://192.168.32.17:8080
# - 查看topic列表
# - 实时监控消息
# - 查看消费者组进度
```

### 2. 消费输出数据
```bash
# 实时查看signal-parsed topic
kafka-console-consumer --bootstrap-server 192.168.32.17:19092 \
  --topic signal-parsed \
  --from-beginning \
  --max-messages 1 \
  --property print.key=true \
  --property print.value=true

# 输出示例：
# null  {
#   "header_code_01": 44,
#   "train_no": 100,
#   "carriage_no": 3,
#   "cur_station": 45,
#   "next_station": 66,
#   "start_station": 291,
#   "terminal_station": 129,
#   "timestamp": "2026-02-19T14:35:42Z",
#   "device_id": "HVAC-00100-03",
#   "environment": {
#     "temp_cabin_u1": 22.5,
#     "humidity_cabin_u1": 45.3
#   },
#   "compressor_u11": {
#     "frequency": 45.2,
#     "current": 12.3,
#     "power": 2500
#   },
#   "fault_detection": {
#     "low_pressure_u11": false,
#     "high_pressure_u11": false
#   },
#   "alert_level": "OK",
#   "metadata": {
#     "parser_version": "nb67-v1",
#     "parsed_at_ms": 1708347342123
#   }
# }
```

### 3. 查看处理吞吐量
```bash
# 查看日志中的采样信息
docker-compose -f deploy/docker-compose.stage1.yml logs connect-nb67 | grep "\[NB67\]"

# 应该看到每100条消息的输出（可配置）
# [NB67] Processed 100 frames: TrainNo=100 Carriage=3 CurStation=45
# [NB67] Processed 200 frames: TrainNo=105 Carriage=3 CurStation=60
```

### 4. 监控消费者进度
```bash
# 查看消费者组offset信息
kafka-consumer-groups --bootstrap-server 192.168.32.17:19092 \
  --group macda-phase1-parser \
  --describe

# 输出示例：
# TOPIC         PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG
# signal-in     0          2345            3456            1111
# signal-in     1          1234            2345            1111
# 总LAG: 2222
```

## ⚙️ 配置参数

### connect/config/nb67-connect.yaml

核心配置项：

```yaml
input:
  kafka:
    addresses: ["192.168.32.17:19092"]
    topics: ["signal-in"]
    consumer_group: macda-phase1-parser
    start_from_oldest: false        # true = 从头回放，false = 从最新开始

pipeline:
  processors:
    - nb67_parser:
        log_sample_every: 100       # 每100条消息输出采样日志

output:
  kafka:
    addresses: ["192.168.32.17:19092"]
    topic: signal-parsed
    partitioner: round_robin        # 分区策略
```

### docker-compose.stage1.yml

环境变量配置：

```yaml
environment:
  - LOG_LEVEL=info                  # debug/info/warn/error
  - LOG_FORMAT=json                 # json/text
  - REDPANDA_BROKERS=...
  - INPUT_TOPIC=signal-in
  - OUTPUT_TOPIC=signal-parsed
```

资源限制：

```yaml
deploy:
  resources:
    limits:
      cpus: '2'                     # 最多2核
      memory: 2G                    # 最多2GB
    reservations:
      cpus: '1'                     # 预留1核
      memory: 1G                    # 预留1GB
```

## 🔧 故障排除

### 1. Connect无法连接Redpanda
```bash
# 检查网络连通性
docker exec macda-connect-nb67 nc -zv 192.168.32.17 19092

# 检查Connect日志
docker logs macda-connect-nb67 | grep -i error
```

### 2. 消息解析失败
```bash
# 查看具体错误
docker logs macda-connect-nb67 | grep "NB67 parse error"

# 检查输入数据格式
kafka-console-consumer --bootstrap-server 192.168.32.17:19092 \
  --topic signal-in \
  --max-messages 1 \
  --property print.partition=true \
  --property print.offset=true
```

### 3. 输出topic为空
```bash
# 检查消费进度
kafka-consumer-groups --bootstrap-server 192.168.32.17:19092 \
  --group macda-phase1-parser \
  --describe

# 检查output topic存在性
kafka-topics --bootstrap-server 192.168.32.17:19092 --list | grep signal-parsed

# 如果topic不存在，创建它
kafka-topics --bootstrap-server 192.168.32.17:19092 \
  --create \
  --topic signal-parsed \
  --partitions 3 \
  --replication-factor 1
```

## 📈 性能指标验收

阶段1验收标准（见docs/11-macda-refactor-execution-plan.md）：

| 指标 | 目标值 | 验证方法 |
|------|--------|---------|
| **吞吐量** | ≥1000 msg/s | 监控日志采样，计算帧数/秒 |
| **延迟 (P99)** | <100ms | Connect metrics或tcpdump分析 |
| **正确率** | ≥99.99% | 对比解析字段与原始二进制验证 |
| **信号保留** | 100% | 180+字段无丢失 |
| **新字段覆盖** | 100% | 车站信息5字段完整输出 |

### 性能测试脚本（待实现）

```bash
#!/bin/bash
# test_stage1_performance.sh

BROKER="192.168.32.17:19092"
TOPIC_PARSED="signal-parsed"

# 记录起始时间
START_TIME=$(date +%s%N)

# 消费100条消息
kafka-console-consumer --bootstrap-server $BROKER \
  --topic $TOPIC_PARSED \
  --max-messages 100 \
  --formatter org.apache.kafka.common.serialization.StringDeserializer \
  > /tmp/output.jsonl

# 计算耗时
END_TIME=$(date +%s%N)
ELAPSED_MS=$(( (END_TIME - START_TIME) / 1000000 ))

echo "Throughput: $(echo "100000 / $ELAPSED_MS" | bc) msg/s"
echo "Latency: ${ELAPSED_MS}ms for 100 messages"

# 验证字段完整性
jq '.route_info | keys | length' /tmp/output.jsonl | sort | uniq -c
# 应该输出全部5个字段
```

## 📝 二进制协议变更说明

本实现已包含以下二进制协议更新：

| 偏移 | 字段 | 类型 | 说明 |
|------|------|------|------|
| 452-453 | `dmp_exu_pos` | U16 | 废排风阀开度百分比 |
| 454-455 | `start_station` | U16 | 起始站ID（枚举值） |
| 456-457 | `terminal_station` | U16 | 终点站ID（枚举值） |
| 458-459 | `cur_station` | U16 | 当前站ID（枚举值） |
| 460-461 | `next_station` | U16 | 下一站ID（枚举值） |

完整定义见：
- `connect/codec/NB67.ksy` - Kaitai Struct定义
- `docs/requirements/binary-spec.md` - 协议规范

## 🔐 安全考虑

- Connect管理端口 (4195) 仅在debug模式暴露
- Kafka认证配置：见yaml中的`sasl_`参数（当前未启用）
- 建议在生产环境启用一下功能：
  - TLS/mTLS for Kafka连接
  -认证和授权（RBAC）
  - 消息加密

## 📚 相关文档

- [docs/11-macda-refactor-execution-plan.md](../../docs/11-macda-refactor-execution-plan.md) - 12周执行计划
- [docs/requirements/binary-spec.md](../../docs/requirements/binary-spec.md) - NB67协议规范
- [Redpanda Connect官方文档](https://docs.redpanda.com/redpanda-connect/home/)
- [Kaitai Struct使用指南](https://kaitai.io/)

## ✅ 下一步（阶段2）

完成阶段1后，下一步工作包括：

1. **性能基线**：记录四大指标的baseline
2. **阶段2**：通用数据API（REST + WebSocket）
3. **故障告警**：从signal-parsed另行处理异常消息
4. **长期存储**：集成TimescaleDB
