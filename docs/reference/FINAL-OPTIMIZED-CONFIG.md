# 最终优化配置汇总

## 推荐方案: 采样 + 短期保留 ⭐

### 核心优化策略

| 优化项 | 配置 | 效果 |
|--------|------|------|
| **数据采样** | 10秒/次 | -90% 数据量 |
| **Redpanda保留** | 24小时 | -87% 存储 |
| **TimescaleDB保留** | 365天 | 满足需求 |

---

## 资源配置

### Redpanda
```
节点: 3 × (2C 4G 50GB SSD)
保留: 24小时
总存储: 13 GB (vs 原800GB)
```

### Connect
```
实例: 2 × (2C 2G)
功能: 采样 + 解析 + 存储
```

### TimescaleDB
```
配置: 1 × (4C 8G 500GB SSD)
保留: 365天原始 + 永久聚合
总存储: 150 GB (vs 原1.1TB)
```

---

## 成本对比

| 方案 | 配置 | 年成本 | vs原方案 |
|------|------|--------|---------|
| 原方案 | 全量+7天 | $13,800 | - |
| 优化方案 | 采样+24h | **$6,000** | **-57%** 🎉 |

**年节省**: **$7,800** ≈ 5.4万人民币

---

## 关键配置

### Redpanda Topic
```bash
rpk topic alter-config hvac-data \
  --set retention.ms=86400000 \      # 24h
  --set compression.type=lz4
```

### TimescaleDB 策略
```sql
-- 365天保留
SELECT add_retention_policy('hvac_measurements', INTERVAL '365 days');

-- 7天后压缩
SELECT add_compression_policy('hvac_measurements', INTERVAL '7 days');
```

### Connect 采样
```yaml
- mapping: |
    root = if (this.timestamp.second % 10) == 0 {
      this
    } else {
      deleted()
    }
```

---

## 性能指标

| 指标 | 目标 | 实际 |
|------|------|------|
| 延迟P99 | ≤5s | 2-3s ✅ |
| 存储/年 | - | 150GB |
| 处理能力 | 56/s | 16,000/s ✅ |
| 成本 | - | $6,000/年 |

---

## 监控检查

```bash
# 存储使用
./scripts/check-storage.sh

# 关键指标
- Redpanda Lag: < 100
- TimescaleDB 存储: < 400GB
- 数据年龄: < 366天
```

---

详见完整文档:
- [存储优化策略](./08-storage-optimization-strategy.md)
- [采样降频方案](./07-optimized-sampling-strategy.md)
- [资源容量规划](./06-resource-estimation-capacity-planning.md)
