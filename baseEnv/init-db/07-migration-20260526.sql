-- =============================================================================
-- Migration: 2026-05-26
-- 变更内容：
--   H1. 新建 hvac.alarm_code_ref 表，存储 alertcode_v2.xlsx PHM v2 规范的告警码
--       HVAC127-HVAC175（49个故障位告警码）
--   H2. 修正 warning_config 中空气质量预警条目：AQ_CO2 触发方式注释补充
--   H3. 更新 hvac.fact_event 中旧格式告警码为 HVAC 格式（若存在历史数据）
-- 说明：所有语句幂等，可重复执行
-- =============================================================================

-- ----------------------------------------------------------------------------
-- H1. 新建 hvac.alarm_code_ref：PHM v2 告警码参考表
--     用于前端显示、BFF 查询时将 HVAC 码还原为中文描述
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS hvac.alarm_code_ref (
    seq          INTEGER       NOT NULL,            -- 告警序号（27-75）
    hvac_code    VARCHAR(16)   NOT NULL,            -- 平台码（HVAC127-HVAC175）
    location     VARCHAR(64)   NOT NULL,            -- 部件位置
    fault_name   VARCHAR(128)  NOT NULL,            -- 故障名称
    phm_signal   VARCHAR(64),                       -- PHM 信号名（alertcode_v2.xlsx name 列）
    level        SMALLINT      NOT NULL DEFAULT 2,  -- 严重程度：1=严重 2=一般 3=轻微
    PRIMARY KEY (seq)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_alarm_code_ref_hvac_code
    ON hvac.alarm_code_ref (hvac_code);

COMMENT ON TABLE hvac.alarm_code_ref IS 'PHM v2 告警码参考表（alertcode_v2.xlsx HVAC127-HVAC175）';

-- 插入 49 个告警码（幂等）
INSERT INTO hvac.alarm_code_ref (seq, hvac_code, location, fault_name, phm_signal, level)
VALUES
    (27, 'HVAC127', '通风机1-1过流', '通风机1-1过流故障', 'bOCFlt_EF_U11', 1),
    (28, 'HVAC128', '通风机1-2过流', '通风机1-2过流故障', 'bOCFlt_EF_U12', 1),
    (29, 'HVAC129', '冷凝风机1-1过流', '冷凝风机1-1过流故障', 'bOCFlt_CF_U11', 2),
    (30, 'HVAC130', '冷凝风机1-2过流', '冷凝风机1-2过流故障', 'bOCFlt_CF_U12', 2),
    (31, 'HVAC131', '变频器1-1', '变频器1-1故障', 'bFlt_VFD_U11', 2),
    (32, 'HVAC132', '压缩机1-1低压', '压缩机1-1低压故障', 'bLPFlt_Comp_U11', 2),
    (33, 'HVAC133', '压缩机1-1高压连锁', '压缩机1-1高压连锁故障', 'bSCFlt_Comp_U11', 2),
    (34, 'HVAC134', '变频器1-2', '变频器1-2故障', 'bFlt_VFD_U12', 2),
    (35, 'HVAC135', '压缩机1-2低压', '压缩机1-2低压故障', 'bLPFlt_Comp_U12', 2),
    (36, 'HVAC136', '压缩机1-2高压连锁', '压缩机1-2高压连锁故障', 'bSCFlt_Comp_U12', 2),
    (37, 'HVAC137', '新风阀U1', '新风阀U1故障', 'bFlt_FAD_U1', 3),
    (38, 'HVAC138', '回风阀U1', '回风阀U1故障', 'bFlt_RAD_U1', 3),
    (39, 'HVAC139', '空气净化U1', '空气净化U1故障', 'bFlt_AirClean_U1', 3),
    (40, 'HVAC140', '扩展模块U1', '扩展模块U1故障', 'bFlt_ExpBoard_U1', 2),
    (41, 'HVAC141', '新风温度传感器U1', '新风温度传感器U1故障', 'bFlt_FrsTemp_U1', 3),
    (42, 'HVAC142', '送风温度传感器1-1', '送风温度传感器1-1故障', 'bFlt_SplyTemp_U11', 3),
    (43, 'HVAC143', '送风温度传感器1-2', '送风温度传感器1-2故障', 'bFlt_SplyTemp_U12', 3),
    (44, 'HVAC144', '回风温度传感器U1', '回风温度传感器U1故障', 'bFlt_RntTemp_U1', 3),
    (45, 'HVAC145', '融霜温度传感器1-1', '融霜温度传感器1-1故障', 'bFlt_DFSTemp_U11', 3),
    (46, 'HVAC146', '融霜温度传感器1-2', '融霜温度传感器1-2故障', 'bFlt_DFSTemp_U12', 3),
    (47, 'HVAC147', '通风机2-1过流', '通风机2-1过流故障', 'bOCFlt_EF_U21', 1),
    (48, 'HVAC148', '通风机2-2过流', '通风机2-2过流故障', 'bOCFlt_EF_U22', 1),
    (49, 'HVAC149', '冷凝风机2-1过流', '冷凝风机2-1过流故障', 'bOCFlt_CF_U21', 2),
    (50, 'HVAC150', '冷凝风机2-2过流', '冷凝风机2-2过流故障', 'bOCFlt_CF_U22', 2),
    (51, 'HVAC151', '变频器2-1', '变频器2-1故障', 'bFlt_VFD_U21', 2),
    (52, 'HVAC152', '压缩机2-1低压', '压缩机2-1低压故障', 'bLPFlt_Comp_U21', 2),
    (53, 'HVAC153', '压缩机2-1高压连锁', '压缩机2-1高压连锁故障', 'bSCFlt_Comp_U21', 2),
    (54, 'HVAC154', '变频器2-2', '变频器2-2故障', 'bFlt_VFD_U22', 2),
    (55, 'HVAC155', '压缩机2-2低压', '压缩机2-2低压故障', 'bLPFlt_Comp_U22', 2),
    (56, 'HVAC156', '压缩机2-2高压连锁', '压缩机2-2高压连锁故障', 'bSCFlt_Comp_U22', 2),
    (57, 'HVAC157', '新风阀U2', '新风阀U2故障', 'bFlt_FAD_U2', 3),
    (58, 'HVAC158', '回风阀U2', '回风阀U2故障', 'bFlt_RAD_U2', 3),
    (59, 'HVAC159', '空气净化U2', '空气净化U2故障', 'bFlt_AirClean_U2', 3),
    (60, 'HVAC160', '扩展模块U2', '扩展模块U2故障', 'bFlt_ExpBoard_U2', 2),
    (61, 'HVAC161', '新风温度传感器U2', '新风温度传感器U2故障', 'bFlt_FrsTemp_U2', 3),
    (62, 'HVAC162', '送风温度传感器2-1', '送风温度传感器2-1故障', 'bFlt_SplyTemp_U21', 3),
    (63, 'HVAC163', '送风温度传感器2-2', '送风温度传感器2-2故障', 'bFlt_SplyTemp_U22', 3),
    (64, 'HVAC164', '回风温度传感器U2', '回风温度传感器U2故障', 'bFlt_RntTemp_U2', 3),
    (65, 'HVAC165', '融霜温度传感器2-1', '融霜温度传感器2-1故障', 'bFlt_DFSTemp_U21', 3),
    (66, 'HVAC166', '融霜温度传感器2-2', '融霜温度传感器2-2故障', 'bFlt_DFSTemp_U22', 3),
    (67, 'HVAC167', '车厢温度传感器1', '车厢温度传感器1故障', 'bFlt_VehTemp', 3),
    (68, 'HVAC168', '车厢温度传感器2', '车厢温度传感器2故障', 'bFlt_SeatTemp', 3),
    (69, 'HVAC169', '紧急逆变器', '紧急逆变器故障', 'bFlt_EmergIVT', 1),
    (70, 'HVAC170', '变频器1-1通讯', '变频器1-1通讯故障', 'bComuFlt_VFD_U11', 2),
    (71, 'HVAC171', '变频器1-2通讯', '变频器1-2通讯故障', 'bComuFlt_VFD_U12', 2),
    (72, 'HVAC172', '变频器2-1通讯', '变频器2-1通讯故障', 'bComuFlt_VFD_U21', 2),
    (73, 'HVAC173', '变频器2-2通讯', '变频器2-2通讯故障', 'bComuFlt_VFD_U22', 2),
    (74, 'HVAC174', '机组1供电', '机组1供电故障', 'bMCBFlt_Pwr_U1', 1),
    (75, 'HVAC175', '机组2供电', '机组2供电故障', 'bMCBFlt_Pwr_U2', 1)
ON CONFLICT (seq) DO UPDATE SET
    hvac_code  = EXCLUDED.hvac_code,
    location   = EXCLUDED.location,
    fault_name = EXCLUDED.fault_name,
    phm_signal = EXCLUDED.phm_signal,
    level      = EXCLUDED.level;

-- ----------------------------------------------------------------------------
-- H2. 补充预警码参考视图，便于 BFF/前端查询完整码表
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW hvac.v_predict_code_ref AS
SELECT
    seq,
    'HVAC' || (100 + seq) AS hvac_code,
    CASE seq
        WHEN 1 THEN '系统1冷媒泄漏预警'
        WHEN 2 THEN '系统2冷媒泄漏预警'
        WHEN 3 THEN '制冷系统预警'
        WHEN 4 THEN '新风温度传感器预警'
        WHEN 5 THEN '回风温度传感器预警'
        WHEN 6 THEN '车厢温度超温预警'
        WHEN 7 THEN '滤网脏堵预警'
        WHEN 8 THEN '通风机1电流预警'
        WHEN 9 THEN '通风机2电流预警'
        WHEN 10 THEN '冷凝风机1电流预警'
        WHEN 11 THEN '冷凝风机2电流预警'
        WHEN 12 THEN '废排风机电流预警'
        WHEN 13 THEN '压缩机1电流预警'
        WHEN 14 THEN '压缩机2电流预警'
        WHEN 15 THEN '空气质量预警'
    END AS fault_name,
    CASE seq
        WHEN 1 THEN 'WARN_REFRIGERANT_LEAK_COOLING'
        WHEN 2 THEN 'WARN_REFRIGERANT_LEAK_COOLING'
        WHEN 3 THEN 'WARN_COOLING_SYSTEM'
        WHEN 4 THEN 'WARN_TEMP_SENSOR'
        WHEN 5 THEN 'WARN_TEMP_SENSOR'
        WHEN 6 THEN 'WARN_CABIN_OVERHEAT'
        WHEN 7 THEN 'WARN_FILTER_CLOG'
        WHEN 8 THEN 'WARN_EF_CURRENT'
        WHEN 9 THEN 'WARN_EF_CURRENT'
        WHEN 10 THEN 'WARN_CF_CURRENT'
        WHEN 11 THEN 'WARN_CF_CURRENT'
        WHEN 12 THEN 'WARN_EXUF_CURRENT'
        WHEN 13 THEN 'WARN_CP_CURRENT'
        WHEN 14 THEN 'WARN_CP_CURRENT'
        WHEN 15 THEN 'WARN_AQ_CO2'
    END AS warn_code
FROM generate_series(1, 15) AS seq;

COMMENT ON VIEW hvac.v_predict_code_ref IS
    'PHM v2 预警码 HVAC101-115 与 warning_config.warn_code 映射关系';

-- ----------------------------------------------------------------------------
-- H3. 历史数据更新：将 fact_event 中旧格式告警码迁移为 HVAC 格式
--     仅影响 event_type='alarm' 且 fault_code 以 'b' 开头（旧格式）的记录
-- ----------------------------------------------------------------------------
-- 注：旧格式为小写 b 开头的字段名（如 blpflt_comp_u11），新格式为 HVAC{n}。
--     以下映射基于 alertcode_v2.xlsx。carriage_id 用于确定 HVAC 前缀。
--     仅对存量数据执行，新数据由 connect-nb67 v2 直接生成正确格式。

DO $$
DECLARE
    mapping RECORD;
BEGIN
    -- 仅在存在旧格式数据时执行
    IF EXISTS (SELECT 1 FROM hvac.fact_event WHERE event_type = 'alarm' AND fault_code ~ '^b') THEN
        FOR mapping IN
            SELECT field, seq FROM (VALUES
                ('blpflt_comp_u11', 32), ('bscflt_comp_u11', 33),
                ('blpflt_comp_u12', 35), ('bscflt_comp_u12', 36),
                ('blpflt_comp_u21', 52), ('bscflt_comp_u21', 53),
                ('blpflt_comp_u22', 55), ('bscflt_comp_u22', 56),
                ('bflt_vfd_u11', 31), ('bflt_vfd_u12', 34),
                ('bflt_vfd_u21', 51), ('bflt_vfd_u22', 54),
                ('bflt_vfd_com_u11', 70), ('bflt_vfd_com_u12', 71),
                ('bflt_vfd_com_u21', 72), ('bflt_vfd_com_u22', 73),
                ('bocflt_ef_u11', 27), ('bocflt_ef_u12', 28),
                ('bocflt_ef_u21', 47), ('bocflt_ef_u22', 48),
                ('bocflt_cf_u11', 29), ('bocflt_cf_u12', 30),
                ('bocflt_cf_u21', 49), ('bocflt_cf_u22', 50),
                ('bflt_powersupply_u1', 74), ('bflt_powersupply_u2', 75),
                ('bflt_emergivt', 69),
                ('bflt_expboard_u1', 40), ('bflt_expboard_u2', 60),
                ('bflt_frstemp_u1', 41), ('bflt_frstemp_u2', 61),
                ('bflt_splytemp_u11', 42), ('bflt_splytemp_u12', 43),
                ('bflt_splytemp_u21', 62), ('bflt_splytemp_u22', 63),
                ('bflt_rnttemp_u1', 44), ('bflt_rnttemp_u2', 64),
                ('bflt_coiltemp_u11', 45), ('bflt_coiltemp_u12', 46),
                ('bflt_coiltemp_u21', 65), ('bflt_coiltemp_u22', 66),
                ('bflt_vehtemp_u1', 67), ('bflt_vehtemp_u2', 68),
                ('bflt_ap_u11', 39), ('bflt_ap_u21', 59)
            ) AS t(field, seq)
        LOOP
            UPDATE hvac.fact_event
            SET fault_code = 'HVAC' || (carriage_id * 100 + mapping.seq)
            WHERE event_type = 'alarm'
              AND fault_code = mapping.field;
        END LOOP;
        RAISE NOTICE '历史告警码迁移完成（旧格式 → HVAC 格式）';
    END IF;
END $$;
