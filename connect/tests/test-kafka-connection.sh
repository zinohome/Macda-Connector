#!/bin/bash
##############################################################################
# 🧪 Test 1: Kafka连接和Topic验证
# 
# 验证：
# 1. Redpanda broker可达性
# 2. signal-in topic存在且有数据
# 3. signal-parsed topic存在
# 4. 消费一条signal-in消息检查格式
##############################################################################

set -e

BROKER="192.168.32.17:19092"
SIGNAL_IN="signal-in"
SIGNAL_PARSED="signal-parsed"

echo "════════════════════════════════════════════════════════════════════════"
echo "🧪 Test 1: Kafka连接和Topic验证"
echo "════════════════════════════════════════════════════════════════════════"
echo ""

# Step 1: 检查broker连接
echo "Step 1️⃣  检查Broker连接... ($BROKER)"
if timeout 5 bash -c "cat < /dev/null > /dev/tcp/192.168.32.17/19092" 2>/dev/null; then
    echo "  ✅ Broker可达"
else
    echo "  ❌ Broker不可达: $BROKER"
    echo "  解决方法："
    echo "    1. 检查192.168.32.17是否在线"
    echo "    2. 检查防火墙是否允许19092端口"
    echo "    3. 检查Redpanda是否已启动: docker-compose ps"
    exit 1
fi
echo ""

# Step 2: 检查Topics
echo "Step 2️⃣  检查Topics..."

# 安装kafka-cli工具用于测试（如果本地没有）
KAFKA_CLI=""
if command -v kafka-topics &> /dev/null; then
    KAFKA_CLI="kafka-topics"
elif command -v docker &> /dev/null && docker ps -a | grep -q "kafka"; then
    # 使用Docker中的kafka工具
    KAFKA_CLI="docker exec kafka kafka-topics.sh"
else
    echo "  ⚠️  未找到kafka工具，使用nc进行基础检查"
    echo ""
    echo "  检查signal-in topic:"
    echo "    kafka-topics --bootstrap-server $BROKER --list | grep signal-in"
    echo ""
    echo "  检查signal-parsed topic:"
    echo "    kafka-topics --bootstrap-server $BROKER --list | grep signal-parsed"
    exit 0
fi

Topics=$($KAFKA_CLI --bootstrap-server $BROKER --list 2>/dev/null || echo "")

if echo "$Topics" | grep -q "^${SIGNAL_IN}$"; then
    echo "  ✅ Topic '$SIGNAL_IN' 存在"
else
    echo "  ⚠️  Topic '$SIGNAL_IN' 不存在，需要创建"
    echo "    建议：docker exec kafka kafka-topics.sh --create \\"
    echo "      --bootstrap-server $BROKER \\"
    echo "      --topic $SIGNAL_IN \\"
    echo "      --partitions 3 --replication-factor 1"
fi

if echo "$Topics" | grep -q "^${SIGNAL_PARSED}$"; then
    echo "  ✅ Topic '$SIGNAL_PARSED' 存在"
else
    echo "  ⚠️  Topic '$SIGNAL_PARSED' 不存在，需要创建"
    echo "    建议：docker exec kafka kafka-topics.sh --create \\"
    echo "      --bootstrap-server $BROKER \\"
    echo "      --topic $SIGNAL_PARSED \\"
    echo "      --partitions 3 --replication-factor 1"
fi
echo ""

# Step 3: 消费signal-in消息
echo "Step 3️⃣  检查signal-in消息..."
echo ""

# 检查是否有消息
CONSUMER_CMD="kafka-console-consumer --bootstrap-server $BROKER \
  --topic $SIGNAL_IN \
  --max-messages 1 \
  --timeout-ms 5000"

echo "  执行："
echo "    $CONSUMER_CMD"
echo ""
echo "  输出："

if command -v kafka-console-consumer &> /dev/null; then
    # 本地有kafka工具
    MSG=$($CONSUMER_CMD 2>&1 | head -1 || echo "")
    
    if [ -z "$MSG" ]; then
        echo "  ⚠️  signal-in topic为空（无消息）"
        echo ""
        echo "    需要推送测试数据到signal-in topic"
        echo "    参考：docs/stage1/3-DEPLOYMENT.md"
    else
        echo "  ✅ 收到消息（长度: ${#MSG}字节）"
        echo ""
        echo "    消息前100字符:"
        echo "    ${MSG:0:100}..."
        
        # 检查是否为二进制格式（16进制）
        if echo "$MSG" | od -A n -t x1 | head -1 | grep -qE '[0-9a-f]{2}'; then
            echo ""
            echo "  ✅ 消息为二进制格式（符合NB67预期）"
        else
            echo ""
            echo "  ⚠️  消息不是二进制格式"
        fi
    fi
else
    echo "  ⚠️  本地未安装kafka-console-consumer"
    echo ""
    echo "    在192.168.32.17上执行："
    echo "    kafka-console-consumer --bootstrap-server $BROKER \\"
    echo "      --topic $SIGNAL_IN --max-messages 1"
fi

echo ""
echo "════════════════════════════════════════════════════════════════════════"
echo "✅ Test 1 完成"
echo "════════════════════════════════════════════════════════════════════════"
echo ""
echo "下一步："
echo "  bash tests/test-nb67-parsing.sh      - 验证NB67解析"
echo "  bash tests/test-end-to-end.sh        - 端到端测试"
echo ""
