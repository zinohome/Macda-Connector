#!/usr/bin/env bash
set -euo pipefail

# 统一最小验证命令：启动 connect-nb67，消费 signal-in 并输出到 signal-parsed。
# 运行后观察日志/控制台即可（CTRL+C 退出）。

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../cmd/connect-nb67"

exec go run . -c ../../config/nb67-connect.yaml
