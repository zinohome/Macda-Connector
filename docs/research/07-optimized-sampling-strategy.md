# 优化版本:数据采样降频方案

## 方案对比

### 版本 1: 全量数据 (原方案)
- **采样频率**: 1 条/秒
- **数据量**: 560 条/秒
- **用途**: 实时监控,故障预警

### 版本 2: 采样降频 (优化方案) ⭐
- **采样频率**: 1 条/10秒 (每个机组)
- **数据量**: 56 条/秒 (**减少 90%**)
- **用途**: 趋势分析,历史数据,成本优化

## 优化后数据规模

### 数据量重新计算

```
原始数据量: 560 条/秒
采样间隔: 10 秒
降频后数据量: 560 ÷ 10 = 56 条/秒

每日消息数: 56 × 86,400 = 4,838,400 条 ≈ 484 万条 (减少 90%)
每日数据量: 4.2 GB/天 (vs 原来 36 GB)
每年数据量: 1.5 TB/年 (vs 原来 13 TB)
```

### 带宽需求

```
JSON 格式: 56 条/秒 × 750 字节 = 42 KB/秒 ≈ 0.34 Mbps (减少 90%)
```

## 实现方式

### 方法 1: Redpanda Connect 采样 (推荐) ⭐

在 Redpanda Connect 中使用 Bloblang 实现采样逻辑:

```yaml
input:
  kafka:
    addresses: ["redpanda:9092"]
    topics: ["hvac-raw-data"]
    consumer_group: "sampled-processor"

pipeline:
  processors:
    # 1. 解析二进制数据
    - nb67_parser:
        debug: false
    
    # 2. 采样逻辑 - 关键步骤!
    - mapping: |
        # 构建设备唯一标识
        let device_key = "%s_%d_%d".format(
          this.device.train_no,
          this.device.carriage_no,
          this.msg_src_dvc_no  # 源设备号,区分同一车厢的两个空调
        )
        
        # 提取时间戳的秒数
        let timestamp_seconds = this.timestamp.year * 31536000 +
                                this.timestamp.month * 2592000 +
                                this.timestamp.day * 86400 +
                                this.timestamp.hour * 3600 +
                                this.timestamp.minute * 60 +
                                this.timestamp.second
        
        # 采样逻辑: 只保留时间戳是10的倍数的数据
        # 例如: 00秒, 10秒, 20秒, 30秒, 40秒, 50秒
        root = if ($timestamp_seconds % 10) == 0 {
          this  # 保留这条数据
        } else {
          deleted()  # 丢弃这条数据
        }
    
    # 3. 后续处理 (只处理保留的数据)
    - mapping: |
        # 时间戳转换
        root.time = "%d-%02d-%02d %02d:%02d:%02d".format(
          this.timestamp.year + 2000,
          this.timestamp.month,
          this.timestamp.day,
          this.timestamp.hour,
          this.timestamp.minute,
          this.timestamp.second
        ).parse_timestamp("2006-01-02 15:04:05")
        
        # 设备信息
        root.device_id = "HVAC-%d-%d-%d".format(
          this.device.train_no,
          this.device.carriage_no,
          this.msg_src_dvc_no
        )
        
        # ... 其他字段映射
        
    # 4. 故障检测 (保持不变)
    - mapping: |
        root = this
        
        # 故障检测逻辑
        root.has_fault = (
          this.faults.cfbk_ef_u11 ||
          this.faults.blpflt_comp_u11
        )
        
        root.alert_level = if this.has_fault {
          "warning"
        } else {
          "normal"
        }

output:
  sql_insert:
    driver: postgres
    dsn: "postgres://user:pass@timescaledb:5432/metro_hvac"
    table: "hvac_measurements"
    # ... 其他配置
    
    batching:
      count: 500   # 保持不变
      period: 10s  # 增加到10秒 (因为数据量减少)
```

**优点**:
- ✅ 简单易维护
- ✅ 灵活可调整 (改配置即可)
- ✅ 保留所有原始数据在 Kafka (可回溯)
- ✅ 无需修改数据采集端

### 方法 2: Kafka Topic 预处理

创建两个 Topic:
- `hvac-raw-data`: 全量数据 (1条/秒)
- `hvac-sampled-data`: 采样数据 (1条/10秒)

使用独立的采样器服务:

```yaml
# sampler-connect.yaml
input:
  kafka:
    addresses: ["redpanda:9092"]
    topics: ["hvac-raw-data"]
    consumer_group: "data-sampler"

pipeline:
  processors:
    - mapping: |
        # 采样逻辑同上
        root = if (this.timestamp.second % 10) == 0 {
          this
        } else {
          deleted()
        }

output:
  kafka:
    addresses: ["redpanda:9092"]
    topic: "hvac-sampled-data"
```

然后主处理器只消费 `hvac-sampled-data`:

```yaml
input:
  kafka:
    addresses: ["redpanda:9092"]
    topics: ["hvac-sampled-data"]  # 只消费采样后的数据
```

**优点**:
- ✅ 关注点分离
- ✅ 可以同时支持全量和采样两种用途
- ✅ 故障检测用全量,历史存储用采样

**缺点**:
- ❌ 多一个组件
- ❌ Kafka 存储两份数据

### 方法 3: 采集端降频 (最彻底)

直接在边缘采集端降低采样频率:

```python
# 边缘采集器修改
import time

last_sent_time = {}

def should_send_data(device_id):
    current_time = int(time.time())
    
    if device_id not in last_sent_time:
        last_sent_time[device_id] = current_time
        return True
    
    # 距离上次发送超过10秒才发送
    if current_time - last_sent_time[device_id] >= 10:
        last_sent_time[device_id] = current_time
        return True
    
    return False

# 采集循环
while True:
    data = collect_sensor_data()
    
    if should_send_data(data.device_id):
        send_to_kafka(data)  # 只发送采样数据
```

**优点**:
- ✅ 最省资源 (网络、存储)
- ✅ 减少边缘网关负载

**缺点**:
- ❌ 丢失实时性 (无法事后回溯)
- ❌ 无法支持实时告警

## 优化后资源配置

### Redpanda 集群

```yaml
节点数: 3 (保持高可用)
每节点配置:
  CPU: 2 核 (vs 原来 4 核) ⬇️
  内存: 4 GB (vs 原来 8 GB) ⬇️
  磁盘: 200 GB SSD (vs 原来 500 GB) ⬇️

集群总资源:
  CPU: 6 核 (vs 原来 12 核)
  内存: 12 GB (vs 原来 24 GB)
  存储: 600 GB (vs 原来 1.5 TB)
```

**存储计算**:
```
7天数据 = 4.2 GB/天 × 7 = 29 GB
副本数 = 3
总存储 = 29 GB × 3 = 87 GB (vs 原来 800 GB)
```

### Connect 实例

```yaml
实例数: 2 (vs 原来 3) ⬇️
每实例配置:
  CPU: 2 核 (保持)
  内存: 2 GB (vs 原来 4 GB) ⬇️

总配置: 4C 4G (vs 原来 6C 12G)
```

**处理能力**:
```
实际负载: 56 条/秒
单实例能力: 8,000 条/秒
每实例负载: 56 / 2 = 28 条/秒
利用率: 28 / 8,000 = 0.35% (极低)
```

### TimescaleDB

```yaml
节点数: 1 主 (可选 1 备)
主节点配置:
  CPU: 4 核 (vs 原来 8 核) ⬇️
  内存: 8 GB (vs 原来 16 GB) ⬇️
  磁盘: 1 TB SSD (vs 原来 2 TB) ⬇️
```

**存储计算**:
```
1年数据: 484万条/天 × 365 × 200字节 = 354 GB
压缩后 (70%): ~120 GB
含索引和聚合: ~200 GB (vs 原来 1.1 TB)
```

## 成本对比

### 云部署成本

| 组件 | 原方案 | 优化方案 | 节省 |
|------|--------|---------|------|
| Redpanda (3节点) | $300/月 | $150/月 | **-50%** |
| Connect | $150/月 | $60/月 | **-60%** |
| TimescaleDB | $600/月 | $300/月 | **-50%** |
| Grafana | $50/月 | $50/月 | - |
| 其他 | $50/月 | $50/月 | - |
| **月总计** | **$1,150** | **$610** | **-47%** |
| **年总计** | **$13,800** | **$7,320** | **-47%** |

**年节省**: $6,480 ≈ 6.5万人民币

### 自建成本

| 项目 | 原方案 | 优化方案 | 节省 |
|------|--------|---------|------|
| 硬件 (一次性) | $21,500 | $12,000 | **-44%** |
| 年运维 | $6,000 | $4,000 | **-33%** |

## 性能指标对比

| 指标 | 原方案 | 优化方案 | 变化 |
|------|--------|---------|------|
| **吞吐量** | 560 条/秒 | 56 条/秒 | -90% |
| **端到端延迟** | 1-2 秒 | 2-3 秒 | +50% (仍满足要求) |
| **存储/年** | 13 TB | 1.5 TB | -88% |
| **处理余量** | 40倍 | 400倍 | +10倍 |
| **成本/年** | $13,800 | $7,320 | -47% |

## 混合方案:双轨制 (最佳实践) ⭐

结合两种方案的优点:

```
┌─────────────────┐
│  原始数据流      │  全量 560条/秒
│  (hvac-raw)     │
└────────┬────────┘
         │
         ├──────────────┐
         │              │
         ▼              ▼
┌─────────────┐  ┌─────────────┐
│ 实时告警     │  │ 采样器       │
│ (全量数据)   │  │ (1/10采样)  │
│             │  │             │
│ 只做检测     │  │ 56条/秒      │
│ 不存储       │  └──────┬──────┘
└─────────────┘         │
                        ▼
                 ┌─────────────┐
                 │ TimescaleDB │
                 │ (长期存储)   │
                 └─────────────┘
```

### 实现配置

#### 实时告警 Connect (无存储)

```yaml
# realtime-alert-connect.yaml
input:
  kafka:
    addresses: ["redpanda:9092"]
    topics: ["hvac-raw-data"]
    consumer_group: "realtime-alert"

pipeline:
  processors:
    - nb67_parser: {}
    
    # 只做故障检测,不做完整映射
    - mapping: |
        root.device_id = this.device.train_no + "-" + this.device.carriage_no
        root.timestamp = now()
        
        # 故障检测
        root.has_critical_fault = (
          this.faults.blpflt_comp_u11 ||
          this.faults.bflt_highpres_u11
        )
        
        # 只保留有故障的数据
        root = if this.has_critical_fault {
          this
        } else {
          deleted()  # 正常数据丢弃,不存储
        }

output:
  # 发送到告警系统 (Email/SMS/钉钉)
  http_client:
    url: "http://alert-system/api/alerts"
    verb: POST
```

#### 采样存储 Connect

```yaml
# sampled-storage-connect.yaml
input:
  kafka:
    addresses: ["redpanda:9092"]
    topics: ["hvac-raw-data"]
    consumer_group: "sampled-storage"

pipeline:
  processors:
    - nb67_parser: {}
    
    # 10秒采样
    - mapping: |
        root = if (this.timestamp.second % 10) == 0 {
          this
        } else {
          deleted()
        }
    
    # 完整数据映射
    - mapping: |
        # ... 完整映射逻辑
        root.time = ...
        root.device_id = ...
        # 所有字段

output:
  sql_insert:
    driver: postgres
    dsn: "postgres://user:pass@timescaledb:5432/metro_hvac"
    table: "hvac_measurements"
    # ...
```

### 混合方案资源配置

```yaml
Redpanda: 3 节点 × (3C 6G + 300GB)  # 介于两者之间
Connect 告警: 2 实例 × (2C 2G)       # 轻量级
Connect 存储: 2 实例 × (2C 2G)       # 轻量级  
TimescaleDB: 1 主 × (4C 8G + 1TB)   # 中等配置
```

**月成本**: ~$800 (vs 原来 $1,150)  
**年节省**: ~$4,200

## 方案选择建议

### 场景 1: 成本敏感,历史分析为主

**推荐**: 纯采样方案  
**配置**: 优化版全套  
**成本**: $7,320/年  
**备注**: 适合预算有限,主要做趋势分析

### 场景 2: 需要实时告警 + 成本优化 (推荐) ⭐

**推荐**: 混合双轨制方案  
**配置**: 中等配置  
**成本**: $9,600/年  
**备注**: 平衡成本和功能,最佳实践

### 场景 3: 追求极致性能,成本不敏感

**推荐**: 原方案 (全量数据)  
**配置**: 原方案全套  
**成本**: $13,800/年  
**备注**: 适合对实时性要求极高的场景

## 实施步骤

### 快速验证 (1周)

1. 在现有系统中添加采样 processor
2. 对比采样前后的数据质量
3. 确认业务需求是否满足

### 完整迁移 (2周)

1. 部署优化配置的新集群
2. 配置采样逻辑
3. 数据迁移和验证
4. 逐步切换流量

### 监控优化 (持续)

1. 监控采样后的数据覆盖率
2. 调整采样间隔 (比如改为5秒或15秒)
3. 根据实际使用情况优化资源

## 注意事项

### ⚠️ 采样可能带来的问题

1. **瞬时故障遗漏**: 如果故障只持续1-2秒,可能被采样遗漏
   - **缓解**: 实时告警用全量数据,存储用采样数据

2. **趋势平滑**: 10秒采样会丢失短期波动
   - **评估**: 对于空调系统,10秒采样通常足够

3. **时间对齐**: 确保所有设备的采样时间对齐
   - **实现**: 基于绝对时间戳采样 (xx:xx:00, xx:xx:10)

### ✅ 建议

- 保留 Kafka 中的全量数据 7 天,以便回溯
- 定期抽查采样数据的质量
- 对关键设备可以不采样 (动态配置)

## 总结

| 维度 | 全量方案 | 采样方案 | 混合方案 ⭐ |
|------|---------|---------|-----------|
| **成本** | $13,800/年 | $7,320/年 | $9,600/年 |
| **实时告警** | ✅ 优秀 | ❌ 有遗漏风险 | ✅ 优秀 |
| **历史分析** | ✅ 完整 | ✅ 足够 | ✅ 足够 |
| **资源使用** | 高 | 极低 | 中等 |
| **实施难度** | 简单 | 简单 | 中等 |
| **推荐指数** | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

**最终建议**: 采用 **混合双轨制方案**,既保证实时告警能力,又大幅降低存储成本!
