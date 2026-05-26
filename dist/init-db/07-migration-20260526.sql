-- =============================================================================
-- Migration: 2026-05-26
-- 变更内容：
--   H1. 恢复 WARN_TEMP_SENSOR 和 WARN_CABIN_OVERHEAT 被测试模式覆盖的阈值
--
-- 背景：05-migration-20260513.sql 的 G5 "测试模式" 节故意将
--   WARN_TEMP_SENSOR trigger_value 从 8.0 降至 0.1（duration=0），
--   WARN_CABIN_OVERHEAT trigger_value 从 4.0 降至 1.0（duration=0）。
--   该变更提交到了仓库并应用于生产，导致#8 新/回风温度传感器预警误报。
--
-- 本次 migration 将两者恢复为 PHM 文档定义的正式阈值，
-- 并保持 default_* 列（出厂值）不变。
-- 所有语句幂等，可重复执行。
-- =============================================================================

-- H1. 恢复 WARN_TEMP_SENSOR：8.0℃，持续 300s
UPDATE hvac.warning_config SET
    trigger_value    = 8.0,
    duration_seconds = 300
WHERE warn_code = 'WARN_TEMP_SENSOR';

-- H2. 恢复 WARN_CABIN_OVERHEAT：4.0℃，持续 1200s，去除测试用 min_cooling_runtime_s=0
UPDATE hvac.warning_config SET
    trigger_value    = 4.0,
    duration_seconds = 1200,
    params = params - 'min_cooling_runtime_s'
WHERE warn_code = 'WARN_CABIN_OVERHEAT'
  AND (params->>'min_cooling_runtime_s')::int = 0;
