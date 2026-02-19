# 简化架构方案:去除 Python 转换服务

## 问题分析

您当前使用 Python来解析二进制数据,然后发到 Redpanda,感觉架构太复杂。

### 您的数据格式特点

基于 NB67.ksy 分析:
- **字节序**: Big-endian (be)
- **数据类型**: 
  - `u1` (unsigned 1 byte)
  - `u2` (unsigned 2 bytes)  
  - `u4` (unsigned 4 bytes)
  - `s2` (signed 2 bytes)
  - `b1le` (1-bit little-endian flag)
- **字段总数**: ~180+ 字段
- **格式特点**: 固定长度,结构简单,字段多

### 当前架构的问题

```
二进制文件 → Python转换服务 → Redpanda → Redpanda Connect → TimescaleDB
             (额外一层的复杂度)
```

## ✅ 推荐方案:Redpanda Connect Go 插件

### 架构简化

```
二进制文件 → Redpanda(原始二进制) → Redpanda Connect(Go插件解析) → TimescaleDB
            直接读取                   一站式处理
```

### 方案对比

| 特性 | Python方案 | Go插件方案 ⭐ |
|------|-----------|------------|
| 组件数量 | 4个 | 3个 (-25%) |
| 额外服务 | Python转换服务 | 无 |
| 性能 | 中等 | 高 (Go编译) |
| 维护成本 | 高 (两套代码) | 低 (一套配置) |
| 部署复杂度 | 高 | 低 |
| 代码量 | Python + YAML | Go + YAML |

### 实施方案

#### 步骤 1: 用 Kaitai Struct 编译器生成 Go 代码

```bash
# 安装 Kaitai Struct 编译器
brew install kaitai-struct-compiler  # macOS
# 或
apt-get install kaitai-struct-compiler  # Linux

# 生成 Go 解析代码
kaitai-struct-compiler --target go --outdir ./kaitai_gen codec/NB67.ksy
```

这会生成一个 `nb67.go` 文件,包含完整的解析逻辑。

#### 步骤 2: 创建 Redpanda Connect Go 插件

```go
// plugins/nb67_parser.go
package main

import (
    "context"
    "bytes"
    
    "github.com/redpanda-data/benthos/v4/public/service"
    "github.com/kaitai-io/kaitai_struct_go_runtime/kaitai"
    
    // 导入生成的 Kaitai 代码
    "./kaitai_gen"  // nb67.go
)

func init() {
    // 注册自定义 processor
    service.RegisterProcessor(
        "nb67_parser",
        nb67ParseProcessorConfig(),
        nb67ParseProcessorConstructor,
    )
}

func nb67ParseProcessorConfig() *service.ConfigSpec {
    return service.NewConfigSpec().
        Summary("Parses NB67 binary format using Kaitai Struct").
        Field(service.NewBoolField("debug").
            Description("Enable debug logging").
            Default(false))
}

func nb67ParseProcessorConstructor(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
    debug, _ := conf.FieldBool("debug")
    return &nb67ParseProcessor{
        logger: mgr.Logger(),
        debug:  debug,
    }, nil
}

type nb67ParseProcessor struct {
    logger *service.Logger
    debug  bool
}

func (p *nb67ParseProcessor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error) {
    // 获取原始二进制数据
    data, err := msg.AsBytes()
    if err != nil {
        return nil, err
    }
    
    // 使用 Kaitai 生成的代码解析
    stream := kaitai.NewStream(bytes.NewReader(data))
    nb67Data := kaitai_gen.NewNb67()
    err = nb67Data.Read(stream, nil, nb67Data)
    if err != nil {
        p.logger.Errorf("Failed to parse NB67 data: %v", err)
        return nil, err
    }
    
    // 转换为 JSON 结构
    result := map[string]interface{}{
        // 消息头
        "msg_header_code01": nb67Data.MsgHeaderCode01,
        "msg_header_code02": nb67Data.MsgHeaderCode02,
        "msg_length":        nb67Data.MsgLength,
        "msg_src_dvc_no":    nb67Data.MsgSrcDvcNo,
        "msg_host_dvc_no":   nb67Data.MsgHostDvcNo,
        "msg_type":          nb67Data.MsgType,
        "msg_frame_no":      nb67Data.MsgFrameNo,
        "msg_line_no":       nb67Data.MsgLineNo,
        "msg_train_type":    nb67Data.MsgTrainType,
        "msg_train_no":      nb67Data.MsgTrainNo,
        "msg_carriage_no":   nb67Data.MsgCarriageNo,
        
        // 时间戳
        "timestamp": map[string]interface{}{
            "year":   nb67Data.MsgSrcDvcYear,
            "month":  nb67Data.MsgSrcDvcMonth,
            "day":    nb67Data.MsgSrcDvcDay,
            "hour":   nb67Data.MsgSrcDvcHour,
            "minute": nb67Data.MsgSrcDvcMinute,
            "second": nb67Data.MsgSrcDvcSecond,
        },
        
        // 设备信息
        "device": map[string]interface{}{
            "flag":        nb67Data.DvcFlag,
            "train_no":    nb67Data.DvcTrainNo,
            "carriage_no": nb67Data.DvcCarriageNo,
        },
        
        // 传感器数据
        "sensors": map[string]interface{}{
            "fas_sys":     nb67Data.FasSys,
            "ras_sys":     nb67Data.RasSys,
            "tic":         nb67Data.Tic,
            "load":        nb67Data.Load,
            "tveh_1":      nb67Data.Tveh1,
            "humidity_1":  nb67Data.Humdity1,
            "tveh_2":      nb67Data.Tveh2,
            "humidity_2":  nb67Data.Humdity2,
            
            // 空气质量 U1
            "aq_u1": map[string]interface{}{
                "temp":     nb67Data.AqTU1,
                "humidity": nb67Data.AqHU1,
                "co2":      nb67Data.AqCo2U1,
                "tvoc":     nb67Data.AqTvocU1,
                "formald":  nb67Data.AqFormaldU1,
                "pm25":     nb67Data.AqPm25U1,
                "pm10":     nb67Data.AqPm10U1,
            },
            
            // 压缩机 U11
            "compressor_u11": map[string]interface{}{
                "frequency": nb67Data.FCpU11,
                "current":   nb67Data.ICpU11,
                "voltage":   nb67Data.VCpU11,
                "power":     nb67Data.PCpU11,
                "suct_temp": nb67Data.SucktU11,
                "suct_pres": nb67Data.SuckpU11,
                "sp":        nb67Data.SpU11,
                "eev_pos":   nb67Data.EevposU11,
                "high_pres": nb67Data.HighpressU11,
            },
            
            // ... 其他100+字段
        },
        
        // 故障标志位
        "faults": map[string]interface{}{
            "cfbk_ef_u11":     nb67Data.CfbkEfU11,
            "cfbk_cf_u11":     nb67Data.CfbkCfU11,
            "cfbk_comp_u11":   nb67Data.CfbkCompU11,
            "bocflt_ef_u11":   nb67Data.BocfltEfU11,
            "blpflt_comp_u11": nb67Data.BlpfltCompU11,
            // ... 其他故障位
        },
        
        // 运行时间统计
        "operation_time": map[string]interface{}{
            "power":        nb67Data.Dwpower,
            "emerg_op_tm":  nb67Data.DwemergOpTm,
            "ef_op_tm_u11": nb67Data.DwefOpTmU11,
            "cf_op_tm_u11": nb67Data.DwcfOpTmU11,
            "cp_op_tm_u11": nb67Data.DwcpOpTmU11,
            // ... 其他运行时间
        },
    }
    
    if p.debug {
        p.logger.Debugf("Parsed NB67 data: %+v", result)
    }
    
    // 创建新消息
    newMsg := msg.Copy()
    newMsg.SetStructured(result)
    
    return service.MessageBatch{newMsg}, nil
}

func (p *nb67ParseProcessor) Close(ctx context.Context) error {
    return nil
}

func main() {
    service.RunCLI(context.Background())
}
```

#### 步骤 3: 编译 Redpanda Connect (包含插件)

```bash
# 创建自定义 main.go
cat > cmd/custom-connect/main.go <<'EOF'
package main

import (
    "context"
    
    "github.com/redpanda-data/benthos/v4/public/service"
    
    // 导入所有标准组件
    _ "github.com/redpanda-data/benthos/v4/public/components/all"
    
    // 导入自定义插件
    _ "github.com/yourusername/macda-connector/plugins"
)

func main() {
    service.RunCLI(context.Background())
}
EOF

# 编译
go build -o rpk-connect-nb67 cmd/custom-connect/main.go
```

#### 步骤 4: Redpanda Connect 配置

```yaml
# config.yaml
input:
  # 方案 A: 直接读取文件
  file:
    paths: ["/data/incoming/*.bin"]
    scanner:
      to_the_end: {}
    codec: all-bytes
    delete_on_finish: true

  # 或方案 B: 从 Kafka/Redpanda 读取原始二进制
  # kafka:
  #   addresses: ["redpanda:9092"]
  #   topics: ["hvac-raw-binary"]
  #   consumer_group: "nb67-parser"

pipeline:
  processors:
    # 使用自定义 NB67 解析器
    - nb67_parser:
        debug: false
    
    # 数据转换和特征提取
    - mapping: |
        # 时间戳转换
        root.time = "%d-%02d-%02d %02d:%02d:%02d".format(
          this.timestamp.year + 2000,
          this.timestamp.month,
          this.timestamp.day,
          this.timestamp.hour,
          this.timestamp.minute,
          this.timestamp.second
        ).parse_timestamp("2006-01-02 15:04:05")
        
        # 设备信息
        root.device_id = "HVAC-%d-%d".format(
          this.device.train_no,
          this.device.carriage_no
        )
        root.train_no = this.device.train_no
        root.carriage_no = this.device.carriage_no
        
        # 传感器数据(原始值 * 0.1,根据您的实际缩放因子调整)
        root.temp_vehicle_1 = this.sensors.tveh_1 / 10.0
        root.temp_vehicle_2 = this.sensors.tveh_2 / 10.0
        root.humidity_1 = this.sensors.humidity_1 / 10.0
        root.humidity_2 = this.sensors.humidity_2 / 10.0
        
        root.aq_temp_u1 = this.sensors.aq_u1.temp / 10.0
        root.aq_humidity_u1 = this.sensors.aq_u1.humidity / 10.0
        root.aq_co2_u1 = this.sensors.aq_u1.co2
        root.aq_pm25_u1 = this.sensors.aq_u1.pm25
        
        # 压缩机数据
        root.comp_u11_freq = this.sensors.compressor_u11.frequency / 10.0
        root.comp_u11_current = this.sensors.compressor_u11.current / 10.0
        root.comp_u11_voltage = this.sensors.compressor_u11.voltage
        root.comp_u11_power = this.sensors.compressor_u11.power
        root.comp_u11_suct_temp = this.sensors.compressor_u11.suct_temp / 10.0
        root.comp_u11_suct_pres = this.sensors.compressor_u11.suct_pres / 10.0
        root.comp_u11_high_pres = this.sensors.compressor_u11.high_pres / 10.0
        
        # 故障检测
        root.has_fault = (
          this.faults.cfbk_ef_u11 ||
          this.faults.cfbk_cf_u11 ||
          this.faults.bocflt_ef_u11 ||
          this.faults.blpflt_comp_u11
        )
        
        root.fault_count = 0
        root.fault_count = if this.faults.cfbk_ef_u11 { $fault_count + 1 } else { $fault_count }
        root.fault_count = if this.faults.cfbk_cf_u11 { $fault_count + 1 } else { $fault_count }
        # ... 统计所有故障位
        
        # 告警级别
        root.alert_level = if this.faults.blpflt_comp_u11 {
          "critical"
        } else if this.has_fault {
          "warning"
        } else {
          "normal"
        }

output:
  sql_insert:
    driver: postgres
    dsn: "postgres://user:pass@timescaledb:5432/metro_hvac"
    table: "hvac_measurements"
    columns:
      - time
      - device_id
      - train_no
      - carriage_no
      - temp_vehicle_1
      - temp_vehicle_2
      - humidity_1
      - humidity_2
      - aq_temp_u1
      - aq_co2_u1
      - aq_pm25_u1
      - comp_u11_freq
      - comp_u11_current
      - comp_u11_voltage
      - comp_u11_power
      - comp_u11_suct_temp
      - comp_u11_high_pres
      - has_fault
      - fault_count
      - alert_level
    
    args_mapping: |
      root = [
        this.time,
        this.device_id,
        this.train_no,
        this.carriage_no,
        this.temp_vehicle_1,
        this.temp_vehicle_2,
        this.humidity_1,
        this.humidity_2,
        this.aq_temp_u1,
        this.aq_co2_u1,
        this.aq_pm25_u1,
        this.comp_u11_freq,
        this.comp_u11_current,
        this.comp_u11_voltage,
        this.comp_u11_power,
        this.comp_u11_suct_temp,
        this.comp_u11_high_pres,
        this.has_fault,
        this.fault_count,
        this.alert_level
      ]
    
    batching:
      count: 1000
      period: 5s
```

#### 步骤 5: 运行

```bash
# 直接运行编译好的二进制
./rpk-connect-nb67 run config.yaml
```

## 方案优势

### ✅ 架构简化

- **组件减少**: 4个 → 3个 (-25%)
- **无额外服务**: 不需要 Python 转换服务
- **配置集中**: 所有逻辑在一个 YAML 文件

### ✅ 性能提升

- **Go 编译**: 比 Python 快 5-10倍
- **无网络开销**: 减少一次网络传输
- **内存效率**: Go 更节省内存

### ✅ 维护简化

- **单一代码库**: 只需要维护 Go 插件
- **类型安全**: Kaitai 生成的代码有类型检查
- **测试简单**: 单元测试 Go 代码即可

### ✅ 部署简化

```yaml
# docker-compose.yml (简化版)
version: '3.8'

services:
  redpanda:
    image: docker.redpanda.com/redpandadata/redpanda:latest
    # ... 配置

  timescaledb:
    image: timescale/timescaledb:latest-pg14
    # ... 配置

  # 只需要一个自定义 connect,不需要转换服务!
  connect:
    build:
      context: .
      dockerfile: Dockerfile.connect
    volumes:
      - ./config.yaml:/config.yaml
      - /data/incoming:/data/incoming
    command: run /config.yaml

  grafana:
    image: grafana/grafana:latest
    # ... 配置
```

## 实施步骤

### 第一周:开发 Go 插件

1. 用 Kaitai 编译器生成 Go 代码
2. 创建 Redpanda Connect 插件
3. 单元测试解析逻辑

### 第二周:集成测试

1. 编译自定义 Redpanda Connect
2. 端到端测试
3. 性能测试

### 第三周:部署上线

1. 容器化
2. 迁移现有数据
3. 监控和调优

## 替代方案:如果不想写 Go

如果您完全不想写 Go 代码,还有一个更简单的方案:

### 在 Redpanda 之前解析,但用 Go 替代 Python

```go
// 简单的文件监控 + 发送工具
package main

import (
    "github.com/fsnotify/fsnotify"
    "github.com/confluentinc/confluent-kafka-go/kafka"
    // 使用 Kaitai 生成的代码
)

func main() {
    watcher, _ := fsnotify.NewWatcher()
    watcher.Add("/data/incoming")
    
    producer, _ := kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": "redpanda:9092",
    })
    
    for event := range watcher.Events {
        if event.Op&fsnotify.Create == fsnotify.Create {
            data, _ := os.ReadFile(event.Name)
            
            // 用 Kaitai 解析
            nb67 := parseNB67(data)
            
            // 转为 JSON 并发送
            jsonData, _ := json.Marshal(nb67)
            producer.Produce(&kafka.Message{
                TopicPartition: kafka.TopicPartition{
                    Topic: "hvac-data",
                },
                Value: jsonData,
            }, nil)
            
            os.Remove(event.Name)
        }
    }
}
```

这样比 Python 更快,但仍然是独立服务。

## 我的建议

**强烈推荐 Go 插件方案** ⭐

**原因**:
1. ✅ 真正简化架构 (减少一个组件)
2. ✅ 性能最优
3. ✅ 维护最简单
4. ✅ 代码量不大 (Kaitai 自动生成大部分)

**开发工作量**:
- Kaitai 生成代码: 1 分钟
- Go 插件开发: 2-3 天
- 测试和集成: 2-3 天
- **总计**: 约 1 周

vs 维护 Python 转换服务的长期成本,这个投入非常值得!

## 下一步

如果您同意这个方案,我可以:

1. 帮您生成完整的 Go 插件代码
2. 创建完整的部署配置
3. 提供测试脚本

您觉得这个方案如何?
