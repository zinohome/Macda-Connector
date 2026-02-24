# =============================================================
# web-nb67-bff Dockerfile
# BFF (Backend For Frontend) 生产镜像
# 策略：
#   - 使用 tsx 直接运行 TypeScript（无需预编译，简化构建）
#   - 多阶段构建分离 devDependencies，仅保留运行时依赖
#   - 所有敏感配置（数据库、Kafka）通过环境变量注入，不进镜像
# =============================================================

# ── Stage 1: 依赖安装阶段 ─────────────────────────────────────
FROM node:22-alpine AS deps

WORKDIR /app

# 优先拷贝 package 文件，利用缓存层
COPY package.json package-lock.json ./

# 安装所有依赖（包含 devDependencies，tsx 在其中）
RUN npm ci

# ── Stage 2: 运行阶段 ─────────────────────────────────────────
FROM node:22-alpine AS runtime

# 设置工作目录
WORKDIR /app

# 从 deps 阶段拷贝完整 node_modules（含 tsx、typescript）
COPY --from=deps /app/node_modules ./node_modules

# 拷贝 package.json（tsx 需要读取 "type": "module" 配置）
COPY package.json ./

# 拷贝 TypeScript 配置文件（tsx 运行时需要）
COPY tsconfig.json ./

# 拷贝全部源码（.dockerignore 已排除无关文件）
COPY src ./src

# ── 环境变量默认值 ────────────────────────────────────────────
# 以下为容器内默认值，生产部署时通过 docker-compose environment: 覆盖
# 注意：敏感信息（密码）应通过 Docker Secrets 或外部配置管理

# BFF 监听配置
ENV PORT=3000
ENV HOST=0.0.0.0
ENV LOG_LEVEL=info

# 时间分析模式（DEV=入库时间 / PRD=设备时间），生产环境按需覆盖
ENV RUNTIME=DEV

# 数据库默认指向 Docker 网络内的 timescaledb 服务
ENV DATABASE_URL=postgres://postgres:passw0rd@timescaledb:5432/postgres?sslmode=disable

# Kafka 默认指向 Docker 网络内的 Redpanda 集群
ENV KAFKA_BROKERS=redpanda-1:9092,redpanda-2:9092,redpanda-3:9092

# 声明对外端口
EXPOSE 3000

# 健康检查：探测 Fastify 服务根路径（10s间隔，5次失败则标记不健康）
HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD wget -qO- http://localhost:3000/health || exit 1

# 启动命令：用 tsx 直接运行 TypeScript 入口文件
CMD ["node_modules/.bin/tsx", "src/index.ts"]
