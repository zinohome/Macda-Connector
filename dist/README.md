# MACDA Connector — 部署手册

> **版本**：v2.5.12  
> **更新**：2026-05-19

---

## 📦 目录结构

```
dist/
├── image-save.sh               ← 【有网络机器】拉取并打包所有镜像为 .tar
├── image-load.sh               ← 【离线服务器】加载镜像 + MD5 完整性校验
├── install.sh                  ← 初始化目录权限 + 复制配置文件（支持 --1panel 模式）
├── start.sh                    ← 一键启动/停止/重启/状态（普通环境使用）
├── docker-compose-Data.yml     ← 基础设施层（Redpanda + TimescaleDB + Connect 流水线）
├── docker-compose-Web.yml      ← 应用层（BFF + 前端 Nginx）
├── docker-compose-report.yml   ← 报送层（ground-reporter + mock-platform）
├── docker-compose-mock.yml     ← Mock 数据源（调试/演示用）
├── docker-compose-Desktop.yml  ← 远程桌面（webtop GUI，按需启动）
├── config/                     ← Redpanda Connect 流水线配置文件
│   ├── nb67-parser.yaml
│   ├── nb67-event-builder.yaml
│   ├── nb67-event-writer.yaml
│   ├── nb67-storage-writer.yaml
│   └── nb67-pg-writer.yaml
├── init-db/                    ← TimescaleDB 初始化与迁移 SQL（start.sh 自动执行）
│   ├── 01-init.sql
│   ├── 02-migration-20260504.sql
│   ├── 03-migration-20260512.sql
│   ├── 04-migration-20260513.sql
│   └── 05-migration-20260513.sql
├── mock-platform/
│   └── main.go                 ← mock-platform 源码（install.sh 复制到 DATA_DIR）
├── mock-data/
│   └── whole_frame-260203      ← NB67 测试帧数据（Mock 模式使用）
├── images/                     ← image-save.sh 生成的 .tar 镜像文件目录
└── README.md                   ← 本文档
```

> `1panel/` 目录由 `install.sh --1panel` 自动生成，不在版本控制中。

---

## 🏗️ 系统架构

```
                        ┌──────────────────────────────────────────────────────┐
                        │                Docker Network: macdanet               │
                        │                                                      │
  [ 设备/信号源 ]  ──►  │  mock-redpanda   (信号模拟，仅调试)                  │
                        │       ↓                                              │
  [ 真实设备   ]  ──►  │  Redpanda 集群   (3节点，生产级消息队列)              │
                        │  ├── connect-topic-in      (信号接入)                │
                        │  ├── connect-parser        (NB67协议解析)            │
                        │  ├── connect-storage-writer (原始数据落盘)           │
                        │  ├── connect-event-builder  (事件检测)               │
                        │  ├── connect-pg-writer      (数据持久化)             │
                        │  └── connect-event-writer   (事件历史写入)           │
                        │       ↓                                              │
                        │  TimescaleDB    (时序数据库)                         │
                        │       ↓                                              │
                        │  nb67-bff       (BFF API + WebSocket)               │
                        │       ↓                                              │
  [ 浏览器 ]  ◄──────  │  nb67-web       (Nginx，对外 :8080)                  │
                        │                                                      │
                        │  ground-reporter  ──►  地面健康管理平台               │
                        │  mock-platform    ◄──  (本地测试接收端，:18188)       │
                        └──────────────────────────────────────────────────────┘
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

---

### 方式一：普通环境（无 1panel）

#### 1. 运行安装脚本

```bash
chmod +x install.sh start.sh

# 以 root 运行（默认数据目录: /data/MACDA2）
sudo ./install.sh

# 自定义数据目录
sudo ./install.sh --data-dir /opt/macda

# 仅更新配置文件，不影响已有数据
sudo ./install.sh --update
```

安装脚本自动完成：
- 生成 `.env`（写入 `DATA_DIR`，docker-compose 自动读取，无需手动修改 yml）
- 创建所有挂载目录并设置权限（Redpanda 101:101 / TimescaleDB 1000:1000 / PgAdmin 5050:5050）
- 复制 Connect 配置、5 个 SQL 文件、mock-platform 源码、Mock 测试数据

#### 2. 准备镜像

**在线部署**（服务器能访问 Harbor）：
```bash
# docker compose 启动时自动拉取，无需额外操作
```

**离线部署**（无外网）：
```bash
# ① 有网络的机器上打包镜像
chmod +x image-save.sh && ./image-save.sh

# ② 传输整个 dist/ 到目标服务器
tar czf macda-dist.tar.gz dist/

# ③ 在目标服务器加载
chmod +x image-load.sh
./image-load.sh --verify   # 先验证 MD5
./image-load.sh            # 加载所有镜像
```

#### 3. 一键启动

```bash
./start.sh            # 启动全部服务（Data + Web + Report）
./start.sh mock       # 同上 + Mock 数据源（调试/演示用）
./start.sh desktop    # 同上 + 远程桌面（:10001）
./start.sh all        # 全部启动（含 Mock + Desktop）
```

#### 4. 验证

```bash
./start.sh status

# 前端访问
http://<服务器IP>:8080

# Redpanda Console
http://<服务器IP>:28080

# mock-platform（本地报送验证）
http://<服务器IP>:18188
```

---

### 方式二：1panel 环境

#### 1. 运行安装脚本（加 --1panel 参数）

```bash
sudo ./install.sh --1panel
```

脚本除完成普通安装外，还会在 `dist/1panel/` 下生成 5 个子目录，每个目录含独立的 `docker-compose.yml`（注入 `name:` 字段）和 `.env` 文件。

安装完成后脚本会打印每个子目录的完整路径。

#### 2. 在 1panel 中按顺序添加编排应用

> **注意：必须先启动 data，再启动其他环境**

| 顺序 | 1panel 应用名 | 路径 |
|------|-------------|------|
| ① | **data** | `<dist路径>/1panel/data` |
| ② | **web** | `<dist路径>/1panel/web` |
| ② | **report** | `<dist路径>/1panel/report` |
| ③按需 | **mock** | `<dist路径>/1panel/mock` |
| ③按需 | **desktop** | `<dist路径>/1panel/desktop` |

在 1panel「容器」→「编排」→「创建编排」中，选择对应目录即可。应用名称由 `name:` 字段决定，与目录名一致。

---

## ⚙️ 关键配置

### 地面平台报送（ground-reporter）

编辑 `docker-compose-report.yml`，修改以下环境变量：

```yaml
environment:
  - PLATFORM_IP=10.12.48.187    # 地面健康管理平台真实 IP
  - PLATFORM_PORT=8188           # 端口
  - PLATFORM_API_KEY=            # X-Api-Key 认证（现场联调时填写）
```

本地测试时将 `PLATFORM_IP` 改为 `mock-platform`，`mock-platform` 会将收到的 JSON 格式化打印到日志。

### 切换时间分析模式

```yaml
# docker-compose-Web.yml 中 nb67-bff 的环境变量
- RUNTIME=DEV   # 使用数据入库时间（设备时钟不可信）
- RUNTIME=PRD   # 使用设备上报事件时间（设备时钟已校准）
```

### 修改外网访问地址（Redpanda）

```bash
sed -i 's/192.168.32.17/<服务器IP>/g' docker-compose-Data.yml
```

---

## 📈 水平扩展

所有 Kafka Topic 均为 **3 分区**，单服务有效扩展上限为 3 个实例。

> ⚠️ 使用 `--scale` 扩展时，需先移除对应服务的 `container_name:`，否则多实例会因名称冲突报错。

```bash
# 扩展 connect-parser 到 3 个实例
docker compose -f docker-compose-Data.yml up -d --scale connect-parser=3
```

| 信号接入量 | 推荐配置 |
|-----------|---------|
| 低 (< 1K/s) | 所有服务 1 实例（默认）|
| 中 (1K~5K/s) | connect-parser=2，其余 1 |
| 高 (> 5K/s) | connect-parser=3，connect-storage-writer=3，其余 2 |

---

## 🛑 停止服务

```bash
./start.sh stop      # 停止所有服务（保留数据）
./start.sh restart   # 重启所有服务
./start.sh status    # 查看状态
```

---

## 🔧 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| `nb67-bff` unhealthy | BFF 正在连接 DB/Kafka，有 30s 宽限期 | 等待 30s 后再检查 |
| Redpanda 无法启动 | 数据目录权限问题 | 检查 `${DATA_DIR}/redpanda/` 权限（需 101:101） |
| TimescaleDB 无数据表 | SQL 未执行 | 重新运行 `sudo ./install.sh` 后重启容器 |
| ground-reporter 连接超时 | 平台 IP 不通 | 检查 `PLATFORM_IP` 配置；本地测试改为 `mock-platform` |
| 1panel 显示应用名错误 | 未使用 `--1panel` 参数生成子目录 | 重新运行 `sudo ./install.sh --1panel` |

---

## 📋 镜像版本

| 服务 | 镜像 |
|------|------|
| 前端 (Nginx) | `harbor.naivehero.top:8443/macda2/nb67-web:v2.5.12` |
| BFF (Node.js) | `harbor.naivehero.top:8443/macda2/nb67-bff:v2.5.11` |
| Connect 流水线 | `harbor.naivehero.top:8443/macda2/nb-parse-connect:v2.5.0` |
| 地面报送服务 | `harbor.naivehero.top:8443/macda2/ground-reporter:v2.5.0` |
| TimescaleDB | `harbor.naivehero.top:8443/macda2/timescaledb-ha:pg14-ts2.19-all` |
| Redpanda | `harbor.naivehero.top:8443/macda2/redpanda:v25.3.7` |
| mock-platform | `golang:1.24-alpine`（运行时镜像，需联网）|
