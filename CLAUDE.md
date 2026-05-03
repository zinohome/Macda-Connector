# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MACDA-Connector 是一个地铁列车 HVAC 系统遥测数据处理平台，核心目标是将现有的 Python/Faust 系统迁移到 Redpanda Connect（Go）。项目处于迁移和开发阶段。

**数据流向总览**：
```
现场设备 → signal-in (Kafka topic, 462字节NB67二进制)
  → connect-parser (Go插件解析) → signal-parsed (JSON, 180+字段)
  → connect-storage-writer → signal-storage (采样数据)
  → connect-event-builder → signal-alarm/predict/life (事件数据)
  → connect-pg-writer/storage-adapter → TimescaleDB
  → nb67-bff (Fastify REST+WebSocket API)
  → nb67-web (Vue 3 前端)
```

## 模块结构

| 目录 | 语言/技术 | 职责 |
|------|-----------|------|
| `connect/cmd/connect-nb67/` | Go | NB67解析器（Redpanda Connect自定义处理器） |
| `connect/cmd/storage-adapter/` | Go | 高性能直写TimescaleDB备用方案（Plan B） |
| `connect/codec/` | Kaitai Struct + Go | NB67二进制协议定义，**唯一真实来源(SSOT)** |
| `connect/config/` | YAML | Redpanda Connect 流水线配置 |
| `web-nb67-bff/` | TypeScript/Fastify | BFF：REST + WebSocket，查询TimescaleDB和Kafka |
| `web-nb67-web/` | Vue 3/Vite | 前端：Element Plus + ECharts |
| `baseEnv/` | Docker Compose | 开发/生产基础设施编排 |
| `dist/` | Docker Compose + Shell | 离线部署包（客户环境） |

## 常用命令

### connect-nb67（Go）

```bash
cd connect/cmd/connect-nb67

# 构建（静态编译，解决动态链接问题）
CGO_ENABLED=0 go build -o connect-nb67 .

# 运行（指定配置文件）
./connect-nb67 -c ../../config/nb67-parser.yaml
```

### storage-adapter（Go）

```bash
cd connect/cmd/storage-adapter
go build -o storage-adapter .
```

### web-nb67-bff（TypeScript）

```bash
cd web-nb67-bff
npm install
npm run dev      # 开发模式（tsx 热重载）
```

### web-nb67-web（Vue 3）

```bash
cd web-nb67-web
npm install
npm run dev      # Vite 开发服务器
npm run build    # 生产构建
```

### 测试脚本

```bash
bash connect/tests/test-nb67-parsing.sh       # 验证NB67解析（唯一现有测试脚本）
```

### 开发环境启动（Docker）

`baseEnv/` 下有四个 Compose 文件，职责各异：

| 文件 | 用途 | 说明 |
|------|------|------|
| `docker-compose-mock.yml` | **网络创建者 + 数据源**（单节点 Redpanda + 数据生成器） | 唯一创建 `macdanet` 网络的文件；数据生成器每 33ms 读一次二进制帧写入 `signal-in`；端口段 4xxxx |
| `docker-compose-Dev.yml` | **完整全栈**（3节点 Redpanda + TimescaleDB + pgAdmin + 全部5个流水线阶段 + nb67-web + nb67-bff） | 依赖已存在的 `macdanet`；内置 `connect-topic-in` 将 mock-redpanda 数据桥接到真实集群 |
| `docker-compose-Prod.yml` | **轻量应用层**（仅 nb67-web + nb67-bff） | 用于独立更新 Web/BFF 而不重启基础设施；与 Dev.yml 不能同时运行（容器名冲突） |
| `docker-compose-desktop.yml` | **远程桌面**（webtop 浏览器 GUI） | 端口 10001，供需要 GUI 的远程访问场景使用 |

**典型启动顺序**：mock 先创建网络，Dev 再拉起全栈：

```bash
cd baseEnv

# 1. 创建网络 + 启动数据源
docker compose -f docker-compose-mock.yml up -d

# 2. 启动完整全栈（基础设施 + 流水线 + Web）
docker compose -f docker-compose-Dev.yml up -d

# 动态扩/缩容 connect-parser（不超过 partition 数）
docker compose -f docker-compose-Dev.yml up -d --scale connect-parser=3

# 仅更新 Web/BFF（先停 Dev 中的同名容器，再用 Prod 拉起）
docker stop nb67-web nb67-bff
docker compose -f docker-compose-Prod.yml up -d
```

### Docker镜像构建与推送

```bash
# web/bff 镜像（统一入口脚本）
./build-and-push.sh v1.2.0       # 构建并推送 web + bff
./build-and-push.sh v1.2.0 web   # 仅 web
./build-and-push.sh v1.2.0 bff   # 仅 bff

# connect-nb67 镜像（在 linux/amd64 上）
docker build -f connect/Dockerfile.connect \
  -t harbor.naivehero.top:8443/macda2/nb-parse-connect:v2.x \
  connect
docker push harbor.naivehero.top:8443/macda2/nb-parse-connect:v2.x

# macOS Apple Silicon 使用 buildx
docker buildx build --platform linux/amd64 \
  -f connect/Dockerfile.connect \
  -t harbor.naivehero.top:8443/macda2/nb-parse-connect:v2.x \
  --push connect
```

### 从 Kaitai 规格文件重新生成 Go 代码

```bash
# 修改 codec/NB67.ksy 后执行
kaitai-struct-compiler -t go codec/NB67.ksy -o connect/cmd/connect-nb67/
```

## 关键架构决策

### NB67 协议解析

- `connect/codec/NB67.ksy` 是协议定义的**唯一真实来源**
- `connect/cmd/connect-nb67/nb67.go` 和 `connect/codec/nb67.go` 均为 **Kaitai 自动生成文件，禁止手工修改**
- 任何协议字段变更必须同时更新：`NB67.ksy` → 重新生成 Go 代码 → 更新 `nb67_processor.go` → 更新存储映射 → 更新 API 契约 → 更新文档

### 三段流水线配置（当前主线）

`nb67-parser.yaml` → `nb67-storage-writer.yaml` → `nb67-event-builder.yaml` → `nb67-event-writer.yaml`

- `nb67-pg-writer.yaml`：将 `signal-storage` 采样数据写入 TimescaleDB（与 storage-writer 配合）
- `nb67-event-writer.yaml`：消费 `signal-alarm`/`signal-predict`/`signal-life`，持久化事件到 TimescaleDB

> `nb67-connect.yaml` 是已废弃的单体配置，仅兼容保留。

### 双轨管道

- **全量轨**（`signal-parsed`）：用于实时告警，无采样
- **采样轨**（`signal-storage`）：用于长期存储，减少90%写入量

### Plan A / Plan B 存储方案

- **Plan A**（默认）：`connect-pg-writer` 用 Redpanda Connect SQL 插件直接写入 TimescaleDB
- **Plan B**（备用）：`storage-adapter`（Go 原生）在 Plan A 达到瓶颈时切换，通过 `profiles: ["plan_b"]` 激活

### RUNTIME：设备时钟可信度开关

`RUNTIME` 不是环境选择器，而是**设备时钟可信度策略**，影响 BFF 所有时序查询以及 `connect-event-builder`：

| 值 | 时间字段 | 适用场景 |
|----|----------|----------|
| `DEV` | `ingest_time`（服务器入库时间） | 现场设备时钟不可信（当前所有 Compose 文件的实际配置） |
| `PRD` | `event_time`（设备上报时间） | 现场设备时钟已校准 |

代码默认值是 `PRD`（见 `web-nb67-bff/src/config/index.ts`），但所有 Compose 文件均显式设为 `DEV`。切换前必须确认现场设备时钟精度。

## 目录与文件组织规则（强制）

- **所有文档**必须放在 `docs/` 下，禁止散落在根目录
- **测试脚本**必须放在 `connect/tests/`，不允许在根目录出现 `tests/`
- **临时文件**必须放在 `temp/` 下
- **根目录**仅保留 `README.md`、`AGENTS.md`、`CLAUDE.md` 等必要入口文件

## 任务跟踪（beads）

```bash
bd onboard       # 初始化（首次使用）
bd ready         # 查看可执行任务
bd show <id>     # 查看任务详情
bd update <id> --status in_progress
bd close <id>    # 完成任务
bd sync          # 与 git 同步
```

## 执行计划参考

所有重大功能开发以 `docs/research/11-macda-refactor-execution-plan.md` 为基准，按以下优先级执行：
1. Kafka Connect 开发（NB67解析、存储、告警、预测）
2. 展示数据 API（契约优先、配置驱动）
3. 前端适配与优化

前端集成主目标是 `web-nb67-web/`（即 `web-nb67.250513`），禁止以 Grafana 为优先设计。
