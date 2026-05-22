-- =============================================================================
-- Migration: 2026-05-21
-- 变更内容：
--   F1. 将 WARN_REFRIGERANT_LEAK 拆分为两个独立可配置条目
--       - WARN_REFRIGERANT_LEAK_COOLING：制冷模式触发条件（吸气压力）
--       - WARN_REFRIGERANT_LEAK_VENT：通风模式触发条件（高压）
--   F2. 为 WARN_COOLING_SYSTEM 补充 raw_scale=10，使阈值可热加载生效
-- 说明：所有语句幂等，可重复执行
-- =============================================================================

-- ----------------------------------------------------------------------------
-- F0. 删除旧的合并条目 WARN_REFRIGERANT_LEAK（已拆分为 _COOLING 和 _VENT 两个独立配置）
-- ----------------------------------------------------------------------------
DELETE FROM hvac.warning_config WHERE warn_code = 'WARN_REFRIGERANT_LEAK';

-- ----------------------------------------------------------------------------
-- F1a. 冷媒泄露预警 - 制冷模式条件（单独可配置）
-- 触发：制冷/弱冷模式 AND 压缩机频率>30Hz AND 吸气压力 < trigger_value bar
-- trigger_value=2.0, raw_scale=10 → Go 代码原始值 = 20（即 2.0bar×10）
-- duration_seconds=300 → 持续5分钟
-- ----------------------------------------------------------------------------
INSERT INTO hvac.warning_config
    (warn_code, component_name, category, threshold_good, threshold_normal, threshold_bad,
     trigger_operator, trigger_value, clear_value, duration_seconds, unit, strategy, params)
SELECT 'WARN_REFRIGERANT_LEAK_COOLING', '冷媒泄露预警（制冷模式）', 'pressure',
       '吸气≥2.0bar', '—', '吸气<2.0bar',
       '<', 2.0, 2.0, 300,
       'bar',
       '制冷模式下冷媒管路可能泄漏，请采用手持式卤素仪检测漏点位置并安排检修',
       '{"raw_scale": 10, "precondition": "cooling_or_weak_cooling AND fcp_gt_30hz",
         "description": "制冷模式：压缩机频率>30Hz且吸气压力<trigger_value bar，持续duration_seconds/60分钟"}'::jsonb
WHERE NOT EXISTS (SELECT 1 FROM hvac.warning_config WHERE warn_code = 'WARN_REFRIGERANT_LEAK_COOLING');

-- ----------------------------------------------------------------------------
-- F1b. 冷媒泄露预警 - 通风模式条件（单独可配置）
-- 触发：通风模式 AND 高压 < trigger_value bar
-- trigger_value=5.0, raw_scale=10 → Go 代码原始值 = 50（即 5.0bar×10）
-- duration_seconds=900 → 持续15分钟
-- ----------------------------------------------------------------------------
INSERT INTO hvac.warning_config
    (warn_code, component_name, category, threshold_good, threshold_normal, threshold_bad,
     trigger_operator, trigger_value, clear_value, duration_seconds, unit, strategy, params)
SELECT 'WARN_REFRIGERANT_LEAK_VENT', '冷媒泄露预警（通风模式）', 'pressure',
       '高压≥5bar', '—', '高压<5bar',
       '<', 5.0, 5.0, 900,
       'bar',
       '通风模式下冷媒管路可能泄漏，请采用手持式卤素仪检测漏点位置并安排检修',
       '{"raw_scale": 10,
         "description": "通风模式：高压<trigger_value bar，持续duration_seconds/60分钟"}'::jsonb
WHERE NOT EXISTS (SELECT 1 FROM hvac.warning_config WHERE warn_code = 'WARN_REFRIGERANT_LEAK_VENT');

-- ----------------------------------------------------------------------------
-- F2. 为 WARN_COOLING_SYSTEM 补充 raw_scale=10
-- trigger_value=2.0A，Go 代码原始值=20（即 2.0A×10）
-- 只在 raw_scale 尚未设置时更新
-- ----------------------------------------------------------------------------
UPDATE hvac.warning_config
SET params = jsonb_set(COALESCE(params, '{}'), '{raw_scale}', '10')
WHERE warn_code = 'WARN_COOLING_SYSTEM'
  AND NOT (params ? 'raw_scale');
