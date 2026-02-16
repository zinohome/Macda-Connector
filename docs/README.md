# 地铁空调数据分析项目研究文档

## 背景

本项目旨在处理和分析地铁空调系统产生的复杂时序数据,实现:
- 实时数据监控
- 故障检测和告警
- 故障预测和预防性维护
- 数据可视化展示

## 研究目标

评估使用 **Redpanda Connect** 组件处理二进制格式的空调数据(约100+字段),并存储到 **TimescaleDB** 时序数据库的技术方案可行性。

## 核心需求

1. **数据采集**: 边缘设备采集空调传感器数据,写入二进制文件
2. **数据汇聚**: 数据汇聚到 Redpanda (Kafka兼容)消息队列
3. **数据解析**: 解析二进制格式,提取100+字段
4. **数据存储**: 存入 TimescaleDB 时序数据库
5. **实时监控**: 通过 Grafana 等工具展示
6. **故障检测**: 实时检测异常状态
7. **故障预测**: 基于历史数据预测潜在故障

## 研究结论

### ✅ 总体评估: **完全可行**

使用 Redpanda Connect + TimescaleDB 方案 **技术可行、架构合理、成本可控**。

### 📊 关键发现

| 评估维度 | 结果 | 说明 |
|---------|------|------|
| **技术可行性** | ✅ 优秀 | 所有需求均可实现 |
| **性能** | ✅ 优秀 | 轻松应对预期负载 |
| **复杂度** | ✅ 低 | 架构简洁,易于维护 |
| **成本** | ✅ 合理 | 约 $10K/年(云) 或更低(自建) |
| **实施周期** | ✅ 短 | 约 2.5 个月可上线 |
| **可扩展性** | ✅ 强 | 支持水平扩展 |

## 文档目录

### [01. Redpanda Connect 概览](./01-redpanda-connect-overview.md)

**内容**:
- Redpanda Connect 核心特性
- 输入/输出连接器
- Bloblang 映射语言
- Schema Registry 集成
- 部署方式
- 与其他方案对比

**适合阅读者**: 所有团队成员

### [02. 二进制数据处理分析](./02-binary-data-processing-analysis.md)

**内容**:
- 二进制数据处理能力分析
- Avro/Protobuf 格式支持
- 自定义解析方案对比
- 推荐实施方案 (转换服务 + Avro)
- 性能评估
- 代码示例

**适合阅读者**: 开发工程师

### [03. TimescaleDB 集成方案](./03-timescaledb-integration.md)

**内容**:
- TimescaleDB 核心特性
- 表结构设计 (Hypertable)
- Redpanda Connect 配置详解
- 压缩和保留策略
- 连续聚合(实时统计)
- 性能优化建议
- 监控和运维

**适合阅读者**: DBA、运维工程师

### [04. 架构可行性评估](./04-architecture-feasibility-assessment.md)

**内容**:
- 整体架构设计
- 详细实施方案
- 故障检测规则引擎
- 故障预测模型训练
- 部署架构 (Docker Compose)
- 性能和成本评估
- 风险和挑战
- 实施路线图

**适合阅读者**: 架构师、项目经理、决策者

## 推荐技术架构

```
┌─────────────┐
│ 边缘采集层   │  二进制数据文件
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ 转换服务     │  Binary → Avro
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Redpanda    │  消息队列 + Schema Registry
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Redpanda    │  数据解析 + 清洗 + 故障检测
│ Connect     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ TimescaleDB │  时序数据存储 + 自动聚合
└──────┬──────┘
       │
       ├──────────┐
       ▼          ▼
┌──────────┐ ┌──────────┐
│ Grafana  │ │  ML模型  │
│ 实时监控  │ │  故障预测 │
└──────────┘ └──────────┘
```

## 核心技术栈

| 组件 | 技术 | 作用 |
|------|------|------|
| 消息队列 | **Redpanda** | 数据汇聚和分发 |
| Schema管理 | **Schema Registry** | Avro schema 版本管理 |
| 流处理 | **Redpanda Connect** | 数据解析、转换、故障检测 |
| 时序数据库 | **TimescaleDB** | 数据存储和查询 |
| 可视化 | **Grafana** | 实时监控面板 |
| 故障预测 | **Python (scikit-learn)** | 机器学习模型 |

## 关键优势

### 1. 架构简洁
- 组件少,易于理解和维护
- 无需复杂的大数据平台 (Hadoop/Spark)
- 声明式配置,无需大量编码

### 2. 性能优秀
- Redpanda: 高吞吐、低延迟
- TimescaleDB: 针对时序数据优化
- 批量写入: 提升数据库性能

### 3. 开发效率高
- Bloblang: 简单的映射语言
- SQL: 熟悉的查询语言
- 丰富的文档和示例

### 4. 运维友好
- 容器化部署 (Docker/Kubernetes)
- 完善的监控指标
- 自动化数据管理 (压缩、保留)

### 5. 成本可控
- 开源组件为主
- 资源需求合理
- 可以自建或使用云服务

## 实施建议

### Phase 1: POC 验证 (2周)

**目标**: 验证技术方案可行性

**任务**:
- [ ] 搭建单机测试环境
- [ ] 开发转换服务原型
- [ ] 验证二进制数据解析
- [ ] 测试端到端数据流
- [ ] 性能基准测试

### Phase 2: MVP 开发 (6周)

**目标**: 实现核心功能

**任务**:
- [ ] 完善转换服务
- [ ] Redpanda Connect 配置和调优
- [ ] TimescaleDB 表结构设计
- [ ] 基础监控面板开发
- [ ] 简单故障检测规则
- [ ] 集成测试

### Phase 3: 试运行 (4周)

**目标**: 小规模上线验证

**任务**:
- [ ] 部署高可用集群
- [ ] 接入少量设备 (10-20个)
- [ ] 监控数据质量和系统性能
- [ ] 调优和问题修复
- [ ] 运维培训

### Phase 4: 全量上线 (持续)

**目标**: 生产环境稳定运行

**任务**:
- [ ] 所有设备接入
- [ ] 完善告警规则
- [ ] 性能持续优化
- [ ] 积累故障数据

### Phase 5: 智能化 (3-6个月后)

**目标**: 故障预测和预防性维护

**任务**:
- [ ] 特征工程
- [ ] ML 模型训练和评估
- [ ] 模型部署和监控
- [ ] 预测性维护建议

## 快速开始

### 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- 8 GB+ RAM
- 50 GB+ 磁盘空间

### 启动测试环境

```bash
# 克隆项目
cd Macda-Connector

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 访问监控面板
open http://localhost:3000  # Grafana (admin/admin)
```

### 发送测试数据

```bash
# 运行转换服务
python converter/binary_to_avro.py

# 查看 Redpanda Connect 日志
docker-compose logs -f connect-1

# 查询 TimescaleDB
docker-compose exec timescaledb psql -U hvac_user -d metro_hvac -c "SELECT * FROM hvac_measurements LIMIT 10;"
```

## 性能指标

### 预期负载

- **设备数量**: 500
- **采样频率**: 5 分钟/次
- **消息速率**: ~1.67 msg/s
- **数据量**: ~14 GB/天, ~5 TB/年

### 系统容量

- **Redpanda**: 支持 100K+ msg/s
- **Redpanda Connect**: 支持 10K+ msg/s  
- **TimescaleDB**: 支持 100K+ inserts/s (批量)

**结论**: 系统容量远超预期负载,有充足的扩展空间。

## 成本估算

### 云部署 (AWS/阿里云)

| 资源 | 配置 | 月成本 | 年成本 |
|------|------|--------|--------|
| Redpanda 集群 | 3 × c5.xlarge (4C8G) | ~$250 | ~$3,000 |
| TimescaleDB | 1 × c5.2xlarge (8C16G) + 备份 | ~$400 | ~$5,000 |
| Redpanda Connect | 2 × t3.medium (2C4G) | ~$80 | ~$1,000 |
| 转换服务 | 2 × t3.medium (2C4G) | ~$80 | ~$1,000 |
| Grafana | 1 × t3.medium (2C4G) | ~$40 | ~$500 |
| 存储 (5TB) | 5TB SSD | ~$500 | ~$6,000 |
| **总计** | | **~$1,350/月** | **~$16,500/年** |

### 自建机房

硬件一次性投入: ~$15,000  
运维成本: ~$5,000/年  
**总计**: 第一年 ~$20,000, 后续 ~$5,000/年

**ROI**: 自建方案约 1.5 年回本

## 风险和缓解

| 风险 | 影响 | 可能性 | 缓解措施 |
|------|------|--------|---------|
| 二进制格式变更 | 高 | 中 | Schema Registry 版本管理 |
| 数据质量问题 | 中 | 中 | 数据验证 + 质量监控 |
| 性能瓶颈 | 中 | 低 | 性能测试 + 预留容量 |
| 系统故障 | 高 | 低 | 高可用部署 + 监控告警 |
| 存储成本 | 中 | 中 | 压缩 + 数据保留策略 |

## 常见问题 (FAQ)

### Q1: 为什么选择 Redpanda 而不是 Kafka?

**A**: Redpanda 是 Kafka 的直接替代品,具有以下优势:
- ✅ 更高的性能 (10倍吞吐量)
- ✅ 更低的延迟
- ✅ 更少的资源消耗 (不依赖 Zookeeper)
- ✅ 更简单的运维
- ✅ 完全兼容 Kafka API

如果团队已有 Kafka 集群,也可以直接使用 Kafka。

### Q2: 为什么不直接用 Bloblang 解析二进制数据?

**A**: Bloblang 对复杂二进制类型 (如 little-endian float32) 支持有限。使用独立的转换服务:
- ✅ 更灵活 (Python/Go 等语言二进制处理能力强)
- ✅ 更容易维护 (代码而非配置)
- ✅ 更好的错误处理
- ✅ Schema Registry 提供版本管理

### Q3: TimescaleDB vs InfluxDB?

**A**: 两者都是时序数据库,但:

| 特性 | TimescaleDB | InfluxDB |
|------|-------------|----------|
| SQL 支持 | ✅ 完整 PostgreSQL | ❌ InfluxQL (受限) |
| 生态系统 | PostgreSQL 生态 | 专有生态 |
| JOIN 查询 | ✅ 支持 | ❌ 有限 |
| 聚合函数 | ✅ 丰富 | ✅ 丰富 |
| 成本 | 开源 | 企业功能收费 |

对于需要复杂查询的场景,TimescaleDB 更合适。

### Q4: 能否实现精确一次 (Exactly-Once) 语义?

**A**: 
- Redpanda Connect → TimescaleDB: 默认 **至少一次** (At-Least-Once)
- 可以通过幂等性保证 (如使用 `device_id + timestamp` 作为主键)
- 真正的精确一次需要分布式事务,但对于监控场景,至少一次通常足够

### Q5: 如何处理数据延迟?

**A**:
- 网络延迟: 优化网络连接
- 处理延迟: 调整批处理参数,减小批次大小
- 写入延迟: 增加 TimescaleDB 连接池

通常延迟在秒级是可以接受的。

## 后续优化方向

1. **实时流计算**
   - 考虑引入 Flink 处理更复杂的窗口聚合

2. **边缘计算**
   - 在边缘端直接进行初步故障检测,减少数据传输

3. **联邦学习**
   - 多车辆数据联合训练故障预测模型

4. **数字孪生**
   - 构建空调系统的数字孪生模型,进行仿真测试

## 参考资源

### 官方文档

- [Redpanda Connect 文档](https://docs.redpanda.com/redpanda-connect/home/)
- [TimescaleDB 文档](https://docs.timescale.com/)
- [Bloblang 语言](https://docs.redpanda.com/redpanda-connect/guides/bloblang/about/)
- [Grafana 文档](https://grafana.com/docs/)

### 相关项目

- [Redpanda GitHub](https://github.com/redpanda-data/redpanda)
- [Redpanda Connect GitHub](https://github.com/redpanda-data/connect)
- [TimescaleDB GitHub](https://github.com/timescale/timescaledb)

### 社区支持

- [Redpanda Community Slack](https://redpanda.com/slack)
- [TimescaleDB Community](https://timescaledb.slack.com)

## 团队联系

如有疑问,请联系项目团队。

---

**文档版本**: 1.0  
**最后更新**: 2026-02-17  
**下次审查**: 2026-03-17
