# Redpanda Connect 概览

## 简介

Redpanda Connect 是一个高性能、具备容错能力的数据流处理器,能够连接各种数据源和目标,并在数据传输过程中执行各种转换、过滤、富化等操作。

## 核心特点

### 1. **声明式配置**
- 使用 YAML 格式定义数据管道
- 配置简洁,易于理解和维护
- 单个配置文件即可定义完整的数据流

### 2. **丰富的连接器支持**

**输入源 (Inputs)**:
- Kafka / Redpanda
- AWS (Kinesis, SQS, S3, DynamoDB)
- GCP (Pub/Sub, Cloud Storage)
- Azure (Blob Storage, Queue Storage)
- MQTT、AMQP、NATS
- Redis Streams
- HTTP/WebSocket
- 文件系统 (CSV, JSON, etc.)
- 数据库 (PostgreSQL, MySQL, MongoDB)

**输出目标 (Outputs)**:
- 同样支持上述所有系统
- PostgreSQL / TimescaleDB (通过 `sql_insert`)
- Elasticsearch
- HTTP API

### 3. **强大的数据处理能力**

**Bloblang 映射语言**:
- 专为数据映射设计的领域特定语言 (DSL)
- 安全、快速、强大
- 支持:
  - 字段映射和转换
  - 条件逻辑
  - 数据验证
  - 编码/解码 (base64, hex等)
  - JSON/XML 解析
  - 正则表达式
  - 数学运算

**处理器 (Processors)**:
- `mapping`: Bloblang 映射处理器
- `schema_registry_decode/encode`: Schema Registry 集成
- `avro`: Avro 格式转换
- `protobuf`: Protobuf 格式转换
- 自定义处理器 (Go 插件)

### 4. **Schema Registry 集成**
- 支持 Confluent Schema Registry
- 支持 Redpanda Schema Registry
- 自动处理模式版本
- 支持 Avro、Protobuf、JSON Schema

### 5. **运维友好**
- 内置健康检查端点
- 完善的监控指标 (Prometheus)
- 分布式追踪支持
- 错误处理和重试机制
- 死信队列支持

## 架构模式

```
┌─────────┐     ┌────────────┐     ┌──────────────┐     ┌─────────┐
│ Input   │────▶│ Processors │────▶│  Batching    │────▶│ Output  │
│ (Kafka) │     │ (Bloblang) │     │  (Optional)  │     │  (DB)   │
└─────────┘     └────────────┘     └──────────────┘     └─────────┘
```

### 典型配置结构

```yaml
input:
  kafka:
    addresses: ["kafka:9092"]
    topics: ["hvac-data"]
    consumer_group: "connect-group"

pipeline:
  processors:
    - mapping: |
        # Bloblang 映射逻辑
        root.field1 = this.data.field1
        root.timestamp = now()

output:
  sql_insert:
    driver: postgres
    dsn: "postgres://user:pass@host:5432/db"
    table: "hvac_measurements"
    columns: ["field1", "timestamp"]
```

## 部署方式

1. **独立二进制文件**
   ```bash
   rpk connect run config.yaml
   ```

2. **Docker 容器**
   ```bash
   docker run docker.redpanda.com/redpandadata/connect run -f /config.yaml
   ```

3. **Kubernetes**
   - 支持 Helm Chart 部署
   - 易于横向扩展

## 性能特点

- **高吞吐量**: 专为流处理优化
- **低延迟**: 实时数据处理能力
- **并行处理**: 支持多线程处理
- **背压处理**: 自动调节数据流速度
- **批处理优化**: 支持批量写入数据库

## 与替代方案对比

| 特性 | Redpanda Connect | Apache Flink | Kafka Streams | Logstash |
|------|-----------------|--------------|---------------|----------|
| 配置复杂度 | 低 (YAML) | 高 (Java/Scala) | 中 (Java) | 中 (配置文件) |
| 部署方式 | 独立/容器 | 集群 | 嵌入应用 | 独立 |
| 学习曲线 | 平缓 | 陡峭 | 中等 | 平缓 |
| 性能 | 高 | 非常高 | 高 | 中 |
| 二进制数据处理 | 支持 | 支持 | 有限 | 支持 |
| Schema Registry | 原生支持 | 需集成 | 需集成 | 插件 |

## 关键优势

1. **声明式配置**: 无需编程，通过 YAML 配置即可构建复杂管道
2. **数据格式灵活**: 支持多种序列化格式的转换
3. **易于运维**: 单一二进制文件，无复杂依赖
4. **可观测性**: 完善的监控和日志支持
5. **可扩展性**: 支持 Go 插件扩展功能

## 适用场景

✅ **非常适合**:
- 流数据的 ETL 处理
- 数据格式转换和标准化
- 实时数据管道
- 多源数据聚合
- 事件驱动架构

⚠️ **需要评估**:
- 复杂的有状态计算 (考虑 Flink)
- 需要窗口聚合函数 (基础支持,复杂场景考虑 Flink)
- 需要精确一次语义 (Exactly-once) 的场景

## 文档和资源

- 官方文档: https://docs.redpanda.com/redpanda-connect/home/
- GitHub: https://github.com/redpanda-data/connect
- Bloblang Playground: https://docs.redpanda.com/redpanda-connect/guides/bloblang/playground/
- Connector Catalog: https://docs.redpanda.com/redpanda-connect/components/about/
