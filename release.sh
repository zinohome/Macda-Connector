#!/usr/bin/env bash
# =============================================================
# release.sh — MACDA Connector 统一发布脚本
#
# 用法:
#   ./release.sh prepare v2.5.0   # Phase 1: 更新版本号 + 生成 CHANGELOG + commit
#   ./release.sh publish           # Phase 2: 构建推送镜像 + 打 tag + push
# =============================================================
set -euo pipefail

REGISTRY="harbor.naivehero.top:8443/macda2"
RELEASE_STATE_FILE=".release-version"

# ── 颜色输出 ───────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
CYAN='\033[1;36m'; NC='\033[0m'
log_step()  { echo -e "\n${CYAN}══════ $* ══════${NC}"; }
log_info()  { echo -e "${GREEN}[OK]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[!!]${NC}  $*"; }
log_error() { echo -e "${RED}[ERR]${NC} $*" >&2; exit 1; }

# ── prepare ────────────────────────────────────────────────────
cmd_prepare() {
    local version="${1:-}"
    [[ -z "$version" ]] && log_error "用法: ./release.sh prepare <version>  例: ./release.sh prepare v2.5.0"
    [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || log_error "版本号格式错误: $version  需为 vX.Y.Z"

    log_step "Phase 1: Prepare $version"

    # 检查 git 工作区干净
    [[ -z "$(git status --porcelain)" ]] || log_error "工作区有未提交变更，请先 commit 或 stash"

    # 检查当前分支
    local branch; branch=$(git rev-parse --abbrev-ref HEAD)
    [[ "$branch" == "main" ]] || log_warn "当前分支: $branch（不是 main，请确认）"

    log_step "更新版本号"

    # 版本匹配模式：v[0-9][0-9.a-z-]* 可精确匹配 v2.5.0 / v2.2.0-full，
    # 不会误匹配 markdown 反引号等后续字符
    local vpat="v[0-9][0-9.a-z-]*"

    # baseEnv/docker-compose-Dev.yml: nb67-web, nb67-bff, nb-parse-connect
    sed -i "s|\(nb67-web:\)${vpat}|\1${version}|g"          baseEnv/docker-compose-Dev.yml
    sed -i "s|\(nb67-bff:\)${vpat}|\1${version}|g"          baseEnv/docker-compose-Dev.yml
    sed -i "s|\(nb-parse-connect:\)${vpat}|\1${version}|g"  baseEnv/docker-compose-Dev.yml

    # baseEnv/docker-compose-Prod.yml: nb67-web, nb67-bff
    sed -i "s|\(nb67-web:\)${vpat}|\1${version}|g"          baseEnv/docker-compose-Prod.yml
    sed -i "s|\(nb67-bff:\)${vpat}|\1${version}|g"          baseEnv/docker-compose-Prod.yml

    # dist/docker-compose-Web.yml: nb67-web, nb67-bff
    sed -i "s|\(nb67-web:\)${vpat}|\1${version}|g"          dist/docker-compose-Web.yml
    sed -i "s|\(nb67-bff:\)${vpat}|\1${version}|g"          dist/docker-compose-Web.yml

    # dist/docker-compose-Data.yml: nb-parse-connect（v2.2.0-full 后缀也一并替换）
    sed -i "s|\(nb-parse-connect:\)${vpat}|\1${version}|g"  dist/docker-compose-Data.yml

    # dist/README.md: 版本头 + 更新日期 + 镜像版本表
    local today; today=$(date +%Y-%m-%d)
    sed -i "s|\*\*版本\*\*：${vpat}|**版本**：${version}|g"   dist/README.md
    sed -i "s|\*\*更新\*\*：[0-9-]*|**更新**：${today}|g"     dist/README.md
    sed -i "s|\(nb67-web:\)${vpat}|\1${version}|g"            dist/README.md
    sed -i "s|\(nb67-bff:\)${vpat}|\1${version}|g"            dist/README.md
    sed -i "s|\(nb-parse-connect:\)${vpat}|\1${version}|g"    dist/README.md

    log_info "版本号更新完成"

    log_step "生成 CHANGELOG"
    _gen_changelog "$version"

    # 保存版本号供 publish 阶段读取
    echo "$version" > "$RELEASE_STATE_FILE"

    log_step "提交变更"
    git add \
        baseEnv/docker-compose-Dev.yml \
        baseEnv/docker-compose-Prod.yml \
        dist/docker-compose-Web.yml \
        dist/docker-compose-Data.yml \
        dist/README.md \
        CHANGELOG.md \
        "$RELEASE_STATE_FILE"
    git commit -m "chore(release): prepare ${version}"

    log_info "Prepare 完成！确认无误后执行: ./release.sh publish"
}

# ── 生成 CHANGELOG ─────────────────────────────────────────────
_gen_changelog() {
    local version="$1"
    local today; today=$(date +%Y-%m-%d)
    local last_tag; last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    local range="${last_tag:+${last_tag}..}HEAD"

    local feats fixes chores
    feats=$(git log  $range --pretty=format:"- %s" --grep="^feat"  2>/dev/null || true)
    fixes=$(git log  $range --pretty=format:"- %s" --grep="^fix"   2>/dev/null || true)
    chores=$(git log $range --pretty=format:"- %s" --grep="^chore" 2>/dev/null || true)

    local entry
    entry="## ${version} (${today})\n"
    [[ -n "$feats"  ]] && entry+="\n### Features\n${feats}\n"
    [[ -n "$fixes"  ]] && entry+="\n### Bug Fixes\n${fixes}\n"
    [[ -n "$chores" ]] && entry+="\n### Chores\n${chores}\n"

    # 插入文件头（第一个 ## 之前）
    local tmp; tmp=$(mktemp)
    awk -v block="${entry}" '
        /^## / && !inserted { printf "%s\n", block; inserted=1 }
        { print }
        END { if (!inserted) printf "\n%s\n", block }
    ' CHANGELOG.md > "$tmp"
    mv "$tmp" CHANGELOG.md

    log_info "CHANGELOG 已更新"
}

# ── publish ────────────────────────────────────────────────────
cmd_publish() {
    [[ -f "$RELEASE_STATE_FILE" ]] || log_error "未找到 .release-version，请先执行 ./release.sh prepare <version>"
    local version; version=$(cat "$RELEASE_STATE_FILE")

    log_step "Phase 2: Publish $version"

    local web_image="${REGISTRY}/nb67-web:${version}"
    local bff_image="${REGISTRY}/nb67-bff:${version}"
    local connect_image="${REGISTRY}/nb-parse-connect:${version}"

    # ── 构建 nb67-web ──────────────────────────────────────────
    log_step "构建 nb67-web"
    docker build --platform linux/amd64 \
        --tag "$web_image" \
        --file web-nb67-web/Dockerfile \
        web-nb67-web
    log_info "nb67-web 构建完成"

    # ── 构建 nb67-bff ──────────────────────────────────────────
    log_step "构建 nb67-bff"
    docker build --platform linux/amd64 \
        --tag "$bff_image" \
        --file web-nb67-bff/Dockerfile \
        web-nb67-bff
    log_info "nb67-bff 构建完成"

    # ── 构建 nb-parse-connect ──────────────────────────────────
    log_step "构建 nb-parse-connect"
    docker build --platform linux/amd64 \
        --tag "$connect_image" \
        --file connect/Dockerfile.connect \
        connect
    log_info "nb-parse-connect 构建完成"

    # ── 推送镜像 ───────────────────────────────────────────────
    log_step "推送镜像"
    docker push "$web_image"     && log_info "推送: $web_image"
    docker push "$bff_image"     && log_info "推送: $bff_image"
    docker push "$connect_image" && log_info "推送: $connect_image"

    # ── git tag + push ─────────────────────────────────────────
    log_step "打 Tag 并推送"
    git tag "$version"
    git push origin main
    git push origin "$version"
    log_info "Tag $version 已推送"

    # ── 清理状态文件 ───────────────────────────────────────────
    rm -f "$RELEASE_STATE_FILE"

    log_step "发布摘要"
    echo ""
    echo "  版本:    $version"
    echo "  镜像:"
    echo "    $web_image"
    echo "    $bff_image"
    echo "    $connect_image"
    echo "  Tag:     $version"
    echo ""
    log_info "发布完成！"
}

# ── 入口 ───────────────────────────────────────────────────────
case "${1:-}" in
    prepare) cmd_prepare "${2:-}" ;;
    publish) cmd_publish ;;
    *)
        echo "用法:"
        echo "  ./release.sh prepare <version>   更新版本号 + 生成 CHANGELOG + commit"
        echo "  ./release.sh publish              构建推送镜像 + 打 tag + push"
        exit 1
        ;;
esac
