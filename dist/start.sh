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
REPORT_COMPOSE="${SCRIPT_DIR}/docker-compose-report.yml"

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

# ── 1panel 检测 ──────────────────────────────────────────────
check_1panel() {
    if command -v 1pctl &>/dev/null 2>&1 || \
       systemctl is-active --quiet 1panel 2>/dev/null || \
       [ -f "/usr/local/bin/1panel" ]; then
        echo -e "\n\033[0;31m[ERR]\033[0m  检测到 1panel，请勿直接使用 start.sh！"
        echo ""
        echo "       应在 1panel 中管理各编排应用："
        echo "         $(dirname "$0")/1panel/data     ← 先启动"
        echo "         $(dirname "$0")/1panel/web"
        echo "         $(dirname "$0")/1panel/report"
        echo "         $(dirname "$0")/1panel/mock     ← 按需"
        echo "         $(dirname "$0")/1panel/desktop  ← 按需"
        echo ""
        echo "       如需强制使用 start.sh，请运行：./start.sh --force"
        echo ""
        exit 1
    fi
}

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

# ── 等待 TimescaleDB 就绪并执行数据库迁移 ─────────────────────
run_migrations() {
    log_step "等待 TimescaleDB 就绪"
    local elapsed=0
    until docker exec timescaledb psql -U postgres postgres -c "SELECT 1" &>/dev/null; do
        sleep 3; elapsed=$((elapsed + 3))
        [[ $elapsed -ge 60 ]] && { log_error "TimescaleDB 启动超时"; return 1; }
        echo -n "."
    done
    echo ""
    log_info "TimescaleDB 就绪，执行数据库迁移..."
    for sql in "${DATA_DIR}/timescaledb/init-db"/*.sql; do
        [[ -f "$sql" ]] || continue
        docker exec -i timescaledb psql -U postgres postgres \
            -v ON_ERROR_STOP=1 < "$sql" >/dev/null
        log_info "  ✓ $(basename "$sql")"
    done
    log_info "数据库迁移完成 ✓"
}

# ── 启动服务 ─────────────────────────────────────────────────
start() {
    log_step "启动基础设施层 (Data)"
    docker compose -f "${DATA_COMPOSE}" up -d
    log_info "基础设施启动完毕 ✓"

    run_migrations

    log_step "启动应用层 (Web)"
    docker compose -f "${WEB_COMPOSE}" up -d
    log_info "应用层启动完毕 ✓"

    log_step "启动报送层 (Report)"
    docker compose -f "${REPORT_COMPOSE}" up -d
    log_info "报送层启动完毕 ✓"

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
    log_step "停止报送层 (Report)"
    docker compose -f "${REPORT_COMPOSE}" down 2>/dev/null || true

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
    # --force 跳过 1panel 检测（高级用户明确知晓风险时使用）
    local force=false
    local args=()
    for arg in "$@"; do
        [[ "$arg" == "--force" ]] && force=true || args+=("$arg")
    done

    [[ "${force}" == "false" ]] && check_1panel
    check

    local cmd="${args[0]:-start}"

    case "${cmd}" in
        start)   start ;;
        mock)    start mock ;;
        stop)    stop ;;
        restart) stop; sleep 2; start ;;
        status)  status ;;
        report)  docker compose -f "${REPORT_COMPOSE}" up -d; log_info "report 启动完毕 ✓" ;;
        *)
            echo "用法: $0 [start|stop|restart|status|mock|report|--force]"
            exit 1
            ;;
    esac
}

main "$@"
