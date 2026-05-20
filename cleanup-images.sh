#!/usr/bin/env bash
# =============================================================
# cleanup-images.sh — 清理开发机上的历史构建镜像
#
# 策略：对每个服务镜像，只保留版本号最新的一个，删除其余历史版本。
#       同时清理所有 dangling 镜像（无标签的中间层镜像）。
#
# 用法:
#   ./cleanup-images.sh          # 预览模式（只显示，不删除）
#   ./cleanup-images.sh --delete # 执行删除
# =============================================================

set -euo pipefail

REGISTRY="harbor.naivehero.top:8443/macda2"
DRY_RUN=true
[[ "${1:-}" == "--delete" ]] && DRY_RUN=false

# docker 命令（自动判断是否需要 sudo）
DOCKER="docker"
if ! docker info &>/dev/null 2>&1; then
    DOCKER="sudo docker"
fi

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[1;36m'; NC='\033[0m'
log_step()  { echo -e "\n${CYAN}══════ $* ══════${NC}"; }
log_info()  { echo -e "${GREEN}[OK]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[!!]${NC}  $*"; }
log_delete(){ echo -e "${RED}[DEL]${NC} $*"; }

# 对 vX.Y.Z 格式版本排序，输出最大版本（最后一行）
latest_version() {
    sort -t. -k1,1V -k2,2n -k3,3n | tail -n1
}

cleanup_service() {
    local service="$1"
    local image_prefix="${REGISTRY}/${service}"

    log_step "处理服务: ${service}"

    # 列出该服务的所有本地镜像标签（过滤掉 <none>）
    local tags
    tags=$($DOCKER images "${image_prefix}" --format "{{.Tag}}" 2>/dev/null \
           | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+' || true)

    if [[ -z "$tags" ]]; then
        log_warn "未找到 ${image_prefix} 的本地镜像，跳过"
        return
    fi

    local count; count=$(echo "$tags" | wc -l | tr -d ' ')
    log_info "找到 ${count} 个标签: $(echo "$tags" | tr '\n' ' ')"

    if [[ "$count" -le 1 ]]; then
        log_info "只有一个版本，无需清理"
        return
    fi

    local latest; latest=$(echo "$tags" | latest_version)
    log_info "保留最新版本: ${latest}"

    while IFS= read -r tag; do
        [[ "$tag" == "$latest" ]] && continue
        local full_image="${image_prefix}:${tag}"
        if $DRY_RUN; then
            log_warn "[预览] 将删除: ${full_image}"
        else
            log_delete "删除: ${full_image}"
            $DOCKER rmi "${full_image}" || log_warn "删除失败（可能正在使用）: ${full_image}"
        fi
    done <<< "$tags"
}

# ── 主流程 ───────────────────────────────────────────────────
echo ""
if $DRY_RUN; then
    echo -e "${YELLOW}【预览模式】不会实际删除任何镜像。确认后请加 --delete 参数执行。${NC}"
else
    echo -e "${RED}【删除模式】将实际删除历史镜像！${NC}"
fi

cleanup_service "nb67-web"
cleanup_service "nb67-bff"
cleanup_service "nb-parse-connect"

# 清理 dangling 镜像（无标签的中间层镜像）
log_step "清理 dangling 镜像"
dangling=$($DOCKER images -f "dangling=true" -q 2>/dev/null || true)
if [[ -z "$dangling" ]]; then
    log_info "无 dangling 镜像"
else
    count=$(echo "$dangling" | wc -l | tr -d ' ')
    if $DRY_RUN; then
        log_warn "[预览] 将清理 ${count} 个 dangling 镜像"
    else
        $DOCKER image prune -f
        log_info "已清理 dangling 镜像"
    fi
fi

log_step "完成"
if $DRY_RUN; then
    echo -e "${YELLOW}预览结束。确认无误后执行: ./cleanup-images.sh --delete${NC}"
else
    echo -e "${GREEN}历史镜像清理完成！${NC}"
    echo ""
    echo "当前镜像列表:"
    $DOCKER images "${REGISTRY}/"* 2>/dev/null || true
fi
