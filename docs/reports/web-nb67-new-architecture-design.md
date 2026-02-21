# Web-NB67 实时数据驱动新架构设计与开发文档

> **设计目标**: 基于 Redpanda + Redpanda Connect + TimescaleDB 新型物联网流批一体基础架构，重构前端数据流与呈现，实现工业级高可用与高性能。
> **适用终端**: 列车 HVAC (空调) 监控大屏及各类运维管理终端。
> **设计日期**: 2026-02

---

## 目录

1. [核心流批一体架构设计](#1-核心流批一体架构设计)
2. [安全性审计与整改策略 (Security Audit)](#2-安全性审计与整改策略-security-audit)
3. [TimescaleDB 数据库模型优化设计](#3-timescaledb-数据库模型优化设计)
4. [API 交互层设计规范](#4-api-交互层设计规范)
5. [前端工程化改造与高效运行方案](#5-前端工程化改造与高效运行方案)

---

## 1. 核心流批一体架构设计

通过整合边缘采集，剥离此前前端每 2 分钟轮询（`setInterval`）的低效架构，重构为“**实时推送 + 历史冷读**”并行的双引擎架构。

### 1.1 数据流转架构图

```mermaid
graph TD
    subgraph 边缘侧/设备侧
        A[NB67列车 HVAC终端] -->|TCP/MQTT二进制流| B(Redpanda: signal-in)
    end

    subgraph 核心流处理层 (Redpanda Connect)
        B --> C[RC: nb67-parser]
        C -->|解析出标准化JSON| D(Redpanda: signal-parsed)
        
        D --> E[RC: nb67-event-builder]
        E -->|检出| F(Redpanda: signal-alarm / predict / life)
        
        D --> G[RC: nb67-storage-writer]
    end

    subgraph 固化存储与API层
        G -->|批量高吞吐写入| H[(TimescaleDB: HyperTable)]
        F -->|持久化| H
        H -->|查询| I[BFF 聚合 API 模块]
        F -->|订阅| J[WebSocket 通知网关]
    end

    subgraph 前端展现层 (Vue3 + Vite)
        J -->|实时故障/指标推送| K[前端: Pinia Store 响应式流]
        I -->|历史/统计接口| L[前端: ECharts分析/数据报表]
    end
```

### 1.2 架构优势
*   **毫秒级故障感知**: 设备端发生异常，走 `signal-alarm` -> `WebSocket` 快速直达前端，消除 2 分钟的数据盲区。
*   **数据库零压**: 高频遥测写入和查询解耦，通过 TimescaleDB `Continuous Aggregates` (连续聚合) 将前端 ECharts 查询复杂度降为 O(1)。
*   **高可靠(HA)**: 依托 Redpanda 副本集容灾支持。

---

## 2. 安全性审计与整改策略 (Security Audit)

针对遗留的直接访问 GraphQL 引擎架构，需坚决执行以下安全隔离隔离原则：

| 风险等级 | 项目 | 现状描述 | 架构重构策略 |
| :--- | :--- | :--- | :--- |
| 🔴 **P0 (致命)** | 超管凭证硬编码 | 前端 `x-hasura-admin-secret` 携带密码裸奔，任何人可直接拖库/删库。 | **建立 BFF 网关**：褫夺前端直接访问数据库/GraphQL的权利，所有请求必须经过后端网关，由后端持有数据库密码。 |
| 🟠 **P1 (高危)** | 无鉴权拦截 | 页面无角色区分，任何访问者可查看任意车厢数据。 | **RBAC + JWT 鉴权**：后端颁发 Token 给前端，前端将其附加在 `Authorization: Bearer <token>` 并在网关层验签，严格约束数据行级权限。 |
| 🟡 **P2 (中度)** | 防重放与限流缺失 | 没有频控，极易被 DDoS 及重复报文消耗性能。 | 在 WebSocket 通道和 REST API 设计前引入 API Gateway 限定连接数与速率。Redpanda Connect 层加入针对时间戳的校验。 |

---

## 3. TimescaleDB 数据库模型优化设计

> **规范准则**: 严格遵循 PostgreSQL / TimescaleDB 开发规范。放弃原有非时序的平铺数据表，采用 **Hypertable (超表)**，针对读写极度分离的物联网特征进行建模。

### 3.1 核心 DDL 脚本设计

```postgresql
-- =============================================================================
-- 核心底层数据表与时序优化设计 
-- 适用环境: PostgreSQL 14 + TimescaleDB
-- 设计依据: 严格对接 Redpanda Connect 的 storage-adapter 写入端 (hvac.fact_raw) 
--          并同步满足前端 web-nb67 历史数据检索与可视化展示需求
-- =============================================================================

CREATE SCHEMA IF NOT EXISTS hvac;

-- 1. 创建核心超表：实时传感器遥测与状态宽表 (Hypertable)
-- 此表结构严格映射 storage-adapter/consumer.go 内定 INSERT 结构
CREATE TABLE IF NOT EXISTS hvac.fact_raw (
    -- 1.1 基线时间体系 (主键核心)
    event_time TIMESTAMPTZ NOT NULL,              -- 报文产生时间
    ingest_time TIMESTAMPTZ NOT NULL,             -- 网关接收时间
    process_time TIMESTAMPTZ,                     -- 解析处理时间
    event_time_valid BOOLEAN,                     -- event_time 是否有效合法(防漂移)
    
    -- 1.2 空间维度特征
    line_id INTEGER,                              -- 线路编号
    train_id INTEGER,                             -- 列车编号
    carriage_id INTEGER,                          -- 车厢编号
    device_id VARCHAR(64) NOT NULL,               -- 设备唯一标识 (如 'HVAC-67-12345-01')
    frame_no BIGINT,                              -- 帧序号
    
    -- 1.3 质量与解析特征
    parser_version VARCHAR(32),                   -- 解析器版本
    quality_code SMALLINT,                        -- 数据质量码
    
    -- 1.4 核心遥测指标 - U1 / U2 机组运行模式与变频机特征
    wmode_u1 SMALLINT, wmode_u2 SMALLINT,         -- 运行模式
    f_cp_u11 NUMERIC(8,2), f_cp_u12 NUMERIC(8,2), -- 压缩机频率 (Hz)
    f_cp_u21 NUMERIC(8,2), f_cp_u22 NUMERIC(8,2),
    i_cp_u11 NUMERIC(8,2), i_cp_u12 NUMERIC(8,2), -- 压缩机电流 (A)
    i_cp_u21 NUMERIC(8,2), i_cp_u22 NUMERIC(8,2),
    suckp_u11 NUMERIC(8,2), suckp_u12 NUMERIC(8,2), -- 吸气压力
    suckp_u21 NUMERIC(8,2), suckp_u22 NUMERIC(8,2),
    highpress_u11 NUMERIC(8,2), highpress_u12 NUMERIC(8,2), -- 高压
    highpress_u21 NUMERIC(8,2), highpress_u22 NUMERIC(8,2),
    
    -- 1.5 核心遥测指标 - 温度与空气质量
    fas_u1 NUMERIC(6,2), fas_u2 NUMERIC(6,2),     -- 综合新风温度
    ras_u1 NUMERIC(6,2), ras_u2 NUMERIC(6,2),     -- 综合回风温度
    presdiff_u1 NUMERIC(8,2), presdiff_u2 NUMERIC(8,2), -- 滤网压差
    aq_co2_u1 NUMERIC(8,2), aq_co2_u2 NUMERIC(8,2), -- CO2
    aq_tvoc_u1 NUMERIC(8,2), aq_tvoc_u2 NUMERIC(8,2), -- TVOC
    aq_pm2_5_u1 NUMERIC(8,2), aq_pm2_5_u2 NUMERIC(8,2), -- PM2.5
    aq_pm10_u1 NUMERIC(8,2), aq_pm10_u2 NUMERIC(8,2), -- PM10

    -- 1.6 布尔状态/故障报警位 (Fault/Alarms)
    bocflt_ef_u11 BOOLEAN, bocflt_ef_u12 BOOLEAN,    -- 通风机过流等保护
    bocflt_ef_u21 BOOLEAN, bocflt_ef_u22 BOOLEAN,
    bocflt_cf_u11 BOOLEAN, bocflt_cf_u12 BOOLEAN,    -- 冷凝风机过流
    bocflt_cf_u21 BOOLEAN, bocflt_cf_u22 BOOLEAN,
    blpflt_comp_u11 BOOLEAN, blpflt_comp_u12 BOOLEAN,-- 压缩机低压故障
    blpflt_comp_u21 BOOLEAN, blpflt_comp_u22 BOOLEAN,
    bscflt_comp_u11 BOOLEAN, bscflt_comp_u12 BOOLEAN,-- 压缩机通讯故障
    bscflt_comp_u21 BOOLEAN, bscflt_comp_u22 BOOLEAN,
    bflt_tempover BOOLEAN,                           -- 整体车厢超级高温报警
    bflt_diffpres_u1 BOOLEAN, bflt_diffpres_u2 BOOLEAN, -- 滤网压差故障

    -- 1.7 预留全量扩展字段与时序防重防重复联合主键声明
    payload_json JSONB,                           -- 万能兜底全量解析 JSON
    UNIQUE (device_id, event_time, ingest_time)
);
COMMENT ON TABLE hvac.fact_raw IS '设备明细底层存根数据宽表';

-- 将表转化为 Hypertable，极速写入及查询。以 event_time 作为时序切片键，分块长度7天
SELECT create_hypertable('hvac.fact_raw', 'event_time', chunk_time_interval => INTERVAL '7 days');

-- 为经常进行 UI 端过滤的 device_id (车厢级别) 加速建立索引
CREATE INDEX IF NOT EXISTS ix_fact_raw_device_time 
    ON hvac.fact_raw (device_id, event_time DESC);

-- =============================================================================
-- 构建服务于前端 Web 展现的连续聚合层 (Continuous Aggregates)
-- 目的: 给前端 web-nb67 的 Echart 和总览界面提供 O(1) 的响应
-- =============================================================================

-- 2. 首页与 ECharts 的机组温度一小时平均趋势视图
CREATE MATERIALIZED VIEW hvac.mv_temperature_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', event_time) AS bucket_time,
    train_id,
    carriage_id,
    AVG(ras_u1) AS avg_ras_u1,   -- 回风温度(环境温度) U1
    AVG(ras_u2) AS avg_ras_u2,   -- 回风温度(环境温度) U2
    MAX(fas_u1) AS max_fas_u1,   -- 新风温度 U1
    MAX(fas_u2) AS max_fas_u2    -- 新风温度 U2
FROM hvac.fact_raw
-- 此处使用有效的数据才进入图表聚合
WHERE event_time_valid = true
GROUP BY bucket_time, train_id, carriage_id;

-- 定时拉起持续聚合刷新任务
SELECT add_continuous_aggregate_policy('hvac.mv_temperature_hourly',
  start_offset => INTERVAL '3 days',
  end_offset => INTERVAL '1 hour',
  schedule_interval => INTERVAL '1 hour');
```

---

## 4. API 交互层设计规范

采用 **Google 资源导向风格 (Resource-Oriented Design)** 重新规划前后端契约，并分为“数据流推送”与“资源状态同步”两大赛道。

### 4.1 WebSocket 实时流接入 (Stream API) -> 数据源: Redpanda

*   **Endpoint 动态寻址**: `ws://${window.location.hostname}:[API端口]/v1/streams/telemetry` (或通过 `vite.config.js` 的 proxy 映射)
*   **用途**: **仅用于 1 秒级的高频实时大屏刷新（SVG 机组图和实时故障弹窗）**。
*   **底层查询路径**: WebSocket 服务器 (BFF网关) 作为 Consumer 组的一个节点，**直接订阅 Redpanda 中的 `signal-parsed` 或 `signal-alarm` Topic**，将最新的 JSON 逐条 Push 给前端。不经过数据库，实现零延迟边缘触达。
*   **交互机制**: 前端握手连入并带上 Bearer Token，网关认证后即建立信道响应式更新 Pinia Store。

### 4.2 资源型 API 标准化设计 (RESTful API) -> 数据源: TimescaleDB

*   **用途**: **用于查询历史数据、报表统计、折线图分析、分页列表等所有非实时场景**。
*   **底层查询路径**: API 服务器 (BFF网关) **直接通过 SQL 查询 TimescaleDB 中的超表（如 `hvac.fact_raw` 或 `hvac.mv_temperature_hourly` 聚合视图）**。因为从 Redpanda 查历史数据既不经济也不支持丰富的条件过滤组合。

```typescript
/**
 * @file src/api/hvac.ts
 * @description 根据规范重构的网络请求层。统一 baseUrl，采用标准 REST 语义。
 */
import { request } from '@/utils/request';

// ==========================================
// 【资源名称: Trains 列车】数据源：TimescaleDB
// ==========================================

/**
 * 获取列车即时运行状态快照聚合
 * 【查库】 select * from hvac.fact_raw order by event_time desc limit 1
 */
export const getTrainRunningState = (trainNo: string) => {
    return request.get(`/v1/trains/${trainNo}/running-state`);
};

// ==========================================
// 【资源名称: Alarms 报警信息】数据源：TimescaleDB
// ==========================================

/**
 * 获取报警信息历史分页 (结合日期条件)
 * 【查库】 select * from hvac.alarms where ... limit/offset
 * @param params 分页与筛选入参
 */
export const listAlarms = (params: { 
    trainNo?: string;
    severity?: number;     
    startTime: string; 
    endTime: string;
    page: number;
    pageSize: number;
}) => {
    return request.get('/v1/alarms', { params });
};

// ==========================================
// 【资源名称: Telemetry 遥测指标分析】数据源：TimescaleDB 连续聚合层
// ==========================================

/**
 * 查询空调机组的时间序列趋势 
 * 【查库】 直接命中 TimescaleDB 的 mv_temperature_hourly 等 Continuous Aggregates 缓存视图
 */
export const getUnitTelemetryTrend = (unitId: string, timeGrain: 'hourly' | 'daily') => {
    return request.get(`/v1/units/${unitId}/telemetry:trend`, { 
        params: { bucket: timeGrain } 
    });
};
```

### 4.3 数据访问层 (DAL) 策略：杜绝“高自由度可配置”，采用安全的“参数化构建”

针对“API 是做成完全可配置（通过前端传 JSON 或规则生成 SQL），还是在代码里写死 SQL 语句”的底层选型问题。结合工业物联网系统对安全和性能的严苛要求，这里的**架构铁律**是：

1.  **绝对禁止“全动态可配置查询” (Anti-Pattern)**：
    此前使用 GraphQL (Hasura) 暴露极高自由度的查询配置，正是导致 P0 级安全漏洞与性能不可控的根源。如果让前端在 Payload 中定义需要查哪些表、拼凑什么 WHERE 条件，本质上就是自己重新造了一个有 SQL 注入风险、极度不安全的 Hasura。一旦遭遇恶意拼接，可能打满 TimescaleDB 内存甚至造成删库。
2.  **适度且安全的“条件可选式” SQL (最佳实践)**：
    不应该把 SQL 拼写逻辑暴露给上层（配置），而是**在 API 路由或 Controller 代码内部，利用安全合规的 Query Builder（如 `Knex.js` 或具有极致 TypeScript 类型推导的 `Kysely`）**进行语句拼装，或是使用 `pg` 库的原生**参数化绑定**查库。

基于这一准则，我们要落地的底层查询机制如下：

*   **对于固定且复杂的时序聚合查询（如 ECharts 大盘数据）**：
    在代码中编写**带预编译防注入参数绑定**的原生 SQL（因为需要用到 TimescaleDB 特有的 `time_bucket()` 等高级函数，Query Builder 表达能力不足）。通过入参仅替换列车编号或起止时间，SQL 的“骨架”是写死（Hardcoded）在底层的，最为安全可靠。
*   **对于多条件聚合搜索（如报警记录分页列表）**：
    在代码中使用轻量级的 Query Builder 去动态组凑 WHERE 块。例如前端仅传递 `{ severity: 2, limit: 10 }`，后端代码中隐性拼接 `SELECT * FROM alarms WHERE severity = $1 LIMIT $2`。前端永远无法决定或看到真实的底层 `hvac.alarms` 架构信息。

---

## 5. 前端工程化改造与高效运行方案

为了支撑现代化流式架构的界面展示，前端 `oldproj/web-nb67.250513` 需执行以下几项针对性改造：

### 5.1 从轮询模式切换为全双工流送模式 (Pinia + WebSocket 最佳实践)

很多开发者会觉得前端处理 WebSocket 和增量更新很复杂，但实际上在 Vue 3 (Composition API) 结合 Pinia 的加持下，代码不仅不复杂，反而比原来到处分散的 `setInterval` 和几十次回调要**简洁得多**。

核心思想是：**所有的实时数据只由一个 Pinia Store 集中管理，各个组件只需简单绑定 Store 的状态，由于 Vue 的响应式机制，组件会自动更新，无需任何手动刷新或事件监听逻辑分发。**

以下是一个可落地的实现参考实例 (使用 Vue 3 + TypeScript)：

```typescript
// src/store/useStreamStore.ts
import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useStreamStore = defineStore('stream', () => {
    // 1. 定义集中式的状态 (State)
    // 假设我们有一个包含了所有车厢最新实时指标的字典，键为 device_id
    const deviceTelemetry = ref<Record<string, any>>({});
    // 实时报警队列（限定最多展示最新的 50 条）
    const realtimeAlarms = ref<any[]>([]);

    // 2. WebSocket 连接单例实例
    let ws: WebSocket | null = null;

    // 3. 通道初始化方法
    const connect = () => {
        if (ws) return;
        
        // 动态计算地址
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        ws = new WebSocket(`${protocol}//${window.location.host}/api/v1/streams/telemetry`);
        
        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            
            // 路由策略: 根据网关发来的指令类型进行数据分流更新
            if (data.type === 'telemetry_update') {
               // 场景A: SVG 监控图上的压缩机温度、频率变化
               // Vue 的响应式将导致绑定了该 device_id 的组件原地极速热更
               const device_id = data.payload.device_id;
               deviceTelemetry.value[device_id] = { 
                   ...deviceTelemetry.value[device_id], 
                   ...data.payload 
               };
            } 
            else if (data.type === 'new_alarm') {
               // 场景B: 弹出新的报警弹窗并在右侧列表置顶
               realtimeAlarms.value.unshift(data.payload);
               if (realtimeAlarms.value.length > 50) {
                   realtimeAlarms.value.pop(); // 防止内存溢出的固定长度队列
               }
               // 可以直接在这里触发一个全屏幕的 ElNotification 提醒！
            }
        };

        ws.onclose = () => {
            // 实现断线自动重连机制
            setTimeout(() => { ws = null; connect(); }, 3000);
        };
    };

    return { deviceTelemetry, realtimeAlarms, connect };
});
```

然后在你的**任意组件中 (例如原来几千行的 `trainUnitMapData.vue`)**，你要做的事情变得无比简单：

```vue
<!-- airConditioner/trainUnitMapData.vue -->
<script setup>
import { useStreamStore } from '@/store/useStreamStore'
import { computed, onMounted } from 'vue'

const store = useStreamStore()

// 当挂载主应用时（比如 App.vue 或 Home.vue），只需调用一次 store.connect()
// 这里我们在单独组件里只需读取当前的 device_id。
const props = defineProps(['currentDeviceId'])

// 直接从 Pinia 取响应式数据！告别乱七八糟的 axios 轮询调用和回调地狱！
const u11_fc = computed(() => store.deviceTelemetry[props.currentDeviceId]?.f_cp_u11 || 0)
const u11_ic = computed(() => store.deviceTelemetry[props.currentDeviceId]?.i_cp_u11 || 0)
</script>

<template>
  <div class="hvac-map">
     <!-- 这里的文本内容会在 WebSocket 每秒接收到新数据时，自动以微秒级重绘，毫无闪烁 -->
     <div class="sensor-tag">1号变频机频率: {{ u11_fc }} Hz</div>
     <div class="sensor-tag">1号变频机电流: {{ u11_ic }} A</div>
  </div>
</template>
```

**为什么这种前端架构更好？**
*   **消除 Callback Hell**：曾经你在 `RunState.vue` 等组件要写无数的 `request(...)`，然后针对各种返回格式手工 `this.data = ...`，现在所有网络同步**全部下沉到 Pinia 层**。组件退化为纯洁的 UI Render（只负责画视图），逻辑异常清晰。
*   **极致的性能**：你不用每隔几分钟用全量 HTTP 取回长达几百 KB 全部未变动的大 JSON 了。WS 连接中，后端网关只用下发变动的几十字节小包。
*   **不再有重复查询**：以前每个组件各拉各的数据，甚至出现了“调用同一 API 返回同样的数据赋值两遍”的地步（我在你之前提交的代码审计中提到过）。现在全局只有单一长链接，带宽利用绝佳。

---

## 6. BFF 网关工程规约与高可读性代码骨架

为了满足“代码结构清晰、可读性好、便于交接和长期迭代”的核心诉求，无论是前端重构还是后端的全新 Node.js BFF (Backend-For-Frontend) 网关代码，都将强推 **“关注点分离分层架构 (Layered Architecture)”**。严禁将路由定义、业务鉴权、SQL 拼接和数据传输揉杂在一份长文件里。

### 6.1 预推荐的 API 服务端精悍目录结构

BFF 项目（比如我们命名为 `web-nb67-bff` 或放在老项目的 `server/` 目录下）将具备极其清晰的业务流向边界：

```text
server/                # (新) BFF 后端 API 网关
├── src/
│   ├── config/        # 环境与常量配置 (ENV/数据库密钥等)
│   ├── router/        # 仅负责挂载 URL、HTTP METHOD (如 /v1/alarms -> GET)
│   ├── controller/    # [控制器层] 验证外部入参规范，并转发命令给 service，组装 JSON 响应
│   ├── service/       # [业务逻辑层] 流媒体数据格式化、权限鉴权、加工复杂算法
│   ├── repository/    # [数据访问层 DAL] 存放原生且安全的 SQL，只返回纯数据查询结果
│   ├── streams/       # [流式处理区] 独立存放 WebSocket 的广播池与 Redpanda/Kafkajs 监听逻辑
│   ├── utils/         # 通用工具函数 (时间处理、加解密)
│   └── index.ts       # 唯一主入口 (Mount Fastify 实例及中间件)
├── packge.json
├── tsconfig.json
└── eslint.config.js
```

### 6.2 代码编写准则与整洁度承诺

1.  **TypeScript 强类型推导**：
    全面摒弃 `.js` 裸写（无异于代码坟场）。在数据访问层 (`repository`) 返回时，显式挂载接口如 `: Promise<TelemetryRow[]>`；API 接收时采用如 TypeBox 或 Zod 验证 Payload 类型。当你用 VSCode 悬停在任何变量上，都能知道这个机组数据是 Number 还是 Boolean。
2.  **胖 Repository，瘦 Controller**：
    所有与 TimescaleDB 打交道的 `$1` 传参拼接，全部死死封装在 `repository` 的对应函数里。Controller 只有几行清爽的调用流：`const data = await hvacRepo.getAlarms(limit)`。
3.  **单向依赖原则**：
    `Router -> Controller -> Service -> Repository`。下层永远不知道上层的存在，`Service` 永远不去直接碰 HTTP（不关心返回的是 Res 还是 404 Header）。这种规矩带来的好处是，将来如果加新业务或者测 Bug，直接能精确定位到对应的层。
4.  **Google 级代码注释**：
    所有关键业务模块上部加入 `/** ... */` JSDoc 块级注释。特别是对于类似 `HighpressU11`（U11高压）这类工业黑话，变量旁必须用中文详释字段实际意义。
*   当前 `request.js` 中的 `timeout: 10000000` (1千万毫秒) 属于极度危险的死锁配置坑位。
*   **改造**: `timeout` 缩减为工业标准的 `15000` 毫秒 (15秒)。
*   在全局 `axios.interceptors.response` 拦截器中，必须添加异常捕获 (`error.response.status`) 并使用 `ElMessage.error()` 提供友好的人机交互指引。

### 5.3 构建链路与代码健壮性升级
*   随着业务逻辑转变为处理繁杂的时序与监控字段，强烈建议引入 **TypeScript**，以接口类型严格声明遥测对象 (例: `interface TelemetryPayload`)，避免诸如原先代码里 `dwcf_op_cnt_u11` 张冠李戴导致数据链崩塌的恶性 Bug。
*   摒弃全局粗放型的 `unplugin-auto-import` 自动导包方式，或者强制为其生成清晰的 `.d.ts` 类型补丁文件，以增强开发阶段编辑器侧的代码追溯可维护能力。

### 5.4 容器化部署架构与 API 技术栈选型 (BFF + SPA)

针对“如何部署 API 网关与 Web 前端”以及“它们是否在一端”的疑问。在结合了你的运行环境（纯 IP 及私有化部署）后，最佳的系统落地策略如下：

**选型结论**：**采用“同容器部署的 Node.js BFF (Backend For Frontend) 架构”。** Web（Vue 的静态分发）和 API（REST & WebSocket 服务端）将被打包进 **同一个 Docker 容器中**。

#### A: 为什么合并成一个容器？
1. **解决纯 IP 的 CORS 跨域痛点**：在无域名、纯局域网 IP 交付下，分开部署带来的跨域甚至 WSS 协议握手将极其头疼（比如前端跑在 80，后端 3000）。合并入同一个 Node 服务后，前端打包的 `dist/` 直接由 Node 取用做静态托管，所有的 `/api` 与 `ws://` 直接由该 Node 进程路由，完全消灭跨域问题，**一个内部端口解决所有服务**。
2. **边缘轻量化**：这个新加入的 API 层定位不是重型业务后端，它是一个纯粹的数据承接口 (BFF 层)，只负责代理请求查 TimescaleDB / 订阅 Redpanda，用单容器极为经济。

#### B: 后端 (API) 技术栈建议
*   **运行时**: **Node.js (LTS Version 20+)**
*   **Web 框架**: **Fastify** 或 **Express** (推荐 Fastify，性能更强且原生支持好 WebSocket)
*   **库与中间件**: 
    1. `pg` (PostgreSQL 客户端，直连 TimescaleDB)
    2. `kafkajs` (极轻量的 JS Kafka Client，负责作为 Consumer 消费 Redpanda 中的 `signal-alarm` 发射 WebSocket)
    3. `ws` (处理 WebSocket 握手及长连接通道管理)

#### C: 全栈运行链路图
当项目交付并在你现成的 `docker-compose-Dev.yml` 跑起来时，实际上该容器承担了“中间件路由器”和“前端服务器”的双栖作用：

```text
[浏览器 (192.168.x.x:8080)]
   │
   ├─ (1) 请求 /index.html ──> [Node.js Fastify] 返回 Vue3 的 SPA 静态产物资源
   │
   ├─ (2) 请求 GET /api/v1/alarms ──> [Node.js Fastify] 
   │                                     └─> 执行 SQL 查询 [TimescaleDB]
   │
   └─ (3) 连接 ws://.../telemetry ──> [Node.js Fastify 长连接握手成功] 
                                         └─> Kafkajs 进程挂起，持续听取 [Redpanda] 并向外 push
```

#### D: 任务落地方案拆解 (研发计划计划)
1. **[Phase 1] 数据库定稿与初始化** (已由 DDL 表结构得出)
2. **[Phase 2] Node.js BFF 项目脚手架搭建**:
    - 初始化 `package.json` 并引入 Fastify, PG, Kafkajs 等依赖。
    - 编写 RESTful 接口进行对 TimescaleDB 历史数据的 CRUD 测试。
    - 编写 WebSocket 并利用 Kafkajs 跑通 Redpanda Stream 消费的闭环。
3. **[Phase 3] 原有 Vue 项目向新型 Store 及 API 迁移**:
    - 废弃原 `setInterval` 代码，向 Vue/Pinia + WebSockets 注入 `dist`，完成组件响应式交互绑定及改造。
4. **[Phase 4] Dockerfile 编写**:
    - 利用多阶段构建 (Multi-stage Build)。阶段一半是 Node.js 组配前端，使用 `npm run build` 出静态包；另一半使用极其精简的 Alpine Node 镜像，将静态包通过服务端框架伺服，暴露出单一 `8080` 端口供局域网直接访问。
