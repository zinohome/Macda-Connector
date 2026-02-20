# Connect 重实现：存储与展示完整规划（V1）

日期：2026-02-20  
范围：`connect/` 主程序重实现（非 oldproj 优化）

---

## 1. 目标与边界

### 1.1 目标

在 `connect/` 内重建一条可演进的数据主链路：

1. NB67 二进制解析（Go + Kaitai）
2. 统一时序入库（支持高吞吐与高效查询）
3. 预测与告警（可配置、可追溯）
4. Dashboard/API 的契约优先查询支持

### 1.2 非目标

- 不沿用 oldproj 的 `pro_* / dev_*` 双表复制策略。
- 不在前端页面层拼 SQL 或拼指标语义。
- 不在第一阶段做复杂机器学习模型上线（先规则+特征框架）。

---

## 2. 总体架构（重实现版）

## 2.1 逻辑分层

采用“单一事实层 + 服务层”的四层结构：

1. **L0 接入层（Connect Ingestion）**
   - 输入：`signal-in` 二进制 topic
   - 输出：`signal.parsed.v1`（标准化 JSON）

2. **L1 单一事实层（Raw Fact Hypertable）**
   - 一张主事实表（车厢时序明细）
   - 保留 `event_time` 与 `ingest_time`
   - 这是所有计算和回溯的唯一来源（SSOT）

3. **L2 特征/事件层（Feature & Event）**
   - 窗口特征（5m/15m/30m 等）
   - 预测事件（开始/恢复）
   - 告警事件（开始/恢复）

4. **L3 展示服务层（API Query Model）**
   - 面向页面的稳定查询视图/物化视图
   - 严格按契约输出，不暴露底层表细节

## 2.2 组件职责

- `connect-nb67`：解析、字段标准化、消息质量标记、写入解析 topic
- `connect-storage`：批量写入事实层（TimescaleDB）
- `connect-rule`：轻规则实时判定（可选）
- DB 连续聚合/定时任务：窗口特征、统计聚合
- API 服务：只读服务层视图

---

## 3. 单一事实层（为什么这样设计）

## 3.1 设计原则

1. **一份明细真相**：不再拆 `pro/dev` 两套明细表。
2. **双时间共存**：事件时间与接入时间同时保存。
3. **宽表 + 必要索引**：满足追溯与在线查询平衡。
4. **可演进字段**：协议扩展时，增加字段不破坏既有查询。

## 3.2 推荐主表（示意）

```sql
CREATE TABLE hvac_fact_raw (
  event_time      TIMESTAMPTZ NOT NULL,
  ingest_time     TIMESTAMPTZ NOT NULL DEFAULT now(),
  line_id         INT NOT NULL,
  train_id        INT NOT NULL,
  carriage_id     INT NOT NULL,
  device_id       TEXT NOT NULL,
  frame_no        BIGINT,
  parser_version  TEXT,
  quality_code    SMALLINT DEFAULT 0,

  -- 关键状态/测点（示例）
  wmode_u1        SMALLINT,
  wmode_u2        SMALLINT,
  f_cp_u11        DOUBLE PRECISION,
  f_cp_u12        DOUBLE PRECISION,
  f_cp_u21        DOUBLE PRECISION,
  f_cp_u22        DOUBLE PRECISION,
  suckp_u11       DOUBLE PRECISION,
  suckp_u12       DOUBLE PRECISION,
  suckp_u21       DOUBLE PRECISION,
  suckp_u22       DOUBLE PRECISION,
  highpress_u11   DOUBLE PRECISION,
  highpress_u12   DOUBLE PRECISION,
  highpress_u21   DOUBLE PRECISION,
  highpress_u22   DOUBLE PRECISION,
  fas_u1          DOUBLE PRECISION,
  fas_u2          DOUBLE PRECISION,
  ras_u1          DOUBLE PRECISION,
  ras_u2          DOUBLE PRECISION,
  presdiff_u1     DOUBLE PRECISION,
  presdiff_u2     DOUBLE PRECISION,

  -- 故障位（示例）
  bocflt_ef_u11   BOOLEAN,
  bocflt_ef_u12   BOOLEAN,
  blpflt_comp_u11 BOOLEAN,
  bscflt_comp_u11 BOOLEAN,
  bflt_tempover   BOOLEAN,

  payload_json    JSONB
);

SELECT create_hypertable('hvac_fact_raw', 'event_time', if_not_exists => TRUE);
CREATE INDEX IF NOT EXISTS idx_hvac_raw_device_event ON hvac_fact_raw(device_id, event_time DESC);
CREATE INDEX IF NOT EXISTS idx_hvac_raw_ingest ON hvac_fact_raw(ingest_time DESC);
```

## 3.3 为什么不是多张 `pro/dev`

- 双表会导致查询口径漂移（同一指标两套时间轴）。
- 写入、索引、保留策略翻倍，维护成本高。
- 展示层最终仍要合并口径，复杂度转移而非消除。

---

## 4. 分层设计（你问的“分层”具体考虑）

## 4.1 L2 特征层（计算友好）

目标：把页面常用计算前置，避免每次扫明细。

示例：

- `hvac_feature_5m`：每设备 5 分钟窗口近似分位数、均值、极值
- `hvac_feature_15m`
- `hvac_feature_30m`

这样预测与统计都直接读 feature 层，不重复跑大 SQL。

## 4.2 L2 事件层（展示友好）

- `hvac_alarm_event`
  - `event_id`, `device_id`, `fault_code`, `start_time`, `end_time`, `status`
- `hvac_predict_event`
  - `predict_code`, `score_or_flag`, `start_time`, `end_time`, `status`

页面展示要的是“事件”，不是“某一分钟 max(flag)=1”。

## 4.3 L3 服务层（契约优先）

按照页面定义固定查询模型，例如：

- `vw_train_overview_rt`：列车在线、报警数、预警数
- `vw_carriage_health_rt`：每节车厢当前状态
- `vw_fault_stat_range`：时间范围内故障占比
- `vw_alarm_list_range`：告警/预警分页列表
- `vw_temperature_trend`：温度趋势曲线

API 只查这些视图，不直接拼明细 SQL。

---

## 5. 预测设计（你问的“预测怎么考虑”）

## 5.1 原则

1. **先规则后模型**：先把规则系统做成可配置，不再把阈值硬编码在代码里。
2. **窗口特征驱动**：预测读取 feature 层，不直接读 raw。
3. **事件化输出**：输出“预测开始/恢复”，而不是重复刷同一条告警。

## 5.2 规则配置表示例

```sql
CREATE TABLE predict_rule_config (
  rule_id        TEXT PRIMARY KEY,
  enabled        BOOLEAN NOT NULL DEFAULT TRUE,
  feature_window TEXT NOT NULL,   -- 5m/15m/30m
  expression     TEXT NOT NULL,   -- 例如 suckp_u11 < 2 AND f_cp_u11 > 30
  severity       SMALLINT,
  cooldown_sec   INT DEFAULT 300
);
```

## 5.3 执行流程

1. Connect 或任务调度读取最新 `feature_*`
2. 按 `predict_rule_config` 计算命中
3. 命中状态变化时写 `hvac_predict_event`
4. 同步更新 `vw_carriage_health_rt` 当前状态

---

## 6. 告警设计（你问的“告警怎么考虑”）

## 6.1 原则

- 告警来自两类：
  1. 设备原生故障位（硬故障）
  2. 规则衍生告警（软故障/预警）
- 都统一进入事件表，统一生命周期管理。

## 6.2 去抖与合并

- 持续 N 个采样点命中才开事件（防毛刺）
- 连续 M 个采样点恢复才关事件
- 相同 `device_id + fault_code` 的连续命中做事件延长，不重复开新事件

## 6.3 对展示的收益

- 报警列表可直接显示 `start_time/end_time/duration`
- 统计图直接按事件聚合，不再遍历全量布尔位数组

---

## 7. 时间规则（你问的“时间规则怎么考虑”）

## 7.1 三个时间字段

建议统一三字段：

1. `event_time`：设备上报时间（业务主时间）
2. `ingest_time`：Connect 接入时间（系统主时间）
3. `process_time`：规则/预测处理时间（可选）

## 7.2 使用规则

- 业务分析、趋势、报表：优先 `event_time`
- 监控延迟、排查堵塞：用 `ingest_time`
- 任务 SLA 评估：用 `process_time`

## 7.3 有效性判定

增加 `event_time_valid`，规则示例：

- `event_time` 为空或超当前时间 ±24h 视为无效
- 无效时查询可回退到 `ingest_time`

---

## 8. 只有一个时间测试数据时，如何保证能看到时序（开发模式）

你这个问题的结论：**开发模式默认用 `ingest_time` 作为可视时间轴，并保留 `event_time` 原值。**

## 8.1 开发模式策略

增加配置项：

- `time_mode = prod | dev`
- `dev_time_axis = ingest_time`

行为：

- `prod`：`query_time = event_time`
- `dev`：`query_time = CASE WHEN event_time_valid THEN event_time ELSE ingest_time END`

## 8.2 单时间样本可视化增强

如果测试数据 `event_time` 全相同：

1. 原始 `event_time` 不改写（保真）
2. 展示层使用 `query_time`（dev 用 ingest）
3. 可选启用“回放偏移”模式：按消息序号给 `query_time` 叠加毫秒偏移，保证曲线可展开

示例：

```sql
query_time = ingest_time + (seq_no * interval '200 milliseconds')
```

这仅用于开发演示，不写回事实层。

---

## 9. Connect 中的重实现落地拆分

## 9.1 子模块

1. `connect-nb67-parser`（已有基础）
   - 输入 `signal-in`
   - 输出 `signal.parsed.v1`

2. `connect-storage-writer`
   - 输入 `signal.parsed.v1`
   - 批量写 `hvac_fact_raw`

3. `connect-feature-builder`
   - 输入 `signal.parsed.v1` 或直接查 DB
   - 产出 `feature` 表或 topic

4. `connect-event-builder`
   - 读取 feature + rule config
   - 写 `hvac_alarm_event` / `hvac_predict_event`

## 9.2 Topic 规划

- `signal-in`
- `signal.parsed.v1`
- `signal.feature.5m.v1`（可选）
- `signal.event.alarm.v1`
- `signal.event.predict.v1`

---

## 10. 与 Dashboard 的契约对齐（存储反推）

按页面拆 API 契约：

1. 首页总览
   - 输入：时间范围 + 线路
   - 输出：在线列车、告警数、预警数、健康数

2. 车厢状态
   - 输入：列车 ID
   - 输出：每车厢当前 alarm_count / warning_count / 最新状态时间

3. 报警列表
   - 输入：列车/车厢 + 时间范围 + 类型 + 分页
   - 输出：事件级记录（start/end/level/advice/signal snapshot）

4. 故障统计图
   - 输入：时间范围
   - 输出：故障码分布占比（按事件或按命中次数）

5. 温度趋势
   - 输入：设备 + 时间范围 + 采样粒度
   - 输出：序列点（query_time, value）

---

## 11. 分阶段实施计划

## 阶段 A（1~2 周）

- 建立单一事实层 `hvac_fact_raw`
- Connect 完成解析 + 入库
- Dev 时间模式（ingest_time 轴）打通

验收：

- 每秒写入稳定
- Dashboard 能看到连续时序曲线（即便 event_time 恒定）

## 阶段 B（1~2 周）

- 建 `feature_5m/15m/30m`
- 建 `alarm_event/predict_event`
- 完成告警去抖与事件生命周期

验收：

- 告警列表展示 start/end
- 故障统计不再扫 raw 宽表

## 阶段 C（1 周）

- API 全部切到服务层视图
- 指标监控（延迟、写入成功率、查询耗时）

验收：

- 页面查询 p95 达标
- 规则可配置上线

---

## 12. 风险与控制

1. **字段爆炸导致写入慢**
   - 控制：热字段列化 + 冷字段 JSONB

2. **规则误报多**
   - 控制：去抖 + 冷却 + 规则灰度

3. **时间混乱**
   - 控制：强制双时间 + 明确 query_time 规则

4. **开发数据单时间无法看趋势**
   - 控制：dev 使用 ingest_time 轴 + 回放偏移

---

## 13. 最终建议（明确回答）

1. 你现在应该在 `connect/` 重建，不做 oldproj 优化。  
2. 单一事实层是必须的，分层是为了把“计算复杂度”从页面和 API 移到稳定数据层。  
3. 开发只有一个时间字段时，**开发模式用 ingest_time 做展示时间轴**，生产模式回到 event_time。  
4. 预测与告警都要事件化，否则展示与统计会一直不稳定。

---

## 14. 已交付执行文件（V1）

1. TimescaleDB 初始化与服务层视图：
   - [docs/requirements/sql/connect-timescaledb-ddl-v1.sql](docs/requirements/sql/connect-timescaledb-ddl-v1.sql)

2. Connect 配置拆分与拓扑：
   - [docs/requirements/connect-config-split-and-topology-v1.md](docs/requirements/connect-config-split-and-topology-v1.md)

3. 字段契约清单（解析/存储/API 对齐）：
   - [docs/requirements/connect-field-contract-v1.md](docs/requirements/connect-field-contract-v1.md)

4. 一键联调顺序（parser/storage/event）：
   - [docs/requirements/connect-one-click-run-sequence-v1.md](docs/requirements/connect-one-click-run-sequence-v1.md)

5. Storage Adapter 全字段映射（binary → intermediate → table）：
   - [docs/requirements/connect-storage-adapter-field-mapping-v1.md](docs/requirements/connect-storage-adapter-field-mapping-v1.md)

6. Storage Adapter 运行说明：
   - [connect/cmd/storage-adapter/README.md](connect/cmd/storage-adapter/README.md)
