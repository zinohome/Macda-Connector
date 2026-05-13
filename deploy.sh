#!/usr/bin/env bash
# =============================================================================
# deploy.sh — MACDA Connector 一键全量部署脚本
#
# 功能：停止所有栈 → 清空数据 → 预置文件 → 按序启动 → 初始化数据库 → 验证
# 用法：bash deploy.sh [--clean]
#   --clean  清空所有数据目录（全新部署），不加则仅重启容器
# =============================================================================
set -euo pipefail

# ── 配置区 ─────────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASENV_DIR="${SCRIPT_DIR}/baseEnv"
CONNECT_CONFIG_DIR="${SCRIPT_DIR}/connect/config"
MOCK_DATA_DIR="${SCRIPT_DIR}/dist/mock-data"

# 宿主机数据挂载根目录
HOST_DATA="/data/MACDA2"

# compose 文件
COMPOSE_MOCK="${BASENV_DIR}/docker-compose-mock.yml"
COMPOSE_DEV="${BASENV_DIR}/docker-compose-Dev.yml"
COMPOSE_DESKTOP="${BASENV_DIR}/docker-compose-desktop.yml"

# docker 命令（自动判断是否需要 sudo）
DOCKER="docker"
if ! docker info &>/dev/null 2>&1; then
    DOCKER="sudo docker"
fi
DC="${DOCKER} compose"

# ── 颜色输出 ───────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[1;36m'; NC='\033[0m'
log_step()  { echo -e "\n${CYAN}══════ $* ══════${NC}"; }
log_info()  { echo -e "${GREEN}[OK]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[!!]${NC}  $*"; }
log_error() { echo -e "${RED}[ERR]${NC} $*" >&2; exit 1; }
log_wait()  { echo -e "${YELLOW}[..]${NC}  $*"; }

# ── 参数解析 ───────────────────────────────────────────────────────────────
DO_CLEAN=false
for arg in "$@"; do
    [[ "$arg" == "--clean" ]] && DO_CLEAN=true
done

# ── Step 1: 停止所有栈 ────────────────────────────────────────────────────
log_step "Step 1: 停止所有容器栈"

for stack_cfg in "desktop:${COMPOSE_DESKTOP}" "dev:${COMPOSE_DEV}" "mock:${COMPOSE_MOCK}"; do
    name="${stack_cfg%%:*}"
    cfg="${stack_cfg##*:}"
    if [[ -f "$cfg" ]]; then
        log_info "停止 ${name} 栈..."
        ${DC} -p "${name}" -f "${cfg}" down --remove-orphans 2>/dev/null || true
    fi
done

# 强制清理可能遗留的 MACDA 相关容器
# （解决不同 project 名 / 手动启动 / baseenv 默认 project 导致的孤儿容器冲突）
log_info "清理遗留容器..."
${DOCKER} ps -a --format "{{.Names}}" | \
    grep -E "^nb67-|^dev-connect|^dev-init|^baseenv-connect|^timescaledb$|^pgadmin$|^redpanda-|^redpanda-console$|^mock-" | \
    xargs -r ${DOCKER} rm -f 2>/dev/null || true
log_info "所有容器已停止并清理"

# ── Step 2: 清空数据目录（仅 --clean 模式）───────────────────────────────
if [[ "${DO_CLEAN}" == "true" ]]; then
    log_step "Step 2: 清空所有数据目录（--clean 模式）"
    sudo rm -rf \
        "${HOST_DATA}/timescaledb/data/data" \
        "${HOST_DATA}/redpanda/redpanda1/"* \
        "${HOST_DATA}/redpanda/redpanda2/"* \
        "${HOST_DATA}/redpanda/redpanda3/"* \
        "${HOST_DATA}/mock/redpanda/data/"* \
        "${HOST_DATA}/mock/connect/data/"* \
        2>/dev/null || true
    log_info "TimescaleDB、Redpanda（3节点）、Mock 数据已清空"
else
    log_step "Step 2: 跳过数据清空（未指定 --clean）"
    log_warn "保留现有数据。如需全新部署请使用: bash deploy.sh --clean"
fi

# ── Step 3: 预置宿主机挂载文件 ────────────────────────────────────────────
log_step "Step 3: 预置宿主机文件"

# 3a. Connect 配置文件（避免 Docker 创建假目录）
mkdir -p "${HOST_DATA}/connect/config"
for yaml in nb67-parser.yaml nb67-storage-writer.yaml nb67-event-builder.yaml \
            nb67-event-writer.yaml nb67-pg-writer.yaml; do
    src="${CONNECT_CONFIG_DIR}/${yaml}"
    dst="${HOST_DATA}/connect/config/${yaml}"
    if [[ -f "$src" ]]; then
        # 若目标是目录（Docker 创建的假文件），先删掉
        [[ -d "$dst" ]] && sudo rm -rf "$dst"
        sudo cp "$src" "$dst"
    else
        log_error "找不到 Connect 配置文件: ${src}"
    fi
done
log_info "Connect 配置文件就位 (5个 yaml)"

# 3b. mock 数据源（二进制录制帧）
mkdir -p "${HOST_DATA}/mock/connect/data/input"
FRAME_FILE="${MOCK_DATA_DIR}/whole_frame-260203"
if [[ -f "$FRAME_FILE" ]]; then
    [[ -d "${HOST_DATA}/mock/connect/data/input/whole_frame-260203" ]] && \
        sudo rm -rf "${HOST_DATA}/mock/connect/data/input/whole_frame-260203"
    sudo cp "$FRAME_FILE" "${HOST_DATA}/mock/connect/data/input/"
    log_info "Mock 录制帧文件就位"
else
    log_error "找不到 Mock 数据文件: ${FRAME_FILE}"
fi

# 3c. init-db SQL 文件
mkdir -p "${HOST_DATA}/timescaledb/init-db"
for sql in 01-init.sql 02-migration-20260504.sql 03-migration-20260512.sql 04-migration-20260513.sql 05-migration-20260513.sql; do
    src="${BASENV_DIR}/init-db/${sql}"
    if [[ -f "$src" ]]; then
        sudo cp "$src" "${HOST_DATA}/timescaledb/init-db/"
    else
        log_error "找不到 SQL 文件: ${src}"
    fi
done
log_info "数据库初始化 SQL 就位 (5个文件)"

# ── Step 4: 按序启动三个栈 ───────────────────────────────────────────────
log_step "Step 4: 启动容器栈"

log_info "启动 mock（创建 macdanet 网络 + 数据源）..."
${DC} -p mock -f "${COMPOSE_MOCK}" up -d
log_info "mock 栈已启动"

log_info "启动 dev（基础设施 + 流水线 + Web）..."
${DC} -p dev -f "${COMPOSE_DEV}" up -d
log_info "dev 栈已启动"

log_info "启动 desktop（远程桌面 GUI）..."
${DC} -p desktop -f "${COMPOSE_DESKTOP}" up -d 2>/dev/null || \
    log_warn "desktop 启动失败或未配置，跳过"

# ── Step 5: 等待 TimescaleDB 就绪 ────────────────────────────────────────
log_step "Step 5: 等待 TimescaleDB 就绪"
log_wait "等待 PostgreSQL 接受连接..."

MAX_WAIT=60
elapsed=0
until ${DOCKER} exec timescaledb psql -U postgres postgres -c "SELECT 1" &>/dev/null; do
    sleep 3
    elapsed=$((elapsed + 3))
    if [[ $elapsed -ge $MAX_WAIT ]]; then
        log_error "等待 TimescaleDB 超时（${MAX_WAIT}s），请检查容器状态"
    fi
    echo -n "."
done
echo ""
log_info "TimescaleDB 就绪"

# ── Step 6: 执行数据库初始化（幂等，始终执行）───────────────────────────
log_step "Step 6: 初始化数据库 Schema"

# 说明：不依赖 initdb 自动执行（数据目录非空时会被跳过）
# SQL 文件全部使用 IF NOT EXISTS / WHERE NOT EXISTS，多次执行安全无副作用

run_sql() {
    local label="$1" file="$2"
    log_info "执行 ${label}..."
    # ON_ERROR_STOP=1：遇 ERROR 立即非零退出，配合外层 set -e 快速失败
    # stdout 丢弃（NOTICE 太多），stderr 保留（ERROR/FATAL 可见）
    if ! ${DOCKER} exec -i timescaledb psql -U postgres postgres \
            -v ON_ERROR_STOP=1 < "${file}" >/dev/null; then
        log_error "SQL 执行失败: ${label}"
    fi
}

run_sql "01-init.sql（Schema + 表 + hypertable）" \
    "${HOST_DATA}/timescaledb/init-db/01-init.sql"
run_sql "02-migration-20260504.sql（recovery_time + warning_config）" \
    "${HOST_DATA}/timescaledb/init-db/02-migration-20260504.sql"
run_sql "03-migration-20260512.sql（raw_scale + duration_seconds）" \
    "${HOST_DATA}/timescaledb/init-db/03-migration-20260512.sql"
run_sql "04-migration-20260513.sql（WARN_CABIN_OVERHEAT 阈值修正）" \
    "${HOST_DATA}/timescaledb/init-db/04-migration-20260513.sql"
run_sql "05-migration-20260513.sql（PHM策略+出厂默认+测试模式）" \
    "${HOST_DATA}/timescaledb/init-db/05-migration-20260513.sql"

# 验证表存在
TABLE_COUNT=$(${DOCKER} exec timescaledb psql -U postgres postgres -tAc \
    "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='hvac'" 2>/dev/null)
if [[ "${TABLE_COUNT}" -ge 4 ]]; then
    log_info "数据库初始化完成（hvac schema 共 ${TABLE_COUNT} 张表）"
else
    log_error "数据库初始化异常，请检查 init SQL 执行结果"
fi

# ── Step 7: 验证全链路运行 ────────────────────────────────────────────────
log_step "Step 7: 验证系统运行状态"

# 7a. 等待 mock 数据流入（最多60秒）
log_wait "等待 Mock 数据写入数据库..."
elapsed=0
until [[ $(${DOCKER} exec timescaledb psql -U postgres postgres -tAc \
    "SELECT COUNT(*) FROM hvac.fact_raw" 2>/dev/null | tr -d ' ') -gt 0 ]]; do
    sleep 3
    elapsed=$((elapsed + 3))
    if [[ $elapsed -ge 60 ]]; then
        log_warn "60秒内未收到数据，请检查 mock-connect 容器状态"
        break
    fi
    echo -n "."
done
echo ""

# 7b. 打印最终状态
RECORD_COUNT=$(${DOCKER} exec timescaledb psql -U postgres postgres -tAc \
    "SELECT COUNT(*) FROM hvac.fact_raw" 2>/dev/null | tr -d ' ')
LATEST_TIME=$(${DOCKER} exec timescaledb psql -U postgres postgres -tAc \
    "SELECT MAX(ingest_time) FROM hvac.fact_raw" 2>/dev/null | tr -d ' ')
WC_COUNT=$(${DOCKER} exec timescaledb psql -U postgres postgres -tAc \
    "SELECT COUNT(*) FROM hvac.warning_config" 2>/dev/null | tr -d ' ')

# 7c. 容器状态
HEALTHY=$(${DOCKER} ps --format "{{.Status}}" | grep -c "healthy" || true)
TOTAL=$(${DOCKER} ps --format "{{.Names}}" | grep -v "1Panel\|cloud9" | wc -l)

# ── 最终报告 ──────────────────────────────────────────────────────────────
log_step "部署完成"
cat << EOF

  ┌──────────────────────────────────────────────┐
  │           MACDA Connector 部署报告            │
  ├──────────────────────────────────────────────┤
  │  容器总数     : ${TOTAL} 个                           │
  │  Healthy 数   : ${HEALTHY} 个                           │
  │  fact_raw 记录: ${RECORD_COUNT} 条                        │
  │  最新数据时间 : ${LATEST_TIME}  │
  │  warning_config: ${WC_COUNT} 条                        │
  ├──────────────────────────────────────────────┤
  │  前端访问地址 : http://$(hostname -I | awk '{print $1}'):8080    │
  │  pgAdmin     : http://$(hostname -I | awk '{print $1}'):5050    │
  └──────────────────────────────────────────────┘

EOF

if [[ "${RECORD_COUNT}" -gt 0 ]]; then
    log_info "系统运行正常 ✓"
else
    log_warn "数据库暂无数据，请检查 mock-connect 容器日志:"
    echo "  ${DOCKER} logs mock-mock-connect-1 --tail 20"
fi
