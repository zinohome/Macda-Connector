# MACDA-Connector 完整系统架构与代码逻辑详解

**文档日期**: 2026-05-01  
**版本**: v2.1.2  
**作者**: Agent (Claude Haiku 4.5) + Team  
**目标受众**: 新接手开发者、二期功能添加

---

## 📋 目录

1. [系统概述](#1-系统概述)
2. [业务流程](#2-业务流程)
3. [架构设计](#3-架构设计)
4. [数据流详解](#4-数据流详解)
5. [代码结构](#5-代码结构)
6. [部署配置详解](#6-部署配置详解)
7. [Dev vs Prod 模式差异](#7-dev-vs-prod-模式差异)
8. [部署前检查清单](#8-部署前检查清单)

---

## 1. 系统概述

### 1.1 项目背景

**MACDA-Connector** 是地铁列车 HVAC (供暖、通风、空调) 系统的遥测数据处理平台。

- **前身**: Faust (Python) + Kafka + TimescaleDB （560 msg/sec）
- **现在**: Redpanda Connect (Go) + Redpanda + TimescaleDB （5,000+ msg/sec 目标）
- **阶段**: Phase 1 已完成并交付客户，Phase 2 正筹备

### 1.2 核心指标

| 指标 | Faust (旧系统) | Redpanda Connect (新系统) |
|------|---|---|
| 吞吐量 | 560 msg/sec | 5,000+ msg/sec |
| 端到端延迟 | ~500ms | ~100ms |
| 内存用量 | ~500MB/实例 | ~150MB/实例 |
| 部署复杂度 | 3个独立进程 | 1个单一二进制 |

### 1.3 客户已测试的内容

✅ 二进制NB67数据解析  
✅ 数据落库TimescaleDB  
✅ 基础Web展示界面  
✅ 告警与预警事件检测  
✅ 离线部署流程  

---

## 2. 业务流程

### 2.1 端到端数据流（5个阶段）

```
┌─────────────────────────────────────────────────────────────────────┐
│  Stage 1: 原始信号获取与解析 (nb67-parser)                          │
│  ─────────────────────────────────────────────────────────────────   │
│  来源: Redpanda topic `signal-in` (462字节二进制帧)                  │
│  处理: NB67.ksy + 自定义Go Processor                                 │
│  输出: topic `signal-parsed` (180+ JSON字段 + 元数据)                │
└─────────────────────────────────────────────────────────────────────┘
            ↓
┌─────────────────────────────────────────────────────────────────────┐
│  Stage 2: 事件检测 (nb67-event-builder)                             │
│  ─────────────────────────────────────────────────────────────────   │
│  消费: topic `signal-parsed`                                         │
│  处理: Go原生处理器分析三类事件                                       │
│  输出: 分发到三个topic (fan_out模式):                                │
│    1. `signal-predict` → 26种预警码 (HVAC算法)                      │
│    2. `signal-alarm`   → 原生故障位                                  │
│    3. `signal-life`    → 部件寿命预警                                │
└─────────────────────────────────────────────────────────────────────┘
            ↓
┌─────────────────────────────────────────────────────────────────────┐
│  Stage 3: 存储持久化 (nb67-storage-writer + nb67-pg-writer)         │
│  ─────────────────────────────────────────────────────────────────   │
│  路径1: `signal-parsed` → TimescaleDB hvac.fact_raw (原始遥测)       │
│         (nb67-storage-writer 消费 signal-in, 采样落库)               │
│                                                                      │
│  路径2: `signal-predict/alarm/life` → TimescaleDB hvac.fact_event   │
│         (nb67-event-writer 消费事件topic，写事件历史)                │
└─────────────────────────────────────────────────────────────────────┘
            ↓
┌─────────────────────────────────────────────────────────────────────┐
│  Stage 4: 数据查询与展示 (nb67-bff + nb67-web)                      │
│  ─────────────────────────────────────────────────────────────────   │
│  BFF: Fastify REST API + WebSocket 实时订阅                         │
│        从 TimescaleDB 查询数据，提供给前端                            │
│                                                                      │
│  Web: Vue 3 + Element Plus + ECharts                                │
│       展示实时状态、历史数据、趋势图、告警信息                        │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 三类事件定义

#### 2.2.1 Predict Event (预警) — 算法预警
- **来源**: 算法规则检测（26种预警码）
- **触发**: 基于传感器数据的阈值/趋势判断
- **例子**: 冷媒泄漏预警、制冷系统压力异常、温度控制偏差
- **Payload**: `{ event_meta, predict_event: { hits: [code, name, severity, ...] } }`

#### 2.2.2 Alarm Event (故障) — 原生故障位
- **来源**: 设备NB67二进制中的故障状态位
- **触发**: 设备自诊断故障位 = 1
- **例子**: 压缩机低压故障、通讯故障、过流保护
- **Payload**: `{ event_meta, alarm_event: { hits: [code, name, level, ...] } }`

#### 2.2.3 Life Event (寿命预警) — 部件累计运行时间
- **来源**: 设备上报的部件累计运行时间 (小时/次数)
- **触发**: 达到额定寿命的 75%/90% 时告警
- **例子**: 风机、压缩机、电子膨胀阀寿命预警
- **Payload**: `{ event_meta, life_event: { hits: [code, name, severity, value, limit, ...] } }`

---

## 3. 架构设计

### 3.1 整体拓扑图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        MACDA Connect 系统拓扑                         │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  [设备]              [Mock数据源]              [真实Kafka]          │
│    ↓                     ↓                       ↓                  │
│    └─────────────────────┴───────────────────────┘                  │
│                         ↓                                            │
│              ┌──────────────────────┐                                │
│              │  Redpanda 消息队列    │ (3节点高可用)                 │
│              │  ├─ signal-in        │ (462字节NB67帧)               │
│              │  ├─ signal-parsed    │ (180+字段JSON)                │
│              │  ├─ signal-predict   │ (预警事件)                     │
│              │  ├─ signal-alarm     │ (故障事件)                     │
│              │  └─ signal-life      │ (寿命事件)                     │
│              └──────────────────────┘                                │
│                ↑      ↑      ↑      ↑                               │
│           ┌────┴──┬───┴──┬───┴──┬───┴────┐                          │
│           │       │      │      │        │                          │
│   ┌──────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────┐       │
│   │nb67-     │ │nb67-   │ │nb67-   │ │nb67-   │ │storage-  │       │
│   │parser    │ │event-  │ │pg-     │ │event-  │ │adapter   │       │
│   │(stage1)  │ │builder │ │writer  │ │writer  │ │(Plan B)  │       │
│   │          │ │(stage2)│ │        │ │        │ │          │       │
│   └────┬─────┘ └───┬────┘ └────┬───┘ └───┬────┘ └────┬─────┘       │
│        │           │           │         │           │             │
│        └───────────┴─ TimescaleDB ──────┴───────────┘             │
│              │           │         │                               │
│              │      ┌────┴────┬────┘                                │
│              │      ↓         ↓                                     │
│         hvac.fact_raw  hvac.fact_event  (Hypertables)             │
│         (原始遥测宽表) (事件历史表)                                 │
│              │              │                                      │
│              └──────┬────────┘                                      │
│                     ↓                                               │
│              ┌─────────────┐                                        │
│              │  nb67-bff   │ (Fastify API + WebSocket)             │
│              │  (BFF层)    │ 查询数据 + 实时推送                    │
│              └──────┬──────┘                                        │
│                     ↓                                               │
│              ┌─────────────┐                                        │
│              │  nb67-web   │ (Vue 3 + Element Plus)                │
│              │  (前端)     │ 展示面板                               │
│              └─────────────┘                                        │
│                     ↑                                               │
│              [浏览器 8080]                                          │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 关键设计决策

#### 3.2.1 双轨管道
系统采用"**全量告警轨 + 采样存储轨**"的混合模式：

- **全量轨**: `signal-parsed → signal-alarm/predict/life` （无采样，确保不丢失告警）
- **采样轨**: `signal-parsed → signal-storage → hvac.fact_raw` （采样率可配，减少90%写入）

这样做的好处：
✅ 实时告警不丢失  
✅ 长期存储成本下降70%+  
✅ 灵活调整采样策略，不影响告警准确度

#### 3.2.2 Plan A / Plan B 存储方案
```yaml
Plan A (默认):
  ├─ connect-pg-writer (Redpanda Connect SQL插件)
  ├─ 使用: docker-compose-Dev.yml / docker-compose-Prod.yml
  └─ 特点: 开箱即用，稳定

Plan B (备用):
  ├─ storage-adapter (Go原生高性能实现)
  ├─ 激活: docker-compose -f ... --profile plan_b up -d
  └─ 特点: 当Plan A达到瓶颈时切换，无需改配置
```

#### 3.2.3 时序字段选择 (DEV vs PRD)
```
DEV 模式 (开发/测试):
  ├─ 使用: ingest_time (网关接收时间)
  ├─ 优点: 不依赖设备时钟准确度
  ├─ 缺点: 分析的是处理延迟，不是业务时间
  └─ 场景: 调试、部署验证

PRD 模式 (生产):
  ├─ 使用: event_time (设备上报时间)
  ├─ 优点: 分析结果反映业务时间，准确
  ├─ 缺点: 依赖设备时钟校准
  └─ 场景: 生产环境、历史分析
```

---

## 4. 数据流详解

### 4.1 二进制解析 (Stage 1: nb67-parser)

**输入**: `signal-in` topic (462字节NB67帧)

**处理流程**:
```
462字节二进制
    ↓
NB67.ksy (Kaitai定义) + Kaitai Runtime
    ↓
nb67_processor.go (Go处理器) 
    ├─ 解析所有字段
    ├─ 时间格式转换 (设备时间 → ISO 8601)
    ├─ 错误检测 (CRC/长度校验)
    └─ 添加元数据 (parse_version, quality_status, parsed_at)
    ↓
nb67-parser.yaml (Bloblang映射)
    ├─ 添加schema_version标记
    ├─ 生成ingest_time (网关接收时间)
    ├─ 生成device_id (HVAC-{line}-{train}-{carriage})
    ├─ 验证event_time是否合法 (month/day范围)
    └─ 质量标记 (quality_code: 0=OK, 1=ERROR)
    ↓
JSON 输出 (180+ 字段)
```

**输出**: `signal-parsed` topic

**示例输出**:
```json
{
  "schema_version": "nb67.parsed",
  "header_code_01": 0xAA,
  "header_code_02": 0x55,
  "message_length": 462,
  "src_device_no": 1,
  "train_no": 12345,
  "carriage_no": 3,
  "line_id": 1,
  "train_id": 12345,
  "carriage_id": 3,
  "device_id": "HVAC-1-12345-3",
  "event_time": "2026-05-01T09:40:00+08:00",
  "event_time_text": "2026-05-01 09:40:00",
  "event_time_valid": true,
  "ingest_time": "2026-05-01T09:40:00.123+08:00",
  "parser_version": "nb67-v1",
  "quality_status": "OK",
  "quality_code": 0,
  // ... 180+ 其他字段 (温度、湿度、故障位等)
  "tveh_1": 2205,  // 实际值需要乘以 0.1，即 22.05°C
  "humdity_1": 5500, // 实际值需要乘以 0.01，即 55.00%
  "dmp_exu_pos": 45, // 新增字段：排气阀位置
  "start_station": 10, // 新增字段：起始站
  "terminal_station": 25, // 新增字段：终点站
  "cur_station": 15, // 新增字段：当前站
  "next_station": 16, // 新增字段：下一站
  "blpflt_comp_u11": false, // 压缩机低压故障位
  "raw": { /* 完整的Kaitai解析对象 */ }
}
```

**关键点**:
- 所有温度/湿度值需要**缩放** (×0.1 或 ×0.01)
- `event_time_valid` 字段用于在DEV模式下识别设备时钟异常
- `raw` 字段包含完整的Kaitai对象，供下游处理器使用

### 4.2 事件检测 (Stage 2: nb67-event-builder)

**输入**: `signal-parsed` topic

**处理流程**:

#### 4.2.1 Predict Event (算法预警)
```
signal-parsed JSON (含 raw 字段)
    ↓
nb67_event_processor.go 检查预警规则
    ├─ 规则1: 冷媒泄漏 (ref_leak_u11/u12/u21/u22)
    │   条件: 吸气压力 < 阈值1 AND 高压 < 阈值2
    │   输出: "HVAC301"
    │
    ├─ 规则2: 制冷能力下降 (cooling_efficiency_down)
    │   条件: 温度下降斜率 < 阈值 (持续5分钟)
    │   输出: "HVAC302"
    │
    ├─ ... (共26种预警码)
    │
    └─ 规则26: 部件寿命预警
        条件: 累计运行时间 >= 75% 寿命
        输出: "LIFE001" 等
    ↓
SubEvent 结构
{
  "event_meta": { /* 元数据 */ },
  "predict_event": {
    "hits": [
      { "code": "HVAC301", "name": "冷媒泄漏预警", "severity": 3 },
      { "code": "HVAC302", "name": "制冷能力下降", "severity": 2 }
    ]
  }
}
```

#### 4.2.2 Alarm Event (原生故障位)
```
检查 signal-parsed 中的故障位字段
    ├─ blpflt_comp_u11 (压缩机低压故障)
    ├─ bscflt_comp_u11 (压缩机通讯故障)
    ├─ bocflt_ef_u11 (通风机过流)
    ├─ ... (共20+ 故障位)
    │
    └─ 只要 bit = 1，就生成对应 AlarmHit
    ↓
SubEvent 结构
{
  "event_meta": { /* 元数据 */ },
  "alarm_event": {
    "hits": [
      { "code": "blpflt_comp_u11", "name": "压缩机低压故障", "level": 1 },
      { "code": "bocflt_ef_u11", "name": "通风机过流保护", "level": 2 }
    ]
  }
}
```

#### 4.2.3 Life Event (寿命预警)
```
signal-parsed → raw字段提取累计运行时间
    ├─ 风机累计运行时间
    ├─ 压缩机累计运行时间  
    ├─ 阀门动作次数
    │
    └─ 与预设阈值对比
        ├─ > 额定寿命的 90% → severity=3 (严重)
        ├─ > 额定寿命的 75% → severity=2 (中等)
        └─ < 额定寿命的 75% → (不告警)
    ↓
SubEvent 结构
{
  "event_meta": { /* 元数据 */ },
  "life_event": {
    "hits": [
      { "code": "LIFE001", "name": "风机即将达到寿命", 
        "severity": 3, "value": 86000, "limit": 90000 },
      { "code": "LIFE002", "name": "压缩机接近寿命预警",
        "severity": 2, "value": 135000, "limit": 180000 }
    ]
  }
}
```

**输出**: 三个 topic (fan_out 分发)
- `signal-predict` (仅包含predict_event)
- `signal-alarm` (仅包含alarm_event)
- `signal-life` (仅包含life_event)

### 4.3 存储持久化 (Stage 3: nb67-storage-writer + nb67-pg-writer)

#### 4.3.1 原始遥测入库 (signal-in → hvac.fact_raw)

**配置**: `nb67-storage-writer.yaml`

```yaml
input:
  kafka:
    topics: [signal-in]  # 消费原始topic，非parsed
    
processors:
  # 只入库合法的帧（采样可选）
  - filter:
      if: this.quality_status == "OK"  # 过滤错误帧
  - filter:
      if: this.event_time_valid == true  # 过滤时间异常帧
  # 可添加采样逻辑（每100条入库1条）
  
output:
  sql:
    driver: postgres
    dsn: postgres://postgres:passw0rd@timescaledb:5432/postgres
    query: |
      INSERT INTO hvac.fact_raw (
        event_time, ingest_time, line_id, train_id, carriage_id, device_id,
        f_cp_u11, i_cp_u11, suckp_u11, highpress_u11, 
        fas_u1, ras_u1, aq_co2_u1, aq_pm2_5_u1, ... 其他字段
      ) VALUES (
        ${parsed_at}, ${ingest_time}, ${line_id}, ${train_id}, ${carriage_id}, ${device_id},
        ${f_cp_u11}, ${i_cp_u11}, ${suckp_u11}, ${highpress_u11},
        ... 对应参数占位符
      )
```

**入库速度**: 560 msg/sec → ~1 条/秒 (采样模式下)

**表结构**: Hypertable，自动按 event_time 分块 (7天/块)

#### 4.3.2 事件历史入库 (signal-predict/alarm/life → hvac.fact_event)

**配置**: `nb67-pg-writer.yaml` 和 `nb67-event-writer.yaml`

```yaml
# nb67-pg-writer: 消费采样的原始数据，落库到 fact_raw
input:
  kafka:
    topics: [signal-in]
    
# nb67-event-writer: 消费事件 topic，落库到 fact_event
input:
  kafka:
    topics: [signal-predict, signal-alarm, signal-life]
    
processors:
  - mapping: |
      root.event_time = this.event_meta.event_time
      root.event_type = if this.exists("predict_event") {
        "predict"
      } else if this.exists("alarm_event") {
        "alarm"
      } else {
        "life"
      }
      # 遍历 hits 数组，为每个 hit 生成一条记录
      
output:
  sql:
    query: |
      INSERT INTO hvac.fact_event (
        event_time, device_id, event_type, fault_code, fault_name, severity, payload_json
      ) VALUES (...)
```

**表结构**: Hypertable，支持历史追溯

---

## 5. 代码结构

### 5.1 目录树

```
Macda-Connector/
├── connect/                          # 后端 Redpanda Connect 流水线
│   ├── cmd/
│   │   ├── connect-nb67/             # NB67 解析器 + 事件检测器
│   │   │   ├── main.go               (47行) Processor注册入口
│   │   │   ├── nb67_processor.go     (250行) 解析逻辑 + JSON输出
│   │   │   ├── nb67_event_processor.go (~500行) 事件检测算法
│   │   │   ├── nb67.go               (1936行) Kaitai生成的二进制解析
│   │   │   └── go.mod
│   │   │
│   │   └── storage-adapter/          # Plan B: 高性能存储适配器（备用）
│   │       ├── main.go               启动入口
│   │       ├── consumer.go           Kafka消费逻辑
│   │       ├── transform.go          数据转换
│   │       ├── config.go             配置加载
│   │       └── types.go              数据结构
│   │
│   ├── codec/
│   │   ├── NB67.ksy                  ⭐ 二进制协议SSOT (Kaitai Struct定义)
│   │   └── nb67.go                   (Kaitai自动生成，勿手改)
│   │
│   ├── config/                       ⭐ 三阶段 Redpanda Connect 配置
│   │   ├── nb67-parser.yaml          (Stage 1) 二进制→JSON解析
│   │   ├── nb67-event-builder.yaml   (Stage 2) 事件检测 (fan_out)
│   │   ├── nb67-storage-writer.yaml  (Stage 3a) 采样原始数据→DB
│   │   ├── nb67-pg-writer.yaml       (Stage 3b) 全量原始数据→DB
│   │   └── nb67-event-writer.yaml    (Stage 3c) 事件历史→DB
│   │
│   └── tests/
│       └── test-nb67-parsing.sh      解析器测试脚本
│
├── web-nb67-bff/                     # 前端 BFF 层 (Fastify REST+WS)
│   ├── src/
│   │   ├── index.ts                  (400+行) Fastify 启动 + 路由注册
│   │   ├── config/
│   │   │   ├── index.ts              配置管理 (DB/Kafka连接)
│   │   │   ├── db.ts                 Kysely ORM + 数据库连接
│   │   │   └── kafka.js              Kafka消费者配置
│   │   │
│   │   ├── repository/               数据层 (查询逻辑)
│   │   │   ├── history.repository.ts 历史数据查询
│   │   │   ├── status.repository.ts  实时状态查询
│   │   │   └── alarm.repository.ts   告警查询
│   │   │
│   │   ├── utils/
│   │   │   └── time.ts               时间格式化
│   │   │
│   │   └── package.json              依赖: fastify, kubernetes, cors等
│   │
│   └── README.md                     BFF部署说明
│
├── web-nb67-web/                     # 前端 Web 界面 (Vue 3)
│   ├── src/
│   │   ├── App.vue                   顶层容器
│   │   ├── main.ts                   Vite入口
│   │   ├── api/
│   │   │   └── api.ts                REST API 调用
│   │   │
│   │   ├── views/                    主要页面
│   │   │   ├── home/                 首页 (实时仪表板)
│   │   │   │   ├── index.vue         布局
│   │   │   │   ├── Left.vue          左侧列车表
│   │   │   │   ├── Center.vue        中央地图/趋势
│   │   │   │   ├── Right.vue         右侧告警
│   │   │   │   └── AirErrorData.vue  底部错误数据表
│   │   │   │
│   │   │   ├── trainInfo/            列车详情页
│   │   │   │   ├── index.vue         列表视图
│   │   │   │   ├── CarTemperature.vue 温度分布
│   │   │   │   ├── AlarmInfo.vue     告警详情
│   │   │   │   └── RunState.vue      运行状态
│   │   │   │
│   │   │   ├── airConditioner/       空调系统详情
│   │   │   │   ├── System.vue        系统运行状态
│   │   │   │   ├── Healthy.vue       健康度分析
│   │   │   │   ├── Echart.vue        实时数据曲线
│   │   │   │   └── TrainUnitMap.vue  单元位置图
│   │   │   │
│   │   │   └── historyData/          历史数据查询
│   │   │       └── index.vue
│   │   │
│   │   ├── components/               公共组件
│   │   │   ├── TitleHeader.vue       页头
│   │   │   ├── BackTop.vue           返回顶部
│   │   │   └── ...
│   │   │
│   │   ├── router/                   路由配置
│   │   │   └── index.ts
│   │   │
│   │   ├── stores/                   Pinia 全局状态
│   │   │   └── ...
│   │   │
│   │   ├── assets/                   静态资源 (图片/图标)
│   │   │
│   │   └── style.css                 全局样式
│   │
│   ├── vite.config.ts                Vite 构建配置
│   ├── tsconfig.json                 TypeScript 配置
│   └── package.json                  依赖: vue@3, element-plus, echarts
│
├── baseEnv/                          # 开发环境编排
│   ├── docker-compose-Dev.yml        ⭐ 开发环境 (完整3节点Redpanda)
│   └── docker-compose-mock.yml       Mock数据源（供测试）
│
├── dist/                             # 离线部署包
│   ├── docker-compose-Data.yml       基础设施层
│   ├── docker-compose-Web.yml        应用层
│   ├── docker-compose-mock.yml       Mock层
│   ├── config/                       Connect配置文件副本
│   ├── init-db/                      TimescaleDB初始化SQL
│   ├── image-save.sh                 镜像打包脚本
│   ├── image-load.sh                 镜像加载脚本
│   ├── install.sh                    初始化脚本
│   ├── start.sh                      启停脚本
│   └── README.md                     部署文档
│
└── docs/                             # 文档
    ├── requirements/                 需求文档
    │   ├── binary-spec.md            NB67二进制规格
    │   ├── connect-*.md              Connect配置说明
    │   └── ...
    │
    ├── research/                     研究文档
    │   ├── 11-macda-refactor-execution-plan.md (12周执行计划)
    │   └── ...
    │
    └── SYSTEM-ARCHITECTURE-AND-CODEBASE-WALKTHROUGH.md (本文件)
```

### 5.2 关键文件详解

#### 5.2.1 NB67.ksy (二进制协议SSOT)

位置: `connect/codec/NB67.ksy`

**作用**: 定义NB67二进制格式的字段布局、类型、偏移量

**结构** (462字节):
```
字节0-1:     消息头标识 (0xAA55)
字节2-3:     消息长度
字节4-5:     源设备号 + 宿设备号
字节6-7:     消息类型 + 帧序号
字节8-11:    列车号 + 车厢号 + 协议版本
字节12-17:   设备时间戳 (YY-MM-DD-HH-mm-SS)
... (180+ 字段)
字节452-461: 新增5个字段 (排气阀位置、起始站、终点站、当前站、下一站)
```

**修改流程** (如需新增字段):
```
1. 编辑 NB67.ksy (Kaitai Struct YAML)
2. 重新生成Go代码:
   kaitai-struct-compiler -t go codec/NB67.ksy -o connect/cmd/connect-nb67/
3. 更新 nb67_processor.go ParsedOutput 结构体
4. 同时更新 nb67-parser.yaml Bloblang 映射
5. 更新 TimescaleDB Schema (if 涉及新字段入库)
6. 提交: git add *.ksy *.go *.yaml
```

⚠️ **关键约束**: NB67.ksy 是**唯一真实来源 (SSOT)**，所有字段变更必须从此处开始。

#### 5.2.2 nb67_processor.go (解析逻辑)

位置: `connect/cmd/connect-nb67/nb67_processor.go`

**职责**:
1. 接收462字节二进制数据
2. 通过Kaitai Runtime解析成Go结构体 (nb67.go)
3. 映射到JSON对象 (ParsedOutput)
4. 添加元数据 (parsed_at, ingest_time等)
5. 返回JSON消息

**核心函数**:
```go
func (p *NB67Processor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error)

// 逻辑:
// 1. 从msg获取二进制字节
// 2. 创建Nb67对象，调用Read() 解析
// 3. 构建ParsedOutput映射
// 4. JSON序列化
// 5. 返回
```

#### 5.2.3 nb67_event_processor.go (事件检测)

位置: `connect/cmd/connect-nb67/nb67_event_processor.go`

**职责**: 包含26种预警规则 + 部件寿命计算逻辑

**数据结构**:
```go
type EventOutput struct {
  PredictEvent SubEvent  // 预警事件
  AlarmEvent   SubEvent  // 故障事件
  LifeEvent    SubEvent  // 寿命事件
}

type SubEvent struct {
  EventMeta EventMeta   // 元数据
  Hits      interface{} // 命中列表
  Source    string      // 来源标识
}
```

**规则引擎** (示例):
```go
// 预警规则: 冷媒泄漏
if suckp < 500 && highpress < 1000 {
  hits = append(hits, PredictHit{
    Code:     "HVAC301",
    Name:     "冷媒泄漏预警",
    Severity: 3,
  })
}
```

#### 5.2.4 nb67-parser.yaml (Bloblang映射)

位置: `connect/config/nb67-parser.yaml`

**职责**: 处理器链配置 + JSON字段映射

```yaml
input:
  kafka:
    addresses: [redpanda-1:9092, ...]
    topics: [signal-in]

pipeline:
  processors:
    # 处理器1: Go自定义处理器
    - nb67_parser:
        log_sample_every: 100
    
    # 处理器2: Bloblang脚本映射
    - mapping: |
        root = this  # 保留原字段
        root.schema_version = "nb67.parsed"
        root.ingest_time = now()  # 添加接收时间
        root.device_id = "HVAC-%v-%v-%v".format(...)  # 生成设备ID
        root.quality_code = if this.quality_status == "OK" { 0 } else { 1 }

output:
  kafka:
    topic: signal-parsed
    key: ${! this.device_id }  # 按设备ID分区
    batching:
      count: 200               # 200条消息或
      period: 100ms            # 100毫秒时自动flush
```

#### 5.2.5 nb67-event-builder.yaml (事件分发)

位置: `connect/config/nb67-event-builder.yaml`

**职责**: 调用事件处理器 + fan_out 分拣

```yaml
input:
  kafka:
    topics: [signal-parsed]

pipeline:
  processors:
    # 调用Go处理器 (nb67_event_builder)
    - nb67_event_builder: {}

output:
  broker:
    pattern: fan_out  # 一进三出
    outputs:
      # 输出1: signal-predict
      - processors:
          - mapping: |
              root = if this.exists("predict_event") && 
                        this.predict_event.hits.length() > 0 {
                this.predict_event
              } else {
                deleted()  # 无命中则删除
              }
        kafka:
          topic: signal-predict
      
      # 输出2: signal-alarm
      - processors:
          - mapping: |
              root = if this.exists("alarm_event") && ... {
                this.alarm_event
              } else {
                deleted()
              }
        kafka:
          topic: signal-alarm
      
      # 输出3: signal-life
      - processors: [...]
        kafka:
          topic: signal-life
```

#### 5.2.6 BFF index.ts (API入口)

位置: `web-nb67-bff/src/index.ts`

**职责**:
1. 启动Fastify服务
2. 注册REST路由 + WebSocket路由
3. 连接TimescaleDB + Kafka
4. 提供API端点给前端

**关键路由** (示例):
```typescript
// REST: 获取在线列车告警总结
GET /api/rest/AirSystem
  → StatusRepository.getTrainAlarmSummary()
  → SELECT train_no, alarm_count, warning_count FROM hvac.fact_event ...

// REST: 获取指定列车实时告警
POST /api/rest/v2/train/RealtimeAlarm
  Body: { train_no: 12345 }
  → HistoryRepository.getRealtimeAlarm(trainNo)
  → 查询最新的alarm/predict/life事件

// WebSocket: 实时推送
WS /ws/signals
  → 订阅Kafka消息，推送给客户端
```

#### 5.2.7 前端 views (Vue 3页面)

**首页** (`views/home/index.vue`)
- Left: 列车列表 + 选择
- Center: 地图 + 趋势图
- Right: 告警/预警汇总

**列车详情** (`views/trainInfo/index.vue`)
- 单列车温度/湿度/故障信息

**空调详情** (`views/airConditioner/index.vue`)
- 系统运行状态、健康度、实时数据曲线

---

## 6. 部署配置详解

### 6.1 开发环境 (baseEnv/docker-compose-Dev.yml)

**目的**: 本地测试和功能验证

**主要服务**:
```yaml
services:
  redpanda-1/2/3:       # 3节点Redpanda集群
  timescaledb:          # 时序数据库
  pgadmin:              # PostgreSQL管理工具
  
  # Connect流水线 (5个处理器)
  connect-topic-in:     # 创建signal-in topic
  connect-parser:       # Stage 1: 解析
  connect-storage-writer:  # Stage 3a: 采样原始→DB
  connect-event-builder:   # Stage 2: 事件检测
  connect-pg-writer:       # Stage 3b: 全量原始→DB
  connect-event-writer:    # Stage 3c: 事件→DB
  
  # 应用层
  nb67-bff:             # Fastify API服务
  nb67-web:             # Nginx前端
```

**运行命令**:
```bash
# 启动所有服务
docker-compose -f baseEnv/docker-compose-Dev.yml up -d

# 查看状态
docker-compose -f baseEnv/docker-compose-Dev.yml ps

# 查看日志
docker-compose -f baseEnv/docker-compose-Dev.yml logs -f connect-parser
```

**端口映射**:
- 8080: 前端Web
- 3000: BFF API (内网)
- 5432: TimescaleDB (内网)
- 19092/29092/39092: Kafka 外网端口

### 6.2 Mock环境 (baseEnv/docker-compose-mock.yml)

**目的**: 独立的数据源模拟环境

**特点**:
- 单节点Redpanda (而非3节点)
- Mock数据生产者 (持续读取mock-data文件，发送到signal-in)
- 独立的Redpanda Console (48080端口)

**运行命令**:
```bash
# 启动Mock环境
docker-compose -f baseEnv/docker-compose-mock.yml up -d

# 查看Mock数据是否发送
docker exec mock-redpanda rpk topic consume signal-in
```

**使用场景**:
- 调试时无真实设备数据
- 演示系统功能
- 压力测试 (快速回放历史数据)

### 6.3 生产环境 (dist/docker-compose-Data.yml + docker-compose-Web.yml)

**目的**: 离线交付给客户

**分层**:
- `docker-compose-Data.yml`: 基础设施层 (Redpanda + TimescaleDB + Connect)
- `docker-compose-Web.yml`: 应用层 (BFF + Web)

**部署流程**:
```bash
# 1. 初始化 (创建目录、权限、复制配置)
sudo ./dist/install.sh --data-dir /data/MACDA2

# 2. 准备镜像 (有网络机器)
cd dist
./image-save.sh  # 生成 images/*.tar

# 3. 传输到离线机器 (U盘/内网)
scp dist/ user@offline-server:/opt/macda/

# 4. 加载镜像 (离线机器)
./image-load.sh --verify  # 验证MD5
./image-load.sh           # 加载

# 5. 启动服务
./start.sh

# 6. 验证
./start.sh status
```

---

## 7. Dev vs Prod 模式差异

### 7.1 时序字段选择

| 模式 | 使用字段 | 优点 | 缺点 | 场景 |
|------|---------|------|------|------|
| **DEV** | `ingest_time` (网关接收时间) | ✅ 不依赖设备时钟，调试方便 | ❌ 分析的是处理延迟 | 开发、部署验证、测试 |
| **PRD** | `event_time` (设备上报时间) | ✅ 业务时间准确 | ❌ 需要设备时钟校准 | 生产、历史分析、SLA |

**配置位置**:
```yaml
# nb67-event-builder.yaml
environment:
  - RUNTIME=DEV  # 切换为 PRD
```

**影响范围**:
- nb67-bff 中的时序查询逻辑
- 前端展示的时间轴
- TimescaleDB 的分析查询

### 7.2 Kafka分区策略

| 模式 | 分区数 | 消费者实例 | 吞吐量 | 部署难度 |
|------|--------|---------|--------|---------|
| **DEV** | 3 (标准) | 1~3 | 560 msg/sec (测试) | ✅ 低 |
| **PRD** | 3+ (可扩展) | 3+ | 5,000+ msg/sec (生产) | ❌ 中高 |

**扩展命令**:
```bash
# PRD环境水平扩展
docker-compose -f docker-compose-Data.yml up -d \
  --scale connect-parser=3 \
  --scale connect-storage-writer=3 \
  --scale connect-event-builder=2
```

---

## 8. 部署前检查清单

在自动部署前，请确认以下项目：

### ✅ 代码与配置
- [ ] NB67.ksy 二进制协议定义已冻结（如需修改字段，务必同步所有代码）
- [ ] 5个Connect YAML配置文件格式正确、Kafka地址准确
- [ ] nb67_processor.go 和 nb67_event_processor.go 编译无误
- [ ] BFF TypeScript 代码无syntax错误
- [ ] 前端Vue组件可正常加载

### ✅ 基础设施
- [ ] Docker daemon 正在运行
- [ ] 剩余磁盘空间 > 50GB (TimescaleDB需要)
- [ ] 网络连通性: 能否访问Harbor镜像仓库 (如在线部署)
- [ ] 防火墙规则: 8080(Web), 3000(BFF), 5432(DB) 等端口开放

### ✅ 数据库
- [ ] TimescaleDB 初始化SQL已准备 (dist/init-db/01-init.sql)
- [ ] hvac Schema, hvac.fact_raw, hvac.fact_event 表已创建
- [ ] 索引已建立 (ix_fact_raw_device_time 等)

### ✅ Kafka
- [ ] Redpanda集群3个节点健康 (health check通过)
- [ ] 已创建 signal-in topic (3分区, 1副本最低)
- [ ] 已创建其他topic: signal-parsed, signal-predict, signal-alarm, signal-life, signal-storage

### ✅ 镜像版本
- [ ] nb-parse-connect:v2.2.0-full (Connect处理器)
- [ ] nb67-bff:v2.1.2 (BFF服务)
- [ ] nb67-web:v2.1.2 (前端Nginx)
- [ ] timescaledb-ha:pg14-ts2.19-all (时序DB)
- [ ] redpanda:v25.3.7 (消息队列)

### ✅ 部署工具
- [ ] dist/image-save.sh 可执行且逻辑正确
- [ ] dist/image-load.sh 可执行且包含MD5校验
- [ ] dist/install.sh 可执行且支持 --data-dir 参数
- [ ] dist/start.sh 可执行且支持 start/stop/status 命令

### ✅ 文档与交接
- [ ] 已生成部署手册 (step-by-step)
- [ ] 已生成post-deployment检查清单
- [ ] 已生成架构说明文档 (本文件)
- [ ] 已生成故障排查手册

---

## 总结

本文档详尽阐述了 MACDA-Connector 的：

1. **业务流程**: 二进制 → 解析 → 事件检测 → 存储 → 查询展示
2. **系统架构**: 5个处理阶段 + 3类事件 + 双轨管道
3. **代码结构**: 后端(Go) + 前端(Vue 3) + 配置(YAML)
4. **部署策略**: Dev环境 vs Mock环境 vs 生产离线部署
5. **关键差异**: DEV时序 vs PRD时序、Plan A vs Plan B存储

**下一步**:

待你确认理解无误后，我将：
1. 自动启动 Dev + Mock 环境
2. 验证各个处理阶段的数据流
3. 记录实际运行情况
4. 为后续功能开发做准备

---
