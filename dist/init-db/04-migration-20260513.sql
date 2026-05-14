-- =============================================================================
-- Migration: 2026-05-13
-- 变更内容：
--   F1. WARN_CABIN_OVERHEAT 修正超温预警阈值语义（按 PHM 文档）
--       原问题：Go 代码把 trigger_value(4.0) 当绝对温度（4℃）比较，永远满足；
--               比较字段用 Tveh（车厢温度），PHM 文档要求用 RasU（回风温度）；
--               制冷模式只判断 mode==2，漏掉弱冷 mode==3。
--       本次修正：
--         - params 补充 target_temp=26.0（制冷目标温度℃，可按现场调整）
--         - trigger_value 保持 4.0（PHM 文档：回风温度 > 目标温度+4℃）
--         - Go 代码使用 (target_temp + trigger_value) × raw_scale = (26+4)×10 = 300
--         - 比较字段改为 RasU1/RasU2（回风温度）
--         - 制冷模式检查改为 mode==2(强冷) OR mode==3(弱冷)
--   F2. 更新 deploy.sh 中 Step 3c / Step 6 的 migration 列表（文件操作由 deploy.sh 维护）
-- 说明：所有语句幂等，可重复执行
-- =============================================================================

-- ----------------------------------------------------------------------------
-- F1. 修正 WARN_CABIN_OVERHEAT 阈值
--     target_temp = 26.0℃（地铁制冷常用目标温度，可按实际现场调整）
--     trigger_value = 1.0℃（超出目标温度1℃即预警，满足用户需求）
--     absolute_threshold = (26.0 + 1.0) × 10 = 270（0.1℃单位）
-- ----------------------------------------------------------------------------
UPDATE hvac.warning_config
SET trigger_value    = 4.0,
    clear_value      = 3.0,
    threshold_good   = '目标温度±2℃',
    threshold_normal = '目标温度+2~4℃',
    threshold_bad    = '>目标温度+4℃',
    params           = COALESCE(params, '{}'::jsonb)
                       || '{"raw_scale": 10, "target_temp": 26.0}'::jsonb
WHERE warn_code = 'WARN_CABIN_OVERHEAT';
