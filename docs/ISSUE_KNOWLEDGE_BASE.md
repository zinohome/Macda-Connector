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

---

## 三、历史 Issue 处理记录

> 最新记录在最前。每条记录包含：问题描述、根因、修复方法、测试验证、经验总结。

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
