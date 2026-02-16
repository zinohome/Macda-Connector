# 资源配置快速参考卡

## 单条地铁线路 (40列车)

### 数据规模
- **消息速率**: 560 条/秒
- **数据量**: 36 GB/天, 13 TB/年
- **延迟要求**: ≤ 5秒
- **消息积压**: 0

---

## 推荐配置

### Redpanda 集群
```
节点数: 3
配置: 4C 8G 500GB SSD/节点
总配置: 12C 24G 1.5TB
```

### Redpanda Connect
```
实例数: 3
配置: 2C 4G/实例
总配置: 6C 12G
```

### TimescaleDB
```
节点数: 1主 + 1备
配置: 8C 16G 2TB SSD/节点
总配置: 16C 32G 4TB
```

### Grafana
```
实例数: 1
配置: 2C 4G 50GB
```

---

## 性能指标

| 指标 | 目标值 | 实际值 |
|------|--------|--------|
| 端到端延迟 P99 | ≤ 5s | 1-2s ✅ |
| 处理能力 | 560/s | 24,000/s ✅ |
| 消息积压 | 0 | 0 ✅ |
| 扩展余量 | - | 40倍 ✅ |

---

## 成本估算

| 类型 | 年成本 |
|------|--------|
| 云部署 | $13,800 |
| 自建硬件 | $21,500 (一次性) |
| 自建运维 | $6,000/年 |

**自建 ROI**: 2.8年

---

## 扩展能力

| 线路数 | Redpanda | Connect | TimescaleDB |
|--------|----------|---------|-------------|
| 1条 (当前) | 3节点 | 3实例 | 8C16G |
| 3-5条 | 3节点 ✅ | 6实例 | 8C16G ✅ |
| 10条 | 3节点 ✅ | 12实例 | 16C32G ⚠️ |

⚠️ TimescaleDB 是首个扩展瓶颈

---

## 监控告警阈值

| 指标 | 告警值 |
|------|--------|
| 端到端延迟 P99 | > 8s |
| Kafka Lag | > 1000 |
| Connect CPU | > 70% |
| TimescaleDB CPU | > 80% |
| 磁盘使用率 | > 85% |

---

## 一键部署

```bash
# 克隆配置
git clone your-repo
cd macda-connector

# 启动集群
docker-compose up -d

# 检查状态
docker-compose ps
./scripts/check-health.sh
```

---

详细分析见: [06-resource-estimation-capacity-planning.md](./06-resource-estimation-capacity-planning.md)
