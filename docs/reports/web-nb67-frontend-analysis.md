# web-nb67 前端应用完整分析报告

> **分析范围**: `oldproj/web-nb67.250513/src/`  
> **分析时间**: 2025-01  
> **文件数量**: 31 个 Vue 组件 + 4 个基础设施文件  
> **技术栈**: Vue 3 (Composition API) + Vue Router + Pinia + Element Plus + ECharts + Axios

---

## 目录

1. [技术架构概览](#1-技术架构概览)
2. [逐页功能拆解](#2-逐页功能拆解)
3. [完整数据字段清单](#3-完整数据字段清单)
4. [API 端点汇总](#4-api-端点汇总)
5. [图表与可视化清单](#5-图表与可视化清单)
6. [代码质量问题与技术债务](#6-代码质量问题与技术债务)
7. [组件依赖关系图](#7-组件依赖关系图)

---

## 1. 技术架构概览

### 1.1 技术栈

| 层 | 技术 | 版本/备注 |
|---|---|---|
| 框架 | Vue 3 | `<script setup>` Composition API |
| 路由 | Vue Router | `createWebHistory`, 基路径 `/KT` |
| 状态管理 | Pinia | 仅存储 `carNum` 和 `carCrew` |
| UI 库 | Element Plus | 表格、标签、弹窗、日期选择器、分页等 |
| 图表 | ECharts | 柱状图、折线图 |
| HTTP | Axios | 自定义封装 `request.js` |
| 构建 | Vite | (推断) |

### 1.2 请求配置

```
baseURL: ''（空，使用相对路径/代理）
timeout: 10,000,000 ms（异常高值）
headers: { 'x-hasura-admin-secret': 'passw0rd' }
响应拦截: 直接返回 response.data
```

### 1.3 业务领域

**地铁/城轨列车空调（HVAC）监控系统** — NB67 线路

- **列车编组**: 6 节车厢 → `Tc1, Mp1, M1, M2, Mp2, Tc2`
- **每节车厢**: 2 套空调机组（1号机组 / 2号机组）
- **车厢 ID 构成**: `列车号 + '0' + 车厢序号`（如 `1234501`）

### 1.4 路由结构

| 路径 | 名称 | 组件 | 说明 |
|---|---|---|---|
| `/` | — | redirect → `/KT` | 根重定向 |
| `/KT` | `home` | `home/index.vue` | 总览首页 |
| `/KT/traininfo` | `trainInfo` | `trainInfo/index.vue` | 列车详情（新版） |
| `/KT/trainDetail/:trainNo` | `trainDetail` | `trainDetail/index.vue` | 列车详情（旧版） |
| `/KT/airConditioner/:trainCoach/:trainNum/:trainCrew` | `airConditioner` | `airConditioner/index.vue` | 机组详情 |
| `/KT/carCrewDetail/:trainNo` | `carCrewDetail` | `carCrewDetail/index.vue` | 车厢机组详情（旧版） |

### 1.5 轮询机制

所有数据页面使用 `setInterval` 每 **2 分钟** 刷新数据，在 `onUnmounted` 时清除定时器。

---

## 2. 逐页功能拆解

### 2.1 首页 — `home/` (5 个组件)

**路由**: `/KT`  
**入口**: `home/index.vue` (1372 行)

#### 布局结构
```
┌─────────────────────────────────────────┐
│  Left (6列)  │  Center (12列)  │ Right (6列) │
├─────────────────────────────────────────┤
│           AirErrorData (24列)            │
└─────────────────────────────────────────┘
```

#### 子组件

| 组件 | 功能 | 数据来源 |
|---|---|---|
| `Left.vue` | 列车列表表格：列车号、状态(正常/异常)、报警数、预警数、操作链接 | `getAirSystemApi()` → `vw_train_alarm_count` |
| `Center.vue` | SVG 背景 + 三个统计盒子：在线正常列车、在线预警列车、在线报警列车 | 父组件计算传入 |
| `Right.vue` | 两个表格：实时报警、实时预警（车厢号、名称、时间、指导建议弹窗） | `getRealtimeAlarm()` + `getRealtimeWarning()` |
| `AirErrorData.vue` (1061行) | 空调故障统计 — ECharts 水平柱状图 + 筛选 | `getFaultStatistics()` → `vw_alarm_info_all_date` |

#### 数据流
1. `getAirSystemApi()` → 获取列车列表和报警计数 → 填充 Left + Center
2. `getRealtimeAlarm({trainNo})` → 实时报警 → 与 `propose[]` 故障码表匹配 → 填充 Right 报警表
3. `getRealtimeWarning()` → 实时预警 → 与 `proposeWarn[]` 预警码表匹配 → 填充 Right 预警表
4. `getFaultStatistics()` → 故障统计 → ECharts 柱状图

#### 导航
- Left.vue 点击"操作"链接 → 导航到 `trainInfo` 页面（传入 `trainNo`）

---

### 2.2 列车信息页（新版）— `trainInfo/` (8 个组件)

**路由**: `/KT/traininfo?trainNo=xxx`  
**入口**: `trainInfo/index.vue` (1678 行 — 项目中最大的文件)

#### 布局结构
```
┌──────────────────────────────────────────┐
│  Header: 当前列车号 + 返回按钮             │
├──────────────┬───────────────────────────┤
│ ActualWarning │      StateWarning        │
│  (实时报警)    │      (实时预警)           │
├──────────────┴───────────────────────────┤
│          CarCrew (车厢选择 SVG 图)         │
├──────────────────────────────────────────┤
│       AlarmInfo (报警信息表格+筛选)         │
├──────────────────────────────────────────┤
│       RunState (运行状态信息表格)           │
├──────────────────────────────────────────┤
│     CarTemperature (车厢温度信息表格)       │
└──────────────────────────────────────────┘
```

#### 子组件

| 组件 | 行数 | 功能 |
|---|---|---|
| `ActualWarning.vue` | ~80 | 实时报警表格：车厢号、报警名称、报警时间、指导建议弹窗 |
| `StateWarning.vue` | ~80 | 实时预警表格：车厢号、预警名称、报警时间、指导建议弹窗 |
| `CarCrew.vue` | 419 | 车厢选择 — SVG 列车图（头车/中间车/尾车图片），状态图标（绿/橙/红），点击跳转空调详情 |
| `AlarmInfo.vue` | 557 | 报警信息表格 + 日期范围选择 + 分页，故障类型标签（轻微/正常/中等/严重/预警） |
| `RunState.vue` | 303 | 运行状态信息表格：6车×2机组=12列，6行设备 |
| `CarTemperature.vue` | ~80 | 车厢温度信息表格：3行×6列（车厢内温度、环境温度、目标温度） |
| `Dropdown.vue` | ~100 | 车厢/机组选择下拉菜单（12选项），当前未使用 |

#### 数据流（5 个 API 调用）
1. `getRealtimeAlarm({trainNo})` → 实时报警数据 → 匹配 `propose[]` → ActualWarning
2. `getStatusAlert(trainNo)` → 状态预警 → 匹配 `propose[]` + `proposeWarn[]` → StateWarning
3. `getAlarmInformation({state, startTime, endTime})` → 遍历 6 车厢 → AlarmInfo
4. `getRunningState(trainNo)` → `vw_running_state_info` → 运行状态数据 → RunState
5. `getTrainTemperature(trainNo)` → `vw_traintemperature` → 温度数据 → CarTemperature

#### CarCrew 导航
- 点击某节车厢 → `router.push({ name: 'airConditioner', query: { trainNo, trainCoach } })`

#### 运行状态设备行
| 行 | API 字段 (U1x) | API 字段 (U2x) |
|---|---|---|
| 压缩机1 | `cfbk_comp_u11` | `cfbk_comp_u21` |
| 压缩机2 | `cfbk_comp_u12` | `cfbk_comp_u22` |  
| 通风机 | `cfbk_ef_u11` | `cfbk_ef_u21` |
| 冷凝风机 | `cfbk_cf_u11` | `cfbk_cf_u21` |
| 新风阀开度 | `fadpos_u1` | `fadpos_u2` |
| 回风阀开度 | `radpos_u1` | `radpos_u2` |

#### 温度行
| 行 | API 字段 |
|---|---|
| 车厢内温度 | `ras_sys` |
| 目标温度 | `tic` |
| 环境温度 | `fas_sys` |

---

### 2.3 列车详情页（旧版）— `trainDetail/` (4 个组件)

**路由**: `/KT/trainDetail/:trainNo`  
**入口**: `trainDetail/index.vue` (655 行)  
**⚠️ 使用旧版 API** (`getAirStartDataApi`, `getTrainDataApi`, `getEaDataBydvcAddrApi`)

#### 布局结构
```
┌──────────────────────────────────────────┐
│  BackTop + 列车号                         │
├──────────────────────────────────────────┤
│  空调系统状态（状态图标说明弹窗）            │
├──────────────────────────────────────────┤
│  运行状态信息 (RunState)                   │
├──────────────────────────────────────────┤
│  车厢温度信息 (Temperature)                │
├──────────────────────────────────────────┤
│  报警信息 (Alarm: title="报警信息")         │
├──────────────────────────────────────────┤
│  预警信息 (Alarm: title="预警信息")         │
└──────────────────────────────────────────┘
```

#### 子组件

| 组件 | 行数 | 功能 |
|---|---|---|
| `Alarm.vue` | ~500 | 报警/预警通用表格 + ECharts 信号详情对话框（折线图）+ TitleHeader 日期筛选 + 分页 |
| `RunState.vue` | ~150 | 8列设备运行状态表格（1号-8号机组），`stateName()` 判断阀门开度(%)或运行/停机 |
| `Temperature.vue` | ~100 | 跨车厢温度表格（TC1~TC2），NaN 值显示 N/A |

#### 空调运行模式 `getAirState()`
| 代码 | 模式 |
|---|---|
| 0 | 停机 |
| 5 | 通风 |
| 6 | 弱冷 |
| 7 | 强冷 |
| 8 | 弱暖 |
| 9 | 强暖 |
| 10 | 紧急通风 |
| 12 | 除霜 |

#### Alarm.vue 信号详情对话框
- 点击"详细数据"→ 弹出 ECharts 折线图
- 调用 `getEaOrPaById({id})` 获取信号时序数据
- **每 5 秒**自动刷新信号数据
- 支持单位：℃（温度）、Hz（频率）、A（电流）、V（电压）、bar（压力）、K（过热度）、%（开度）

#### 旧版 API 字段映射
| 设备 | API 字段 |
|---|---|
| 压缩机 | `last_cFBK_Comp_U` |
| 通风机 | `last_cFBK_EF_U` |
| 冷凝风机 | `last_cFBK_CF_U` |
| 新风阀开度 | `last_wPosFAD_U` |
| 回风阀开度 | `last_wPosRAD_U` |
| 车厢温度 | `last_iInnerTemp` (÷10) |
| 环境温度 | `last_iOuterTemp` (÷10) |
| 目标温度 | `last_iSetTemp` (÷10) |

---

### 2.4 空调机组详情页 — `airConditioner/` (7 个组件)

**路由**: `/KT/airConditioner?trainNo=xxx&trainCoach=yyy`  
**入口**: `airConditioner/index.vue` (624 行)

#### 布局结构
```
┌──────────────────────────────────────────┐
│  BackTop + 当前机组: Dropdown              │
├──────────────────────────────────────────┤
│  Healthy (健康评估信息表格)                 │
├──────────────────────────────────────────┤
│  TrainUnitMap → trainUnitMapData          │
│  (空调机组详情 — SVG 结构示意图)            │
├──────────────────────────────────────────┤
│  Echart (温度趋势折线图)                   │
└──────────────────────────────────────────┘
```

#### 子组件

| 组件 | 行数 | 功能 |
|---|---|---|
| `Dropdown.vue` | ~170 | 车厢/机组选择器（12选项），调用 `getTrainSelection` 获取车厢 ID 映射 |
| `Healthy.vue` | ~180 | 健康评估信息表格：设备名称、健康状态(标签)、累计运行时间、动作次数、指导建议 |
| `TrainUnitMap.vue` | ~180 | 空调机组详情容器（包裹 trainUnitMapData） |
| `trainUnitMapData.vue` | 750 | **空调机组 SVG 示意图** — 在 SVG 背景图上叠加实时数据点（绝对定位），hover 显示标签 |
| `System.vue` | ~60 | 通用参数表格（已被注释掉未使用） |
| `Echart.vue` | 333 | 温度趋势折线图（ECharts），支持时/天/月切换 |

#### 健康评估逻辑 (`getData()`)

调用 `getHealthAssessment(carriageID)` → `vw_health_assessment`

**8 个被评估设备**:
| 设备 | 运行时间字段 (U1) | 运行时间字段 (U2) | 动作次数字段 (U1) | 动作次数字段 (U2) | 寿命阈值 |
|---|---|---|---|---|---|
| 通风机 | `dwef_op_tm_u11` | `dwef_op_tm_u21` | `dwef_op_cnt_u11` | `dwef_op_cnt_u21` | 25,000h |
| 冷凝机 | `dwcf_op_tm_u11` | `dwcf_op_tm_u21` | `dwcf_op_cnt_u11` | `dwcf_op_cnt_u11`⚠️ | 25,000h |
| 新风阀 | — (计次) | — (计次) | `dwfad_op_cnt_u1` | `dwfad_op_cnt_u2` | 1,000,000次 |
| 回风阀 | — (计次) | — (计次) | `dwrad_op_cnt_u1` | `dwrad_op_cnt_u2` | 1,000,000次 |
| 压缩机1 | `dwcp_op_tm_u11` | `dwcp_op_tm_u21` | `dwcp_op_cnt_u11` | `dwcp_op_cnt_u21` | 50,000h |
| 压缩机2 | `dwcp_op_tm_u12` | `dwcp_op_tm_u22` | `dwcp_op_cnt_u12` | `dwcp_op_cnt_u22` | 50,000h |
| 废排风机 | `dwexufan_op_tm` | `dwexufan_op_tm` | `dwexufan_op_cnt` | `dwexufan_op_cnt` | 25,000h |
| 废排风阀 | — (计次) | — (计次) | `dwdmpexu_op_cnt` | `dwdmpexu_op_cnt` | 1,000,000次 |

> ⚠️ Bug: 2号机组冷凝机的 `number_actions` 错误引用了 `dwcf_op_cnt_u11`（应为 `dwcf_op_cnt_u21`）

**健康等级判定**:
| 等级 | 条件（运行时间类） | 条件（动作次数类） | 建议 |
|---|---|---|---|
| 健康 | < 寿命×70% | < 次数×70% | — |
| 亚健康 | 70%~85% | 70%~85% | 设备处于亚健康状态，即将进入检修期 |
| 非健康 | > 85% | > 85% | 更换相应部件 |

#### trainUnitMapData.vue — SVG 数据点（30+ 个实时数据项）

调用 `gettemperature(carriageID)` → `vw_system_info[0]`

| 数据点 | 字段 (U1) | 字段 (U2) | 单位 |
|---|---|---|---|
| 冷凝风机1电流 | `i_cf_u11` | `i_cf_u21` | A |
| 冷凝风机2电流 | `i_cf_u12` | `i_cf_u22` | A |
| 压缩机1电流 | `i_cp_u11` | `i_cp_u21` | A |
| 压缩机1电压 | `v_cp_u11` | `v_cp_u21` | V |
| 压缩机1频率 | `f_cp_u11` | `f_cp_u21` | Hz |
| 压缩机2电流 | `i_cp_u12` | `i_cp_u22` | A |
| 压缩机2电压 | `v_cp_u12` | `v_cp_u22` | V |
| 压缩机2频率 | `f_cp_u12` | `f_cp_u22` | Hz |
| 系统1高压 | `highpress_u11` | `highpress_u21` | bar |
| 系统2高压 | `highpress_u12` | `highpress_u22` | bar |
| 系统1吸气温度 | `suckt_u11` | `suckt_u21` | ℃ |
| 系统1吸气压力 | `suckp_u11` | `suckp_u21` | bar |
| 系统2吸气温度 | `suckt_u12` | `suckt_u22` | ℃ |
| 系统2吸气压力 | `suckp_u12` | `suckp_u22` | bar |
| 送风温度1 | `sas_u11` | `sas_u21` | ℃ |
| 送风温度2 | `sas_u12` | `sas_u22` | ℃ |
| 通风机1电流 | `i_ef_u11` | `i_ef_u21` | A |
| 通风机2电流 | `i_ef_u12` | `i_ef_u22` | A |
| 系统1膨胀阀开度 | `eevpos_u11` | `eevpos_u21` | 步 |
| 系统2膨胀阀开度 | `eevpos_u12` | `eevpos_u22` | 步 |
| 空气监测终端温度 | `aq_t_u1` | `aq_t_u2` | ℃ |
| 空气监测终端湿度 | `aq_h_u1` | `aq_h_u2` | RH |
| 空气监测终端CO2 | `aq_co2_u1` | `aq_co2_u2` | ppm |
| 空气监测终端TVOC | `aq_tvoc_u1` | `aq_tvoc_u2` | ppb |
| 空气监测终端PM2.5 | `aq_pm2_5_u1` | `aq_pm2_5_u2` | μg/m³ |
| 空气监测终端PM10 | `aq_pm10_u1` | `aq_pm10_u2` | μg/m³ |
| 新风温度 | `fas_u1` | `fas_u2` | ℃ |
| 新风阀开度 | `fadpos_u1` | `fadpos_u2` | % |
| 回风温度 | `ras_u1` | `ras_u2` | ℃ |
| 回风阀开度 | `radpos_u1` | `radpos_u2` | % |
| 滤网前后压差值 | `presdiff_u1` | `presdiff_u2` | par |

#### Echart.vue — 温度趋势

调用 `getThDataByDvcAddrApi({ carriageNo, limit: 5 })` → 返回 `hourly`, `daily`, `monthly` 数据

三条折线：
| 名称 | 字段 |
|---|---|
| 客室温度 | `ras_sys` |
| 新风温度 | `fas_sys` |
| 目标温度 | `tic` |

时间轴切换：时 / 天 / 月

---

### 2.5 车厢机组详情页（旧版）— `carCrewDetail/` (4 个组件)

**路由**: `/KT/carCrewDetail/:trainNo`  
**入口**: `carCrewDetail/index.vue` (614 行)  
**⚠️ 使用旧版 API** + **存在严重代码问题**

#### 布局结构
```
┌──────────────────────────────────────────┐
│  BackTop + 当前列车号 + 号机组             │
├──────────────────────────────────────────┤
│  健康评估信息标题                          │
│  Health (健康评估) ← 实际显示温度数据       │
│  Temperature (温度参数) ← 实际显示温度数据  │
│  System (系统参数) ← 实际显示温度数据       │
└──────────────────────────────────────────┘
```

#### 子组件

| 组件 | 功能 | 问题 |
|---|---|---|
| `Health.vue` | 表格：设备名称、健康状态、累计运行时间、动作次数 | 列标签与数据不匹配（列标签是健康评估，实际数据是温度） |
| `Temperature.vue` | 表格：系统名称、温度参数 | 只显示 M2 列的数据 |
| `System.vue` | 表格：系统名称、系统1参数 | 只显示 M2 列的数据 |

#### ⚠️ 严重问题
1. `setHealthData`, `setTemperatureDataData`, `setSystemDataData` 三个函数**代码完全相同** — 都映射 `last_iInnerTemp`, `last_iOuterTemp`, `last_iSetTemp`
2. 引用了未定义的函数 `setTemperatureData` 和 `setSystemData`（应为 `setTemperatureDataData` 和 `setSystemDataData`）→ **运行时必然报错**
3. 对同一 API `getTrainDataApi` 调用了 **3 次**，返回相同数据
4. 此页面明显是**半成品/废弃页面**

---

### 2.6 公共组件 — `components/` (2 个)

| 组件 | 路径 | 功能 |
|---|---|---|
| `BackTop.vue` | `src/components/BackTop.vue` | 返回按钮 — 左箭头图标 + "返回"文字，`router.go(-1)` |
| `TitleHeader.vue` | `src/components/TitleHeader.vue` (213行) | 日期筛选标题栏 — 快捷按钮（近7天/30天/12个月）+ 日期范围选择器 |

#### TitleHeader 工作机制
- 通过 `inject('provideChooseDay')` 获取父组件回调函数
- 选择日期后调用回调，传递 `{ startTime, endTime }` 参数
- `onMounted` 时自动触发"近7天"查询
- 被 `trainDetail/Alarm.vue` 使用

---

## 3. 完整数据字段清单

### 3.1 视图层数据源（数据库视图）

| 视图名 | 用途 | 使用页面 |
|---|---|---|
| `vw_train_alarm_count` | 列车报警计数汇总 | home |
| `vw_alarm_info_all_date` | 全日期故障统计 | home/AirErrorData |
| `vw_running_state_info` | 运行状态信息 | trainInfo |
| `vw_traintemperature` | 列车温度 | trainInfo |
| `vw_carriage_status` | 车厢状态（报警数） | trainInfo/CarCrew, airConditioner/Dropdown |
| `vw_carriage_predict_status` | 车厢预测状态（预警数） | trainInfo/CarCrew |
| `vw_health_assessment` | 健康评估 | airConditioner |
| `vw_system_info` | 系统信息（机组实时参数） | airConditioner |

### 3.2 温度相关字段

| 字段 | 含义 | 单位 | 来源视图 |
|---|---|---|---|
| `ras_sys` | 车厢内温度（回风温度系统值） | ℃ | `vw_traintemperature` |
| `fas_sys` | 环境温度（新风温度系统值） | ℃ | `vw_traintemperature` |
| `tic` | 目标温度 | ℃ | `vw_traintemperature` |
| `ras_u1/u2` | 回风温度（机组级） | ℃ | `vw_system_info` |
| `fas_u1/u2` | 新风温度（机组级） | ℃ | `vw_system_info` |
| `sas_u11/u12/u21/u22` | 送风温度 | ℃ | `vw_system_info` |
| `i_outer_temp` | 室外温度 | ℃ | `vw_system_info` |
| `i_rat_u1/u2` | 回风温度（旧字段） | ℃ | `vw_system_info` |
| `i_sat_u11/u12/u21/u22` | 送风温度（旧字段） | ℃ | `vw_system_info` |
| `suckt_u11/u12/u21/u22` | 吸气温度 | ℃ | `vw_system_info` |
| `aq_t_u1/u2` | 空气监测终端温度 | ℃ | `vw_system_info` |

### 3.3 压力相关字段

| 字段 | 含义 | 单位 |
|---|---|---|
| `highpress_u11/u12/u21/u22` | 系统高压 | bar |
| `suckp_u11/u12/u21/u22` | 吸气压力 | bar |
| `i_suck_pres_u11/u21/u22` | 吸气压力（旧字段） | bar |
| `presdiff_u1/u2` | 滤网前后压差 | par |

### 3.4 电气相关字段

| 字段 | 含义 | 单位 |
|---|---|---|
| `i_cp_u11/u12/u21/u22` | 压缩机电流 | A |
| `v_cp_u11/u12/u21/u22` | 压缩机电压 | V |
| `f_cp_u11/u12/u21/u22` | 压缩机频率 | Hz |
| `i_cf_u11/u12/u21/u22` | 冷凝风机电流 | A |
| `i_ef_u11/u12/u21/u22` | 通风机电流 | A |
| `w_freq_u11/u12/u21/u22` | 变频器频率 | Hz |
| `w_crnt_u11/u12/u21/u22` | 变频器电流 | A |

### 3.5 阀门/开度相关字段

| 字段 | 含义 | 单位 |
|---|---|---|
| `fadpos_u1/u2` | 新风阀开度 | % |
| `radpos_u1/u2` | 回风阀开度 | % |
| `eevpos_u11/u12/u21/u22` | 膨胀阀开度 | 步 |

### 3.6 空气质量相关字段

| 字段 | 含义 | 单位 |
|---|---|---|
| `aq_h_u1/u2` | 湿度 | RH |
| `aq_co2_u1/u2` | CO2 浓度 | ppm |
| `aq_tvoc_u1/u2` | TVOC 浓度 | ppb |
| `aq_pm2_5_u1/u2` | PM2.5 | μg/m³ |
| `aq_pm10_u1/u2` | PM10 | μg/m³ |

### 3.7 运行状态反馈字段

| 字段 | 含义 | 值域 |
|---|---|---|
| `cfbk_comp_u11/u12/u21/u22` | 压缩机运行反馈 | 0=停机, 1=运行 |
| `cfbk_ef_u11/u21` | 通风机运行反馈 | 0=停机, 1=运行 |
| `cfbk_cf_u11/u21` | 冷凝风机运行反馈 | 0=停机, 1=运行 |

### 3.8 健康评估字段（累计运行数据）

| 字段 | 含义 | 单位 |
|---|---|---|
| `dwef_op_tm_u11/u21` | 通风机运行时间 | 秒 |
| `dwef_op_cnt_u11/u21` | 通风机动作次数 | 次 |
| `dwcf_op_tm_u11/u21` | 冷凝风机运行时间 | 秒 |
| `dwcf_op_cnt_u11/u21` | 冷凝风机动作次数 | 次 |
| `dwcp_op_tm_u11/u12/u21/u22` | 压缩机运行时间 | 秒 |
| `dwcp_op_cnt_u11/u12/u21/u22` | 压缩机动作次数 | 次 |
| `dwfad_op_cnt_u1/u2` | 新风阀动作次数 | 次 |
| `dwrad_op_cnt_u1/u2` | 回风阀动作次数 | 次 |
| `dwexufan_op_tm` | 废排风机运行时间 | 秒 |
| `dwexufan_op_cnt` | 废排风机动作次数 | 次 |
| `dwdmpexu_op_cnt` | 废排风阀动作次数 | 次 |

### 3.9 过热度字段

| 字段 | 含义 | 单位 |
|---|---|---|
| `i_sup_heat_u11/u12/u21/u22` | 过热度 | K |

---

## 4. API 端点汇总

### 4.1 活跃端点（当前页面使用）

| # | 函数名 | 方法 | 路径 | 参数 | 返回数据键 | 使用页面 |
|---|---|---|---|---|---|---|
| 1 | `getAirSystemApi` | GET | `/api/rest/AirSystem` | — | `vw_train_alarm_count` | home |
| 2 | `getRealtimeAlarm` | POST | `/api/rest/v2/train/RealtimeAlarm` | `{trainNo: string[]}` | 报警数组 | home, trainInfo |
| 3 | `getRealtimeWarning` | GET | `/api/rest/RealtimeWarning` | — | 预警数组 | home |
| 4 | `getFaultStatistics` | POST | `/api/rest/v2/FaultStatistics` | `{...}` | `vw_alarm_info_all_date` | home/AirErrorData |
| 5 | `getRealtimeAlarmDetail` | GET | `/api/rest/train/RealtimeAlarm/{trainID}` | path: trainID | — | — |
| 6 | `getStatusAlert` | GET | `/api/rest/train/StatusAlert/{trainID}` | path: trainID | — | trainInfo |
| 7 | `getTrainSelection` | GET | `/api/rest/rain/TrainSelection/{trainId}` ⚠️ | path: trainId | `vw_carriage_status` | trainInfo/CarCrew, airConditioner/Dropdown |
| 8 | `getAlarmInformation` | POST | `/api/rest/train/AlarmInformation` | `{state, startTime, endTime}` | 报警信息数组 | trainInfo |
| 9 | `getRunningState` | GET | `/api/rest/train/RunningState/{trainID}` | path: trainID | `vw_running_state_info` | trainInfo |
| 10 | `getTrainTemperature` | GET | `/api/rest/train/TrainTemperature/{trainID}` | path: trainID | `vw_traintemperature` | trainInfo |
| 11 | `getHealthAssessment` | GET | `/api/rest/carriage/HealthAssessment/{carriageID}` | path: carriageID | `vw_health_assessment` | airConditioner |
| 12 | `gettemperature` | GET | `/api/rest/carriage/SystemInfo/{carriageID}` | path: carriageID | `vw_system_info` | airConditioner |
| 13 | `getSystemFirst` | GET | `/api/rest/carriage/SystemInfo/{carriageID}` | path: carriageID | `vw_system_info` | (重复) |
| 14 | `getSystemSecond` | GET | `/api/rest/carriage/SystemSecond/{carriageID}` | path: carriageID | — | — |
| 15 | `getThDataByDvcAddrApi` | POST | `/api/rest/carriage/TemperatureTrend` | `{carriageNo, limit}` | `{hourly, daily, monthly}` | airConditioner/Echart |
| 16 | `getActiveCarApi` | GET | `/api/rest/train` | — | — | — |
| 17 | `getEaOrPaById` | — | — | `{id}` | 信号时序数据 | trainDetail/Alarm |

### 4.2 旧版端点（Legacy，被 trainDetail 和 carCrewDetail 使用）

| 函数名 | 路径 | 使用页面 |
|---|---|---|
| `getAirStartDataApi` | `/api/airStartData` | trainDetail, carCrewDetail |
| `getTrainDataApi` | (未明确) | trainDetail, carCrewDetail |
| `getEaDataBydvcAddrApi` | (未明确) | trainDetail, carCrewDetail |

### 4.3 未使用/注释掉的端点

`getActiveCarNumApi`, `getCarEaDataApi`, `getCarPaDataApi`, `getEaDataApi`, `getHealthDataApi`, `getNewDataApi`, `getPaDataByDvcAddrApi`, `getStatisDataByDvcAddrApi`

### 4.4 API 问题

| 问题 | 详情 |
|---|---|
| **路径拼写错误** | `getTrainSelection` 路径为 `/api/rest/rain/...`，应为 `/api/rest/train/...` |
| **重复端点** | `gettemperature` 和 `getSystemFirst` 指向相同路径 |
| **安全风险** | Hasura admin secret 硬编码在前端代码中 |
| **超时设置** | 10,000,000 ms ≈ 2.78 小时，显然不合理 |

---

## 5. 图表与可视化清单

### 5.1 ECharts 图表

| # | 位置 | 图表类型 | 数据内容 | 交互 |
|---|---|---|---|---|
| 1 | `home/AirErrorData.vue` | **水平柱状图** | 故障类型百分比统计 | 故障类型筛选、颜色编码 |
| 2 | `airConditioner/Echart.vue` | **折线图** (3线) | 客室温度、新风温度、目标温度趋势 | 时/天/月切换 |
| 3 | `trainDetail/Alarm.vue` | **折线图** (对话框) | 单条信号时序数据（报警/预警详情） | 5秒自动刷新、单位自适应 |

### 5.2 SVG / 自定义可视化

| # | 位置 | 可视化类型 | 说明 |
|---|---|---|---|
| 4 | `home/Center.vue` | **SVG 背景 + 统计盒子** | 列车概览三色统计 |
| 5 | `trainInfo/CarCrew.vue` | **SVG 列车编组图** | 6节车厢示意图（头车/中间车/尾车），状态图标（绿✓/橙△/红✕） |
| 6 | `airConditioner/trainUnitMapData.vue` | **SVG 机组结构图 + 数据覆盖层** | 空调机组示意图背景 + 30+ 个绝对定位的实时数据点，hover 显示标签名 |

### 5.3 Element Plus 表格（数据表格汇总）

| # | 位置 | 表格内容 | 列数 |
|---|---|---|---|
| 7 | `home/Left.vue` | 列车列表 | 4列 |
| 8 | `home/Right.vue` | 实时报警 + 实时预警 | 4列 × 2 |
| 9 | `trainInfo/ActualWarning.vue` | 实时报警 | 4列 |
| 10 | `trainInfo/StateWarning.vue` | 实时预警 | 4列 |
| 11 | `trainInfo/AlarmInfo.vue` | 报警信息（含日期筛选+分页） | 7列 |
| 12 | `trainInfo/RunState.vue` | 运行状态（6车×2机组） | 13列 |
| 13 | `trainInfo/CarTemperature.vue` | 车厢温度 | 7列 |
| 14 | `trainDetail/RunState.vue` | 运行状态（8机组） | 9列 |
| 15 | `trainDetail/Temperature.vue` | 车厢温度 | 9列 |
| 16 | `trainDetail/Alarm.vue` | 报警/预警信息 + 分页 | 7列 |
| 17 | `airConditioner/Healthy.vue` | 健康评估信息 | 5列 |

---

## 6. 代码质量问题与技术债务

### 6.1 严重问题

| 编号 | 类型 | 位置 | 描述 |
|---|---|---|---|
| S1 | **安全漏洞** | `utils/request.js` | Hasura admin secret (`passw0rd`) 硬编码在前端 |
| S2 | **运行时错误** | `carCrewDetail/index.vue` | 调用未定义函数 `setTemperatureData`/`setSystemData` |
| S3 | **数据错误** | `airConditioner/index.vue` | 2号机组冷凝机动作次数引用了1号机组字段 `dwcf_op_cnt_u11` |
| S4 | **API 路径错误** | `api/api.js` | `getTrainSelection` 路径 `/api/rest/rain/...` 少了 `t` |

### 6.2 架构问题

| 编号 | 类型 | 描述 |
|---|---|---|
| A1 | **大量代码重复** | 故障码查找表 (`propose[]`, `proposeWarn[]`) 在 3 个文件中重复（home/index.vue 70项, trainInfo/index.vue 70+项, trainInfo/AlarmInfo.vue 50+项），且各处条目不完全一致 |
| A2 | **新旧 API 并存** | trainInfo 使用新版 REST API，trainDetail/carCrewDetail 使用旧版 API，数据模型不统一 |
| A3 | **半成品页面** | carCrewDetail 页面三个子组件显示完全相同的数据，函数名拼写错误导致无法运行 |
| A4 | **System.vue 注释掉** | airConditioner 中的 System.vue 组件已在模板中被注释，但文件仍保留 |

### 6.3 代码质量问题

| 编号 | 类型 | 描述 |
|---|---|---|
| Q1 | **console.log 泛滥** | 几乎每个文件都有大量调试 console.log 未清理 |
| Q2 | **注释代码堆积** | 多个文件存在大段注释掉的旧代码 |
| Q3 | **超时配置不当** | Axios timeout 设为 10,000,000ms（约2.78小时） |
| Q4 | **列标签不一致** | trainDetail/Temperature.vue 列标签 M4车/MP5车 不匹配 6 车编组 |
| Q5 | **重复 API 调用** | `gettemperature` 和 `getSystemFirst` 指向同一端点 |
| Q6 | **数据转换不一致** | 部分温度字段需÷10，部分直接使用原值（旧版 vs 新版 API） |
| Q7 | **Pinia 几乎未使用** | 全局状态管理仅存储 `carNum` 和 `carCrew`，大部分状态通过 props/emit/provide-inject 传递 |
| Q8 | **无 TypeScript** | 整个项目无类型定义，纯 JavaScript |
| Q9 | **无错误边界** | 无全局错误处理、无加载状态管理（除 ECharts 外） |
| Q10 | **CSS 变量重复定义** | `$s-primary`, `$s-danger` 等 SCSS 变量在多个组件中重复声明 |

### 6.4 CarCrew.vue Bug

当 `trainSelectionData.value.slice(2).length == 0` 时（即后4节车不存在），fallback 逻辑错误地重新赋值 `names` 数组，导致车厢名称映射混乱。

---

## 7. 组件依赖关系图

```
App.vue
└── router-view
    ├── home/index.vue ──────────────────── getAirSystemApi, getRealtimeAlarm, getRealtimeWarning
    │   ├── Left.vue                         → 导航到 trainInfo
    │   ├── Center.vue
    │   ├── Right.vue
    │   └── AirErrorData.vue ──────────────── getFaultStatistics
    │
    ├── trainInfo/index.vue ─────────────── getRealtimeAlarm, getStatusAlert, getAlarmInformation,
    │   │                                    getRunningState, getTrainTemperature
    │   ├── ActualWarning.vue
    │   ├── StateWarning.vue
    │   ├── CarCrew.vue ───────────────────── getTrainSelection
    │   │                                     → 导航到 airConditioner
    │   ├── AlarmInfo.vue
    │   ├── RunState.vue
    │   ├── CarTemperature.vue
    │   └── (Dropdown.vue) ────────────────── 未使用
    │
    ├── trainDetail/index.vue ──────────── getAirStartDataApi, getTrainDataApi, getEaDataBydvcAddrApi
    │   │                                   → 导航到 carCrewDetail
    │   ├── Alarm.vue ──────────────────── getEaOrPaById
    │   │   └── components/TitleHeader.vue   inject('provideChooseDay')
    │   ├── RunState.vue
    │   └── Temperature.vue
    │
    ├── airConditioner/index.vue ──────── getHealthAssessment, gettemperature
    │   ├── components/BackTop.vue
    │   ├── Dropdown.vue ──────────────── getTrainSelection
    │   ├── Healthy.vue
    │   ├── TrainUnitMap.vue
    │   │   └── trainUnitMapData.vue        (SVG 数据覆盖层)
    │   ├── (System.vue) ──────────────── 已注释掉
    │   └── Echart.vue ───────────────── getThDataByDvcAddrApi
    │
    └── carCrewDetail/index.vue ──────── getAirStartDataApi, getTrainDataApi, getEaDataBydvcAddrApi
        │                                  ⚠️ 半成品/运行时报错
        ├── components/BackTop.vue
        ├── Health.vue
        ├── Temperature.vue
        └── System.vue
```

---

## 附录：故障码查找表概览

### 故障分级

| 级别 | 标签颜色 | 说明 |
|---|---|---|
| 轻微故障 | 蓝色 `info` | 不影响运行 |
| 中等故障 | 橙色 `warning` | 需关注 |
| 严重故障 | 红色 `danger` | 需立即处理 |
| 预警 | — | 预防性提示 |

### 故障类型覆盖范围（~70+ 种）

- 压缩机相关：过流、过压、欠压、频率异常、高压保护、低压保护、排气温度超限
- 风机相关：通风机过流、冷凝风机过流、废排风机过流
- 传感器相关：送风温度传感器、回风温度传感器、新风温度传感器、吸气温度传感器、吸气压力传感器、排气温度传感器、差压传感器故障
- 阀门相关：新风阀故障、回风阀故障、膨胀阀故障
- 控制器相关：通讯故障、电源故障、紧急通风、空气监测终端故障、电流监测单元故障
- 环境相关：高温报警、低温报警、CO2 超限、PM2.5 超限

---

*报告结束。此文件位于 `docs/reports/` 目录下，遵循项目文档管理规范。*
