#!/usr/bin/env bash
# =============================================================
# start.sh — MACDA Connector 一键启动脚本
#
# 启动顺序：
#   1. docker-compose-Data.yml  ── 基础设施层
#      (Redpanda 集群 + TimescaleDB + Connect 流水线)
#   2. docker-compose-Web.yml   ── 应用层
#      (nb67-bff + nb67-web)
#
# 用法:
#   ./start.sh          # 启动全部服务
#   ./start.sh stop     # 停止全部服务
#   ./start.sh restart  # 重启全部服务
#   ./start.sh status   # 查看所有容器状态
#   ./start.sh mock     # 同时启动 mock 数据源（调试用）
# =============================================================

set -euo pipefail

# ── 配置 ─────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_COMPOSE="${SCRIPT_DIR}/docker-compose-Data.yml"
WEB_COMPOSE="${SCRIPT_DIR}/docker-compose-Web.yml"
MOCK_COMPOSE="${SCRIPT_DIR}/docker-compose-mock.yml"

# 加载 .env（由 install.sh 生成），将 DATA_DIR export 给 docker-compose
if [[ -f "${SCRIPT_DIR}/.env" ]]; then
    # shellcheck disable=SC1091
    set -a; source "${SCRIPT_DIR}/.env"; set +a
else
    echo -e "\033[0;33m[WARN]\033[0m  未找到 .env 文件，请先运行 sudo ./install.sh"
    echo -e "         使用默认数据目录: /data/MACDA2"
    export DATA_DIR="/data/MACDA2"
fi


# ── 工具函数 ─────────────────────────────────────────────────
log_info()  { echo -e "\033[0;32m[INFO]\033[0m  $*"; }
log_step()  { echo -e "\n\033[1;36m══════ $* ══════\033[0m"; }
log_error() { echo -e "\033[0;31m[ERROR]\033[0m $*" >&2; }

# ── 前置检查 ─────────────────────────────────────────────────
check() {
    if ! command -v docker &>/dev/null; then
        log_error "未找到 docker 命令"
        exit 1
    fi
    if ! docker info &>/dev/null; then
        log_error "Docker daemon 未运行"
        exit 1
    fi
}

# ── 启动服务 ─────────────────────────────────────────────────
start() {
    log_step "启动基础设施层 (Data)"
    docker compose -f "${DATA_COMPOSE}" up -d
    log_info "基础设施启动完毕 ✓"

    log_step "启动应用层 (Web)"
    docker compose -f "${WEB_COMPOSE}" up -d
    log_info "应用层启动完毕 ✓"

    if [[ "${1:-}" == "mock" ]]; then
        log_step "启动 Mock 数据源"
        docker compose -f "${MOCK_COMPOSE}" up -d
        log_info "Mock 数据源启动完毕 ✓"
    fi

    log_step "所有服务状态"
    status
}

# ── 停止服务 ─────────────────────────────────────────────────
stop() {
    log_step "停止应用层 (Web)"
    docker compose -f "${WEB_COMPOSE}" down

    log_step "停止基础设施层 (Data)"
    docker compose -f "${DATA_COMPOSE}" down

    # 如果 mock 在运行，也一并停止
    if docker compose -f "${MOCK_COMPOSE}" ps -q 2>/dev/null | grep -q .; then
        log_step "停止 Mock 数据源"
        docker compose -f "${MOCK_COMPOSE}" down
    fi

    log_info "所有服务已停止 ✓"
}

# ── 查看状态 ─────────────────────────────────────────────────
status() {
    echo ""
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" \
        --filter "network=macdanet" 2>/dev/null || docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""
}

# ── 主入口 ───────────────────────────────────────────────────
main() {
    check
    local cmd="${1:-start}"

    case "${cmd}" in
        start)   start ;;
        mock)    start mock ;;
        stop)    stop ;;
        restart) stop; sleep 2; start ;;
        status)  status ;;
        *)
            echo "用法: $0 [start|stop|restart|status|mock]"
            exit 1
            ;;
    esac
}

main "$@"
