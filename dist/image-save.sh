#!/usr/bin/env bash
# =============================================================
# image-save.sh — 镜像离线打包脚本（在有网络的机器上运行）
#
# 功能：
#   1. 从私有 Harbor 拉取所有 compose 文件依赖的镜像
#   2. 将镜像打包为 .tar 文件保存到 images/ 目录
#   3. 生成 manifest.txt 记录镜像清单和校验值
#
# 用法：
#   ./image-save.sh             # 拉取并打包所有镜像
#   ./image-save.sh --no-pull   # 跳过 pull（使用本地已有镜像直接打包）
# =============================================================

set -euo pipefail

# ── 配置 ─────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGES_DIR="${SCRIPT_DIR}/images"
MANIFEST="${IMAGES_DIR}/manifest.txt"
NO_PULL=false
[[ "${1:-}" == "--no-pull" ]] && NO_PULL=true

# ── 所有 compose 文件依赖的镜像清单 ──────────────────────────
# 如升级版本，修改此处镜像 tag 即可
IMAGES=(
    # ── 基础工具 ──────────────────────────────────────────────
    "harbor.naivehero.top:8443/macda2/alpine:3.20"

    # ── 消息队列 (Redpanda) ───────────────────────────────────
    "harbor.naivehero.top:8443/macda2/redpanda:v25.3.7"
    "harbor.naivehero.top:8443/macda2/redpanda-console:v3.5.2"

    # ── 数据流处理 (Redpanda Connect) ────────────────────────
    "harbor.naivehero.top:8443/macda2/nb-parse-connect:v2.2.0-full"
    "harbor.naivehero.top:8443/macda2/connect:latest"

    # ── 时序数据库 ────────────────────────────────────────────
    "harbor.naivehero.top:8443/macda2/timescaledb-ha:pg14-ts2.19-all"
    "harbor.naivehero.top:8443/macda2/pgadmin4:9.12"

    # ── 存储适配器（备用方案，通常不启用）───────────────────
    "harbor.naivehero.top:8443/macda2/storage-adapter:v2.1.2"

    # ── 前端应用 ──────────────────────────────────────────────
    "harbor.naivehero.top:8443/macda2/nb67-bff:v2.1.2"
    "harbor.naivehero.top:8443/macda2/nb67-web:v2.1.2"
)

# ── 工具函数 ─────────────────────────────────────────────────
log_info()    { echo -e "\033[0;32m[INFO]\033[0m  $*"; }
log_warn()    { echo -e "\033[0;33m[WARN]\033[0m  $*"; }
log_step()    { echo -e "\n\033[1;36m══════ $* ══════\033[0m"; }
log_success() { echo -e "\033[0;32m[✓]\033[0m $*"; }
log_error()   { echo -e "\033[0;31m[ERROR]\033[0m $*" >&2; }

# 把镜像名转为安全的文件名（替换 :// / : 为 _）
image_to_filename() {
    echo "$1" | sed 's|[/:.]|_|g'
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

# ── 主流程 ───────────────────────────────────────────────────
main() {
    check

    log_step "准备镜像目录"
    mkdir -p "${IMAGES_DIR}"
    log_info "镜像保存目录: ${IMAGES_DIR}"

    # 初始化 manifest
    {
        echo "# MACDA Connector 镜像清单"
        echo "# 生成时间: $(date '+%Y-%m-%d %H:%M:%S')"
        echo "# 共 ${#IMAGES[@]} 个镜像"
        echo ""
    } > "${MANIFEST}"

    local success_count=0
    local fail_count=0

    for image in "${IMAGES[@]}"; do
        local filename
        filename="$(image_to_filename "${image}").tar"
        local output_path="${IMAGES_DIR}/${filename}"

        log_step "处理: ${image}"

        # 1. 拉取镜像
        if [[ "${NO_PULL}" == "false" ]]; then
            log_info "拉取镜像..."
            if ! docker pull "${image}"; then
                log_warn "拉取失败，跳过: ${image}"
                ((fail_count++)) || true
                continue
            fi
        else
            if ! docker image inspect "${image}" &>/dev/null; then
                log_warn "本地不存在镜像，跳过: ${image}"
                ((fail_count++)) || true
                continue
            fi
            log_info "使用本地已有镜像（跳过 pull）"
        fi

        # 2. 保存为 tar 文件
        log_info "打包镜像 → ${filename}"
        docker save "${image}" -o "${output_path}"

        # 3. 计算文件大小和 MD5
        local size
        size=$(du -sh "${output_path}" | cut -f1)
        local md5
        md5=$(md5sum "${output_path}" 2>/dev/null | cut -d' ' -f1 \
              || md5 -q "${output_path}" 2>/dev/null \
              || echo "N/A")

        log_success "打包完成: ${filename} (${size})"

        # 写入 manifest
        {
            echo "IMAGE=${image}"
            echo "FILE=${filename}"
            echo "SIZE=${size}"
            echo "MD5=${md5}"
            echo ""
        } >> "${MANIFEST}"

        ((success_count++)) || true
    done

    # ── 打印摘要 ─────────────────────────────────────────────
    log_step "打包完成"
    echo ""
    echo "  成功: ${success_count} 个"
    echo "  失败: ${fail_count} 个"
    echo ""
    echo "  镜像文件 (${IMAGES_DIR}/):"
    ls -lh "${IMAGES_DIR}"/*.tar 2>/dev/null | awk '{print "    " $5 "\t" $9}' || echo "    (无文件)"
    echo ""
    echo "  清单文件: ${MANIFEST}"
    echo ""
    echo "  ──────────────────────────────────────────────────"
    echo "  将整个 dist/ 目录（含 images/）拷贝到离线服务器后，运行："
    echo "    sudo ./image-load.sh    ← 加载镜像"
    echo "    sudo ./install.sh       ← 初始化目录和配置"
    echo "    ./start.sh              ← 启动所有服务"
    echo "  ──────────────────────────────────────────────────"
    echo ""
}

main "$@"
