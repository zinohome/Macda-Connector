# Redpanda 社区版可行性清单（本项目）

## 结论

在当前方案范围内（消息汇聚 + Connect 处理 + TimescaleDB 存储 + Grafana 展示 + Python 预测），
**仅使用 Redpanda 社区版即可满足需求**。

本结论基于现有文档需求与配置，不包含新增企业安全/云数据湖场景。

---

## 需求映射（是否需要企业版）

| 项目需求 | 当前实现方式 | 社区版可满足 | 备注 |
|---|---|---|---|
| Kafka 兼容消息队列 | Redpanda topic/consumer group | ✅ | 当前吞吐目标远低于基础能力上限 |
| Schema 管理 | Schema Registry + Avro | ✅ | 文档使用 `schema_registry_decode` |
| 流处理与规则检测 | Redpanda Connect + Bloblang + Go 插件 | ✅ | 属于 Connect 常规能力 |
| 时序存储 | TimescaleDB `sql_insert` 批写入 | ✅ | 与 Redpanda 许可无耦合 |
| 数据保留与压缩 | topic retention + lz4 + TimescaleDB policy | ✅ | 当前未使用企业专属存储能力 |
| 采样降频/双轨 | Connect 采样与分流 | ✅ | 文档推荐路径 |
| 实时监控与告警 | Grafana + 规则检测 | ✅ | 不依赖企业特性 |
| 故障预测 | Python 离线/准实时建模 | ✅ | 不依赖企业特性 |

---

## 当前范围内“不要启用”的企业功能

为确保保持社区版路径，避免在集群中启用以下企业能力：

- Iceberg / Datalake 相关能力
- Tiered Storage / cloud_storage 相关能力
- Enterprise 安全能力（例如 RBAC、OIDC、GSSAPI、FIPS）
- 审计日志等企业合规能力

只要不启用上述能力，社区版路线可持续成立。

---

## 上线前检查清单（建议）

### 1) 许可证状态检查

```bash
rpk cluster license info
```

建议检查结果：
- `license_violation = false`
- 未出现企业功能违规提示

### 2) 配置合规检查

确认 topic 与存储策略使用的是文档中的社区版配置：

- `retention.ms=86400000`（24h，可按方案调整）
- `compression.type=lz4`
- TimescaleDB 原始数据保留 365 天，聚合长期保留

### 3) 流程能力验证

完成端到端压测与回归：

- 二进制解析（Go 插件或转换服务）
- Connect 映射/采样/分流
- TimescaleDB 批写入与聚合查询
- 告警链路与可视化面板

---

## 何时需要升级到企业版

出现以下任一目标时，再评估企业版即可：

- 需要 Iceberg / 数据湖流式写入
- 需要云分层存储、跨云容灾等高级能力
- 需要统一企业级身份与权限控制（RBAC/OIDC/GSSAPI/FIPS）
- 需要审计合规能力与对应支持服务

---

## 适用范围声明

本清单适用于当前仓库中的地铁空调数据处理方案与文档版本（2026-02）。
若后续新增安全合规或数据湖目标，请重新评审本清单。
