# MACDA-Connector

> 地铁列车 HVAC 系统遥测数据处理平台

当前版本：**v2.5.12**（2026-05-21）

---

## 项目概述

MACDA-Connector 是一套用于地铁列车 HVAC（暖通空调）系统的实时遥测数据处理平台。系统基于 **Redpanda Connect（Go）** 构建流处理管道，将现场设备上报的 NB67 二进制协议帧解析、存储并生成运维事件，配合 Web 前后端提供实时监控、历史数据查询、预警管理等功能。

**核心技术栈：**

| 层级 | 技术 |
|------|------|
| 消息队列 | Redpanda（Kafka 兼容，3 节点集群） |
| 流处理 | Redpanda Connect + Go 自定义处理器 |
| 数据库 | TimescaleDB（PostgreSQL 时序扩展） |
| BFF | TypeScript / Fastify（REST + WebSocket） |
| 前端 | Vue 3 / Vite / Element Plus / ECharts |
| 容器化 | Docker Compose（开发 + 生产 + 离线部署） |

---

## 系统架构

```
[ 现场设备 / Mock 信号源 ]
         │  NB67 二进制帧 (462字节, 1Hz)
         ▼
   Redpanda 集群
         │
         ├─► connect-parser        ── NB67 解析 → signal-parsed (JSON, 180+字段)
         │         │
         │         ├─► connect-storage-writer ── 采样落盘 → signal-storage (10s间隔)
         │         │         └─► connect-pg-writer ── 写入 TimescaleDB
         │         │
         │         └─► connect-event-builder  ── 告警/预测/寿命事件检测
         │                   └─► connect-event-writer ── 写入 TimescaleDB
         │
   TimescaleDB
         │
   nb67-bff (Fastify REST + WebSocket)
         │
   nb67-web (Vue 3 前端)
```

**双轨数据流：**
- **全量轨**（`signal-parsed`）：用于实时告警检测，无采样
- **采样轨**（`signal-storage`）：10 秒间隔采样，用于长期趋势存储，减少 90% 写入量

---

## 模块结构

```
Macda-Connector/
├── connect/                        # Go 流处理核心
│   ├── cmd/connect-nb67/           # Redpanda Connect 自定义处理器（NB67解析、事件检测）
│   ├── cmd/storage-adapter/        # Plan B 高性能直写方案（备用）
│   ├── cmd/ground-reporter/        # 数据上报组件
│   ├── codec/                      # NB67 协议定义（Kaitai Struct，SSOT）
│   │   └── NB67.ksy                # ⚠️ 唯一真实来源，不可手工改生成文件
│   ├── config/                     # Redpanda Connect 流水线 YAML 配置
│   │   ├── nb67-parser.yaml
│   │   ├── nb67-storage-writer.yaml
│   │   ├── nb67-event-builder.yaml
│   │   ├── nb67-event-writer.yaml
│   │   └── nb67-pg-writer.yaml
│   └── tests/                      # 测试脚本
├── web-nb67-bff/                   # BFF 服务（TypeScript / Fastify）
├── web-nb67-web/                   # 前端（Vue 3 / Element Plus / ECharts）
│   └── src/views/                  # 页面：首页、列车详情、历史数据、历史告警/预警、预警配置等
├── baseEnv/                        # 开发环境 Docker Compose 编排
│   ├── docker-compose-mock.yml     # 网络创建 + Mock 数据源（必须首先启动）
│   └── docker-compose-Dev.yml      # 完整全栈（基础设施 + 流水线 + Web）
├── reportEnv/                      # 数据上报环境
├── dist/                           # 离线部署包（客户生产环境）
│   ├── install.sh                  # 初始化安装
│   ├── start.sh                    # 一键启停
│   ├── image-save.sh / image-load.sh  # 镜像打包/加载
│   └── init-db/                    # SQL 初始化 + 历次迁移脚本
├── deploy.sh                       # 快速部署脚本
├── release.sh                      # 版本发布脚本
├── build-and-push.sh               # 镜像构建推送（web/bff）
└── docs/                           # 技术研究文档（架构评估、方案对比等）
```

---

## 快速开始（开发环境）

### 前置依赖

- Docker + Docker Compose v2
- Go 1.21+（仅 connect 模块本地构建时需要）
- Node.js 18+（仅前端本地开发时需要）

### 启动完整全栈

```bash
cd baseEnv

# 1. 创建 Docker 网络 + 启动 Mock 数据源
docker compose -f docker-compose-mock.yml up -d

# 2. 启动完整全栈（基础设施 + 流水线 + Web）
docker compose -f docker-compose-Dev.yml up -d
```

启动后访问：
- 前端：`http://localhost:3000`（或查看 Compose 文件中的具体端口）
- BFF API：`http://localhost:3001`

### 动态扩容 connect-parser

```bash
docker compose -f docker-compose-Dev.yml up -d --scale connect-parser=3
```

### 仅更新 Web/BFF（不重启基础设施）

```bash
docker stop nb67-web nb67-bff
docker compose -f baseEnv/docker-compose-Prod.yml up -d
```

---

## 本地开发

### connect-nb67（Go）

```bash
cd connect/cmd/connect-nb67

# 构建（静态编译）
CGO_ENABLED=0 go build -o connect-nb67 .

# 运行
./connect-nb67 -c ../../config/nb67-parser.yaml
```

### web-nb67-bff（TypeScript）

```bash
cd web-nb67-bff
npm install
npm run dev
```

### web-nb67-web（Vue 3）

```bash
cd web-nb67-web
npm install
npm run dev    # 开发服务器
npm run build  # 生产构建
```

### 验证 NB67 解析

```bash
bash connect/tests/test-nb67-parsing.sh
```

---

## 镜像构建与发布

```bash
# 构建并推送 web + bff
./build-and-push.sh v2.5.12

# 仅构建 web 或 bff
./build-and-push.sh v2.5.12 web
./build-and-push.sh v2.5.12 bff

# 构建 connect-nb67（linux/amd64）
docker buildx build --platform linux/amd64 \
  -f connect/Dockerfile.connect \
  -t harbor.naivehero.top:8443/macda2/nb-parse-connect:v2.x \
  --push connect

# 发版（更新 CHANGELOG + 打 tag）
./release.sh prepare v2.5.13
./release.sh publish
```

---

## NB67 协议维护

`connect/codec/NB67.ksy` 是协议定义的**唯一真实来源（SSOT）**。

> ⚠️ `connect/cmd/connect-nb67/nb67.go` 和 `connect/codec/nb67.go` 均为 Kaitai 自动生成文件，**禁止手工修改**。

修改协议字段时，按顺序执行：

```bash
# 修改 NB67.ksy 后重新生成 Go 代码
kaitai-struct-compiler -t go codec/NB67.ksy -o connect/cmd/connect-nb67/
```

然后同步更新：`nb67_processor.go` → 存储映射 → API 契约 → 文档。

---

## 关键配置说明

### RUNTIME（设备时钟策略）

| 值 | 时间字段 | 适用场景 |
|----|----------|----------|
| `DEV` | `ingest_time`（服务器入库时间） | 现场设备时钟不可信（当前默认） |
| `PRD` | `event_time`（设备上报时间） | 设备时钟已校准 |

所有 Compose 文件默认设为 `DEV`，切换前需确认现场设备时钟精度。

### Plan A / Plan B 存储方案

- **Plan A（默认）**：`connect-pg-writer` 通过 Redpanda Connect SQL 插件直写 TimescaleDB
- **Plan B（备用）**：`storage-adapter`（Go 原生），通过 `profiles: ["plan_b"]` 激活

---

## 生产部署（离线环境）

详见 [dist/README.md](dist/README.md)。

```bash
# 有网络机器：拉取并打包镜像
cd dist && bash image-save.sh

# 离线服务器：加载镜像 + 初始化
bash image-load.sh
bash install.sh
bash start.sh
```

---

## 相关文档

- [架构与研究文档](docs/README.md)
- [Connect 模块说明](connect/README.md)
- [部署手册](dist/README.md)
- [Changelog](CHANGELOG.md)
