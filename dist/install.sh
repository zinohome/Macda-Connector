#!/usr/bin/env bash
# =============================================================
# install.sh — MACDA Connector 初始化安装脚本
#
# 功能：
#   1. 接受可选的 --data-dir 参数自定义数据根目录
#   2. 生成 .env 文件（docker-compose 自动读取，无需手动修改 yml）
#   3. 创建所有 docker-compose 挂载所需的宿主机目录
#   4. 复制配置文件到对应挂载目录
#   5. 设置正确的目录权限（Redpanda 需要 101:101）
#
# 用法：
#   sudo ./install.sh                            # 使用默认目录 /data/MACDA2
#   sudo ./install.sh --data-dir /opt/macda      # 自定义数据根目录
#   sudo ./install.sh --update                   # 仅更新配置文件（不覆盖已有数据）
#   sudo ./install.sh --data-dir /opt/macda --update
# =============================================================

set -euo pipefail

# ── 参数解析 ─────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DATA_DIR="/data/MACDA2"   # 默认数据根目录
UPDATE_ONLY=false
ENV_FILE="${SCRIPT_DIR}/.env"

while [[ $# -gt 0 ]]; do
    case "$1" in
        --data-dir)
            BASE_DATA_DIR="${2:?--data-dir 需要指定路径}"
            shift 2
            ;;
        --update)
            UPDATE_ONLY=true
            shift
            ;;
        *)
            echo "用法: $0 [--data-dir <路径>] [--update]"
            exit 1
            ;;
    esac
done

# ── 工具函数 ─────────────────────────────────────────────────
log_info()    { echo -e "\033[0;32m[INFO]\033[0m  $*"; }
log_warn()    { echo -e "\033[0;33m[WARN]\033[0m  $*"; }
log_step()    { echo -e "\n\033[1;36m══════ $* ══════\033[0m"; }
log_success() { echo -e "\033[0;32m[✓]\033[0m $*"; }

# 创建目录（不存在时才创建）
mk_dir() {
    local dir="$1"
    if [[ ! -d "${dir}" ]]; then
        mkdir -p "${dir}"
        log_success "创建目录: ${dir}"
    else
        log_info "已存在，跳过: ${dir}"
    fi
}

# 复制文件（--update 模式强制覆盖，否则目标存在时跳过）
cp_file() {
    local src="$1"
    local dst="$2"
    if [[ ! -f "${src}" ]]; then
        log_warn "源文件不存在，跳过: ${src}"
        return
    fi
    if [[ -f "${dst}" ]] && [[ "${UPDATE_ONLY}" == "false" ]]; then
        log_info "已存在，跳过: ${dst}"
    else
        cp "${src}" "${dst}"
        log_success "复制: $(basename "${src}") → ${dst}"
    fi
}

# ── 0. 生成 .env 文件 ─────────────────────────────────────────
# docker-compose 会自动读取同目录下的 .env，
# 所有 compose 文件里的 ${DATA_DIR} 会被自动替换。
log_step "生成 .env 配置"

if [[ -f "${ENV_FILE}" ]] && [[ "${UPDATE_ONLY}" == "false" ]]; then
    # 已经存在时，读取已有的 DATA_DIR（除非用户显式指定了 --data-dir）
    existing_dir=$(grep "^DATA_DIR=" "${ENV_FILE}" 2>/dev/null | cut -d= -f2 || echo "")
    if [[ -n "${existing_dir}" && "${existing_dir}" != "${BASE_DATA_DIR}" \
          && "${BASE_DATA_DIR}" == "/data/MACDA2" ]]; then
        # 没有显式指定 --data-dir，沿用已有配置
        BASE_DATA_DIR="${existing_dir}"
        log_info ".env 已存在，沿用数据目录: ${BASE_DATA_DIR}"
    fi
fi

cat > "${ENV_FILE}" <<EOF
# MACDA Connector 环境配置
# 由 install.sh 自动生成，请勿手动编辑（重新运行 install.sh 可更新）
# 生成时间: $(date '+%Y-%m-%d %H:%M:%S')

# 数据根目录：所有 docker-compose 挂载路径的前缀
DATA_DIR=${BASE_DATA_DIR}
EOF

log_success ".env 已生成: DATA_DIR=${BASE_DATA_DIR}"

# ── 1. 创建数据目录 ───────────────────────────────────────────
log_step "创建数据目录 (基于 ${BASE_DATA_DIR})"

# Redpanda 3节点集群数据目录
mk_dir "${BASE_DATA_DIR}/redpanda/redpanda1"
mk_dir "${BASE_DATA_DIR}/redpanda/redpanda2"
mk_dir "${BASE_DATA_DIR}/redpanda/redpanda3"

# TimescaleDB 数据目录和初始化脚本目录
mk_dir "${BASE_DATA_DIR}/timescaledb/data"
mk_dir "${BASE_DATA_DIR}/timescaledb/init-db"

# PgAdmin 数据目录
mk_dir "${BASE_DATA_DIR}/pgadmin"

# Connect 流水线配置目录
mk_dir "${BASE_DATA_DIR}/connect/config"

# Mock 环境目录（用于演示/调试）
mk_dir "${BASE_DATA_DIR}/mock/redpanda/data"
mk_dir "${BASE_DATA_DIR}/mock/connect/data"
mk_dir "${BASE_DATA_DIR}/mock/connect/data/input"

# report 环境目录（mock-platform 源码）
mk_dir "${BASE_DATA_DIR}/connect/tests/mock-platform"

# ── 2. 设置目录权限 ───────────────────────────────────────────
log_step "设置目录权限"

# Redpanda 容器以 uid=101 运行
chown -R 101:101 "${BASE_DATA_DIR}/redpanda/redpanda1" \
                  "${BASE_DATA_DIR}/redpanda/redpanda2" \
                  "${BASE_DATA_DIR}/redpanda/redpanda3" \
                  "${BASE_DATA_DIR}/mock/redpanda/data"
chmod -R ug+rwX  "${BASE_DATA_DIR}/redpanda/redpanda1" \
                  "${BASE_DATA_DIR}/redpanda/redpanda2" \
                  "${BASE_DATA_DIR}/redpanda/redpanda3" \
                  "${BASE_DATA_DIR}/mock/redpanda/data"
log_success "Redpanda 目录权限设置完成 (101:101)"

# TimescaleDB 容器以 uid=1000 运行
chown -R 1000:1000 "${BASE_DATA_DIR}/timescaledb/data"
log_success "TimescaleDB 数据目录权限设置完成 (1000:1000)"

# PgAdmin 容器以 uid=5050 运行
chown -R 5050:5050 "${BASE_DATA_DIR}/pgadmin"
log_success "PgAdmin 目录权限设置完成 (5050:5050)"

chmod -R a+rwX "${BASE_DATA_DIR}/connect/config"
chmod -R a+rwX "${BASE_DATA_DIR}/mock/connect/data"
log_success "Connect 目录权限设置完成"

# ── 3. 复制配置文件 ───────────────────────────────────────────
log_step "复制 Connect 流水线配置文件"

for yaml_file in "${SCRIPT_DIR}"/config/*.yaml; do
    cp_file "${yaml_file}" "${BASE_DATA_DIR}/connect/config/$(basename "${yaml_file}")"
done

# ── 4. 复制数据库初始化 SQL ───────────────────────────────────
log_step "复制数据库初始化脚本"

for sql_file in "${SCRIPT_DIR}"/init-db/*.sql; do
    cp_file "${sql_file}" "${BASE_DATA_DIR}/timescaledb/init-db/$(basename "${sql_file}")"
done

# ── 4b. 复制 mock-platform 源码（report 环境用）─────────────────
log_step "复制 mock-platform 源码"

cp_file "${SCRIPT_DIR}/mock-platform/main.go" \
        "${BASE_DATA_DIR}/connect/tests/mock-platform/main.go"

# ── 5. 复制 Mock 测试数据 ─────────────────────────────────────
log_step "复制 Mock 测试数据"

for mock_file in "${SCRIPT_DIR}"/mock-data/*; do
    [[ -f "${mock_file}" ]] || continue
    cp_file "${mock_file}" "${BASE_DATA_DIR}/mock/connect/data/input/$(basename "${mock_file}")"
done

# ── 6. 生成 1panel 编排目录 ──────────────────────────────────
log_step "生成 1panel 编排目录"

PANEL_ROOT="${SCRIPT_DIR}/1panel"

declare -A PANEL_ENVS=(
    [data]="docker-compose-Data.yml"
    [web]="docker-compose-Web.yml"
    [mock]="docker-compose-mock.yml"
    [report]="docker-compose-report.yml"
    [desktop]="docker-compose-Desktop.yml"
)

for env_name in data web mock report desktop; do
    src_file="${SCRIPT_DIR}/${PANEL_ENVS[$env_name]}"
    env_dir="${PANEL_ROOT}/${env_name}"
    mkdir -p "${env_dir}"
    # 在文件头注入 name: 字段，1panel 和 docker compose 都以此作为应用名称
    { echo "name: ${env_name}"; echo ""; cat "${src_file}"; } > "${env_dir}/docker-compose.yml"
    echo "DATA_DIR=${BASE_DATA_DIR}" > "${env_dir}/.env"
    log_success "1panel/${env_name}/docker-compose.yml  (name: ${env_name})"
done

log_info "1panel 目录就绪: ${PANEL_ROOT}"

# ── 7. 检测 1panel，输出对应指引 ─────────────────────────────
HAS_1PANEL=false
if command -v 1pctl &>/dev/null 2>&1 || \
   systemctl is-active --quiet 1panel 2>/dev/null || \
   [ -f "/usr/local/bin/1panel" ]; then
    HAS_1PANEL=true
fi

log_step "安装完成"
echo ""
echo "  数据根目录 : ${BASE_DATA_DIR}"
echo ""

if [[ "${HAS_1PANEL}" == "true" ]]; then
    echo -e "  \033[1;32m✓ 检测到 1panel，请在 1panel 中按以下顺序添加编排应用：\033[0m"
    echo ""
    echo "  ① 必须最先启动："
    echo "     $(pwd)/1panel/data"
    echo ""
    echo "  ② 数据层就绪后启动："
    echo "     $(pwd)/1panel/web"
    echo "     $(pwd)/1panel/report"
    echo ""
    echo "  ③ 按需启动："
    echo "     $(pwd)/1panel/mock     ← Mock 数据源"
    echo "     $(pwd)/1panel/desktop  ← 远程桌面"
    echo ""
    echo "  ⚠️  请勿直接运行 start.sh（与 1panel 管理冲突）"
else
    echo -e "  \033[1;32m✓ 未检测到 1panel，使用 start.sh 管理服务：\033[0m"
    echo ""
    echo "    chmod +x start.sh"
    echo "    ./start.sh        # 启动全部服务（Data + Web + Report）"
    echo "    ./start.sh mock   # 同上 + Mock 数据源"
fi
echo ""
