# TimescaleDB 集成方案

## TimescaleDB 简介

TimescaleDB 是基于 PostgreSQL 的时序数据库扩展,非常适合地铁空调这类时序数据场景。

### 核心特性

- **完全 PostgreSQL 兼容**: 支持所有 SQL 功能
- **自动分区**: 按时间自动分片 (hypertables)  
- **高性能写入**: 针对时序数据优化
- **数据压缩**: 自动压缩历史数据
- **连续聚合**: 物化视图自动更新
- **数据保留策略**: 自动删除旧数据

## Redpanda Connect 与 TimescaleDB 集成

### 1. 连接方式

Redpanda Connect 通过标准 PostgreSQL 驱动连接 TimescaleDB:

```yaml
output:
  sql_insert:
    driver: postgres
    dsn: "postgres://username:password@timescaledb-host:5432/database?sslmode=require"
    table: "hvac_measurements"
    columns: [...]
```

### 2. TimescaleDB 表结构设计

#### 创建 Hypertable

```sql
-- 创建普通 PostgreSQL 表
CREATE TABLE hvac_measurements (
    time TIMESTAMPTZ NOT NULL,
    device_id TEXT NOT NULL,
    car_number INTEGER,
    
    -- 温度相关字段
    temp_supply_air REAL,
    temp_return_air REAL,
    temp_outdoor REAL,
    temp_setpoint REAL,
    
    -- 压力相关字段
    pressure_compressor REAL,
    pressure_condenser REAL,
    pressure_evaporator REAL,
    
    -- 电气相关字段
    voltage_input REAL,
    current_compressor REAL,
    current_fan REAL,
    power_total REAL,
    
    -- 运行状态
    compressor_status INTEGER,
    fan_speed INTEGER,
    mode INTEGER,
    
    -- 故障代码
    fault_code INTEGER,
    warning_code INTEGER,
    
    -- 计算字段
    temp_diff REAL,
    alert_level TEXT,
    needs_maintenance BOOLEAN,
    
    -- 元数据
    processed_at TIMESTAMPTZ DEFAULT NOW(),
    data_quality REAL
);

-- 转换为 Hypertable (核心步骤)
SELECT create_hypertable('hvac_measurements', 'time');

-- 创建索引
CREATE INDEX idx_device_time ON hvac_measurements (device_id, time DESC);
CREATE INDEX idx_alert_level ON hvac_measurements (alert_level, time DESC) 
    WHERE alert_level IN ('warning', 'critical');
```

#### 设备元数据表

```sql
CREATE TABLE devices (
    device_id TEXT PRIMARY KEY,
    car_number INTEGER,
    line_number INTEGER,
    train_number TEXT,
    installation_date DATE,
    model TEXT,
    manufacturer TEXT,
    last_maintenance_date DATE,
    status TEXT
);
```

### 3. Redpanda Connect 完整配置

```yaml
input:
  kafka:
    addresses: ["redpanda:9092"]
    topics: ["hvac-data"]
    consumer_group: "timescaledb-writer"
    start_from_oldest: false

pipeline:
  threads: 4  # 并行处理
  
  processors:
    # 解码 Avro 数据
    - schema_registry_decode:
        url: http://schema-registry:8081
    
    # 数据转换和验证
    - mapping: |
        # 时间字段转换
        root.time = this.timestamp.ts_unix()
        
        # 设备信息
        root.device_id = this.device_id
        root.car_number = this.car_number
        
        # 传感器数据
        root.temp_supply_air = this.temp_supply_air
        root.temp_return_air = this.temp_return_air
        root.temp_outdoor = this.temp_outdoor
        root.temp_setpoint = this.temp_setpoint
        
        root.pressure_compressor = this.pressure_compressor
        root.pressure_condenser = this.pressure_condenser  
        root.pressure_evaporator = this.pressure_evaporator
        
        root.voltage_input = this.voltage_input
        root.current_compressor = this.current_compressor
        root.current_fan = this.current_fan
        
        # 计算派生字段
        root.temp_diff = this.temp_supply_air - this.temp_return_air
        root.power_total = this.voltage_input * (this.current_compressor + this.current_fan)
        
        # 状态字段
        root.compressor_status = this.compressor_status
        root.fan_speed = this.fan_speed
        root.mode = this.mode
        root.fault_code = this.fault_code.or(0)
        root.warning_code = this.warning_code.or(0)
        
        # 数据质量评分 (0-1)
        let missing_fields = 0
        let missing_fields = if this.temp_supply_air == null { $missing_fields + 1 } else { $missing_fields }
        let missing_fields = if this.pressure_compressor == null { $missing_fields + 1 } else { $missing_fields }
        root.data_quality = 1.0 - ($missing_fields / 20.0)  # 假设20个关键字段
    
    # 故障检测
    - mapping: |
        root = this
        
        # 定义阈值
        let temp_high = this.temp_supply_air > 35
        let temp_low = this.temp_supply_air < 10
        let pressure_high = this.pressure_compressor > 2000
        let pressure_low = this.pressure_compressor < 500
        let current_high = this.current_compressor > 50
        
        # 综合判断
        root.alert_level = if $temp_high && $pressure_high && $current_high {
          "critical"
        } else if $temp_high || $pressure_high || $current_high {
          "warning"  
        } else if $temp_low || $pressure_low {
          "warning"
        } else {
          "normal"
        }
        
        # 维护建议
        root.needs_maintenance = (
          $current_high || 
          this.fault_code != 0 ||
          this.data_quality < 0.8
        )
        
        root.processed_at = now()

output:
  sql_insert:
    driver: postgres
    dsn: "postgres://hvac_user:secure_password@timescaledb:5432/metro_hvac?sslmode=require"
    table: "hvac_measurements"
    columns:
      - time
      - device_id
      - car_number
      - temp_supply_air
      - temp_return_air
      - temp_outdoor
      - temp_setpoint
      - pressure_compressor
      - pressure_condenser
      - pressure_evaporator
      - voltage_input
      - current_compressor
      - current_fan
      - power_total
      - compressor_status
      - fan_speed
      - mode
      - fault_code
      - warning_code
      - temp_diff
      - alert_level
      - needs_maintenance
      - processed_at
      - data_quality
    
    args_mapping: |
      root = [
        this.time,
        this.device_id,
        this.car_number,
        this.temp_supply_air,
        this.temp_return_air,
        this.temp_outdoor,
        this.temp_setpoint,
        this.pressure_compressor,
        this.pressure_condenser,
        this.pressure_evaporator,
        this.voltage_input,
        this.current_compressor,
        this.current_fan,
        this.power_total,
        this.compressor_status,
        this.fan_speed,
        this.mode,
        this.fault_code,
        this.warning_code,
        this.temp_diff,
        this.alert_level,
        this.needs_maintenance,
        this.processed_at,
        this.data_quality
      ]
    
    # 批处理配置 - 提升写入性能
    batching:
      count: 1000          # 1000条记录批量插入
      period: 5s           # 或每5秒插入一次
      byte_size: 5MB       # 或累积到5MB
    
    # 并发控制
    max_in_flight: 10      # 最多10个并发批次
    
    # 连接池配置
    conn_max_idle: 5
    conn_max_open: 20
    conn_max_idle_time: 30s
    conn_max_life_time: 1h

# 监控配置
metrics:
  prometheus:
    path: /metrics
    port: 9090

logger:
  level: INFO
  format: json
```

### 4. TimescaleDB 优化配置

#### 压缩策略

```sql
-- 启用压缩 (历史数据)
ALTER TABLE hvac_measurements SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'device_id',
  timescaledb.compress_orderby = 'time DESC'
);

-- 自动压缩 7天以前的数据
SELECT add_compression_policy('hvac_measurements', INTERVAL '7 days');
```

#### 数据保留策略

```sql
-- 自动删除 1年以前的数据
SELECT add_retention_policy('hvac_measurements', INTERVAL '1 year');
```

#### 连续聚合 (实时统计)

```sql
-- 创建1小时聚合视图
CREATE MATERIALIZED VIEW hvac_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', time) AS hour,
    device_id,
    AVG(temp_supply_air) AS avg_temp_supply,
    AVG(temp_return_air) AS avg_temp_return,
    AVG(pressure_compressor) AS avg_pressure,
    AVG(power_total) AS avg_power,
    MAX(temp_supply_air) AS max_temp_supply,
    MIN(temp_supply_air) AS min_temp_supply,
    COUNT(*) AS measurement_count,
    SUM(CASE WHEN alert_level = 'critical' THEN 1 ELSE 0 END) AS critical_alerts,
    SUM(CASE WHEN alert_level = 'warning' THEN 1 ELSE 0 END) AS warning_alerts
FROM hvac_measurements
GROUP BY hour, device_id;

-- 自动刷新策略
SELECT add_continuous_aggregate_policy('hvac_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');
```

#### 日聚合视图

```sql
CREATE MATERIALIZED VIEW hvac_daily
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', time) AS day,
    device_id,
    car_number,
    AVG(temp_supply_air) AS avg_temp,
    STDDEV(temp_supply_air) AS stddev_temp,
    AVG(power_total) AS avg_power,
    SUM(power_total) / 1000.0 AS total_kwh,  # 总耗电量
    COUNT(*) AS measurement_count,
    SUM(CASE WHEN needs_maintenance = true THEN 1 ELSE 0 END) AS maintenance_alerts,
    AVG(data_quality) AS avg_data_quality
FROM hvac_measurements
GROUP BY day, device_id, car_number;
```

### 5. 查询示例

#### 实时监控查询

```sql
-- 获取最新状态
SELECT 
    device_id,
    car_number,
    temp_supply_air,
    pressure_compressor,
    alert_level,
    time
FROM hvac_measurements
WHERE time > NOW() - INTERVAL '5 minutes'
ORDER BY time DESC;

-- 当前告警
SELECT 
    device_id,
    car_number,
    alert_level,
    fault_code,
    time
FROM hvac_measurements
WHERE time > NOW() - INTERVAL '1 hour'
  AND alert_level IN ('warning', 'critical')
ORDER BY time DESC;
```

#### 历史趋势分析

```sql
-- 温度趋势 (使用聚合视图)
SELECT 
    hour,
    device_id,
    avg_temp_supply,
    max_temp_supply,
    min_temp_supply
FROM hvac_hourly
WHERE device_id = 'HVAC-001'
  AND hour > NOW() - INTERVAL '7 days'
ORDER BY hour;

-- 能耗分析
SELECT 
    day,
    device_id,
    total_kwh,
    avg_power
FROM hvac_daily
WHERE day > NOW() - INTERVAL '30 days'
ORDER BY day;
```

#### 故障分析

```sql
-- 故障频率统计
SELECT 
    device_id,
    fault_code,
    COUNT(*) AS occurrence_count,
    MIN(time) AS first_occurrence,
    MAX(time) AS last_occurrence,
    AVG(EXTRACT(EPOCH FROM (
        LEAD(time) OVER (PARTITION BY device_id, fault_code ORDER BY time) - time
    ))) / 3600 AS avg_hours_between_faults
FROM hvac_measurements
WHERE fault_code != 0
  AND time > NOW() - INTERVAL '90 days'
GROUP BY device_id, fault_code
ORDER BY occurrence_count DESC;
```

### 6. 性能优化建议

#### 批量写入策略

```yaml
# Redpanda Connect 配置
output:
  sql_insert:
    batching:
      count: 1000     # 根据数据频率调整
      period: 5s      # 根据实时性要求调整
```

**权衡**:
- 批次越大,吞吐量越高,但延迟增加
- 建议: 1000-5000 条/批,或 1-10秒/批

#### 分区策略

```sql
-- 按时间和设备 ID 分区 (如果设备很多)
SELECT create_hypertable('hvac_measurements', 'time', 
    partitioning_column => 'device_id',
    number_partitions => 4);
```

#### 并行写入

```yaml
# 多个 Redpanda Connect 实例
# 通过 Kafka consumer group 自动负载均衡
```

### 7. 监控和运维

#### 监控指标

```sql
-- 表大小
SELECT 
    pg_size_pretty(pg_total_relation_size('hvac_measurements')) AS total_size,
    pg_size_pretty(pg_relation_size('hvac_measurements')) AS table_size,
    pg_size_pretty(pg_indexes_size('hvac_measurements')) AS index_size;

-- 压缩效果
SELECT 
    pg_size_pretty(before_compression_total_bytes) AS before,
    pg_size_pretty(after_compression_total_bytes) AS after,
    ROUND(100.0 * (1 - after_compression_total_bytes::numeric / 
                        before_compression_total_bytes::numeric), 2) AS compression_ratio
FROM timescaledb_information.compression_settings
WHERE hypertable_name = 'hvac_measurements';

-- 写入速率
SELECT 
    time_bucket('1 minute', time) AS minute,
    COUNT(*) AS inserts_per_minute
FROM hvac_measurements
WHERE time > NOW() - INTERVAL '1 hour'
GROUP BY minute
ORDER BY minute DESC;
```

#### Grafana 集成

TimescaleDB 完全兼容 PostgreSQL,可直接使用 Grafana PostgreSQL 数据源:

```sql
-- Grafana 查询示例
SELECT 
    $__timeGroup(time, $__interval),
    device_id,
    AVG(temp_supply_air) AS temperature
FROM hvac_measurements
WHERE 
    $__timeFilter(time)
    AND device_id IN ($device_id)
GROUP BY 1, 2
ORDER BY 1;
```

## 总结

### 优势

✅ **无缝集成**: Redpanda Connect 通过标准 PostgreSQL 驱动直接连接  
✅ **高性能**: 批量写入 + TimescaleDB 优化,可达 100K+ inserts/s  
✅ **SQL 支持**: 完整 SQL 功能,易于查询和分析  
✅ **自动优化**: 压缩、分区、聚合自动化  
✅ **可扩展**: 支持分布式部署

### 注意事项

⚠️ **时间戳格式**: 确保使用 TIMESTAMPTZ 类型  
⚠️ **批处理调优**: 根据数据量和延迟要求调整批次大小  
⚠️ **索引策略**: 根据查询模式创建合适的索引  
⚠️ **资源规划**: TimescaleDB 需要足够的内存和磁盘空间

### 推荐配置

- **批次大小**: 1000-2000 条
- **批次间隔**: 5-10 秒
- **并发连接**: 10-20
- **压缩策略**: 7天后自动压缩
- **保留策略**: 1年后自动删除
