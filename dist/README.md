# MACDA Connector — 部署手册

> **版本**：v2.5.0  
> **更新**：2026-05-13

---

## 📦 目录结构

```
dist/
├── image-save.sh               ← 【有网络机器】拉取并打包所有镜像为 .tar
├── image-load.sh               ← 【离线服务器】加载镜像 + MD5 完整性校验
├── install.sh                  ← 初始化目录权限 + 复制配置文件
├── start.sh                    ← 一键启动/停止/重启/状态
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
├── mock-data/
│   └── whole_frame-260203      ← NB67 测试帧数据（Mock 模式使用）
├── images/                     ← image-save.sh 生成的 .tar 镜像文件目录
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

# 以 root 运行安装脚本（默认数据目录: /data/MACDA2）
sudo ./install.sh

# 自定义数据目录（所有 docker-compose 挂载路径随之更改）
sudo ./install.sh --data-dir /opt/macda

# 仅更新配置文件，不影响已有数据
sudo ./install.sh --update
```

安装脚本会自动完成：
- 生成 `.env` 文件（写入 `DATA_DIR=<路径>`，docker-compose 自动读取，**无需手动修改任何 yml**）
- 创建 `${DATA_DIR}/` 下所有必要的挂载目录
- 设置 Redpanda (101:101)、TimescaleDB (1000:1000)、PgAdmin (5050:5050) 目录权限
- 复制 `config/*.yaml` → `${DATA_DIR}/connect/config/`
- 复制 `init-db/01-init.sql` → `${DATA_DIR}/timescaledb/init-db/`
- 复制 `mock-data/*` → `${DATA_DIR}/mock/connect/data/input/`

### 2. 准备镜像

**方式 A：在线部署**（服务器能访问 Harbor 镜像仓库）

```bash
# Harbor 对外开放，无需登录，启动时 docker compose 会自动拉取镜像
./start.sh
```

**方式 B：离线部署**（服务器无外网，需提前准备镜像包）

```bash
# ① 在有网络的机器上打包镜像
chmod +x image-save.sh
./image-save.sh                # 拉取并打包到 images/*.tar

# ② 将整个 dist/ 目录传输到离线服务器
scp -r dist/ user@server:/opt/macda/
# 或者打包后通过 U 盘/内网传输
tar czf macda-dist.tar.gz dist/

# ③ 在离线服务器上加载镜像
chmod +x image-load.sh
./image-load.sh --verify       # 先验证文件完整性（MD5 校验）
./image-load.sh                # 加载所有镜像到 Docker
```

### 3. 一键启动

```bash
# 启动全部服务（基础设施 + 应用层）
./start.sh

# 如需同时启动 mock 数据源（演示/调试模式，会自动读取 mock-data/ 下的测试帧）
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

## 📈 水平扩展（Scale）

### 扩展原则

系统所有 Kafka Topic 均为 **3 分区**，这决定了 Connect 处理服务的有效扩展上限：

> **黄金法则：单个消费者服务的实例数 ≤ Topic 分区数（3）**
>
> 超过分区数的实例会空转，浪费资源但不会报错。

### 可扩展的服务

以下 5 个 Connect 服务支持水平扩展（故意未设置 `container_name`）：

| 服务名 | 消费 Topic | 生产 Topic | 建议实例数 |
|--------|-----------|-----------|-----------|
| `connect-parser` | `signal-in` | `signal-parsed` | 1 ~ 3 |
| `connect-storage-writer` | `signal-in` | `signal-storage` | 1 ~ 3 |
| `connect-event-builder` | `signal-parsed` | `signal-alarm`, `signal-life` | 1 ~ 3 |
| `connect-pg-writer` | `signal-alarm`, `signal-life` | — (写 TimescaleDB) | 1 ~ 3 |
| `connect-event-writer` | `signal-alarm` | — (写 TimescaleDB) | 1 ~ 3 |

> ⚠️ `connect-topic-in`、`timescaledb`、`redpanda-*` 等服务**不支持**通过 `--scale` 扩展。

### 操作命令

```bash
# 查看当前各服务实例数
docker compose -f docker-compose-Data.yml ps

# 将 connect-parser 扩展到 3 个实例（生产高流量推荐）
docker compose -f docker-compose-Data.yml up -d --scale connect-parser=3

# 同时扩展多个服务（一条命令）
docker compose -f docker-compose-Data.yml up -d \
  --scale connect-parser=3 \
  --scale connect-storage-writer=3 \
  --scale connect-event-builder=2 \
  --scale connect-pg-writer=2 \
  --scale connect-event-writer=2

# 缩减回 1 个实例
docker compose -f docker-compose-Data.yml up -d --scale connect-parser=1
```

### 扩展策略参考

```
信号接入量        推荐配置
──────────────────────────────────────────────────
低负载 (< 1K/s)   所有服务保持 1 实例（默认）
中负载 (1K~5K/s)  connect-parser=2, 其余 1
高负载 (> 5K/s)   connect-parser=3, connect-storage-writer=3,
                  connect-event-builder=2, 其余 2
```

> 💡 **扩展时无需重启其他服务**，新实例启动后 Kafka 会自动触发 Rebalance，将分区分配给新实例。

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
| 前端 (Nginx) | `harbor.naivehero.top:8443/macda2/nb67-web:v2.5.0` |
| BFF (Node.js) | `harbor.naivehero.top:8443/macda2/nb67-bff:v2.5.0` |
| Connect 流水线 | `harbor.naivehero.top:8443/macda2/nb-parse-connect:v2.5.0` |
| TimescaleDB | `harbor.naivehero.top:8443/macda2/timescaledb-ha:pg14-ts2.19-all` |
| Redpanda | `harbor.naivehero.top:8443/macda2/redpanda:v25.3.7` |
