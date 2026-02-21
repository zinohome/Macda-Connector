import type { Generated } from 'kysely';

/**
 * TimescaleDB 中的 hvac.fact_raw 表结构定义
 * 与 baseEnv/init-db/01-init.sql 严格对齐
 */
export interface FactRawTable {
    event_time: Date;
    ingest_time: Date;
    line_id: number;
    train_id: number;
    carriage_id: number;
    device_id: string;

    // 核心环境参数
    tveh_1: number | null;
    tveh_2: number | null;
    humdity_1: number | null;
    humdity_2: number | null;

    // 空气质量
    aq_t_u1: number | null;
    aq_h_u1: number | null;
    aq_co2_u1: number | null;
    aq_tvoc_u1: number | null;
    aq_pm2_5_u1: number | null;
    aq_pm10_u1: number | null;

    // 机组压缩机数据
    f_cp_u11: number | null;
    i_cp_u11: number | null;
    v_cp_u11: number | null;
    p_cp_u11: number | null;
    suckt_u11: number | null;
    highpress_u11: number | null;

    // 故障状态
    blpflt_comp_u11: boolean | null;
    bscflt_comp_u11: boolean | null;
    bscflt_vent_u11: boolean | null;

    // 元数据
    parser_version: string;
    quality_code: number;
    payload_json: any; // JSONB
}

/**
 * 数据库 Schema 统一入口
 */
export interface Database {
    'hvac.fact_raw': FactRawTable;
}
