# 指标显示缩放参考

> **重要**：以下所有指标在前端/BFF 展示时，**显示值 = 原始值 ÷ 10**。
> DB / Kafka 中存储的是原始整数值，展示层必须除以 10 才能还原真实物理量。

来源：`history.repository.ts:TREND_PARAM_DEFS`、`web-nb67-bff/src/index.ts`、`trainUnitMapData.vue:formatVal()`、`nb67-connect.yaml`

---

## ÷10 指标列表

### 温度类

| 指标名（协议/DB字段） | 中文名 | 除数 |
|---|---|:---:|
| `Tic` / `tic` | 客室温度 | 10 |
| `FasU1` / `fas_u1` | 新风温度（机组1） | 10 |
| `FasU2` / `fas_u2` | 新风温度（机组2） | 10 |
| `RasU1` / `ras_u1` | 回风温度（机组1） | 10 |
| `RasU2` / `ras_u2` | 回风温度（机组2） | 10 |
| `SpU11` / `sp_u11` | 设定温度（机组1系统1） | 10 |
| `SpU21` / `sp_u21` | 设定温度（机组2系统1） | 10 |
| `SasU11` / `sas_u11` | 送风温度1（机组1） | 10 |
| `SasU12` / `sas_u12` | 送风温度2（机组1） | 10 |
| `SasU21` / `sas_u21` | 送风温度1（机组2） | 10 |
| `SasU22` / `sas_u22` | 送风温度2（机组2） | 10 |
| `SucktU11` / `suckt_u11` | 吸气温度（机组1系统1） | 10 |
| `SucktU12` / `suckt_u12` | 吸气温度（机组1系统2） | 10 |
| `SucktU21` / `suckt_u21` | 吸气温度（机组2系统1） | 10 |
| `SucktU22` / `suckt_u22` | 吸气温度（机组2系统2） | 10 |
| `Tveh1` / `tveh_1` | 车厢温度1 ① | 10 |
| `Tveh2` / `tveh_2` | 车厢温度2 ① | 10 |
| `Humdity1` / `humdity_1` | 车厢湿度1 ① | 10 |
| `Humdity2` / `humdity_2` | 车厢湿度2 ① | 10 |
| `AqTU1` / `aq_t_u1` | 空气质量温度（机组1） ① | 10 |
| `AqHU1` / `aq_h_u1` | 空气质量湿度（机组1） ① | 10 |

### 压力类

| 指标名（协议/DB字段） | 中文名 | 除数 |
|---|---|:---:|
| `SuckpU11` / `suckp_u11` | 吸气压力（机组1系统1） | 10 |
| `SuckpU12` / `suckp_u12` | 吸气压力（机组1系统2） | 10 |
| `SuckpU21` / `suckp_u21` | 吸气压力（机组2系统1） | 10 |
| `SuckpU22` / `suckp_u22` | 吸气压力（机组2系统2） | 10 |
| `HighpressU11` / `highpress_u11` | 高压（机组1系统1） | 10 |
| `HighpressU12` / `highpress_u12` | 高压（机组1系统2） | 10 |
| `HighpressU21` / `highpress_u21` | 高压（机组2系统1） | 10 |
| `HighpressU22` / `highpress_u22` | 高压（机组2系统2） | 10 |
| `PresdiffU1` / `presdiff_u1` | 滤网压差（机组1） | 10 |
| `PresdiffU2` / `presdiff_u2` | 滤网压差（机组2） | 10 |

### 电气类（电流 / 电压 / 频率）

| 指标名（协议/DB字段） | 中文名 | 除数 |
|---|---|:---:|
| `ICpU11` / `i_cp_u11` | 压缩机1电流（机组1） | 10 |
| `ICpU12` / `i_cp_u12` | 压缩机2电流（机组1） | 10 |
| `ICpU21` / `i_cp_u21` | 压缩机1电流（机组2） | 10 |
| `ICpU22` / `i_cp_u22` | 压缩机2电流（机组2） | 10 |
| `VCpU11` / `v_cp_u11` | 压缩机1电压（机组1） | 10 |
| `VCpU12` / `v_cp_u12` | 压缩机2电压（机组1） | 10 |
| `VCpU21` / `v_cp_u21` | 压缩机1电压（机组2） | 10 |
| `VCpU22` / `v_cp_u22` | 压缩机2电压（机组2） | 10 |
| `FCpU11` / `f_cp_u11` | 压缩机1频率（机组1） | 10 |
| `FCpU12` / `f_cp_u12` | 压缩机2频率（机组1） | 10 |
| `FCpU21` / `f_cp_u21` | 压缩机1频率（机组2） | 10 |
| `FCpU22` / `f_cp_u22` | 压缩机2频率（机组2） | 10 |
| `ICfU11` / `i_cf_u11` | 冷凝风机1电流（机组1） | 10 |
| `ICfU12` / `i_cf_u12` | 冷凝风机2电流（机组1） | 10 |
| `ICfU21` / `i_cf_u21` | 冷凝风机1电流（机组2） | 10 |
| `ICfU22` / `i_cf_u22` | 冷凝风机2电流（机组2） | 10 |
| `IEfU11` / `i_ef_u11` | 通风机1电流（机组1） | 10 |
| `IEfU12` / `i_ef_u12` | 通风机2电流（机组1） | 10 |
| `IEfU21` / `i_ef_u21` | 通风机1电流（机组2） | 10 |
| `IEfU22` / `i_ef_u22` | 通风机2电流（机组2） | 10 |

**共 43 个指标，全部为 ÷10。**

> ① 仅在已废弃的 `nb67-connect.yaml` 中出现。当前主流水线（三段分拆配置）中 `Tveh`/`Humdity`/`AqT`/`AqH` 以原始值存入 `payload_json.raw`，BFF 展示层尚未做缩放处理，后续新增展示时需同步加上 ÷10。

---

## 不需要缩放的指标

以下指标 DB 存储值即为显示值，**无需缩放**：

| 指标名 | 中文名 | 单位 |
|---|---|---|
| `AqCo2` | CO₂浓度 | ppm |
| `AqTvoc` | TVOC 浓度 | ppb |
| `AqPm25` | PM2.5 | μg/m³ |
| `AqPm10` | PM10 | μg/m³ |
| `WmodeU` | 工作模式 | 枚举 |
| `FadposU` | 新风阀开度 | % |
| `RadposU` | 回风阀开度 | % |
| `EevposU` | 膨胀阀开度 | 步 |

---

## 代码中的缩放实现位置

- **趋势图**：`web-nb67-bff/src/history.repository.ts` — `TREND_PARAM_DEFS` 中每个参数的 `scale` 字段
- **列车单元地图**：`web-nb67-web/src/views/trainUnitMapData.vue` — `formatVal()` 函数
- **BFF API 响应**：`web-nb67-bff/src/index.ts` — 返回前对原始值应用缩放
- **协议定义**（已废弃参考）：`connect/nb67-connect.yaml`

新增展示指标时，请确认该字段是否在上表中，并在以上对应位置同步添加 `÷ 10` 处理。
