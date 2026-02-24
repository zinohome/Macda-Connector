# =============================================================
# web-nb67-web Dockerfile
# 多阶段构建（Multi-stage Build）策略：
#   Stage 1 (builder): Node.js 环境执行 npm run build，产出 dist/
#   Stage 2 (runtime): Nginx Alpine 纯静态托管，镜像极小（~30MB）
# =============================================================

# ── Stage 1: 构建阶段 ─────────────────────────────────────────
FROM harbor.naivehero.top:8443/macda2/node:22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 优先拷贝 package 文件并安装依赖
# 利用 Docker 层缓存：只有 package*.json 变化时才重新 npm ci
COPY package.json package-lock.json ./

# 使用 npm ci（比 npm install 更快更可靠，专为 CI/CD 设计）
RUN npm ci

# 拷贝全部源码（.dockerignore 会排除 node_modules、dist 等）
COPY . .

# 执行生产构建，产物输出到 /app/dist/
# vue-tsc 做类型检查，vite build 做打包
RUN npm run build

# ── Stage 2: 运行阶段 ─────────────────────────────────────────
FROM harbor.naivehero.top:8443/macda2/nginx:alpine AS runtime

# 删除 Nginx 默认配置，替换为项目定制配置
RUN rm /etc/nginx/conf.d/default.conf

# 拷贝我们的 Nginx 配置（含 /api /ws 代理规则）
COPY nginx.conf /etc/nginx/nginx.conf

# 从 builder 阶段拷贝 Vue 打包产物到 Nginx 静态目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 声明对外端口
EXPOSE 8080

# 健康检查：探测 Nginx /health 端点（10s间隔，3次失败则标记不健康）
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# Nginx 前台运行（不 daemonize），保持容器存活
CMD ["nginx", "-g", "daemon off;"]
