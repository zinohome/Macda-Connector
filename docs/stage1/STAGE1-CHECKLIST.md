# 阶段1交付清单 ✅ 完成

## 文件结构验证

### ✅ 核心源代码 (connect/cmd/connect-nb67/)
```
✅ main.go                 (47行)  启动器+处理器注册
✅ nb67_processor.go       (250行) Processor接口实现+解析逻辑
✅ nb67.go                 (1936行) Kaitai生成的二进制解析器
✅ go.mod                  (15行)  依赖声明
```

### ✅ 配置与编解码 (connect/codec/ & connect/config/)
```
✅ NB67.ksy                 Kaitai Struct协议定义（含新增5字段）
✅ nb67.go                  Kaitai编译输出备份
✅ nb67-connect.yaml        完整Redpanda Connect配置（180+字段映射）
```

### ✅ 容器化部署
```
✅ connect/Dockerfile.connect        多阶段构建文件（builder+runtime）
✅ deploy/docker-compose.stage1.yml  完整部署编排配置
✅ deploy/README-STAGE1.md           详细部署指南
```

---

## 准备清单（用户侧执行）

### 第一步：本地验证
- [ ] 检查所有源代码文件存在
- [ ] 检查nb67-connect.yaml包含180+字段映射
- [ ] 检查新增5个字段在ParsedOutput结构体中

### 第二步：文件打包
```bash
# 打包所有必需文件到服务器
tar -czf macda-phase1.tar.gz \
  connect/cmd/connect-nb67/ \
  connect/codec/ \
  connect/config/ \
  connect/Dockerfile.connect \
  deploy/docker-compose.stage1.yml

# 上传到192.168.32.17
scp macda-phase1.tar.gz user@192.168.32.17:/opt/
```

### 第三步：服务器端构建
```bash
# SSH到192.168.32.17
ssh user@192.168.32.17

# 解包
cd /opt
tar -xzf macda-phase1.tar.gz

# 构建镜像
cd Macda-Connector
docker-compose -f deploy/docker-compose.stage1.yml build connect-nb67

# 应该看到：
# Building connect-nb67
# [+] Building 45.2s (10/10) FINISHED docker:default
```

### 第四步：启动服务
```bash
# 启动所有服务
docker-compose -f deploy/docker-compose.stage1.yml up -d

# 验证运行状态
docker-compose -f deploy/docker-compose.stage1.yml ps

# 应该显示 connect-nb67 状态为 UP (healthy)
```

### 第五步：验证输出
```bash
# 查看日志
docker-compose -f deploy/docker-compose.stage1.yml logs -f connect-nb67

# 应该看到类似：
# [NB67] Processed 100 frames: TrainNo=100 Carriage=3 CurStation=45
# [NB67] Heartbeat every 30s

# 消费输出topic
kafka-console-consumer --bootstrap-server 192.168.32.17:19092 \
  --topic signal-parsed \
  --max-messages 1

# 应该看到JSON格式的数据，包含：
# - timestamp, train_no, carriage_no, cur_station
# - route_info (新增字段组)
# - compressor_u11, environment等传感器数据
```

---

## 关键文件内容摘要

### 1️⃣ main.go - 启动器
```go
// 注册自定义处理器
service.RegisterProcessor(
    "nb67_parser",
    NewNB67Processor,
    newConfigSpec(),
)
// 启动Connect
service.Run(context.Background())
```

### 2️⃣ nb67_processor.go - 处理器核心逻辑
```go
type Processor struct {}

func (p *Processor) Process(ctx context.Context, msg *service.Message) service.MessageBatch {
    // 1. 读取二进制消息体
    bytes := msg.AsBytes()
    
    // 2. 创建Kaitai Reader
    reader := kaitai.NewReader(bytes)
    
    // 3. 调用Kaitai生成的解析器
    nb67 := ParseNB67(reader)
    
    // 4. 映射到ParsedOutput（180+字段）
    output := &ParsedOutput{
        HeaderCode01:      nb67.HeaderCode01,
        TrainNo:           nb67.TrainNo,
        // ... 180+字段映射 ...
        RouteInfo: RouteInfo{
            StartStation:      nb67.StartStation,     // 新增
            TerminalStation:   nb67.TerminalStation,  // 新增
            CurStation:        nb67.CurStation,
            NextStation:       nb67.NextStation,      // 新增
            ExhaustDamper:     nb67.DmpExuPos,        // 新增
        },
    }
    
    // 5. 序列化为JSON
    json := output.ToJSON()
    
    // 6. 返回结果
    return service.MessageBatch{msg.Copy().SetStructuredMut(json)}
}
```

### 3️⃣ nb67-connect.yaml - 配置示例片段
```yaml
input:
  kafka:
    addresses: ["192.168.32.17:19092"]
    topics: ["signal-in"]
    consumer_group: macda-phase1-parser

pipeline:
  processors:
    - nb67_parser:
        log_sample_every: 100
    
    - mapping: |
        root = .
        root.route_info = {
          "start_station": .start_station,
          "terminal_station": .terminal_station,
          "current_station": .cur_station,
          "next_station": .next_station,
          "exhaust_damper_position": .dmp_exu_pos
        }

output:
  kafka:
    addresses: ["192.168.32.17:19092"]
    topic: signal-parsed
    partitioner: round_robin
```

### 4️⃣ Dockerfile.connect - 多阶段构建
```dockerfile
# Stage 1: Builder
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY connect/cmd/connect-nb67 .
RUN CGO_ENABLED=0 GOOS=linux go build -o connect-nb67 ./main.go

# Stage 2: Runtime
FROM ubuntu:22.04
COPY --from=builder /app/connect-nb67 /usr/bin/
COPY connect/config/nb67-connect.yaml /etc/connect/
ENTRYPOINT ["/usr/bin/connect-nb67"]
```

### 5️⃣ NB67.ksy - 新增字段定义片段
```yaml
types:
  # ... 此前450字段 ...
  
  # 新增5个字段（offset 452-461）
  - id: dmp_exu_pos
    offset: 452
    size: 2
    type: u2
    doc: "废排风阀开度百分比"
  
  - id: start_station
    offset: 454  
    size: 2
    type: u2
    doc: "起始站ID"
  
  - id: terminal_station
    offset: 456
    size: 2
    type: u2
    doc: "终点站ID"
  
  - id: cur_station
    offset: 458
    size: 2
    type: u2
    doc: "当前站ID"
    
  - id: next_station
    offset: 460
    size: 2
    type: u2
    doc: "下一站ID"
```

---

## 输出数据格式示例

Connect处理后输出到`signal-parsed` topic的JSON示例：

```json
{
  "header_code_01": 44,
  "train_no": 100,
  "carriage_no": 3,
  "reserved_001": null,
  
  "route_info": {
    "start_station": 291,
    "terminal_station": 129,
    "current_station": 45,
    "next_station": 66,
    "exhaust_damper_position": 78
  },
  
  "timestamp": "2026-02-19T14:35:42.123Z",
  "device_id": "HVAC-00100-03",
  
  "environment": {
    "temp_cabin_u1": 22.5,
    "humidity_cabin_u1": 45.3,
    "pressure_u1": 101.325
  },
  
  "compressor_u11": {
    "frequency": 45.2,
    "current": 12.3,
    "power": 2500,
    "status": "running"
  },
  
  "fault_detection": {
    "low_pressure_u11": false,
    "high_pressure_u11": false,
    "temperature_warning_cabin": false
  },
  
  "alert_level": "OK",
  
  "metadata": {
    "parser_version": "nb67-v1",
    "parsed_at_ms": 1708347342123,
    "raw_frame_size": 462
  }
}
```

---

## 验收指标

在192.168.32.17服务器上启动后，需验证以下指标：

| 指标 | 验证方法 | 预期结果 |
|------|---------|--------|
| **镜像构建** | `docker images \| grep macda-connect` | 镜像存在且大小 ~150-200MB |
| **容器运行** | `docker ps \| grep connect-nb67` | 状态为 UP(healthy) |
| **消费输入** | 监控日志 `[NB67] Processed XXX frames` | 每100条输出一条日志 |
| **生产输出** | `kafka-console-consumer ... signal-parsed` | 收到JSON格式数据 |
| **字段完整** | 检查JSON中的route_info | 包含5个新增字段 |
| **吞吐量** | 计算 processed frames / 时间 | ≥1000 msg/s |
| **正确性** | 对比字段值与二进制 | ≥99% 准确率 |

---

## 故障速查表

| 问题 | 原因 | 解决方案 |
|------|------|--------|
| `docker-compose: command not found` | 未安装docker-compose v2 | 升级: `docker-compose --version` |
| `Cannot connect to Docker daemon` | Docker未启动 | 启动Docker或检查权限 |
| `failed to build image` | Go编译错误 | 查看日志: `docker-compose logs --build` |
| 容器启动但无日志输出 | 配置问题或Kafka无连接 | 检查环境变量和Kafka地址 |
| signal-parsed topic为空 | 处理器未启动或有异常 | 查看 [查看处理器日志](#查看处理器日志) |
| 解析错误 | 输入二进制格式不对 | 验证signal-in数据 |

---

## 下一步

当阶段1成功部署并验证后：

1. **性能基线记录** - 在监控工具中记录初始的4大指标
2. **阶段2规划** - 设计通用REST API接口
3. **故障告警系统** - 基于signal-parsed的异常检测
4. **长期存储** - 集成TimescaleDB
5. **前端适配** - 连接web-nb67.250513

详见：[docs/11-macda-refactor-execution-plan.md](../docs/11-macda-refactor-execution-plan.md)
