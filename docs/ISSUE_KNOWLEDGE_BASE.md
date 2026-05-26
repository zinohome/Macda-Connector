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
[2026-05-26] #7 - 预警报文code不随车厢变化/fault_name多余/完整预警描述未显示 - ground-reporter/BFF/前端

---

## 三、历史 Issue 处理记录

> 最新记录在最前。每条记录包含：问题描述、根因、修复方法、测试验证、经验总结。

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
