#!/usr/bin/env bash
# =============================================================
# image-load.sh — 镜像离线加载脚本（在离线服务器上运行）
#
# 功能：
#   1. 读取 images/ 目录下的所有 .tar 文件
#   2. 逐一加载到本地 Docker
#   3. 对照 manifest.txt 验证完整性
#
# 用法：
#   sudo ./image-load.sh          # 加载所有镜像
#   sudo ./image-load.sh --verify # 仅验证 MD5，不加载
# =============================================================

set -euo pipefail

# ── 配置 ─────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGES_DIR="${SCRIPT_DIR}/images"
MANIFEST="${IMAGES_DIR}/manifest.txt"
VERIFY_ONLY=false
[[ "${1:-}" == "--verify" ]] && VERIFY_ONLY=true

# ── 工具函数 ─────────────────────────────────────────────────
log_info()    { echo -e "\033[0;32m[INFO]\033[0m  $*"; }
log_warn()    { echo -e "\033[0;33m[WARN]\033[0m  $*"; }
log_step()    { echo -e "\n\033[1;36m══════ $* ══════\033[0m"; }
log_success() { echo -e "\033[0;32m[✓]\033[0m $*"; }
log_error()   { echo -e "\033[0;31m[ERROR]\033[0m $*" >&2; }

# 计算文件 MD5（兼容 Linux md5sum 和 macOS md5）
calc_md5() {
    md5sum "$1" 2>/dev/null | cut -d' ' -f1 \
    || md5 -q "$1" 2>/dev/null \
    || echo "N/A"
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
    if [[ ! -d "${IMAGES_DIR}" ]]; then
        log_error "images/ 目录不存在: ${IMAGES_DIR}"
        exit 1
    fi
}

# ── MD5 验证 ─────────────────────────────────────────────────
verify() {
    if [[ ! -f "${MANIFEST}" ]]; then
        log_warn "manifest.txt 不存在，跳过完整性校验"
        return
    fi

    log_step "验证镜像文件完整性"

    local pass=0
    local fail=0

    # 按块解析 manifest（每个镜像占 4 行 + 空行）
    local image="" file="" expected_md5=""
    while IFS= read -r line || [[ -n "${line}" ]]; do
        [[ "${line}" == \#* ]] && continue  # 跳过注释行
        [[ -z "${line}" ]] && {
            # 空行 = 一个镜像块结束，执行校验
            if [[ -n "${file}" && -n "${expected_md5}" && "${expected_md5}" != "N/A" ]]; then
                local tar_path="${IMAGES_DIR}/${file}"
                if [[ ! -f "${tar_path}" ]]; then
                    log_warn "文件不存在: ${file}"
                    ((fail++)) || true
                elif [[ "$(calc_md5 "${tar_path}")" == "${expected_md5}" ]]; then
                    log_success "MD5 校验通过: ${file}"
                    ((pass++)) || true
                else
                    log_error "MD5 校验失败: ${file}（文件可能损坏）"
                    ((fail++)) || true
                fi
            fi
            image=""; file=""; expected_md5=""
            continue
        }

        case "${line}" in
            IMAGE=*) image="${line#IMAGE=}" ;;
            FILE=*)  file="${line#FILE=}" ;;
            MD5=*)   expected_md5="${line#MD5=}" ;;
        esac
    done < "${MANIFEST}"

    echo ""
    echo "  校验结果: 通过 ${pass} 个，失败 ${fail} 个"
    [[ ${fail} -gt 0 ]] && log_error "存在文件损坏，请重新传输损坏的 .tar 文件" && exit 1
}

# ── 加载镜像 ─────────────────────────────────────────────────
load_images() {
    log_step "加载 Docker 镜像"

    local tar_files=("${IMAGES_DIR}"/*.tar)
    if [[ ! -e "${tar_files[0]}" ]]; then
        log_error "images/ 目录中没有找到 .tar 文件"
        exit 1
    fi

    local success_count=0
    local fail_count=0

    for tar_file in "${IMAGES_DIR}"/*.tar; do
        local filename
        filename="$(basename "${tar_file}")"
        log_info "加载: ${filename}"

        if docker load -i "${tar_file}"; then
            log_success "加载成功: ${filename}"
            ((success_count++)) || true
        else
            log_error "加载失败: ${filename}"
            ((fail_count++)) || true
        fi
    done

    echo ""
    echo "  加载结果: 成功 ${success_count} 个，失败 ${fail_count} 个"

    if [[ ${fail_count} -gt 0 ]]; then
        log_error "部分镜像加载失败，请检查上方错误信息"
        exit 1
    fi
}

# ── 打印已加载镜像 ───────────────────────────────────────────
show_loaded() {
    log_step "当前 Docker 镜像列表"
    docker images --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" \
        | grep -E "macda2|REPOSITORY" || true
}

# ── 主流程 ───────────────────────────────────────────────────
main() {
    check

    log_step "MACDA Connector 离线镜像加载"
    log_info "镜像目录: ${IMAGES_DIR}"

    # 先验证完整性
    verify

    if [[ "${VERIFY_ONLY}" == "true" ]]; then
        log_info "仅验证模式，跳过加载"
        exit 0
    fi

    # 加载镜像
    load_images

    # 显示加载结果
    show_loaded

    echo ""
    echo "  ──────────────────────────────────────────────────"
    echo "  镜像加载完成！下一步："
    echo "    sudo ./install.sh   ← 初始化目录和配置（如未执行）"
    echo "    ./start.sh          ← 启动所有服务"
    echo "  ──────────────────────────────────────────────────"
    echo ""
}

main "$@"
