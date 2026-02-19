# Redpanda Connect 二进制数据处理分析

## 概述

地铁空调数据项目的核心挑战是处理包含100多个字段的二进制格式数据。本文档分析 Redpanda Connect 处理此类二进制数据的能力和方案。

## 二进制数据处理能力

### 1. **原生二进制处理支持**

Redpanda Connect 提供以下二进制数据处理能力:

#### Bloblang 函数和方法

**内容处理**:
- `content()`: 获取消息的原始字节内容
- `bytes()`: 将值转换为字节数组
- `string()`: 将字节转换为字符串

**编码/解码**:
- `.encode("hex")`: 十六进制编码
- `.decode("hex")`: 十六进制解码
- `.encode("base64")`: Base64 编码
- `.decode("base64")`: Base64 解码
- `.encode("ascii")`: ASCII 编码
- `.decode("ascii")`: ASCII 解码

**位操作**:
- 字节切片: `this.bytes().slice(0, 4)`
- 字节拼接: `this.bytes().append(other)`

### 2. **结构化数据格式支持**

#### Schema Registry 集成

```yaml
pipeline:
  processors:
    - schema_registry_decode:
        url: http://schema-registry:8081
        # 自动从 Kafka 消息头读取 schema ID
        # 使用对应的 schema 解码二进制数据
    
    - mapping: |
        # 解码后的数据已经是 JSON 格式
        # 可以直接进行字段映射
        root.temperature = this.temperature
        root.pressure = this.pressure
```

#### Avro 格式处理

```yaml
pipeline:
  processors:
    - avro:
        operator: from_json
        encoding: binary
        schema: |
          {
            "type": "record",
            "name": "HVACData",
            "fields": [
              {"name": "temperature", "type": "float"},
              {"name": "pressure", "type": "float"}
            ]
          }
```

#### Protobuf 格式处理

```yaml
pipeline:
  processors:
    - protobuf:
        operator: from_json
        message: ".HVACData"
        import_paths: ["/schemas"]
```

### 3. **自定义二进制解析方案**

对于自定义二进制格式,有以下几种方案:

#### 方案 A: Bloblang 直接解析 (简单格式)

适用于结构简单、固定长度的二进制格式:

```yaml
pipeline:
  processors:
    - mapping: |
        # 假设前4字节是时间戳 (uint32)
        let timestamp_bytes = content().slice(0, 4)
        
        # 接下来4字节是温度 (float32)
        let temp_bytes = content().slice(4, 8)
        
        # 注意: Bloblang 对复杂二进制解析支持有限
        # 适合简单的字节切片和基础类型转换
        root.raw_timestamp = $timestamp_bytes.encode("hex")
        root.raw_temperature = $temp_bytes.encode("hex")
```

> **限制**: Bloblang 不直接支持复杂的二进制类型转换(如 little-endian float32),需要额外处理。

#### 方案 B: 预处理 + Schema Registry (推荐)

**步骤 1**: 在数据生产端,将自定义二进制格式转换为标准格式 (Avro/Protobuf)

**步骤 2**: 使用 Schema Registry 管理数据模式

**步骤 3**: Redpanda Connect 使用 schema_registry_decode 自动解析

```mermaid
graph LR
    A[二进制数据源] --> B[转换服务<br/>Binary → Avro]
    B --> C[Kafka/Redpanda<br/>with Schema Registry]
    C --> D[Redpanda Connect<br/>schema_registry_decode]
    D --> E[TimescaleDB]
```

#### 方案 C: 自定义 Go 插件 (最灵活)

对于非常复杂的二进制格式:

```go
// 编写自定义 processor 插件
package main

import (
    "github.com/redpanda-data/connect/v4/public/service"
    "encoding/binary"
)

func parseHVACBinary(msg *service.Message) (*service.Message, error) {
    data, err := msg.AsBytes()
    if err != nil {
        return nil, err
    }
    
    // 自定义解析逻辑
    timestamp := binary.LittleEndian.Uint32(data[0:4])
    temperature := binary.LittleEndian.Float32(data[4:8])
    // ... 解析其余100+字段
    
    // 构建 JSON 输出
    result := map[string]interface{}{
        "timestamp": timestamp,
        "temperature": temperature,
        // ...
    }
    
    return service.NewMessage(result), nil
}
```

#### 方案 D: WASM 数据转换

使用 WebAssembly 模块进行自定义转换:

```yaml
pipeline:
  processors:
    - wasm:
        module: ./parsers/hvac_parser.wasm
        function: parse_binary_data
```

## 地铁空调场景的推荐方案

### 场景分析

- **数据特点**: 100+ 字段的复杂二进制格式
- **数据源**: 汇聚到 Redpanda
- **目标**: 存储到 TimescaleDB

### 推荐架构

```mermaid
graph TB
    subgraph "数据采集层"
        A[空调设备] --> B[边缘采集器<br/>写入二进制文件]
    end
    
    subgraph "数据转换层"
        B --> C[转换服务<br/>Binary → Avro/JSON]
        C --> D[Redpanda<br/>with Schema Registry]
    end
    
    subgraph "数据处理层"
        D --> E[Redpanda Connect<br/>解析 + 映射 + 计算]
        E --> E1[实时故障检测]
        E --> E2[数据清洗]
        E --> E3[特征提取]
    end
    
    subgraph "存储层"
        E --> F[TimescaleDB<br/>时序数据存储]
    end
    
    subgraph "分析层"
        F --> G[数据展示<br/>Grafana/Dashboard]
        F --> H[故障预测模型<br/>ML Pipeline]
    end
```

### 具体实施方案

#### 选项 1: 转换服务 + Avro (推荐 ⭐)

**优点**:
- 数据格式标准化
- Schema Registry 提供schema版本管理
- Redpanda Connect 原生支持,配置简单
- 易于维护和扩展

**实施步骤**:

1. **定义 Avro Schema**:

```json
{
  "type": "record",
  "name": "MetroHVACData",
  "namespace": "com.metro.hvac",
  "fields": [
    {"name": "device_id", "type": "string"},
    {"name": "timestamp", "type": "long"},
    {"name": "car_number", "type": "int"},
    {"name": "temp_supply_air", "type": "float"},
    {"name": "temp_return_air", "type": "float"},
    {"name": "pressure_compressor", "type": "float"},
    {"name": "current_compressor", "type": "float"},
    {"name": "voltage_input", "type": "float"},
    // ... 其余90+字段
    {"name": "fault_code", "type": ["null", "int"], "default": null}
  ]
}
```

2. **转换服务** (Python/Go/Java):

```python
# 伪代码示例
import struct
from confluent_kafka import Producer
from confluent_kafka.schema_registry import SchemaRegistryClient
from confluent_kafka.schema_registry.avro import AvroSerializer

def parse_binary_file(binary_data):
    # 解析二进制数据
    offset = 0
    device_id = binary_data[offset:offset+16].decode('utf-8')
    offset += 16
    
    timestamp, = struct.unpack('<Q', binary_data[offset:offset+8])
    offset += 8
    
    car_number, = struct.unpack('<I', binary_data[offset:offset+4])
    offset += 4
    
    # ... 解析所有字段
    
    return {
        'device_id': device_id,
        'timestamp': timestamp,
        'car_number': car_number,
        # ...
    }

# 发送到 Redpanda
def send_to_redpanda(data):
    avro_serializer = AvroSerializer(schema_registry_client, schema_str)
    producer.produce(
        topic='hvac-data',
        value=avro_serializer(data, SerializationContext('hvac-data', MessageField.VALUE))
    )
```

3. **Redpanda Connect 配置**:

```yaml
input:
  kafka:
    addresses: ["redpanda:9092"]
    topics: ["hvac-data"]
    consumer_group: "hvac-processor"

pipeline:
  processors:
    # 自动解码 Avro
    - schema_registry_decode:
        url: http://schema-registry:8081
    
    # 数据转换和特征提取
    - mapping: |
        # 保留原始数据
        root = this
        
        # 计算派生字段
        root.temp_diff = this.temp_supply_air - this.temp_return_air
        root.power = this.voltage_input * this.current_compressor
        
        # 添加处理时间戳
        root.processed_at = now()
        
        # 数据验证
        root.is_valid = this.temp_supply_air > -50 && this.temp_supply_air < 100
    
    # 故障检测逻辑
    - mapping: |
        # 简单的规则引擎
        let high_temp = this.temp_supply_air > 35
        let high_pressure = this.pressure_compressor > 2000
        let high_current = this.current_compressor > 50
        
        root = this
        root.alert_level = if $high_temp && $high_pressure && $high_current {
          "critical"
        } else if $high_temp || $high_pressure {
          "warning"
        } else {
          "normal"
        }
        
        root.needs_maintenance = $high_current

output:
  sql_insert:
    driver: postgres
    dsn: "postgres://user:password@timescaledb:5432/metro_hvac?sslmode=disable"
    table: "hvac_measurements"
    columns:
      - device_id
      - timestamp
      - car_number
      - temp_supply_air
      - temp_return_air
      - pressure_compressor
      - current_compressor
      - voltage_input
      - temp_diff
      - power
      - alert_level
      - needs_maintenance
      - processed_at
    args_mapping: |
      root = [
        this.device_id,
        this.timestamp,
        this.car_number,
        this.temp_supply_air,
        this.temp_return_air,
        this.pressure_compressor,
        this.current_compressor,
        this.voltage_input,
        this.temp_diff,
        this.power,
        this.alert_level,
        this.needs_maintenance,
        this.processed_at
      ]
```

#### 选项 2: 自定义 Go 插件

如果不想引入转换服务,直接在 Redpanda Connect 中处理:

**优点**:
- 减少一个组件
- 更好的性能

**缺点**:
- 需要编写和维护 Go 代码
- 升级较为复杂

#### 选项 3: JSON 格式 (如果性能可接受)

最简单的方案:转换服务直接生成 JSON:

**优点**:
- 最简单,无需 Schema Registry
- 易于调试

**缺点**:
- JSON 体积较大
- 缺少 schema 验证

## 性能考虑

### 吞吐量评估

- **Avro 格式**: 比 JSON 小 30-50%
- **Redpanda Connect**: 单实例可处理 10,000+ msg/s
- **批量写入**: 提升数据库写入性能

### 批处理配置

```yaml
output:
  sql_insert:
    # ... 其他配置
    batching:
      count: 100       # 100条消息批量写入
      period: 1s       # 或每秒写入一次
      byte_size": 10MB # 或达到10MB
```

## 总结

对于地铁空调项目的100+字段二进制数据:

✅ **推荐方案**: 转换服务 + Avro + Schema Registry + Redpanda Connect
- 架构清晰
- 易于维护
- 性能优秀
- 未来可扩展

🔄 **替代方案**: 自定义 Go 插件 (如果团队有 Go 开发能力)

❌ **不推荐**: 纯 Bloblang 解析复杂二进制格式 (功能受限)
