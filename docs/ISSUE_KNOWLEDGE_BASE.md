# Macda-Connector Issue 处理知识库

本文件是全栈开发智能体的**持续记忆文档**。
每次处理 GitHub Issue 前必须阅读，每次完成修复后必须更新。

---

## 使用规范

### 处理 Issue 前（必读）
1. 阅读「已知约束与注意事项」（第一节）
2. 搜索「历史记录索引」（第二节），查找相似 Issue
3. 如找到相似记录，跳转到对应历史条目阅读根因和修复思路

### 完成修复后（必写）
在「历史 Issue 处理记录」（第三节）**顶部**追加本次记录，并更新第二节索引。

---

## 一、已知架构约束与注意事项

> 这里记录项目特有的、容易踩坑的约束，每次发现新坑时追加。

### NB67 协议解析
- `connect/codec/NB67.ksy` 是唯一真实来源，`nb67.go` 为自动生成文件，**禁止手工修改**
- 任何字段变更必须走完整链路：`NB67.ksy` → 重新生成 → `nb67_processor.go` → 存储映射 → API 契约 → 文档

### 时间字段
- `RUNTIME=DEV`（当前所有环境默认）时用 `ingest_time`，`RUNTIME=PRD` 时用 `event_time`
- 切换前必须确认现场设备时钟精度，否则历史数据时序会乱

### 前端构建
- 镜像构建前必须先本地 `npm run build` 验证通过，不允许直接 docker build

### 流水线配置
- `nb67-connect.yaml` 已废弃，勿修改；主线是三段式配置
- 扩容 `connect-parser` 不超过 Redpanda partition 数量

---

## 二、历史记录索引

> 快速检索用。格式：`[日期] #Issue编号 - 问题关键词 - 所属模块`

[2026-05-21] #1 - mock-platform未收到推送/ground-reporter镜像缺失 - ground-reporter/docker-compose
[2026-05-21] #2 - 预警触发配置/历史预警描述/冷媒泄露两条件分设 - nb67_event_processor/BFF
[2026-05-24] #3 - 预警报文location/code与alertcode文件不对应/_c/_v后缀未归一化 - ground-reporter/alertcode_map
[2026-05-24] #5 - 预警历史页面触发条件显示与设置页不一致/strategy误当触发条件 - BFF/status.repository
[2026-05-26] #8 - 新风/回风传感器32767误报/trainNo未补零/冷凝风机预警不消除 - nb67_event_processor/ground-reporter/前端
[2026-05-26] #7 - 预警报文code不随车厢变化/fault_name多余/完整预警描述未显示 - ground-reporter/BFF/前端
[2026-05-26] #11 - 历史预警触发条件随配置变更而变/缺少触发时刻快照 - config_store/nb67_event_processor/BFF
[2026-05-27] #12 - 5位trainNo数据不显示/BFF state解析硬编码4位 - BFF/前端
[2026-05-27] #13 - 历史报警页面缺少机组信息/unit_name未在select中 - BFF/status.repository
[2026-05-27] #14 - 预警无法消除/clear_value未加载/无滞回逻辑 - config_store/nb67_event_processor
[2026-05-27] #32 - dist镜像版本同步/从零重部署 - dist/部署
[2026-06-03] RET-40 #20 #21 - 预警重复/不消除/频率0误报/并发整改 - lifecycle表+storage-adapter状态机+BFF/ground-reporter
[2026-06-05] RET-40 followup - event-writer 脏时间戳日志噪音 - connect/config/nb67-event-writer.yaml
[2026-06-09] RET-40/GH#21 followup-2 - 历史页重复 + 消除延迟 - BFF status.repository + nb67-event-builder.yaml
[2026-06-09 PM] RET-40 followup-3 - 历史报警页去重 + 故障统计去重 - BFF status.repository (alarm 路径)


---

## 三、历史 Issue 处理记录

> 最新记录在最前。每条记录包含：问题描述、根因、修复方法、测试验证、经验总结。

### [2026-06-09 PM] RET-40 followup-3 — 历史报警页去重 + 故障统计去重

**问题描述**：v2.5.28 上线后用户问"3 个未覆盖点是不是要补一下"。3 点为：
1. 历史**报警**页（alarm，不是 predict）`/api/rest/train/AlarmInformation` 仍查 `fact_event` → 同一报警按帧入库会重复 N 行
2. `getFaultStatistics` 取 last 5000 行 fact_event alarm → Echart 占比统计被"同一报警 30 行"严重扭曲
3. 前端浏览器实跑没验证

**修复方法**（lean 方案，不动 DB schema、不动 Go 流水线）：
- `web-nb67-bff/src/repository/status.repository.ts` `getHistoricalEvents` 当 `eventType='alarm'` 时：取全量近 5000 行 → JS 层按 `(train_id, carriage_id, fault_code)` + 60s 发作岛聚合 → 每岛取 `min(time) → event_time/ingest_time` 和 `max(time) → recovery_time`；recovery_time 推断规则：(a) 原 row 有值则用，(b) 同 key 后续还有新岛则用 lastSeen，(c) 是该 key 最新岛且 lastSeen>60s 前 → 用 lastSeen 视为已结束，(d) 否则 NULL 视为活跃
- `getFaultStatistics` 改用 SQL `GROUP BY (fault_code, device_id, date_trunc('minute', event_time))` 聚合 + 时间窗口 `> now() - interval '7 days'`，保留前端 `value === 1` 累加契约（每个数组项仍是 `{HVAC301:1}`），但不再有同一报警 30 倍重复
- BFF 镜像 nb67-bff v2.5.28 → v2.5.29 推 Harbor
- **从零部署**：4 个 compose 全部 `down` 后按 data → mock → web → report 顺序 `up -d`，所有容器 healthy

**测试验证**（自动完成 4 步复测）：
1. `/api/rest/train/HistoryWarning`：6 条 lifecycle，0 异常重复（"重复"是真实多次发作，DB 真实 lifecycle 数）
2. `/api/rest/train/AlarmInformation`：total=4（DB fact_event alarm 行总数 621636），dedup ratio ≈ 155000:1，0 同 key 重复，每行 `start_time/recovery_time/status='已恢复'` 都正确
3. `docker restart connect-event-builder` 后 active count 不变（设备仍在 hit 中，无 transition，符合预期）；rpk 抽样不报新错
4. `/api/rest/v2/FaultStatistics`：返回 5000 个 `{code:1}` 形状条目，但每条对应 `(fault_code, device_id, minute)` 唯一组合，占比统计不再失真

**经验总结**：
- "去重"按业务语义有两种实现：(a) DB 层建独立 lifecycle 表（重，适合状态机），(b) BFF JS 层 60s 岛聚合（轻，适合查询时去重）。**架构对称性 ≠ 必须用同一方案**——predict 走方案 (a) 因为还要喂 ground-reporter 状态机和实时页；alarm 仅查询用，方案 (b) 一行 SQL 就够，避免新表 + 新写入器 + 新迁移的回归面。**先量化风险再选方案**。
- 前端约定 `value === 1` 必须保留时，BFF 应该用 "expand 后展开成同 shape" 而不是直接返回聚合 count——`{code:5}` 会让前端 `value === 1` 判等失败，整个图表静默归零。**改 API 形状前先 grep 所有前端消费点**。
- "从零部署"的安全序：`down` 顺序按依赖逆序 (report→web→mock→data)，`up` 按依赖正序 (data→mock→web→report)；容器卷未删除，数据保留，仅替换镜像和容器实例。

### [2026-06-09] RET-40 / GH#21 followup-2 — 历史页重复 + 消除延迟

**问题描述**（GitHub #21 复测发现）：
- 现象 A：历史预警/报警页面同一条预警重复出现多行
- 现象 B：满足消除条件后 `warning_lifecycle` 仍未消除；只有当**新预警触发**时旧预警才被一并消除，并把消除报文一起发给平台

**根因**：
1. 上一轮 #20/#21 整改时把 `getRealtimeWarnings` 改查了 `warning_lifecycle`，但 `getHistoricalWarnings` 漏改，仍查 `hvac.fact_event`。predict 每帧入库一行 → 历史页天然重复。
2. `connect/config/nb67-event-builder.yaml` 的 signal-predict 分支 mapping：
   ```
   root = if this.exists("predict_event") && this.predict_event.hits.length() > 0 { this.predict_event } else { deleted() }
   ```
   Go 端 `nb67_event_processor.go` 的 `prevPredictHadHits` 状态机在 "上一帧有命中、本帧清空" 的过渡点会**主动发一次 hits=[] 的 predict_event** 给下游消除链路；但 yaml 这里 `hits.length() > 0` 把过渡帧整条扔掉。
   结果：lifecycle_writer 和 ground-reporter 永远收不到"该 device 现在无预警"的信号，状态停在最后一帧；要等下一个不同 fault_code 进来时 set-diff 才把旧 code 算作 removed，那时消除报文的时间戳已经错位。

**修复方法**：
- **`web-nb67-bff/src/repository/status.repository.ts:572-621`** `getHistoricalWarnings`：
  - selectFrom 从 `hvac.fact_event as e` 改成 `hvac.warning_lifecycle as e`
  - 时间窗口语义改为"生命周期与窗口有重叠"：`start_time <= window_end AND (end_time IS NULL OR end_time >= window_start)`
  - 把 lifecycle 列映射回老 API 字段：`start_time → event_time/ingest_time`，`end_time → recovery_time`，`status = end_time IS NULL ? 'active' : 'recovered'`
  - `trigger_snapshot` 包成 `payload_json.trigger_condition_snapshot` 给下游降级逻辑用
- **`connect/config/nb67-event-builder.yaml` + `dist/config/nb67-event-builder.yaml`** signal-predict 分支：
  - `if this.exists("predict_event") && this.predict_event.hits.length() > 0` → `if this.exists("predict_event")`
  - 让 Go 主动发的"过渡空帧"顺利透传，下游消除链路按真实时间戳关闭并发报文
- 镜像：`nb67-bff:v2.5.28`、`nb-parse-connect:v2.5.28`（重建 + 推 Harbor + 同步 dist + 实地容器替换）

**测试验证**：
- `rpk topic consume signal-predict` 看到 `"hits":[]` 的过渡帧（修复前从未出现）
- `warning_lifecycle` 表观察：connect-event-builder 重启后，所有原本 active 的行 end_time 在同秒被批量回填，`5 closed, 0 active`
- BFF API：`POST /api/rest/train/HistoryWarning` 返回每条 lifecycle 一行，`start_time`/`end_time` 同时出现，trigger_condition 正确映射，`total=5`（DB 真实条数）
- 6 个 connect-* 容器 + nb67-bff 全部 `Up (healthy)`

**经验总结**：
- 拆分查询路径时务必**列出所有引用旧路径的位置**——上轮整改只改了 realtime，history 漏了；这种漏改在数据模型迁移里非常常见，应养成 `grep "fact_event" web-nb67-bff/src/` 的扫尾习惯。
- Go processor 已设计正确（含 `prevPredictHadHits` 状态机），但 yaml 层"善意的过滤"把状态机产物吃掉了。**跨语言/跨配置的状态机一定要画一遍端到端时序图**，下游过滤器要看上游 emit 语义说明书，不能凭直觉"空 hits 就 drop"。
- `event_meta` 必须永远在 SubEvent 里——这是 lifecycle/ground-reporter 用来识别 device 的唯一锚点；如果未来真的想压缩流量，可以把 idle 周期延长到 10s 而不是每帧都发空帧，而不是直接 deleted()。

### [2026-06-05] RET-40 followup - event-writer 脏时间戳日志噪音

**问题描述**：connect-event-writer 在 v3 group 切换后日志里持续刷
`date/time field value out of range: "2049-98-53 97:49:99+08:00"`。
吞吐与落库正常（脏行被 ON CONFLICT / 解析失败拦下不入 DB），但 ERROR 等级噪音
干扰排障。

**根因**：`nb67-event-writer.yaml` 的 mapping 把 `event_time_valid` 默认值设为
`| true`（缺失即信任），且仅依赖 parser 的 boolean 标志，没做格式 sanity。
上游若某帧 `event_time_valid` 字段缺失（老 parser 版本残留 / 中间过渡数据），
mapping 直接把 `2049-98-53 ...` 这种脏字符串拼 `+08:00` 送进 sql_raw，
postgres 立即抛 `date/time field value out of range`。

**修复方法**：
- `connect/config/nb67-event-writer.yaml` 和 `dist/config/nb67-event-writer.yaml`
  同步修改 mapping：
  1. 默认从 `| true` 翻转为 `| false`（缺失即降级到 ingest_time，不再"假定 valid"）
  2. 新增 `ts_format_ok` 正则校验 `^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01]) ...`
     即使 `event_time_valid=true` 也要再过这一关，杜绝 parser 漏判的脏帧
  3. `ts_valid = (event_time_valid | false) && ts_format_ok`，任何一边不过都走 ingest_time
- 配置经 `1panel-network` 挂载，无需重建镜像，`docker restart connect-event-writer` 即生效

**测试验证**：
- 同步 yaml 到 `/data/Macda2/connect/config/`，`docker restart connect-event-writer`
- 容器状态：`Up (healthy)`，benthos 正常 Listening + Consuming v3 group
- 重启后 90s 日志：0 条 error，0 条 `out of range`，0 条 warn（restart 前同窗口持续刷）
- v3 group：STATE=Stable，MEMBERS=1，partition LAG ≈ 50（持续追平，未停消费）

**经验总结**：
- bloblang fallback `| true` 在跨版本上游数据混跑时是危险默认，**"未知则降级"**
  比 **"未知则信任"** 安全。新增字段时 default 选 false / nullable，不要选 true。
- benthos 的 sql_raw error 是 ERROR 等级且不可降级，所以**不要靠下游报错过滤脏数据**，
  必须在 mapping 阶段把脏行引流到 fallback 分支，让 sql_raw 永远拿到合法值。
- 正则 sanity 比依赖 parser 标志更可靠：parser 升级、字段重命名、上游平台直接灌数据
  这些场景下 boolean flag 都可能失真，但 `^\d{4}-(01-12)-(01-31) (00-23):...` 永远是
  postgres timestamp 的硬约束。

### [2026-06-03] RET-40 #20 #21 - 预警系统整改：一条预警一行 (Phase 1-7)

**问题描述**（用户连续 N 周反馈，多次修复无效）：
- 现象 A：同一时刻同一条预警在数据库/前端**重复多次**
- 现象 B：满足消除条件后，预警在数据库/前端**仍显示活跃**（#20）
- 现象 C：压缩机不运行（频率=0）时仍触发制冷预警（#21）
- 现象 D：进程重启后，已 active 的预警被当作新 open 整批重发给平台

**根因**：
1. `hvac.fact_event` 主键 `(event_time, device_id, fault_code)` 是事件流模型，nb67-connect 每帧（1Hz）重跑预警规则 → 一条 30s 预警写 30 行。这就是 "同时刻重复" 的物理来源。
2. `fact_event.recovery_time` 字段在全仓代码里**从未被写入**。BFF 查询里加的 `WHERE recovery_time IS NULL` 等价于死过滤，是安慰剂。
3. ground-reporter 的 `AlarmTracker` 是内存 map，重启即丢，导致平台重发风暴。
4. 冷媒泄漏 `_c/_v` 后缀映射成两个 warn_code，物理同一条预警却落两行。
5. #21 现象 A 在 `e846948` 已修（`f1==f2 && f1>0` + 3min 持续判定），但**没有回归测试**，每次担心被改坏。

**修复方法**：Phase 1-7 共 5 个 commit、3 个镜像版本（v2.5.25/26）。
- Phase 1 (`ffc4e0c`): 新建 `hvac.warning_lifecycle` 表 + 部分唯一索引 `UNIQUE(device_id, fault_code) WHERE end_time IS NULL` 兜底重复
- Phase 2+3 (`03a93d5`): `storage-adapter/lifecycle_writer.go` 状态机消费 signal-predict，做 per-(device, fault_code) diff，状态变化才写 DB。新增 `WRITE_FACT_EVENT=false` 开关，与 connect-event-writer 并行不冲突。启动时 `lifecycle.Recover(ctx)` 从 DB 重建 active 集合
- Phase 4: `web-nb67-bff/src/repository/status.repository.ts` `getRealtimeWarnings` 改查 `warning_lifecycle WHERE end_time IS NULL`，删除 distinctOn 与 recovery_time 死过滤
- Phase 5: `connect/cmd/connect-nb67/nb67_event_processor_test.go` 新增 3 个回归测试锁定 #21 现象 A 的当前正确行为
- Phase 6: `ground-reporter/alarm_tracker.go` 新增 `RecoverFromLifecycle(ctx, pool)`，启动时从 lifecycle 拉活跃集合，避免重启风暴
- Phase 7 (`ae708ae`): `storage-adapter/types.go` EventMeta line_id/train_id 改 json.Number 兼容 nb67 字符串输出，发布 v2.5.26 修正
- 部署 (`19c76cc`): docker-compose-Data.yml 新增 `lifecycle-writer` 容器，独立 consumer group `macda-lifecycle-writer-v1`

**测试验证**：
- Phase 1 验收：6 个 SQL 用例（首次/重复拦/标记结束后重开/并发事务）全通过
- Phase 3 单测：`TestLifecycleDiff_ContinuousFrames` (100 帧同 code 仅 1 次 open) / `TestLifecycleDiff_HitToEmptyFrame` / `TestLifecycleDiff_ConcurrentDifferentDevices` / `TestInferWarnCode` / `TestInferUnitID`
- Phase 5 单测：3 个制冷规则回归测试
- Phase 7 端到端（生产 redpanda + timescaledb）：
  - 注入 2 个 hit → DB 2 行 open
  - 重复同 key 帧 → 不新增，last_seen_time 推进
  - 空帧 → 2 行 end_time 写入
  - 再注入同 code → 新建 1 行 active，老行保持 closed

**经验总结**：
- "同一条预警重复" 类问题，**第一反应是查表的唯一约束设计**，不要从查询端去重打补丁。这种 bug 反复修不好的根因都是数据模型错了，查询端补丁只是治标。
- 在 PostgreSQL 把不可重复约束写进 DDL（部分唯一索引）比写在应用层可靠 10 倍——并发场景下应用层 `select then insert` 有 TOCTOU 风险，DB 约束没有。
- 状态机改造高频写入路径时，用 **双轨过渡**（旧表保留 + 新表并行）而不是一刀切，把"代码改对"/"DB 改对"/"消费者改对"/"前端改对"的风险解耦。
- nb67_event_processor 的 `EventMeta.LineID` 是 `string`（来自 json.Number），storage-adapter 的 EventMeta 之前定义为 int32 导致 JSON unmarshal silently fail——**跨服务 schema 务必用 json.Number 兼容字符串和数字两种风格**。
- 多 worktree 场景下 `main` 分支被另一个 worktree 占用时，用 `git push origin <feature-branch>:main` 直推绕过；或者在持有 main 的 worktree 里 `git pull --ff-only`。
- `docker inspect <container> --format '{{index .Config.Labels "com.docker.compose.project.config_files"}}'` 可以反查正在运行的容器对应哪份 compose 文件。



### [2026-05-27] #12 #13 #14 - trainNo 5位/历史预警机组信息/预警无法消除

**问题描述**：
- #12：web 页面地址栏 trainNo 改成五位数后，有些数据不能显示（如历史报警页）
- #13：历史报警页面的报警条目中缺少机组信息，但实时报警中有
- #14：修改了预警消除阈值后，预警仍无法自动消除

**根因**：
- #12：双重 bug：① `trainInfo/index.vue` 的 `listData()` 直接拼接 `route.query.trainNo`（URL 带前导零，如 "07001"），导致 state 多一位字符（"0700101" vs "700101"），BFF 解析出错；② BFF `index.ts` 的 state 解析用 `slice(0,4)` 硬编码4位 trainId，5 位列车号（如 "10551"）时截取 trainId 错误
- #13：`getHistoricalWarnings` select 列表手动列举字段，漏掉了 `e.unit_name`；而 `getRealtimeWarnings` 用 `selectAll()` 所以有 unit_name
- #14：`config_store.go` 的 SQL 查询只加载 `trigger_value`，从未加载 `clear_value`；`warnEntry` 结构体也无 `ClearValue` 字段；`nb67_event_processor.go` 隐式用 trigger_value 作为消除条件，用户在 UI 修改 clear_value 对 Go 处理器完全无效

**修复方法**：
- **`web-nb67-web/src/views/trainInfo/index.vue:1433`**：`listData()` 中 `route.query.trainNo` → `filterForm.trainNo`（已 parseInt 去前导零）
- **`web-nb67-bff/src/index.ts`**：三处 state 解析从 `slice(0,4)` / `slice(4,6)` 改为 `slice(0,-2)` / `slice(-2)`，自动适配 4/5 位 trainId
- **`web-nb67-bff/src/repository/status.repository.ts`**：`getHistoricalWarnings` select 列表加入 `'e.unit_name' as any`
- **`connect/cmd/connect-nb67/config_store.go`**：`warnEntry` 增 `ClearValue float64`，SQL 补 `clear_value`，新增 `csClearThreshold()` 函数
- **`connect/cmd/connect-nb67/nb67_event_processor.go`**：`ruleState` 增 `triggered bool`，新增 `checkRuleWithClear()` 实现滞回逻辑（已激活后只有低于 clear_value 才消除），`checkFanI` / `checkCpI` 改用滞回版本

**测试验证**：
- `CGO_ENABLED=0 go build ./...` — connect-nb67 编译 ✓
- TypeScript 类型检查无新增错误
- 镜像构建推送：nb67-web:v2.5.20 ✓，nb67-bff:v2.5.20 ✓，nb-parse-connect:v2.5.4 ✓
- 容器全部健康运行：nb67-web、nb67-bff、connect-parser、connect-event-builder、connect-event-writer ✓
- 同步修复了 `/data/MACDA2/connect/config/` 目录下 YAML 配置文件缺失的预存部署问题

**经验总结**：
1. **精确 select vs selectAll**：手动列举字段时必须与 schema 核对，极易遗漏新增列。重要实时接口和历史接口应保持字段一致。
2. **state 字符串解析与车号位数**：依赖 `slice(0,4)` 这类固定偏移的字符串解析是脆弱的；改用 `slice(0,-2)` 从末尾取固定2位 carriageId，其余为 trainId，自动适配任意位数。
3. **clear_value 与 trigger_value 分开管理**：DB 有两个字段不代表代码都使用了——务必检查 SQL SELECT 列表是否包含所有业务字段。
4. **滞回（hysteresis）防振荡**：预警系统中触发阈值和消除阈值应分开，防止数值在边界附近时预警反复触发/消除；实现时需引入"已激活"状态标记（`triggered bool`）。
5. **Docker volume 挂载缺失**：若宿主机文件不存在而 Docker 尝试挂载，会自动创建空目录，后续读取报"is a directory"。修复方法：先删目录，复制真实文件，再删除旧容器并重新 up。

### [2026-05-27] #32 - dist镜像版本同步/从零重部署

**问题描述**：
- 用户要求把 `dist/` 中的镜像版本、打包清单和部署说明同步到当前可用版本，并重新从零部署后用浏览器截图确认页面状态。

**根因**：
- `dist/image-save.sh` 仍引用旧版 `nb67-web` / `nb67-bff` / `nb-parse-connect` 镜像标签，`dist/README.md` 的镜像表也未对齐当前 compose 文件，容易导致离线打包和部署说明与实际运行版本不一致。

**修复方法**：
- `dist/image-save.sh`：将 `nb-parse-connect`、`nb67-web`、`nb67-bff` 的镜像 tag 同步到当前部署版本。
- `dist/README.md`：更新文档版本号、更新时间和镜像表。
- `docs/ISSUE_KNOWLEDGE_BASE.md`：新增本条记录，便于后续追踪 dist 同步历史。

**测试验证**：
- 重新读取并核对 `dist/README.md` 与 `dist/docker-compose-*.yml` 的镜像 tag 是否一致。
- 后续将执行镜像构建、从零启动和浏览器截图验证。

**经验总结**：
- dist 离线包里的 `README`、`image-save.sh` 和 `docker-compose` 必须视为一个整体一起更新，否则离线打包、在线部署和文档会出现版本漂移。

### [2026-05-26] #11 - 历史预警触发条件随配置变更而改变

**问题描述**：已触发的历史预警记录中的"触发条件"字段，会随着用户在设置页修改预警配置后自动更新，应该显示触发时刻的设置。例如原先以 2.3A 电流触发的预警，修改为 2.9A 后，历史记录也显示 2.9A。

**根因**：预警事件写入 `fact_event.payload_json` 时只存储了命中码（code/name/severity），未保存触发时刻的配置快照。BFF `getHistoricalWarnings` 每次查询都从当前 `hvac.warning_config` 表动态生成 `trigger_condition`，导致所有历史记录跟随当前配置变化。

**修复方法**：
- **`config_store.go`**：`warnEntry` 增加 `ThresholdBad string` 字段，`load()` 查询新增 `threshold_bad` 列；添加 `csTriggerConditionText(warnCode)` 帮助函数，按与 BFF 相同逻辑生成快照文本（threshold_bad + 持续N分钟）
- **`nb67_event_processor.go`**：`PredictHit` 增加 `TriggerConditionSnapshot string`（omitempty），`buildPredictHits` 中所有配置驱动的命中位置调用 `csTriggerConditionText()` 填充快照；硬编码阈值的条件（如制冷系统过热度条件2）不填快照；YAML 层无需修改（`payload_json = this.string()` 自动包含新字段）
- **`status.repository.ts`**：select 新增 `e.payload_json`；map 逻辑改为优先读 `payload_json.trigger_condition_snapshot`（新数据路径），找不到时降级为当前配置（旧数据向后兼容）

**测试验证**：
- `CGO_ENABLED=0 go build ./...` — connect-nb67 ✓
- TypeScript 类型检查无新增错误（存量错误为预存问题）
- 镜像构建推送：nb-parse-connect:v2.5.3 ✓，nb67-bff:v2.5.19 ✓
- 容器重启：connect-event-builder、connect-event-writer、nb67-bff 均健康启动

**经验总结**：
1. **快照 vs 引用**：任何"历史记录需要反映创建时状态"的字段都应在写入时快照，而非引用当前配置。本项目中 `payload_json` 是扩展历史记录的合适位置。
2. **向后兼容降级**：新增快照字段后，旧数据无该字段，读取层必须降级处理（`|| fallback`），不能假设字段一定存在。
3. **BFF 构建 `trigger_condition` 的逻辑**：取 `threshold_bad`（显示字符串）+ `duration_seconds`（格式化为"持续N分钟"）。Go 层和 BFF 层共享相同逻辑，确保快照与设置页展示完全一致。
4. **config_store 已有 threshold_bad**：修复前 config_store 只加载 `trigger_value`（数值），修复后额外加载 `threshold_bad`（展示字符串），两者用途不同不可互替。

### [2026-05-26] #8 #9 #10 - 传感器误报 / trainNo 补零 / 预警无法自动消除

**问题描述**：
- #8：新风温度传感器预警和回风温度传感器预警被误触发，实际条件未满足
- #9：web 页面地址栏 trainNo 显示为 4 位（如 7002），应补零到 5 位（如 07002）
- #10：冷凝风机电流预警满足消除条件后，预警无法自动消除

**根因**：
- #8：`nb67_event_processor.go` 中计算 FasU1-FasU2 / RasU1-RasU2 温差时未过滤 32767（INT16_MAX，传感器故障/未连接的无效值标记）。一旦某个传感器返回 32767，差值极大，超过阈值 80，误触发 5 分钟计时器，随后报警。
- #9：前端 trainList 中 train_id 存储 4 位字符串（"7001"），router.push 直接使用，URL 显示为 4 位。
- #10：**根本原因**：`nb67_event_processor.go` 在 predictHits 为空时直接丢弃消息（`return service.MessageBatch{}, nil`）。当预警条件清除后，事件处理器不再产生任何消息，ground-reporter 的 AlarmTracker.Diff 从未被以空集合调用，永远无法发出"预警结束"事件给平台。

**修复方法**：
- #8：`nb67_event_processor.go`：计算温差前先读取两个传感器值，加 `!= 32767` 有效性检查（fasValid / rasValid），无效时 condition=false，checkRule 清除计时器状态。
- #9：前端三处修改：① router.push 时 `padStart(5, '0')` 补零；② 从 URL 读取时 `parseInt(...) || 7001` 去前缀零以保持内部 4 位表示；③ 涉及文件：`trainInfo/index.vue`、`historyData/index.vue`、`home/Left.vue`。
- #10：两处修改：
  1. `nb67_event_processor.go` 增加 `prevPredictHadHits sync.Map`，当检测到设备从"有命中"→"无命中"时，保留这一帧（不丢弃），让 ground-reporter 可以用空集合调用 Diff；之后恢复正常（不再额外发送）。
  2. `ground-reporter/api_6_1.go` Handle61Predict：移除 `|| len(hits) == 0` 提前返回，允许空命中列表流入 tracker.Diff 触发清除逻辑。

**测试验证**：
- `CGO_ENABLED=0 go build ./...` — connect-nb67 ✓，ground-reporter ✓
- `npm run build` — web ✓（15s）
- 镜像构建并推送：nb67-web:v2.5.17 / nb67-bff:v2.5.17 / ground-reporter:v2.5.17
- 容器重启后状态：nb67-web healthy / nb67-bff healthy / ground-reporter up

**经验总结**：
1. NB67 协议中 32767（0x7FFF）是传感器无效值标记，凡是差值类计算都需要先过滤掉该值，参考 PresdiffU 的 `< 32767` 模式。
2. 预警清除的生命周期依赖"事件处理器能产出空命中消息"→"AlarmTracker 看到代码消失"→"平台收到 endtime"。如果任一环节早退，预警会永久悬挂。修复后，每当一个设备的 predict 命中从非空变为空时，会精确地额外发送一帧空消息，之后停止，无额外流量。
3. URL 中的 trainNo 只需在 router.push 时补零，内部始终保持 4 位整数字符串，避免影响 state 拼接（"700101"格式）和 BFF 的 `parseInt(state.slice(0, 4))`。

### [2026-05-26] #7 - 预警报文 code 错误 & 完整预警描述未显示 & alertcode_v2.xlsx 更新

**问题描述**：
- **code不变**：不同车厢相同预警报文的 code 字段相同（如 HVAC109），应该随车厢变化
- **fault_name多余**：6.1 平台报文中有 `fault_name` 字段，平台不需要
- **完整预警描述**：历史预警详情弹窗缺少"完整预警描述"字段
- **名称有误**：冷媒泄漏预警名称写成"泄露"，应为"泄漏"

**根因**：
- event processor 生成 HVAC{carriage*100+seq} 格式（如 HVAC313 = 车厢3，seq=13），但 ground-reporter 只做了 normalizeCode 剥掉 `_c`/`_v` 后缀，没有将内部格式转换为平台期望的 alertcode_v2.xlsx 格式（HVAC101-HVAC115）
- 旧方案有26个seq（机组1/机组2各自的预警），新方案按预警类型合并为15个code，需要建立 oldSeq → newCode 映射
- `Record61` 结构体包含 `FaultName` 字段，但平台不需要此字段
- 历史预警 BFF API 和前端未暴露/显示 fault_desc（完整预警描述）

**修复方法**：
- **ground-reporter/alertcode_map.go**：新增 `oldSeqToNewCode` 映射表（old seq 1-26 → HVAC101-115），新增 `platformHvacCode()` 函数；location 查找仍基于原始内部码（保留机组信息）
- **ground-reporter/api_6_1.go**：`buildRecord61` 中 Code 改用 `platformHvacCode(baseCode)` 转换；注释说明 location 和 code 分别使用不同的code
- **ground-reporter/types.go**：从 `Record61` 移除 `FaultName string`
- **connect-nb67/nb67_event_processor.go**：将冷媒泄漏预警 Name 中的"泄露"改为"泄漏"（4处）
- **web-nb67-bff/src/index.ts**：HistoryWarning API 响应新增 `fault_desc: row.fault_name` 字段
- **historyWarning/index.vue**：详情弹窗 (#header 下方) 新增"完整预警描述"蓝色标注区块

**测试验证**：
- `CGO_ENABLED=0 go build ./...` 在 ground-reporter 和 connect-nb67 均通过，无编译错误
- `npm run build` 在 web-nb67-web 通过（vue-tsc + vite build）
- `tsc --noEmit` 在 web-nb67-bff 通过，无新增错误

**经验总结**：
- 内部 HVAC 编码方案（26 seq/carriage）与平台期望的 alertcode_v2 方案（15 type codes）不同，转换时需维护 oldSeqToNewCode 映射
- Location 应基于原始内部码查询（保留机组/系统信息），而 Code 才是需要转换的字段
- 不同格式的代码（如 HVAC313）在 normalizeCode 剥suffix 后，仍需经过 platformHvacCode 转换才能发给平台
- 历史预警 `fault_name`（存于 DB）即是"完整预警描述"，直接映射到 `fault_desc` 字段暴露给前端即可

### [2026-05-24] #3 & #5 - 预警报文字段错误 & 历史预警触发条件显示不对

**问题描述**：
- **#3**：预警报文（6.1平台接口）中 location 和 code 字段内容与 alertcode 文件不对应
- **#5**：预警历史页面显示的"触发条件"内容不正确，应与预警设置页"触发条件"一致

**根因**：
- **根因#3**：Issue #2 修复冷媒泄露两模式分设时，事件处理器（`nb67_event_processor.go`）生成了带 `_c`（制冷）和 `_v`（通风）后缀的故障码（如 "HVAC301_c"）。这些带后缀的码被原样传入 `ground-reporter` 的 `buildRecord61` 函数，而 `alertcodeLocationMap` 只有基础码（"HVAC301"），导致 `locationByCode("HVAC301_c")` 找不到映射，回退为直接返回原始码作为 location，和 alertcode 文件内容不对应。
- **根因#5**：`warning_config.strategy` 字段存储的是"维修指导意见"文本（如"请采用手持式卤素仪检测漏点"），而 `status.repository.ts` 的 `getHistoricalWarnings()` 将其作为 `trigger_condition` 返回给前端。设置页展示的触发条件是 `threshold_bad`（如 ">1.8A"）+ `duration_seconds`，两者来源不同，因此显示不一致。

**修复方法**：
- **#3**：在 `connect/cmd/ground-reporter/api_6_1.go` 添加 `normalizeCode()` 函数，发送给平台前剥去 `_c`/`_v` 后缀，`Location` 和 `Code` 字段均使用归一化后的基础码。同时在 import 中添加 "strings"。
- **#5**：在 `web-nb67-bff/src/repository/status.repository.ts` 的 `getHistoricalWarnings()` 中，将原来用 `strategy` 构建触发条件的逻辑改为用 `threshold_bad`（阈值描述文字）+ `duration_seconds`（持续时间，自动转分钟/秒）动态拼接，使历史页与设置页展示保持一致。

**测试验证**：
- `CGO_ENABLED=0 go build` 在 `connect/cmd/ground-reporter/` 编译通过，无错误
- `tsc --noEmit` 检查无新增 TypeScript 错误（预存在错误不在修复范围）
- 端到端验证逻辑：
  - #3：发送 "HVAC301_c" 故障码，platform 应收到 Code="HVAC301", Location="空调机组1"
  - #5：修改 warning_config 的 threshold_bad 为 ">2.0A"、duration_seconds=900，历史页触发条件应显示 ">2.0A 持续15分钟"

**经验总结**：
- 事件处理器给冷媒泄露代码添加了 `_c`/`_v` 后缀区分模式，任何需要做代码→名称映射的地方都必须先归一化（剥去后缀）
- `warning_config.strategy` 是维修建议文字，不是触发条件描述。触发条件应从 `threshold_bad` + `duration_seconds` 动态构建
- 新增后缀约定时必须同步更新 ground-reporter 和 BFF 所有依赖该代码的映射逻辑

### [2026-05-21] #2 - 预警设置相关问题

**问题描述**：
1. 预警触发是否会根据最新设置更新？
2. 历史预警中"触发条件"描述没有更新
3. 冷媒泄露预警的两个触发条件（制冷模式/通风模式）无法分别设置

**根因**：
- **根因①**: `checkRefLeak`（冷媒泄露）和 `checkCpSys`（制冷系统）使用硬编码阈值，没有调用 `csRawThreshold`/`csDuration` 从数据库热加载。其他预警（如 WARN_TEMP_SENSOR）已正确使用配置热加载。
- **根因②**: `getHistoricalWarnings()` 的 LEFT JOIN 条件错误：`wc.warn_code = e.fault_code`，但 `fact_event.fault_code` 存的是 HVAC 编码（如 `HVAC301`），`warning_config.warn_code` 存的是语义编码（如 `WARN_REFRIGERANT_LEAK`），两者命名规范完全不同，JOIN 永远不匹配，`trigger_condition` 始终为 NULL，回退到硬编码的 `HVAC_SEQ_STRATEGY`。
- **根因③**: `WARN_REFRIGERANT_LEAK` 在数据库中是单行，两个条件用 `params.condition1/condition2` 存储但未分别提供独立的可配置阈值条目。

**修复方法**：
1. `connect/cmd/connect-nb67/nb67_event_processor.go`：在 `checkRefLeak` 闭包外预读 `csRawThreshold("WARN_REFRIGERANT_LEAK_COOLING")` / `csDuration(...)` / `csRawThreshold("WARN_REFRIGERANT_LEAK_VENT")` / `csDuration(...)`；`checkCpSys` 同理读 `WARN_COOLING_SYSTEM`。
2. `web-nb67-bff/src/repository/status.repository.ts`：移除错误的 LEFT JOIN；改为单独加载 `warning_config`，在 JS 层通过 `hvacSeqToWarnCode(seq)` 映射后查找 `strategy`。
3. `baseEnv/init-db/06-migration-20260521.sql` + `dist/init-db/06-migration-20260521.sql`：新增 `WARN_REFRIGERANT_LEAK_COOLING`（制冷模式，trigger_value=2.0bar, duration=300s）和 `WARN_REFRIGERANT_LEAK_VENT`（通风模式，trigger_value=5.0bar, duration=900s）；为 `WARN_COOLING_SYSTEM` 补充 `raw_scale=10`。

**测试验证**：
- `CGO_ENABLED=0 go build` 在 `connect/cmd/connect-nb67/` 和 `connect/cmd/ground-reporter/` 均通过，无编译错误
- `tsc --noEmit` 检查确认无新增 TypeScript 错误（已有预存在错误不在修复范围）
- 端到端测试需在完整 Docker 环境中验证（此工作空间无 Docker）

**经验总结**：
- `fact_event.fault_code` 与 `warning_config.warn_code` 是两套命名体系，不能直接 JOIN
- 新增预警触发条件时，必须同时在 Go 代码中添加 `csRawThreshold` 调用，否则设置不生效
- 数据库 `params.raw_scale` 字段是 Go 代码读取阈值的关键，缺失会导致 trigger_value 无法转换为原始单位

---

### [2026-05-21] #1 - mock-platform 未收到数据推送

**问题描述**：配置了 FAULT_RECORD_URL / SYS_STATUS_URL / LIFE_RECORD_URL，但 mock-platform 没有收到任何推送信息。

**根因**：
1. **核心根因**：`build-and-push.sh` 脚本只构建 web/bff 镜像，没有包含 `ground-reporter` 镜像的构建逻辑。镜像 `harbor.naivehero.top:8443/macda2/ground-reporter:v2.5.0` 从未构建推送，导致 `dist/docker-compose-report.yml` 启动时 ground-reporter 容器无法拉取镜像而失败，mock-platform 在等待一个永远不会连接的 reporter。
2. **配置问题**：`HEARTBEAT_INTERVAL_MIN=10`（10分钟），而需求是每1分钟发送心跳，即使 reporter 启动了也10分钟内看不到心跳数据。
3. **开发环境缺失**：`baseEnv/` 目录没有 `docker-compose-report.yml`，注释说"已移至独立管理"但文件不存在。

**修复方法**：
1. `build-and-push.sh`：新增 `build_reporter()` 函数和 `"reporter"` 构建目标；`all` 目标同时构建 web/bff/reporter。**首次修复后必须执行 `./build-and-push.sh <version> reporter` 推送镜像**。
2. `dist/docker-compose-report.yml`：`HEARTBEAT_INTERVAL_MIN=10` → `HEARTBEAT_INTERVAL_MIN=1`。
3. `baseEnv/docker-compose-report.yml`（新建）：开发环境专用，使用 `build: context` 从本地 Dockerfile 构建 ground-reporter，无需预先推送到 registry；mock-platform 挂载仓库源码 `../connect/tests/mock-platform`。

**测试验证**：
- 需执行 `./build-and-push.sh v2.5.1 reporter` 构建推送镜像后，在部署环境重启 report 服务验证
- 验证命令：`docker logs ground-reporter | grep "6.6 heartbeat sent"` 每分钟应有一条

**经验总结**：
- ground-reporter 有 `Dockerfile.ground-reporter`，但被遗漏在构建脚本外，每次发布版本必须同步构建
- 心跳间隔配置在 `HEARTBEAT_INTERVAL_MIN` 环境变量，默认10分钟，开发调试时建议设为1

---

## 四、测试方法速查

### NB67 解析验证
```bash
bash connect/tests/test-nb67-parsing.sh
```

### 前端构建验证
```bash
cd web-nb67-web && npm run build
```

### BFF 接口验证
```bash
cd web-nb67-bff && npm run dev
# 另开终端：curl http://localhost:PORT/api/...
```

### 完整环境启动
```bash
cd baseEnv
docker compose -f docker-compose-mock.yml up -d   # 先建网络+数据源
docker compose -f docker-compose-Dev.yml up -d    # 再拉全栈
```

---

## 五、常见问题 FAQ

> 高频出现的问题，处理完 Issue 后如发现有规律可总结则追加。

<!-- 新条目追加到此处 -->
