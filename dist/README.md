# MACDA Connector — 部署手册

> **版本**：v2.1.2  
> **更新**：2026-02-24

---

## 📦 目录结构

```
dist/
├── start.sh                    ← 一键启动/停止脚本
├── docker-compose-Data.yml     ← 基础设施层（Redpanda + TimescaleDB + Connect 流水线）
├── docker-compose-Web.yml      ← 应用层（BFF + 前端 Nginx）
├── docker-compose-mock.yml     ← Mock 数据源（调试/演示用）
├── docker-compose-Desktop.yml  ← 本地单机开发环境
├── config/                     ← Redpanda Connect 流水线配置文件
│   ├── nb67-parser.yaml
│   ├── nb67-event-builder.yaml
│   ├── nb67-event-writer.yaml
│   ├── nb67-storage-writer.yaml
│   └── nb67-pg-writer.yaml
├── init-db/
│   └── 01-init.sql             ← TimescaleDB 初始化 SQL（首次部署执行）
└── README.md                   ← 本文档
```

---

## 🏗️ 系统架构

```
                        ┌─────────────────────────────────────────────┐
                        │               Docker Network: macdanet       │
                        │                                             │
  [ 设备/信号源 ]  ──►  │  mock-redpanda  (信号模拟, 仅调试)          │
                        │       ↓                                     │
  [ 真实设备   ]  ──►  │  Redpanda 集群  (3节点, 生产级消息队列)      │
                        │  ├── connect-topic-in    (信号接入)          │
                        │  ├── connect-parser      (NB67协议解析)     │
                        │  ├── connect-storage-writer (原始数据落盘)  │
                        │  ├── connect-event-builder  (事件检测)      │
                        │  ├── connect-pg-writer      (事件持久化)    │
                        │  └── connect-event-writer   (事件历史写入)  │
                        │       ↓                                     │
                        │  TimescaleDB  (时序数据库)                  │
                        │       ↓                                     │
                        │  nb67-bff     (BFF API + WebSocket)         │
                        │       ↓                                     │
  [ 浏览器 ]  ◄──────  │  nb67-web     (Nginx, 对外 :8080)           │
                        └─────────────────────────────────────────────┘
```

---

## 🚀 部署步骤

### 前置要求

| 依赖 | 版本要求 |
|------|---------|
| Docker | ≥ 24.0 |
| Docker Compose | ≥ 2.20 |
| 服务器内存 | ≥ 8 GB |
| 服务器磁盘 | ≥ 50 GB |

### 1. 运行安装脚本（首次部署）

```bash
# 赋予执行权限
chmod +x install.sh start.sh

# 以 root 运行安装脚本（需要创建系统目录并设置权限）
sudo ./install.sh

# 如需更新配置文件（不影响已有数据）
sudo ./install.sh --update
```

安装脚本会自动完成：
- 创建 `/data/MACDA2/` 下所有必要的挂载目录
- 设置 Redpanda (101:101)、TimescaleDB (1000:1000)、PgAdmin (5050:5050) 目录权限
- 复制 `config/*.yaml` → `/data/MACDA2/connect/config/`
- 复制 `init-db/01-init.sql` → `/data/MACDA2/timescaledb/init-db/`
- 复制 `mock-data/*` → `/data/MACDA2/mock/connect/data/input/`

### 2. 登录私有镜像仓库

```bash
docker login harbor.naivehero.top:8443
```

### 4. 一键启动

```bash
# 赋予执行权限（首次）
chmod +x start.sh

# 启动全部服务（基础设施 + 应用层）
./start.sh

# 如需同时启动 mock 数据源（演示/调试模式）
./start.sh mock
```

### 5. 验证部署

```bash
# 查看所有容器状态
./start.sh status

# 正常状态下应看到所有容器为 Up (healthy)
# 访问前端
http://<服务器IP>:8080

# 访问 API 文档
http://<服务器IP>:8080/api/docs  (通过 Nginx 代理)

# 访问 Redpanda Console（消息队列管理）
http://<服务器IP>:28080
```

---

## ⚙️ 关键配置说明

### 修改数据库连接（BFF）

编辑 `docker-compose-Web.yml`，修改以下环境变量：

```yaml
environment:
  - DATABASE_URL=postgres://postgres:passw0rd@timescaledb:5432/postgres?sslmode=disable
  - KAFKA_BROKERS=redpanda-1:9092,redpanda-2:9092,redpanda-3:9092
```

### 切换时间分析模式

```yaml
environment:
  # DEV: 使用数据入库时间（设备时钟不可信时使用）
  # PRD: 使用设备上报事件时间（设备时钟已校准时使用）
  - RUNTIME=DEV
```

### 修改外网访问地址（Redpanda）

在 `docker-compose-Data.yml` 中，将所有 `192.168.32.17` 替换为实际服务器 IP：

```bash
sed -i 's/192.168.32.17/<服务器IP>/g' docker-compose-Data.yml
```

---

## 🛑 停止服务

```bash
# 停止所有服务（保留数据）
./start.sh stop

# 重启所有服务
./start.sh restart
```

---

## 🔧 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| `nb67-bff` 容器 unhealthy | BFF 正在连接 DB/Kafka，有 30s 启动宽限期 | 等待 30s 后再检查状态 |
| `nb67-web` 启动失败 | 依赖 `nb67-bff` 健康检查通过 | 等 BFF 变为 healthy 后自动启动 |
| Redpanda 无法启动 | 数据目录权限问题 | 检查 `/data/MACDA2/redpanda/` 目录权限 (需 101:101) |
| TimescaleDB 无数据表 | 首次部署未执行初始化 SQL | 复制 `init-db/01-init.sql` 到 `/data/MACDA2/timescaledb/init-db/` 后重建容器 |
| 前端图片丢失 | 镜像版本过旧 | 重新拉取最新镜像 `docker compose pull` |

---

## 📋 镜像版本

| 服务 | 镜像 |
|------|------|
| 前端 (Nginx) | `harbor.naivehero.top:8443/macda2/nb67-web:v2.1.2` |
| BFF (Node.js) | `harbor.naivehero.top:8443/macda2/nb67-bff:v2.1.2` |
| Connect 流水线 | `harbor.naivehero.top:8443/macda2/nb-parse-connect:v2.2.0-full` |
| TimescaleDB | `harbor.naivehero.top:8443/macda2/timescaledb-ha:pg14-ts2.19-all` |
| Redpanda | `harbor.naivehero.top:8443/macda2/redpanda:v25.3.7` |
