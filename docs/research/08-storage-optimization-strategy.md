# 存储策略优化配置

## 数据保留策略

### Redpanda 保留策略

#### 选项 1: 24小时保留 (推荐) ⭐

```yaml
# Redpanda Topic 配置
topic: hvac-data
retention.ms: 86400000  # 24小时 (vs 原来 7天)
```

**存储需求计算**:

**全量方案 (560条/秒)**:
```
24小时数据 = 36 GB/天 × 1 = 36 GB
副本数 = 3
总存储 = 36 GB × 3 = 108 GB (vs 原来 800 GB) 
减少 87% ⬇️
```

**采样方案 (56条/秒)**:
```
24小时数据 = 4.2 GB/天 × 1 = 4.2 GB
副本数 = 3
总存储 = 4.2 GB × 3 = 12.6 GB ≈ 13 GB
减少 98% ⬇️
```

#### 选项 2: 极限优化 - 6小时保留

如果确认数据已及时入库，可以进一步缩短：

```yaml
retention.ms: 21600000  # 6小时
```

**存储需求**:
- 全量: 54 GB (18 GB × 3副本)
- 采样: 6.3 GB (2.1 GB × 3副本)

#### 选项 3: 激进优化 - 1小时保留

仅用于缓冲,确保数据不丢失：

```yaml
retention.ms: 3600000  # 1小时
```

**存储需求**:
- 全量: 9 GB (3 GB × 3副本)
- 采样: 1 GB

**⚠️ 注意**: 需要确保 TimescaleDB 写入可靠,否则有数据丢失风险

### TimescaleDB 保留策略

#### 配置方法

```sql
-- 设置365天自动删除策略
SELECT add_retention_policy('hvac_measurements', INTERVAL '365 days');

-- 查看当前策略
SELECT * FROM timescaledb_information.jobs 
WHERE proc_name = 'policy_retention';

-- 修改策略 (如果需要调整)
SELECT remove_retention_policy('hvac_measurements');
SELECT add_retention_policy('hvac_measurements', INTERVAL '365 days');
```

#### 多级保留策略 (推荐) ⭐

**原始数据**: 保留365天后自动删除
**聚合数据**: 永久保留

```sql
-- 原始数据表: 365天保留
SELECT add_retention_policy('hvac_measurements', INTERVAL '365 days');

-- 1小时聚合视图: 永久保留
-- (连续聚合视图不会被retention policy影响)
CREATE MATERIALIZED VIEW hvac_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', time) AS hour,
    device_id,
    AVG(temp_vehicle_1) AS avg_temp_1,
    AVG(comp_u11_power) AS avg_power,
    COUNT(*) AS measurement_count,
    SUM(CASE WHEN alert_level = 'critical' THEN 1 ELSE 0 END) AS critical_alerts
FROM hvac_measurements
GROUP BY hour, device_id;

-- 1天聚合视图: 永久保留
CREATE MATERIALIZED VIEW hvac_daily
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', time) AS day,
    device_id,
    AVG(temp_vehicle_1) AS avg_temp,
    MAX(comp_u11_power) AS max_power,
    SUM(comp_u11_power) / 1000.0 AS total_kwh,
    COUNT(*) AS data_points
FROM hvac_measurements
GROUP BY day, device_id;
```

**数据生命周期**:
```
0-365天:    原始数据 (详细) + 聚合视图
365天-永久: 仅聚合视图 (小时/天级)
```

## 优化后存储配置

### 方案对比

| 方案 | Redpanda保留 | Redpanda存储 | TimescaleDB存储(1年) | 总存储 |
|------|------------|-------------|---------------------|--------|
| **原方案** | 7天 | 800 GB | 1.1 TB | **1.9 TB** |
| **优化1: 24h保留** | 24h | 108 GB | 1.1 TB | **1.2 TB** ⬇️ 37% |
| **优化2: 24h+采样** | 24h | 13 GB | 150 GB | **163 GB** ⬇️ 91% 🎉 |
| **极限: 6h+采样** | 6h | 6.3 GB | 150 GB | **156 GB** ⬇️ 92% |

### 推荐配置: 24小时 + 采样 ⭐

#### Redpanda 集群配置

```yaml
节点数: 3
每节点配置:
  CPU: 2 核
  内存: 4 GB
  磁盘: 50 GB SSD  # vs 原来 500 GB
  网络: 1 Gbps

Topic 配置:
  retention.ms: 86400000      # 24小时
  retention.bytes: 5368709120 # 5 GB/分区 作为backup限制
  segment.ms: 3600000         # 1小时一个segment
  compression.type: lz4
```

#### TimescaleDB 配置

```yaml
主节点配置:
  CPU: 4 核
  内存: 8 GB
  磁盘: 500 GB SSD  # vs 原来 2 TB
  
压缩策略:
  - 7天后自动压缩 (节省70%空间)
  
保留策略:
  - 原始数据: 365天自动删除
  - 聚合视图: 永久保留
```

## 成本影响

### 云部署成本 (采样方案 + 24h保留)

| 组件 | 原配置 | 优化配置 | 月成本 | 年成本 |
|------|--------|---------|--------|--------|
| Redpanda (3节点) | 4C8G 500GB | 2C4G 50GB | $120 | $1,440 |
| Connect (2实例) | 2C4G | 2C2G | $50 | $600 |
| TimescaleDB | 8C16G 2TB | 4C8G 500GB | $250 | $3,000 |
| Grafana | 2C4G | 2C4G | $50 | $600 |
| 其他 | - | - | $30 | $360 |
| **总计** | - | - | **$500** | **$6,000** |

**vs 原方案**: 节省 **$7,800/年** (57%) 🎉

### 自建成本

**硬件 (一次性)**:
| 组件 | 配置 | 成本 |
|------|------|------|
| 3 × Redpanda | 2C4G 50GB | $3,000 |
| 2 × Connect | 2C2G | $2,000 |
| 1 × TimescaleDB | 4C8G 500GB | $3,000 |
| 1 × Grafana | 2C4G | $800 |
| 网络设备 | - | $1,000 |
| **总计** | - | **$9,800** |

vs 原方案 $21,500: 节省 **$11,700** (54%)

**年运维**: $3,000 (vs 原来 $6,000)

## 配置示例

### Redpanda 配置文件

```yaml
# redpanda.yaml
redpanda:
  data_directory: /var/lib/redpanda/data
  
  kafka_api:
    - address: 0.0.0.0
      port: 9092
  
  admin:
    - address: 0.0.0.0
      port: 9644
  
  # 默认 Topic 配置
  default_topic_replications: 3
  log_segment_size: 536870912  # 512 MB
  
  # 存储配置
  retention_bytes: 10737418240  # 10 GB 总限制

# Topic 配置 (通过 rpk 命令)
rpk topic alter-config hvac-data \
  --set retention.ms=86400000 \
  --set retention.bytes=5368709120 \
  --set segment.ms=3600000 \
  --set compression.type=lz4
```

### TimescaleDB 初始化脚本

```sql
-- init.sql

-- 创建 Hypertable
SELECT create_hypertable('hvac_measurements', 'time');

-- 设置压缩策略 (7天后压缩)
ALTER TABLE hvac_measurements SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'device_id',
  timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('hvac_measurements', INTERVAL '7 days');

-- 设置保留策略 (365天后删除)
SELECT add_retention_policy('hvac_measurements', INTERVAL '365 days');

-- 创建连续聚合视图
CREATE MATERIALIZED VIEW hvac_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', time) AS hour,
    device_id,
    AVG(temp_vehicle_1) AS avg_temp_1,
    AVG(temp_vehicle_2) AS avg_temp_2,
    AVG(comp_u11_power) AS avg_power,
    MAX(comp_u11_power) AS max_power,
    COUNT(*) AS measurement_count,
    SUM(CASE WHEN alert_level != 'normal' THEN 1 ELSE 0 END) AS alert_count
FROM hvac_measurements
GROUP BY hour, device_id;

-- 聚合视图刷新策略
SELECT add_continuous_aggregate_policy('hvac_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

-- 创建日聚合
CREATE MATERIALIZED VIEW hvac_daily
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', time) AS day,
    device_id,
    AVG(temp_vehicle_1) AS avg_temp,
    STDDEV(temp_vehicle_1) AS stddev_temp,
    AVG(comp_u11_power) AS avg_power,
    SUM(comp_u11_power) / 1000.0 AS total_kwh,
    COUNT(*) AS data_points,
    SUM(CASE WHEN alert_level = 'critical' THEN 1 ELSE 0 END) AS critical_alerts,
    SUM(CASE WHEN alert_level = 'warning' THEN 1 ELSE 0 END) AS warning_alerts
FROM hvac_measurements
GROUP BY day, device_id;

SELECT add_continuous_aggregate_policy('hvac_daily',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');

-- 创建索引
CREATE INDEX idx_device_time ON hvac_measurements (device_id, time DESC);
CREATE INDEX idx_alert ON hvac_measurements (alert_level, time DESC) 
    WHERE alert_level != 'normal';

-- 查看配置
SELECT * FROM timescaledb_information.compression_settings;
SELECT * FROM timescaledb_information.jobs;
```

## 监控和告警

### 存储监控指标

```yaml
# Prometheus 告警规则
groups:
  - name: storage_alerts
    rules:
      # Redpanda 存储告警
      - alert: RedpandaDiskUsageHigh
        expr: redpanda_storage_disk_free_bytes < 10737418240  # < 10 GB
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Redpanda disk usage high"
          
      # TimescaleDB 存储告警
      - alert: TimescaleDBDiskUsageHigh
        expr: pg_database_size_bytes{datname="metro_hvac"} > 429496729600  # > 400 GB
        for: 30m
        labels:
          severity: warning
          
      # 数据保留检查
      - alert: OldDataNotDeleted
        expr: |
          max(time_bucket('1 day', time)) - min(time_bucket('1 day', time)) > 370
        for: 1h
        labels:
          severity: critical
        annotations:
          summary: "Retention policy not working"
```

### 定期检查脚本

```bash
#!/bin/bash
# check-storage.sh

echo "=== Redpanda Storage Usage ==="
rpk cluster storage list

echo ""
echo "=== TimescaleDB Storage Usage ==="
docker exec timescaledb psql -U hvac_user -d metro_hvac -c "
SELECT 
    pg_size_pretty(pg_total_relation_size('hvac_measurements')) as total_size,
    pg_size_pretty(pg_relation_size('hvac_measurements')) as table_size,
    pg_size_pretty(pg_indexes_size('hvac_measurements')) as index_size;
"

echo ""
echo "=== Data Age Check ==="
docker exec timescaledb psql -U hvac_user -d metro_hvac -c "
SELECT 
    MIN(time) as oldest_data,
    MAX(time) as newest_data,
    MAX(time) - MIN(time) as data_span
FROM hvac_measurements;
"

echo ""
echo "=== Compression Stats ==="
docker exec timescaledb psql -U hvac_user -d metro_hvac -c "
SELECT 
    pg_size_pretty(before_compression_total_bytes) as before,
    pg_size_pretty(after_compression_total_bytes) as after,
    ROUND(100.0 * (1 - after_compression_total_bytes::numeric / 
                        before_compression_total_bytes::numeric), 2) as compression_ratio
FROM timescaledb_information.compression_settings
WHERE hypertable_name = 'hvac_measurements';
"
```

## 迁移建议

### 从长保留迁移到短保留

```bash
# 1. 修改 Redpanda 保留策略
rpk topic alter-config hvac-data --set retention.ms=86400000

# 2. 等待旧数据自动过期 (会在下个segment删除时生效)
# 或手动触发
rpk topic delete-records hvac-data --offset-before <timestamp>

# 3. 验证
rpk topic describe hvac-data
```

### TimescaleDB 保留策略迁移

```sql
-- 1. 添加新策略前,先删除超过365天的数据
DELETE FROM hvac_measurements WHERE time < NOW() - INTERVAL '365 days';

-- 2. 手动 VACUUM 回收空间
VACUUM FULL hvac_measurements;

-- 3. 添加自动保留策略
SELECT add_retention_policy('hvac_measurements', INTERVAL '365 days');

-- 4. 验证
SELECT * FROM timescaledb_information.jobs WHERE proc_name = 'policy_retention';
```

## 总结

### 最终推荐配置 ⭐

| 配置项 | 值 | 说明 |
|--------|---|------|
| **Redpanda保留** | 24小时 | 足够容错,大幅减少存储 |
| **TimescaleDB原始数据** | 365天 | 满足合规要求 |
| **TimescaleDB聚合** | 永久 | 历史趋势分析 |
| **采样策略** | 10秒 | 90%成本节省 |

### 存储节省

```
原方案总存储: 1.9 TB
优化后总存储: 163 GB
节省比例: 91% 🎉
```

### 成本节省

```
云部署年成本: $13,800 → $6,000 (节省 57%)
自建硬件成本: $21,500 → $9,800 (节省 54%)
```

### 注意事项

⚠️ **Redpanda 24小时保留的影响**:
- 数据入库必须在24小时内完成 ✅ (实际延迟<5秒)
- 无法回溯24小时前的原始消息 ⚠️
- 建议监控 Kafka Lag,确保<1000

✅ **推荐做法**:
- 启用 Redpanda Connect 的错误重试机制
- 监控 TimescaleDB 写入成功率
- 定期检查数据完整性
- 保留7天的数据库备份
