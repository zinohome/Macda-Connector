# MACDA-Connector

> 地铁列车 HVAC 系统数据处理平台 - Redpanda Connect 技术研究与 Faust 迁移方案

## 项目概述

本项目旨在研究和设计基于 **Redpanda Connect** 的现代化流处理架构，用于替代现有的 Faust-based 地铁 HVAC 遥测数据处理系统。

### 核心目标

- 🔍 **技术评估**：研究 Redpanda Connect 在复杂二进制数据处理场景的可行性
- 📊 **架构设计**：设计高性能、低成本的流处理架构
- 🔄 **迁移策略**：制定从 Faust (Python) 到 Redpanda Connect (Go) 的完整迁移方案
- 💰 **成本优化**：通过数据采样、存储优化实现 50-70% 成本降低

## 系统架构

### 当前系统 (Faust)
- **技术栈**：Python + Faust + Kafka + TimescaleDB
- **数据量**：560 设备 × 1 Hz = 560 msg/sec
- **二进制格式**：NB67 (632 字节，200+ 字段)
- **处理模式**：Parse → Store → Predict (3 个独立进程)

### 目标系统 (Redpanda Connect)
- **技术栈**：Go + Redpanda Connect + Redpanda + TimescaleDB
- **预期性能**：5,000+ msg/sec (9x 提升)
- **二进制解析**：Go 插件 + Kaitai Struct
- **架构**：单二进制文件，多管道并行

## 项目结构

```
Macda-Connector/
├── docs/                           # 研究文档
│   ├── README.md                   # 文档导航
│   ├── 01-redpanda-connect-overview.md              # Redpanda Connect 概述
│   ├── 02-binary-data-processing-analysis.md        # 二进制数据处理分析
│   ├── 03-timescaledb-integration.md                # TimescaleDB 集成方案
│   ├── 04-architecture-feasibility-assessment.md    # 架构可行性评估
│   ├── 05-simplified-go-plugin-architecture.md      # Go 插件架构（推荐）
│   ├── 06-resource-estimation-capacity-planning.md  # 资源估算与容量规划
│   ├── 07-optimized-sampling-strategy.md            # 数据采样优化策略
│   ├── 08-storage-optimization-strategy.md          # 存储优化策略
│   ├── 09-faust-implementation-analysis.md          # Faust 实现分析（英文）
│   ├── 09-faust-implementation-analysis-cn.md       # Faust 实现分析（中文）
│   ├── RESOURCE-QUICK-REF.md                        # 资源配置速查卡
│   ├── SOLUTION-COMPARISON.md                       # 方案对比表
│   └── FINAL-OPTIMIZED-CONFIG.md                    # 最终优化配置
├── oldproj/                        # 旧版 Faust 实现（不包含在 Git）
│   └── MACDA-NB67/                # 原 Python Faust 项目
└── .gitignore                      # Git 忽略配置
```

## 研究成果总结

### 📈 性能提升预期

| 指标 | Faust (当前) | Redpanda Connect (预计) | 提升 |
|------|--------------|------------------------|------|
| **吞吐量** | 560 msg/sec | 5,000+ msg/sec | **9x** |
| **延迟 (p99)** | ~500ms | ~100ms | **5x 更快** |
| **内存用量** | ~500MB/实例 | ~150MB/实例 | **减少 70%** |
| **CPU 用量** | 8-12 核 (3 进程) | 2-4 核 (1 进程) | **减少 60%** |

### 💰 成本优化策略

1. **数据采样优化** (07-optimized-sampling-strategy.md)
   - 10 秒采样间隔 → 数据量减少 90%
   - 混合管道：实时告警 + 采样存储
   - **成本降低**：月成本从 $2,100 降至 $900 (云) / ¥14,000 降至 ¥6,500 (硬件)

2. **存储优化策略** (08-storage-optimization-strategy.md)
   - Redpanda 保留期：7 天 → 24 小时
   - TimescaleDB 保留期：永久 → 365 天 (原始数据)
   - **存储减少**：91% (43.4 TB → 3.8 TB)
   - **成本降低**：57% (云) / 54% (硬件)

3. **最终优化配置** (FINAL-OPTIMIZED-CONFIG.md)
   - **Redpanda**: 3 节点 × (2C4G + 50GB SSD)
   - **Connect**: 2 实例 × (2C2G)
   - **TimescaleDB**: 1 主节点 × (4C8G + 500GB SSD)
   - **总成本**: $903/月 (云) / ¥6,520/月 (硬件)

### 🔑 关键技术决策

1. **二进制解析方式** ✅ **Go 插件 + Kaitai Struct**
   - 原因：NB67 格式复杂 (632 字节，200+ 字段)，Bloblang 不适合
   - 优势：类型安全、高性能、单一 schema 来源

2. **故障预测逻辑位置** ✅ **分层实现**
   - 简单阈值 → Bloblang (Connect 内)
   - 多窗口聚合 → TimescaleDB 连续聚合
   - 复杂规则 → PostgreSQL 存储过程

3. **有状态操作** ✅ **TimescaleDB 查询**
   - 避免引入 Redis
   - 利用现有 hypertable 索引
   - 运维更简单

## 迁移计划

### 时间线 (14 周 / ~3.5 个月)

| 阶段 | 周数 | 关键交付 |
|------|------|---------|
| **阶段 1** | 2 周 | Go 插件开发 (NB67 解析器) |
| **阶段 2** | 3 周 | 核心管道实现 (Parse, Store) |
| **阶段 3** | 2 周 | 预测逻辑迁移 (26 种故障类型) |
| **阶段 4** | 1 周 | 外部集成 (HTTP APIs) |
| **阶段 5** | 4 周 | 并行测试 (影子模式) |
| **阶段 6** | 2 周 | 逐步切换 + Faust 下线 |

### 风险缓解

- ✅ **二进制解析**：广泛单元测试
- ✅ **预测准确性**：Faust/Connect 并排比较 1 个月
- ✅ **性能验证**：2x 吞吐量负载测试 (1,000 msg/sec)
- ✅ **外部 API**：重试逻辑 + 死信队列
- ✅ **数据库连接**：连接池 + 熔断器

## 快速开始

### 阅读研究文档

1. **入门**：从 [docs/README.md](docs/README.md) 开始
2. **架构理解**：阅读 [05-simplified-go-plugin-architecture.md](docs/05-simplified-go-plugin-architecture.md)
3. **Faust 对比**：阅读 [09-faust-implementation-analysis-cn.md](docs/09-faust-implementation-analysis-cn.md)
4. **快速参考**：查看 [FINAL-OPTIMIZED-CONFIG.md](docs/FINAL-OPTIMIZED-CONFIG.md)

### 核心文档导航

- 🎯 **推荐方案**：[05-simplified-go-plugin-architecture.md](docs/05-simplified-go-plugin-architecture.md)
- 💰 **成本优化**：[08-storage-optimization-strategy.md](docs/08-storage-optimization-strategy.md)
- 📊 **容量规划**：[06-resource-estimation-capacity-planning.md](docs/06-resource-estimation-capacity-planning.md)
- 🔄 **Faust 分析**：[09-faust-implementation-analysis-cn.md](docs/09-faust-implementation-analysis-cn.md)

## 技术栈

### 当前系统
- **Stream Processing**: Faust (Python)
- **Message Broker**: Apache Kafka
- **Database**: TimescaleDB (PostgreSQL)
- **Binary Parser**: Kaitai Struct (Python runtime)

### 目标系统
- **Stream Processing**: Redpanda Connect (Go)
- **Message Broker**: Redpanda
- **Database**: TimescaleDB (PostgreSQL)
- **Binary Parser**: Kaitai Struct (Go plugin)
- **Monitoring**: Prometheus + Grafana

## 主要发现

### ✅ 可行性确认

Redpanda Connect **完全可行**用于此场景：
- ✅ 支持复杂二进制解析 (通过 Go 插件)
- ✅ 原生 TimescaleDB 集成 (`sql_insert` 输出)
- ✅ 强大的批处理能力 (`batch_policy`)
- ✅ 灵活的数据转换 (Bloblang DSL)
- ✅ 内置监控 (Prometheus 指标)

### 🎯 关键优势

1. **性能**: 9x 吞吐量提升，5x 延迟降低
2. **成本**: 50-70% 总成本降低
3. **可维护性**: YAML 配置 vs 3,000+ 行 Python
4. **可观测性**: 开箱即用的 Prometheus 指标
5. **可扩展性**: 无状态设计，易于水平扩展

### ⚠️ 挑战

1. **Go 插件开发**: 需要 1-2 周开发 NB67 解析器
2. **预测逻辑迁移**: 26 种规则需要转换为 Bloblang/SQL
3. **测试验证**: 需要全面测试以匹配 Faust 行为
4. **学习曲线**: 团队需要学习 Redpanda Connect YAML DSL

## 贡献者

- **研究与架构设计**: Google Deepmind Antigravity AI
- **原 Faust 实现**: [MACDA-NB67 项目](oldproj/MACDA-NB67)
- **业务场景**: 地铁列车 HVAC 系统监控

## 许可证

研究文档采用 MIT License。

原 Faust 实现 (oldproj/) 版权归原作者所有。

---

## 下一步行动

1. ✅ 团队审查研究成果
2. [ ] 原型 Go 插件 (NB67 解析器)
3. [ ] 基准测试 Redpanda Connect
4. [ ] 实现 POC 管道 (Parse 模式)
5. [ ] 定义验收标准
6. [ ] 启动阶段 1 迁移

---

*最后更新: 2026-02-17*
