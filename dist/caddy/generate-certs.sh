#!/usr/bin/env bash
# =============================================================
# generate-certs.sh — 为 MACDA Connector Caddy 生成自签名证书
#
# 用法：
#   ./generate-certs.sh                       # 使用默认数据目录 /data/MACDA2
#   DATA_DIR=/opt/macda ./generate-certs.sh   # 自定义数据目录
#   ./generate-certs.sh --force               # 强制覆盖已有证书
#
# 证书信息：
#   - 算法：RSA 4096
#   - 有效期：10 年（3650 天）
#   - SAN：IP:192.168.66.12、DNS:test.macda.org、localhost
#
# 替换证书：
#   将自定义的 server.crt 和 server.key 放到 ${DATA_DIR}/caddy/certs/
#   然后执行 docker restart macda-caddy 即可生效
# =============================================================

set -euo pipefail

DATA_DIR="${DATA_DIR:-/data/MACDA2}"
CERT_DIR="${DATA_DIR}/caddy/certs"
FORCE=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --force) FORCE=true; shift ;;
        *) echo "用法: $0 [--force]"; exit 1 ;;
    esac
done

log_info()    { echo -e "\033[0;32m[INFO]\033[0m  $*"; }
log_success() { echo -e "\033[0;32m[✓]\033[0m $*"; }
log_warn()    { echo -e "\033[0;33m[WARN]\033[0m  $*"; }

# ── 检查 openssl ──────────────────────────────────────────────
if ! command -v openssl &>/dev/null; then
    echo "[ERROR] 未找到 openssl，请先安装：apt install openssl"
    exit 1
fi

# ── 创建证书目录 ──────────────────────────────────────────────
mkdir -p "${CERT_DIR}"

if [[ -f "${CERT_DIR}/server.crt" ]] && [[ "${FORCE}" == "false" ]]; then
    log_warn "证书已存在: ${CERT_DIR}/server.crt"
    log_warn "如需重新生成，请使用 --force 参数"
    exit 0
fi

# ── 生成 OpenSSL 配置（含 SAN） ───────────────────────────────
TMP_CNF=$(mktemp /tmp/macda-cert-XXXXXX.cnf)
trap 'rm -f "${TMP_CNF}"' EXIT

cat > "${TMP_CNF}" << 'EOF'
[req]
default_bits       = 4096
prompt             = no
default_md         = sha256
distinguished_name = dn
x509_extensions    = v3_req

[dn]
CN = macda-connector

[v3_req]
subjectAltName = @alt_names
keyUsage       = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth

[alt_names]
IP.1  = 192.168.66.12
IP.2  = 127.0.0.1
DNS.1 = test.macda.org
DNS.2 = localhost
EOF

# ── 生成自签名证书 ────────────────────────────────────────────
log_info "正在生成 RSA 4096 自签名证书..."

openssl req -x509 \
    -newkey rsa:4096 \
    -keyout "${CERT_DIR}/server.key" \
    -out    "${CERT_DIR}/server.crt" \
    -days   3650 \
    -nodes \
    -config "${TMP_CNF}" 2>/dev/null

# ── 设置文件权限 ──────────────────────────────────────────────
chmod 644 "${CERT_DIR}/server.crt"
chmod 600 "${CERT_DIR}/server.key"

log_success "证书已生成到: ${CERT_DIR}"
log_success "  server.crt（有效期 10 年，SAN: 192.168.66.12, test.macda.org）"
log_success "  server.key（请妥善保管，不要外传）"
echo ""
log_info "替换证书：将自定义 server.crt/server.key 放到 ${CERT_DIR}/"
log_info "然后执行：docker restart macda-caddy"
