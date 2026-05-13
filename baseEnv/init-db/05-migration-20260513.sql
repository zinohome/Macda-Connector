-- =============================================================================
-- Migration: 2026-05-13 (patch 2)
-- 变更内容：
--   G1. hvac.warning_config 新增 default_* 列，存出厂默认值（Reset 功能用）
--   G2. 按 PHM 预警码表修正各条目 strategy（处理建议）
--   G3. 按 PHM 码表修正错误阈值：
--       - WARN_EF_CURRENT: 1.8→2.0A
--       - WARN_CF_CURRENT: 2.3→2.9A
--       - WARN_EXUF_CURRENT: 2.3→4.0A
--       - WARN_AQ_CO2: 4500→1200ppm, duration 900→1200s
--   G4. 写入出厂默认值（PHM 文档值）到 default_* 列
--   G5. 测试模式：降低阈值 + duration=0，让现有 mock 数据尽可能触发预警
--       WARN_CABIN_OVERHEAT params 增加 min_cooling_runtime_s=0（测试用，正式设为 1200）
-- 说明：所有语句幂等，可重复执行
-- =============================================================================

-- ----------------------------------------------------------------------------
-- G1. 新增 default_* 列（已存在则忽略）
-- ----------------------------------------------------------------------------
ALTER TABLE hvac.warning_config
    ADD COLUMN IF NOT EXISTS default_trigger_value    NUMERIC(12,3),
    ADD COLUMN IF NOT EXISTS default_clear_value      NUMERIC(12,3),
    ADD COLUMN IF NOT EXISTS default_duration_seconds INTEGER,
    ADD COLUMN IF NOT EXISTS default_threshold_good   VARCHAR(64),
    ADD COLUMN IF NOT EXISTS default_threshold_normal VARCHAR(64),
    ADD COLUMN IF NOT EXISTS default_threshold_bad    VARCHAR(64),
    ADD COLUMN IF NOT EXISTS default_strategy         TEXT;

-- ----------------------------------------------------------------------------
-- G2 + G3. 按 PHM 码表修正阈值与 strategy
-- ----------------------------------------------------------------------------

-- 3.1 冷媒泄漏 (WARN_REFRIGERANT_LEAK)
UPDATE hvac.warning_config SET
    strategy = '1、检查风阀开启状态\n2、检查风机运行电流及工作状态\n3、确认制冷系统运行压力\n4、检查电子膨胀阀系统'
WHERE warn_code = 'WARN_REFRIGERANT_LEAK';

-- 3.2 制冷系统预警 (WARN_COOLING_SYSTEM)
UPDATE hvac.warning_config SET
    strategy = '1、检查风机及风阀\n2、确认压缩机运行状态\n3、检查确认制冷系统冷媒是否泄漏\n4、检查电子膨胀阀系统'
WHERE warn_code = 'WARN_COOLING_SYSTEM';

-- 3.3 温度传感器 (WARN_TEMP_SENSOR)
UPDATE hvac.warning_config SET
    strategy = '1、检查温度传感器接线是否良好\n2、重新上电复位，查看故障是否消除\n3、更换新传感器'
WHERE warn_code = 'WARN_TEMP_SENSOR';

-- 3.4 车厢超温 (WARN_CABIN_OVERHEAT)
UPDATE hvac.warning_config SET
    strategy = '1、检查2个机组的制冷系统是否异常'
WHERE warn_code = 'WARN_CABIN_OVERHEAT';

-- 3.5 滤网脏堵 (WARN_FILTER_CLOG)
UPDATE hvac.warning_config SET
    strategy = '1、清洗混合风滤网\n2、如果滤网不脏，需要检查微压差传感器是否正常'
WHERE warn_code = 'WARN_FILTER_CLOG';

-- 3.6 通风机电流 (WARN_EF_CURRENT) — PHM: >2.0A（原 1.8A）
UPDATE hvac.warning_config SET
    trigger_value = 2.0,
    clear_value   = 1.8,
    threshold_good   = '≤1.8A',
    threshold_normal = '1.8~2.0A',
    threshold_bad    = '>2.0A',
    strategy = '1、检查风阀状态\n2、检查滤网是否安装\n3、检查风机轴承'
WHERE warn_code = 'WARN_EF_CURRENT';

-- 3.7 冷凝风机电流 (WARN_CF_CURRENT) — PHM: >2.9A（原 2.3A）
UPDATE hvac.warning_config SET
    trigger_value = 2.9,
    clear_value   = 2.5,
    threshold_good   = '≤2.5A',
    threshold_normal = '2.5~2.9A',
    threshold_bad    = '>2.9A',
    strategy = '1、检查风阀状态\n2、检查滤网是否安装\n3、检查风机轴承'
WHERE warn_code = 'WARN_CF_CURRENT';

-- 3.8 废排风机电流 (WARN_EXUF_CURRENT) — PHM: >4.0A（原 2.3A）
UPDATE hvac.warning_config SET
    trigger_value = 4.0,
    clear_value   = 3.5,
    threshold_good   = '≤3.5A',
    threshold_normal = '3.5~4.0A',
    threshold_bad    = '>4.0A',
    strategy = '1、检查风阀状态\n2、检查滤网是否安装\n3、检查风机轴承'
WHERE warn_code = 'WARN_EXUF_CURRENT';

-- 3.9 压缩机电流 (WARN_CP_CURRENT)
UPDATE hvac.warning_config SET
    strategy = '1、检查冷凝风机\n2、检查冷凝器\n3、检查压缩机'
WHERE warn_code = 'WARN_CP_CURRENT';

-- 3.10 空气质量 CO2 — PHM: >1200ppm（原 4500ppm），持续 20min（原 15min）
UPDATE hvac.warning_config SET
    trigger_value    = 1200.0,
    clear_value      = 1200.0,
    duration_seconds = 1200,
    threshold_good   = '<1200ppm',
    threshold_bad    = '≥1200ppm',
    strategy = '1、检查新风阀开度是否正常\n2、检查车厢内空气环境'
WHERE warn_code = 'WARN_AQ_CO2';

UPDATE hvac.warning_config SET
    strategy = '1、检查新风阀开度是否正常\n2、检查车厢内空气环境'
WHERE warn_code IN ('WARN_AQ_PM25', 'WARN_AQ_PM10', 'WARN_AQ_TVOC');

-- ----------------------------------------------------------------------------
-- G4. 写入出厂默认值（PHM 修正后的值）到 default_* 列
--     仅写入尚未设置的行（default_trigger_value IS NULL 时），保证幂等
-- ----------------------------------------------------------------------------
UPDATE hvac.warning_config SET
    default_trigger_value    = trigger_value,
    default_clear_value      = clear_value,
    default_duration_seconds = duration_seconds,
    default_threshold_good   = threshold_good,
    default_threshold_normal = threshold_normal,
    default_threshold_bad    = threshold_bad,
    default_strategy         = strategy
WHERE default_trigger_value IS NULL;

-- ----------------------------------------------------------------------------
-- G5. 测试模式：降低阈值 + duration=0
--     目的：让现有 mock 帧（ras_u1=278, ras_u2=211, fas_u1=219, fas_u2=221）
--     尽可能触发预警，方便验证管道连通性。
--     通过 Reset 按钮（PUT /api/rest/warning/config/:id/reset）恢复出厂值。
--
--     能触发的预警（依赖现有数据）：
--       HVAC_07: |FasU1-FasU2|=2 > 1 ✓
--       HVAC_08: |RasU1-RasU2|=67 > 60 ✓
--       HVAC_09: RasU1=278 > 270=(26+1.0)×10 ✓（需 min_cooling_runtime_s=0）
--
--     无法通过阈值触发（mock 帧传感器为 0 或无效值）：
--       HVAC_01~04: f_cp=0（压缩机未运行），无法满足频率>30Hz 条件
--       HVAC_05~06: f_cp=0，压缩机未运行
--       HVAC_10~11: presdiff=32767（无效值），代码过滤掉
--       HVAC_12~20: 需要风机运行且电流>阈值，mock 帧中风机电流未知
--       HVAC_21~24: i_cp=0（压缩机电流为 0）
--       HVAC_125~126: aq_co2/pm25/pm10/tvoc 全为 0
-- ----------------------------------------------------------------------------

-- WARN_TEMP_SENSOR: 触发阈值从 8℃ 降到 0.1℃，duration=0 → HVAC_07/08 立即触发
UPDATE hvac.warning_config SET
    trigger_value    = 0.1,
    duration_seconds = 0
WHERE warn_code = 'WARN_TEMP_SENSOR';

-- WARN_CABIN_OVERHEAT: 目标26℃ + delta 1.0℃ → 绝对阈值 270，低于 ras_u1=278
-- 同时将 min_cooling_runtime_s 设为 0，跳过 20min 前置等待
UPDATE hvac.warning_config SET
    trigger_value    = 1.0,
    duration_seconds = 0,
    params = COALESCE(params, '{}'::jsonb)
             || '{"min_cooling_runtime_s": 0}'::jsonb
WHERE warn_code = 'WARN_CABIN_OVERHEAT';
