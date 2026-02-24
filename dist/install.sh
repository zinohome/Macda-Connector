#!/usr/bin/env bash
# =============================================================
# install.sh — MACDA Connector 初始化安装脚本
#
# 功能：
#   1. 创建所有 docker-compose 挂载所需的宿主机目录
#   2. 复制配置文件到对应挂载目录
#   3. 设置正确的目录权限（Redpanda 需要 101:101）
#
# 用法：
#   sudo ./install.sh           # 首次部署，完整安装
#   sudo ./install.sh --update  # 仅更新配置文件（不覆盖已有数据目录）
# =============================================================

set -euo pipefail

# ── 配置 ─────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DATA_DIR="/data/MACDA2"
UPDATE_ONLY=false

[[ "${1:-}" == "--update" ]] && UPDATE_ONLY=true

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

# 复制文件（目标存在时给出提示，--update 模式下强制覆盖）
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

# ── 1. 创建数据目录 ───────────────────────────────────────────
log_step "创建数据目录"

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

# ── 2. 设置目录权限 ───────────────────────────────────────────
log_step "设置目录权限"

# Redpanda 容器以 uid=101 运行，需要对应权限
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

# Connect 配置目录和 mock-connect 数据目录设置为全可读写
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

cp_file "${SCRIPT_DIR}/init-db/01-init.sql" \
        "${BASE_DATA_DIR}/timescaledb/init-db/01-init.sql"

# ── 5. 复制 Mock 测试数据 ─────────────────────────────────────
log_step "复制 Mock 测试数据"

for mock_file in "${SCRIPT_DIR}"/mock-data/*; do
    [[ -f "${mock_file}" ]] || continue
    cp_file "${mock_file}" "${BASE_DATA_DIR}/mock/connect/data/input/$(basename "${mock_file}")"
done

# ── 6. 完成摘要 ───────────────────────────────────────────────
log_step "安装完成"
echo ""
echo "  数据根目录: ${BASE_DATA_DIR}"
echo ""
echo "  目录结构:"
find "${BASE_DATA_DIR}" -maxdepth 3 -type d | sort | sed 's|'"${BASE_DATA_DIR}"'|  /data/MACDA2|'
echo ""
echo "  下一步："
echo "    chmod +x start.sh"
echo "    ./start.sh          # 启动所有服务"
echo "    ./start.sh mock     # 启动服务 + Mock 数据源"
echo ""
