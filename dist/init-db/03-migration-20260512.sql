-- =============================================================================
-- Migration: 2026-05-12
-- 变更内容：
--   E1. hvac.warning_config 补充 params.raw_scale 字段
--       供 Go event-builder 的 ConfigStore 做单位换算：
--       raw_threshold = trigger_value × raw_scale
--   E2. 修正 duration_seconds = 0 的行（trigger_value 已正确，只补 duration）
-- 说明：所有语句幂等，可重复执行
-- =============================================================================

-- ----------------------------------------------------------------------------
-- E1. 为需要 ×10 换算的预警码补充 raw_scale = 10
--     trigger_value 存 UI 显示单位（A / ℃ / pa / bar），
--     传感器原始数据单位为 0.1A / 0.1℃ / 0.1pa / 0.1bar，
--     故 raw_threshold = trigger_value × 10
-- ----------------------------------------------------------------------------
UPDATE hvac.warning_config
SET params = COALESCE(params, '{}'::jsonb) || '{"raw_scale": 10}'::jsonb
WHERE warn_code IN (
    'WARN_EF_CURRENT',       -- 通风机 1.8A → 18 (0.1A 单位)
    'WARN_CF_CURRENT',       -- 冷凝风机 2.3A → 23
    'WARN_EXUF_CURRENT',     -- 废排风机 2.3A → 23
    'WARN_CP_CURRENT',       -- 压缩机 18A → 180
    'WARN_CABIN_OVERHEAT',   -- 车厢超温 4℃ → 40 (0.1℃)
    'WARN_FILTER_CLOG',      -- 滤网压差 300pa → 3000 (0.1pa)
    'WARN_REFRIGERANT_LEAK', -- 冷媒泄露 2.0bar → 20 (0.1bar)
    'WARN_TEMP_SENSOR',      -- 温度传感器 8℃ → 80
    'WARN_COOLING_SYSTEM'    -- 制冷系统电流差 2A → 20
);

-- ----------------------------------------------------------------------------
-- E2. 修正 duration_seconds = 0 的行（初次插入时未写入持续时间）
--     以下值来自 PHM 文档，与硬编码保持一致
-- ----------------------------------------------------------------------------
UPDATE hvac.warning_config SET duration_seconds = 600
WHERE warn_code IN ('WARN_EF_CURRENT', 'WARN_CF_CURRENT', 'WARN_EXUF_CURRENT', 'WARN_CP_CURRENT')
  AND duration_seconds = 0;

UPDATE hvac.warning_config SET duration_seconds = 120
WHERE warn_code = 'WARN_CABIN_OVERHEAT' AND duration_seconds = 0;

UPDATE hvac.warning_config SET duration_seconds = 1800
WHERE warn_code = 'WARN_FILTER_CLOG' AND duration_seconds = 0;

UPDATE hvac.warning_config SET duration_seconds = 300
WHERE warn_code IN ('WARN_REFRIGERANT_LEAK', 'WARN_TEMP_SENSOR') AND duration_seconds = 0;

UPDATE hvac.warning_config SET duration_seconds = 180
WHERE warn_code = 'WARN_COOLING_SYSTEM' AND duration_seconds = 0;

UPDATE hvac.warning_config SET duration_seconds = 900
WHERE warn_code = 'WARN_AQ_CO2' AND duration_seconds = 0;

UPDATE hvac.warning_config SET duration_seconds = 1200
WHERE warn_code IN ('WARN_AQ_PM25', 'WARN_AQ_PM10', 'WARN_AQ_TVOC') AND duration_seconds = 0;
