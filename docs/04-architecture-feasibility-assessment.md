# 架构可行性评估与实施建议

## 执行摘要

**结论**: ✅ 使用 Redpanda Connect 处理地铁空调数据并存入 TimescaleDB 的方案 **完全可行**,且具有良好的技术优势。

**关键发现**:
- Redpanda Connect 支持多种二进制数据处理方式
- TimescaleDB 完美支持时序数据场景
- 架构简洁、可维护性高
- 可以在管道中实现基础故障检测
- 为故障预测提供良好的数据基础

## 整体架构方案

### 推荐架构

```mermaid
graph TB
    subgraph "边缘采集层"
        A1[车厢1空调] --> B[边缘网关]
        A2[车厢2空调] --> B
        A3[车厢N空调] --> B
        B --> C[二进制数据文件]
    end
    
    subgraph "数据转换层"
        C --> D[转换服务<br/>Python/Go]
        D --> E[Avro Schema]
        E --> F[Redpanda Cluster]
        F --> G[Schema Registry]
    end
    
    subgraph "流处理层"
        F --> H1[Redpanda Connect 实例1<br/>数据解析+清洗]
        F --> H2[Redpanda Connect 实例2<br/>数据解析+清洗]
        F --> H3[Redpanda Connect 实例N<br/>数据解析+清洗]
        
        H1 --> I[实时故障检测]
        H2 --> I
        H3 --> I
    end
    
    subgraph "存储层"
        I --> J[TimescaleDB<br/>Hypertable]
        J --> K1[原始数据表]
        J --> K2[1小时聚合]
        J --> K3[1天聚合]
    end
    
    subgraph "应用层"
        K1 --> L1[Grafana<br/>实时监控]
        K2 --> L1
        K3 --> L1
        
        K1 --> L2[故障预测模型<br/>ML Pipeline]
        K2 --> L2
        K3 --> L2
        
        K1 --> L3[告警系统]
    end
```

### 技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| 消息队列 | Redpanda | Kafka 兼容,更高性能 |
| Schema 管理 | Schema Registry | Avro schema 版本管理 |
| 流处理 | Redpanda Connect | 轻量级,易于配置 |
| 时序数据库 | TimescaleDB | PostgreSQL 扩展 |
| 可视化 | Grafana | 开源监控面板 |
| 故障预测 | Python ML | scikit-learn/TensorFlow |

## 详细方案设计

### 阶段 1: 数据采集和转换

#### 1.1 边缘采集

```python
# 边缘网关伪代码
import struct
import time
from pathlib import Path

class HVACDataCollector:
    def collect_binary_data(self, device_id, sensors):
        """采集传感器数据并写入二进制文件"""
        data = struct.pack(
            '<16sQI100f',  # 格式: 设备ID(16字节), 时间戳(8字节), 车厢号(4字节), 100个float
            device_id.encode('utf-8').ljust(16),
            int(time.time() * 1000),
            sensors['car_number'],
            *[sensors.get(f'field_{i}', 0.0) for i in range(100)]
        )
        
        # 写入文件
        filename = Path(f'/data/hvac_{device_id}_{timestamp}.bin')
        filename.write_bytes(data)
        
        return filename
```

#### 1.2 转换服务

```python
# 转换服务: 二进制 -> Avro -> Redpanda
import struct
from confluent_kafka import SerializingProducer
from confluent_kafka.schema_registry import SchemaRegistryClient
from confluent_kafka.schema_registry.avro import AvroSerializer
from confluent_kafka.serialization import StringSerializer
import glob
import os

class BinaryToAvroConverter:
    def __init__(self, schema_registry_url, kafka_bootstrap_servers):
        self.schema_registry_client = SchemaRegistryClient({
            'url': schema_registry_url
        })
        
        # 加载 Avro Schema
        with open('hvac_schema.avsc', 'r') as f:
            schema_str = f.read()
        
        avro_serializer = AvroSerializer(
            self.schema_registry_client,
            schema_str
        )
        
        self.producer = SerializingProducer({
            'bootstrap.servers': kafka_bootstrap_servers,
            'key.serializer': StringSerializer('utf_8'),
            'value.serializer': avro_serializer
        })
    
    def parse_binary_file(self, filepath):
        """解析二进制文件"""
        with open(filepath, 'rb') as f:
            data = f.read()
        
        # 解析固定格式
        device_id = data[0:16].decode('utf-8').strip('\x00')
        timestamp = struct.unpack('<Q', data[16:24])[0]
        car_number = struct.unpack('<I', data[24:28])[0]
        
        # 解析100个字段
        offset = 28
        fields = struct.unpack('<100f', data[offset:offset+400])
        
        return {
            'device_id': device_id,
            'timestamp': timestamp,
            'car_number': car_number,
            'temp_supply_air': fields[0],
            'temp_return_air': fields[1],
            'pressure_compressor': fields[2],
            # ... 映射所有字段
        }
    
    def process_files(self, data_dir):
        """处理目录下的所有二进制文件"""
        for filepath in glob.glob(f'{data_dir}/*.bin'):
            try:
                record = self.parse_binary_file(filepath)
                
                # 发送到 Redpanda
                self.producer.produce(
                    topic='hvac-raw-data',
                    key=record['device_id'],
                    value=record,
                    on_delivery=self.delivery_report
                )
                
                # 成功后删除文件
                os.remove(filepath)
                
            except Exception as e:
                print(f'Error processing {filepath}: {e}')
                # 移动到错误目录
                os.rename(filepath, f'/data/errors/{os.path.basename(filepath)}')
        
        self.producer.flush()
    
    def delivery_report(self, err, msg):
        if err:
            print(f'Delivery failed: {err}')
        else:
            print(f'Message delivered to {msg.topic()} [{msg.partition()}]')

# 运行转换服务
converter = BinaryToAvroConverter(
    schema_registry_url='http://schema-registry:8081',
    kafka_bootstrap_servers='redpanda:9092'
)

# 定期扫描并处理文件
import schedule
schedule.every(10).seconds.do(lambda: converter.process_files('/data/incoming'))

while True:
    schedule.run_pending()
    time.sleep(1)
```

### 阶段 2: 流处理和故障检测

#### 2.1 Redpanda Connect 配置

参见 `03-timescaledb-integration.md` 中的完整配置。

#### 2.2 故障检测规则引擎

```yaml
# 在 Redpanda Connect 中实现规则引擎
pipeline:
  processors:
    - mapping: |
        # 定义故障检测规则
        root = this
        
        # 规则1: 温度异常
        let temp_too_high = this.temp_supply_air > 35
        let temp_too_low = this.temp_supply_air < 10
        let temp_abnormal = $temp_too_high || $temp_too_low
        
        # 规则2: 压缩机异常
        let pressure_abnormal = this.pressure_compressor > 2000 || this.pressure_compressor < 500
        let current_abnormal = this.current_compressor > 50 || this.current_compressor < 5
        
        # 规则3: 温差异常
        let temp_diff = this.temp_supply_air - this.temp_return_air
        let temp_diff_abnormal = $temp_diff < 3 || $temp_diff > 15
        
        # 规则4: 频繁启停 (需要状态)
        # 这个需要更复杂的逻辑,可以在后续的 ML 模型中处理
        
        # 综合判断
        root.fault_detected = $temp_abnormal || $pressure_abnormal || $current_abnormal || $temp_diff_abnormal
        
        root.fault_reasons = []
        root.fault_reasons = if $temp_too_high { $fault_reasons.append("HIGH_TEMP") } else { $fault_reasons }
        root.fault_reasons = if $temp_too_low { $fault_reasons.append("LOW_TEMP") } else { $fault_reasons }
        root.fault_reasons = if $pressure_abnormal { $fault_reasons.append("ABNORMAL_PRESSURE") } else { $fault_reasons }
        root.fault_reasons = if $current_abnormal { $fault_reasons.append("ABNORMAL_CURRENT") } else { $fault_reasons }
        root.fault_reasons = if $temp_diff_abnormal { $fault_reasons.append("ABNORMAL_TEMP_DIFF") } else { $fault_reasons }
        
        # 设置告警级别
        root.alert_level = if this.fault_code != 0 {
          "critical"
        } else if this.fault_detected {
          "warning"
        } else {
          "normal"
        }
```

### 阶段 3: 故障预测

#### 3.1 数据提取

```python
# 从 TimescaleDB 提取训练数据
import pandas as pd
from sqlalchemy import create_engine

engine = create_engine('postgresql://user:pass@timescaledb:5432/metro_hvac')

query = """
SELECT 
    device_id,
    time,
    temp_supply_air,
    temp_return_air,
    pressure_compressor,
    current_compressor,
    power_total,
    temp_diff,
    -- 滞后特征 (前1小时的平均值)
    AVG(temp_supply_air) OVER (
        PARTITION BY device_id 
        ORDER BY time 
        RANGE BETWEEN INTERVAL '1 hour' PRECEDING AND CURRENT ROW
    ) AS temp_1h_avg,
    -- 趋势特征
    temp_supply_air - LAG(temp_supply_air, 12) OVER (
        PARTITION BY device_id ORDER BY time
    ) AS temp_1h_change,  -- 假设每5分钟一条数据
    -- 目标变量: 未来2小时是否发生故障
    LEAD(fault_code, 24) OVER (
        PARTITION BY device_id ORDER BY time
    ) != 0 AS fault_next_2h
FROM hvac_measurements
WHERE time > NOW() - INTERVAL '90 days'
  AND data_quality > 0.8
ORDER BY device_id, time
"""

df = pd.read_sql(query, engine)
```

#### 3.2 特征工程

```python
from sklearn.preprocessing import StandardScaler
import numpy as np

# 创建更多特征
df['temp_diff'] = df['temp_supply_air'] - df['temp_return_air']
df['power_efficiency'] = df['power_total'] / (df['temp_diff'] + 1)  # 避免除零

# rolling 特征
df['temp_rolling_std'] = df.groupby('device_id')['temp_supply_air'].transform(
    lambda x: x.rolling(window=12, min_periods=1).std()
)

# 时间特征  
df['hour'] = df['time'].dt.hour
df['day_of_week'] = df['time'].dt.dayofweek
df['is_summer'] = df['time'].dt.month.isin([6, 7, 8]).astype(int)

# 标准化
scaler = StandardScaler()
feature_cols = [
    'temp_supply_air', 'temp_return_air', 'pressure_compressor',
    'current_compressor', 'power_total', 'temp_diff', 
    'temp_1h_avg', 'temp_1h_change', 'temp_rolling_std'
]
df[feature_cols] = scaler.fit_transform(df[feature_cols])
```

#### 3.3 模型训练

```python
from sklearn.ensemble import RandomForestClassifier, GradientBoostingClassifier
from sklearn.model_selection import train_test_split, cross_val_score
from sklearn.metrics import classification_report, roc_auc_score

# 准备数据
X = df[feature_cols + ['hour', 'day_of_week', 'is_summer']]
y = df['fault_next_2h']

# 删除包含 NaN 的行
mask = ~(X.isna().any(axis=1) | y.isna())
X = X[mask]
y = y[mask]

# 划分训练集和测试集
X_train, X_test, y_train, y_test = train_test_split(
    X, y, test_size=0.2, random_state=42, stratify=y
)

# 训练模型
model = GradientBoostingClassifier(
    n_estimators=100,
    learning_rate=0.1,
    max_depth=5,
    random_state=42
)

model.fit(X_train, y_train)

# 评估
y_pred = model.predict(X_test)
y_pred_proba = model.predict_proba(X_test)[:, 1]

print(classification_report(y_test, y_pred))
print(f'AUC-ROC: {roc_auc_score(y_test, y_pred_proba):.4f}')

# 特征重要性
feature_importance = pd.DataFrame({
    'feature': X.columns,
    'importance': model.feature_importances_
}).sort_values('importance', ascending=False)

print(feature_importance)
```

#### 3.4 模型部署

```python
# 保存模型
import joblib
joblib.dump(model, 'hvac_fault_prediction_model.pkl')
joblib.dump(scaler, 'hvac_scaler.pkl')

# 创建预测服务
from flask import Flask, request, jsonify

app = Flask(__name__)

model = joblib.load('hvac_fault_prediction_model.pkl')
scaler = joblib.load('hvac_scaler.pkl')

@app.route('/predict', methods=['POST'])
def predict():
    data = request.json
    
    # 从 TimescaleDB 提取最新数据并构建特征
    # ... (查询逻辑)
    
    # 预测
    features = [...]  # 提取的特征
    prediction_proba = model.predict_proba([features])[0][1]
    
    return jsonify({
        'device_id': data['device_id'],
        'fault_probability': float(prediction_proba),
        'prediction': 'high_risk' if prediction_proba > 0.7 else 'normal',
        'timestamp': time.time()
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
```

## 部署架构

### 容器化部署 (推荐)

```yaml
# docker-compose.yml
version: '3.8'

services:
  # Redpanda (Kafka 兼容)
  redpanda:
    image: docker.redpanda.com/redpandadata/redpanda:latest
    command:
      - redpanda start
      - --smp 2
      - --memory 2G
      - --advertise-kafka-addr redpanda:9092
    ports:
      - "9092:9092"
      - "9644:9644"  # Admin API
    volumes:
      - redpanda-data:/var/lib/redpanda/data

  # Schema Registry
  schema-registry:
    image: confluentinc/cp-schema-registry:latest
    depends_on:
      - redpanda
    environment:
      SCHEMA_REGISTRY_HOST_NAME: schema-registry
      SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS: redpanda:9092
    ports:
      - "8081:8081"

  # TimescaleDB
  timescaledb:
    image: timescale/timescaledb:latest-pg14
    environment:
      POSTGRES_USER: hvac_user
      POSTGRES_PASSWORD: secure_password
      POSTGRES_DB: metro_hvac
    ports:
      - "5432:5432"
    volumes:
      - timescale-data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql

  # Redpanda Connect (多实例)
  connect-1:
    image: docker.redpanda.com/redpandadata/connect:latest
    depends_on:
      - redpanda
      - timescaledb
      - schema-registry
    volumes:
      - ./connect-config.yaml:/connect.yaml
    command: run /connect.yaml
    environment:
      - INSTANCE_ID=1

  connect-2:
    image: docker.redpanda.com/redpandadata/connect:latest
    depends_on:
      - redpanda
      - timescaledb
      - schema-registry
    volumes:
      - ./connect-config.yaml:/connect.yaml
    command: run /connect.yaml
    environment:
      - INSTANCE_ID=2

  # 转换服务
  converter:
    build: ./converter
    depends_on:
      - redpanda
      - schema-registry
    volumes:
      - /data/incoming:/data/incoming
      - /data/errors:/data/errors
    environment:
      KAFKA_BOOTSTRAP_SERVERS: redpanda:9092
      SCHEMA_REGISTRY_URL: http://schema-registry:8081

  # Grafana
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    depends_on:
      - timescaledb
    volumes:
      - grafana-data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin

volumes:
  redpanda-data:
  timescale-data:
  grafana-data:
```

## 可行性评估

### ✅ 技术可行性

| 需求 | 评估 | 说明 |
|------|------|------|
| 二进制数据处理 | ✅ 可行 | 通过转换服务转为 Avro 格式 |
| 100+ 字段存储 | ✅ 可行 | TimescaleDB 支持宽表 |
| 实时处理 | ✅ 可行 | Redpanda Connect 低延迟 |  
| 故障检测 | ✅ 可行 | Bloblang 规则引擎 |
| 故障预测 | ✅ 可行 | ML 模型基于历史数据训练 |
| 数据展示 | ✅ 可行 | Grafana + TimescaleDB |
| 水平扩展 | ✅ 可行 | 所有组件支持分布式部署 |

### 💡 性能评估

**数据规模假设**:
- 车厢数: 500 个设备
- 采样频率: 每 5 分钟
- 数据量: ~100 KB/条 (100字段)

**吞吐量计算**:
```
每秒消息数 = 500 设备 / (5 * 60 秒) = 1.67 msg/s
每秒数据量 = 1.67 * 100 KB ≈ 167 KB/s
每天数据量 = 167 KB/s * 86400 s ≈ 14 GB/天
每年数据量 ≈ 5 TB/年 (压缩后约 1-2 TB)
```

**性能结论**:
- ✅ Redpanda: 轻松处理 1.67 msg/s (支持 100K+ msg/s)
- ✅ Redpanda Connect: 轻松处理 (支持 10K+ msg/s)
- ✅ TimescaleDB: 轻松处理 (批量写入支持 100K+ rows/s)

### 💰 成本评估

| 组件 | 推荐配置 | 年成本估算 (云) |
|------|---------|---------------|
| Redpanda | 3 节点,4C8G | ~$3,000 |
| TimescaleDB | 1 主 + 1 备,8C16G | ~$5,000 |
| Redpanda Connect | 2 实例,2C4G | ~$1,000 |
| 转换服务 | 2 实例,2C4G | ~$1,000 |
| Grafana | 1 实例,2C4G | ~$500 |
| **总计** | | **~$10,500/年** |

自建成本会更低 (约 30-50%)。

### ⏱️ 实施周期

| 阶段 | 工作内容 | 预计时间 |
|------|---------|---------|
| 1 | 环境搭建和测试 | 1 周 |
| 2 | 转换服务开发 | 2 周 |
| 3 | Redpanda Connect 配置 | 1 周 |
| 4 | TimescaleDB 表设计和优化 | 1 周 |
| 5 | 监控面板开发 | 1 周 |
| 6 | 故障检测规则开发 | 1 周 |
| 7 | 集成测试 | 2 周 |
| 8 | 试运行和调优 | 2 周 |
| **总计** | | **~11 周 (2.5个月)** |

故障预测模型可以在后期迭代中逐步完善。

## 风险和挑战

### ⚠️ 潜在风险

1. **数据质量**
   - 风险: 二进制格式变更导致解析失败
   - 缓解: Schema Registry 版本管理 + 严格的数据验证

2. **数据延迟**
   - 风险: 文件传输延迟
   - 缓解: 优化文件传输策略,考虑实时流式传输

3. **存储成本**  
   - 风险: 历史数据量大
   - 缓解: TimescaleDB 压缩策略 + 数据保留策略

4. **系统可用性**
   - 风险: 单点故障
   - 缓解: 所有组件部署高可用集群

### 🔧 技术挑战

1. **复杂二进制格式解析**
   - 挑战: 100+ 字段的解析逻辑复杂
   - 解决: 详细的数据字典文档 + 单元测试

2. **故障预测模型准确性**
   - 挑战: 需要大量历史故障数据训练
   - 解决: 先从简单规则引擎开始,逐步积累数据训练 ML 模型

3. **实时性要求**
   - 挑战: 告警需要及时
   - 解决: 优化批处理参数,平衡吞吐量和延迟

## 总结和建议

### ✅ 整体结论

使用 **Redpanda Connect + TimescaleDB** 方案处理地铁空调数据 **完全可行**,且具有以下优势:

1. **架构简洁**: 组件少,易于维护
2. **技术成熟**: 所有组件生产级别
3. **性能优秀**: 满足实时性和吞吐量要求
4. **成本可控**: 相比大数据平台成本低
5. **易于扩展**: 支持水平扩展

### 🎯 实施建议

#### 短期 (1-3个月)

1. **POC 验证**
   - 搭建单机测试环境
   - 验证二进制数据解析逻辑
   - 测试端到端数据流

2. **核心功能开发**
   - 转换服务开发和测试
   - Redpanda Connect 配置
   - TimescaleDB 表结构设计
   - 基础监控面板

3. **试运行**
   - 少量设备接入
   - 数据质量监控
   - 性能调优

#### 中期 (3-6个月)

1. **全量部署**
   - 所有设备接入
   - 高可用集群部署
   - 完善监控告警

2. **基础故障检测**
   - 规则引擎开发
   - 告警推送集成
   - 运维团队培训

#### 长期 (6-12个月)

1. **故障预测模型**
   - 积累历史故障数据
   - 特征工程迭代
   - ML 模型训练和优化

2. **高级功能**
   - 预测性维护建议
   - 能耗优化建议
   - 设备生命周期管理

### 📋 下一步行动

1. ✅ **立即**: 准备 POC 环境和测试数据
2. ✅ **本周**: 搭建单机测试环境
3. ✅ **下周**: 开发转换服务原型
4. ✅ **2周内**: 完成端到端测试

---

**本方案已经过充分研究,技术路线清晰可行,建议尽快启动 POC 验证。**
