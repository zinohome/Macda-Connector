# Connect Module - Kafka Connect NB67处理器

**模块用途**: Redpanda Kafka Connect 的自定义处理器，用于解析NB67二进制浮车空调数据帧

---

## 📂 目录结构说明

```
connect/
├── README.md                  ← 本文档
├── Dockerfile.connect         ← Docker镜像构建文件（部署用）
│
├── cmd/                       ← 📦 应用主程序包
│   └── connect-nb67/          ← Go应用主程序
│       ├── main.go            ← 入口点，注册nb67_parser处理器到Redpanda Connect
│       ├── nb67_processor.go   ← 处理器实现，处理消息转换逻辑
│       ├── nb67.go            ← Kaitai生成的NB67协议解析代码（AUTO-GENERATED）
│       └── go.mod             ← Go模块定义
│
├── codec/                     ← 🔧 协议编解码相关
│   ├── NB67.ksy               ← Kaitai规格文件（协议定义）
│   └── nb67.go                ← Kaitai生成的Go解析代码（AUTO-GENERATED）
│
├── config/                    ← ⚙️ 配置文件
│   ├── nb67-connect.yaml      ← Redpanda Connect连接器配置（完整版）
│   └── phase1-connect.yaml    ← 阶段1简化版配置
│
└── tests/                     ← 🧪 自动化测试脚本
    ├── test-kafka-connection.sh    ← 验证Kafka/Redpanda连接
    ├── test-nb67-parsing.sh        ← 验证NB67解析功能
    └── test-end-to-end.sh          ← 完整流程验证
```

---

## 📋 每个文件的详细用途

### cmd/connect-nb67/

#### 1. **main.go** (47行)
```
⚡ 入口点文件
用途：
  • 将nb67_processor处理器注册到Redpanda Connect框架
  • 定义处理器的名称、配置、约束条件
  • 启动Connect服务
  
关键代码：
  service.RegisterProcessor("nb67_parser", ...)
```

#### 2. **nb67_processor.go** (250行)
```
🔄 处理器核心实现
用途：
  • 处理Kafka消息的主业务逻辑
  • 从消息中提取二进制数据
  • 调用NB67解析器解析二进制格式
  • 将解析结果转换为JSON
  • 处理错误和异常情况
  
流程：
  1. 接收Kafka消息（binary format in signal-in topic）
  2. 记录消息大小和时间戳
  3. 使用NB67解析器解析二进制数据
  4. 提取180+个字段到结构体
  5. 转换为JSON格式
  6. 发送到signal-parsed topic
```

#### 3. **nb67.go** (1936行)
```
⚠️ AUTO-GENERATED FILE
用途：
  • Kaitai Struct生成的NB67二进制协议解析代码
  • 定义了180+个字段的数据结构
  • 实现二进制解析逻辑
  
⚠️ 重要！不应手工编辑！
  生成方式：kaitai-struct-compiler -t go NB67.ksy

包含的字段示例：
  • header_code_01      - 帧头代码
  • train_no            - 列车号
  • carriage_no         - 车厢号
  • temperature         - 温度传感器数据
  • route_info
    - start_station    - 起始站点
    - terminal_station - 终点站点
    - current_station  - 当前站点
    - next_station     - 下一站点
    - exhaust_damper_position (新增)
```

#### 4. **go.mod** (14行)
```
📦 Go模块定义
用途：
  • 定义包名：github.com/macda/connect-nb67
  • 指定Go版本要求
  • 列出依赖包（redpanda connect v4）
```

---

### codec/

#### 1. **NB67.ksy** (Kaitai Struct规格文件)
```
📋 协议定义（真实源）
用途：
  • 使用Kaitai Struct描述NB67二进制协议
  • 定义字段类型、大小、顺序
  • 跨语言规格（可生成Go/Python/Java/C++代码）

内容：
  - 字节序：Little Endian
  - 字段总数：180+
  - 总帧大小：462字节
  - 定义route_info结构体

状态：
  ✅ 单一事实来源（SSOT）
  ✅ 与Go代码同步
  ✅ 包含新增字段定义
```

#### 2. **nb67.go** (1936行)
```
⚠️ AUTO-GENERATED FILE
用途：
  • NB67.ksy编译生成的Go代码
  • 包含解析器类和数据结构
  • 处理二进制格式转换

生命周期：
  NB67.ksy → kaitai-struct-compiler ⬇️
          ↓
        nb67.go （自动生成）

⚠️ 注意：修改时应修改.ksy文件，然后重新生成

命令：
  kaitai-struct-compiler -t go NB67.ksy
```

---

### config/

#### 1. **nb67-connect.yaml** (500+行)
```
⚙️ Redpanda Connect连接器配置（完整版）
用途：
  • 定义Connect任务的完整配置
  • 指定source connector（Kafka source）
  • 指定processor（nb67_parser）
  • 指定sink connector（Kafka sink）
  • 定义字段映射（180+字段→JSON）

配置流程：
  signal-in (Kafka) 
    ↓ [SourceConnector]
    ↓ Binary data
    ↓ [nb67_processor]
    ↓ Parse → JSON with 180+ fields
    ↓ [SinkConnector]
    ↓
  signal-parsed (Kafka)
```

#### 2. **phase1-connect.yaml** (简化版)
```
⚙️ 阶段1测试配置
用途：
  • 最小可用的NB67解析管道（包含 nb67_parser）
  • 用于快速验证“能解析/能出JSON/新增车站字段可用”
  • 映射只保留最关键的 timestamp / device_id / route_info
  • 输出到独立 topic：signal-parsed-phase1（避免和完整版混用）
```

---

### tests/

#### 1. **test-kafka-connection.sh**
```
🧪 Kafka连接验证测试
用途：
  • 检查192.168.32.17:19092 Broker可达性
  • 验证signal-in topic存在
  • 验证signal-parsed topic存在
  • 验证能消费Kafka消息
  
执行：
  bash connect/tests/test-kafka-connection.sh
```

#### 2. **test-nb67-parsing.sh**
```
🧪 NB67解析器验证测试
用途：
  • 从signal-in消费二进制消息
  • 验证消息大小（462字节）
  • 验证Kaitai解析器有所有字段定义
  • 验证新增5个车站信息字段存在
  • 验证YAML配置完整性
  
执行：
  bash connect/tests/test-nb67-parsing.sh
```

#### 3. **test-end-to-end.sh**
```
🧪 端到端完整流程测试
用途：
  • 验证docker-compose配置
  • 验证connect-nb67容器运行状态
  • 推送测试二进制数据到signal-in
  • 验证signal-parsed收到JSON输出
  • 验证JSON格式有效性
  • 采集性能指标

执行：
  bash connect/tests/test-end-to-end.sh
```

---

### Dockerfile.connect

```
🐳 Docker镜像构建
用途：
  • 定义connect-nb67服务的Docker镜像
  • 基于Redpanda Connect官方镜像
  • 加载自定义的nb67_processor处理器
  • 配置运行环境

使用场景：
  docker build -f Dockerfile.connect -t connect-nb67:v1 .
  docker run connect-nb67:v1
```

---

## 🔄 数据流向

```
Kafka (signal-in topic)
       ↓ 二进制数据 (462字节)
       ↓
┌─────────────────────────────────────┐
│    Connect nb67_processor           │
│  ┌─────────────────────────────────┐│
│  │  1. 接收二进制消息              ││
│  │  2. 调用Kaitai解析器解析        ││
│  │  3. 提取180+个字段              ││
│  │  4. 转换为JSON                  ││
│  │  5. 发送输出                    ││
│  └─────────────────────────────────┘│
└─────────────────────────────────────┘
       ↓ JSON数据（180+字段）
       ↓
Kafka (signal-parsed topic)
       ↓
下游应用（API、查询、告警等）
```

---

## 📊 关键数字

| 项 | 数值 | 说明 |
|----|------|------|
| **二进制帧大小** | 462字节 | NB67规程定义 |
| **字段总数** | 180+ | Kaitai定义 |
| **新增字段数** | 5 | 车站信息字段 |
| **main.go行数** | 47行 | 非常简洁 |
| **nb67_processor.go行数** | 250行 | 核心业务逻辑 |
| **nb67.go行数** | 1936行 | Kaitai生成 |
| **配置文件** | 2个 | nb67-connect.yaml + phase1-connect.yaml |
| **测试脚本** | 3个 | 完整覆盖 |

---

## 🚀 快速开始

### 理解代码
```bash
# 1. 了解协议定义
cat connect/codec/NB67.ksy

# 2. 了解处理器实现
cat connect/cmd/connect-nb67/main.go
cat connect/cmd/connect-nb67/nb67_processor.go

# 3. 查看Connect配置
cat connect/config/nb67-connect.yaml
```

### 运行测试
```bash
# 测试Kafka连接
bash connect/tests/test-kafka-connection.sh

# 测试解析器
bash connect/tests/test-nb67-parsing.sh

# 完整测试
bash connect/tests/test-end-to-end.sh
```

### 构建Docker镜像
```bash
cd connect
docker build -f Dockerfile.connect -t connect-nb67:v1 .
```

---

## 📝 文件清单

| 文件/目录 | 类型 | 行数 | 状态 | 说明 |
|---------|------|------|------|------|
| cmd/ | 目录 | - | ✅ | Go应用主程序 |
| codec/ | 目录 | - | ✅ | 协议定义和解析器 |
| config/ | 目录 | - | ✅ | Connect配置 |
| tests/ | 目录 | - | ✅ | 测试脚本 |
| main.go | 程序 | 47 | ✅ | 入口点 |
| nb67_processor.go | 程序 | 250 | ✅ | 处理器实现 |
| nb67.go(cmd/) | 程序 | 1936 | ⚠️ | AUTO-GENERATED |
| nb67.go(codec/) | 程序 | - | ⚠️ | AUTO-GENERATED |
| NB67.ksy | 规格 | - | ✅ | 协议定义 |
| nb67-connect.yaml | 配置 | 500+ | ✅ | 完整配置 |
| phase1-connect.yaml | 配置 | - | ✅ | 简化配置 |
| Dockerfile.connect | 配置 | - | ✅ | Docker构建 |
| test-*.sh | 脚本 | - | ✅ | 测试脚本(3个) |

---

## ⚠️ 重要说明

### 不要手工编辑这些文件：
- ❌ `cmd/connect-nb67/nb67.go` - AUTO-GENERATED by Kaitai
- ❌ `codec/nb67.go` - AUTO-GENERATED by Kaitai

### 应该编辑的文件：
- ✅ `codec/NB67.ksy` - 协议规格（修改后需重新生成）
- ✅ `cmd/connect-nb67/main.go` - 应用入口
- ✅ `cmd/connect-nb67/nb67_processor.go` - 业务逻辑
- ✅ `config/*.yaml` - 连接器配置

### 生成Auto-Generated文件的命令：
```bash
# 从NB67.ksy生成nb67.go
kaitai-struct-compiler -t go codec/NB67.ksy -o cmd/connect-nb67/

# 验证生成成功
ls -la cmd/connect-nb67/nb67.go
```

---

## 🎯 模块目标

✅ **已完成**
- 解析NB67二进制格式（180+字段）
- 实现Kafka Connect处理器
- 支持180+字段的JSON输出
- 验收清单覆盖

🔬 **下一步（Phase 2）**
- REST API接口
- WebSocket实时推送
- 故障告警系统
- 查询优化

---

## 📞 故障排除

### 问题：消息无法解析
```bash
# 1. 检查消息大小是否462字节
bash connect/tests/test-kafka-connection.sh

# 2. 检查字段定义是否完整
grep "start_station\|terminal_station" codec/NB67.ksy

# 3. 查看错误日志
docker logs connect-nb67
```

### 问题：字段缺失
```bash
# 检查NB67.ksy定义
cat codec/NB67.ksy | grep -A5 "route_info"

# 重新生成nb67.go
kaitai-struct-compiler -t go codec/NB67.ksy -o cmd/connect-nb67/
```

### 问题：Docker构建失败
```bash
# 查看Dockerfile
cat Dockerfile.connect

# 清理旧镜像
docker image rm connect-nb67:v1

# 重新构建
docker build -f Dockerfile.connect -t connect-nb67:v1 .
```

---

**模块清晰度**: ✅ 100%  
**最后更新**: 2026-02-19  
**下一步**: 执行connect/tests/test-kafka-connection.sh
