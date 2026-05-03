import { db } from '../config/db.js';
import { sql } from 'kysely';
import { config } from '../config/index.js';

// 参数趋势分析支持的字段列表及其映射配置
// label: 前端显示名, col: DB列名, scale: 原始值÷scale=真实值
export const TREND_PARAM_DEFS: Record<string, { label: string; col: string; scale: number; unit: string }> = {
    ras_u1:      { label: '回风温度(机组1)', col: 'ras_u1',      scale: 10, unit: '℃' },
    ras_u2:      { label: '回风温度(机组2)', col: 'ras_u2',      scale: 10, unit: '℃' },
    fas_u1:      { label: '新风温度(机组1)', col: 'fas_u1',      scale: 10, unit: '℃' },
    fas_u2:      { label: '新风温度(机组2)', col: 'fas_u2',      scale: 10, unit: '℃' },
    tic:         { label: '客室温度',        col: "payload_json->'raw'->>'Tic'", scale: 10, unit: '℃' },
    i_cp_u11:    { label: '压缩机1电流(机组1)', col: 'i_cp_u11', scale: 10, unit: 'A' },
    i_cp_u21:    { label: '压缩机1电流(机组2)', col: 'i_cp_u21', scale: 10, unit: 'A' },
    f_cp_u11:    { label: '压缩机1频率(机组1)', col: 'f_cp_u11', scale: 10, unit: 'Hz' },
    f_cp_u21:    { label: '压缩机1频率(机组2)', col: 'f_cp_u21', scale: 10, unit: 'Hz' },
    suckp_u11:   { label: '吸气压力(机组1系统1)', col: 'suckp_u11', scale: 10, unit: 'bar' },
    suckp_u21:   { label: '吸气压力(机组2系统1)', col: 'suckp_u21', scale: 10, unit: 'bar' },
    highpress_u11: { label: '高压(机组1系统1)', col: 'highpress_u11', scale: 10, unit: 'bar' },
    highpress_u21: { label: '高压(机组2系统1)', col: 'highpress_u21', scale: 10, unit: 'bar' },
    presdiff_u1: { label: '滤网压差(机组1)', col: 'presdiff_u1', scale: 10, unit: 'pa' },
    presdiff_u2: { label: '滤网压差(机组2)', col: 'presdiff_u2', scale: 10, unit: 'pa' },
    aq_co2_u1:   { label: 'CO₂浓度(机组1)',  col: 'aq_co2_u1',  scale: 1,  unit: 'ppm' },
    aq_co2_u2:   { label: 'CO₂浓度(机组2)',  col: 'aq_co2_u2',  scale: 1,  unit: 'ppm' },
    aq_pm2_5_u1: { label: 'PM2.5(机组1)',    col: 'aq_pm2_5_u1', scale: 1, unit: 'ug/m³' },
    aq_pm2_5_u2: { label: 'PM2.5(机组2)',    col: 'aq_pm2_5_u2', scale: 1, unit: 'ug/m³' },
    aq_pm10_u1:  { label: 'PM10(机组1)',     col: 'aq_pm10_u1',  scale: 1, unit: 'ug/m³' },
    aq_pm10_u2:  { label: 'PM10(机组2)',     col: 'aq_pm10_u2',  scale: 1, unit: 'ug/m³' },
    aq_tvoc_u1:  { label: 'TVOC(机组1)',     col: 'aq_tvoc_u1',  scale: 1, unit: 'ppb' },
    aq_tvoc_u2:  { label: 'TVOC(机组2)',     col: 'aq_tvoc_u2',  scale: 1, unit: 'ppb' },
};

export class HistoryRepository {
    private static get timeCol() {
        return config.runtime === 'DEV' ? 'ingest_time' : 'event_time';
    }

    // B6: 多参数趋势查询（params 为 TREND_PARAM_DEFS 的 key 数组）
    static async getParamTrend(
        trainId: number,
        carriageId: number,
        type: string,
        params: string[]
    ) {
        const typeMap: Record<string, { bucketSize: string; lookback: string }> = {
            hour:  { bucketSize: '10 seconds', lookback: '1 hour' },
            day:   { bucketSize: '1 minute',   lookback: '24 hours' },
            week:  { bucketSize: '15 minutes', lookback: '168 hours' },
            month: { bucketSize: '30 minutes', lookback: '720 hours' },
        };
        const { bucketSize, lookback } = typeMap[type] || typeMap.hour;

        // 只保留白名单内的参数，防止 SQL 注入
        const validParams = params.filter(p => TREND_PARAM_DEFS[p]);
        if (validParams.length === 0) validParams.push('ras_u1', 'fas_u1', 'tic');

        const selects: any[] = [
            sql.raw(`time_bucket('${bucketSize}', ${this.timeCol})`).as('bucket'),
        ];
        for (const key of validParams) {
            const def = TREND_PARAM_DEFS[key];
            const colExpr = def.col.includes("->")
                ? `CAST(${def.col} AS FLOAT)`
                : `CAST(${def.col} AS FLOAT)`;
            selects.push(
                sql.raw(`ROUND(AVG(${colExpr}) / ${def.scale}, 1)`).as(key)
            );
        }

        return await db
            .selectFrom('hvac.fact_raw')
            .select(selects)
            .where('train_id', '=', trainId)
            .where('carriage_id', '=', carriageId)
            .where(sql.raw(this.timeCol), '>', sql`NOW() - INTERVAL '${sql.raw(lookback)}'`)
            .groupBy('bucket')
            .orderBy('bucket', 'asc')
            .execute();
    }

    // B5: 预警触发前30分钟参数曲线（固定查与该预警相关的参数）
    static async getPredictDetailCurve(
        trainId: number,
        carriageId: number,
        triggerTime: Date,
        warnCode: string
    ) {
        const startTime = new Date(triggerTime.getTime() - 30 * 60 * 1000);

        // 根据 warn_code 确定展示哪些参数
        const paramsByWarnCode: Record<string, string[]> = {
            WARN_CABIN_OVERHEAT:   ['ras_u1', 'ras_u2', 'fas_u1', 'fas_u2', 'tic'],
            WARN_REFRIGERANT_LEAK: ['suckp_u11', 'suckp_u21', 'highpress_u11', 'highpress_u21'],
            WARN_COOLING_SYSTEM:   ['i_cp_u11', 'i_cp_u21', 'f_cp_u11', 'f_cp_u21'],
            WARN_TEMP_SENSOR:      ['fas_u1', 'fas_u2', 'ras_u1', 'ras_u2'],
            WARN_FILTER_CLOG:      ['presdiff_u1', 'presdiff_u2'],
            WARN_EF_CURRENT:       ['ras_u1', 'ras_u2'],
            WARN_CF_CURRENT:       ['ras_u1', 'ras_u2'],
            WARN_EXUF_CURRENT:     ['ras_u1', 'ras_u2'],
            WARN_CP_CURRENT:       ['i_cp_u11', 'i_cp_u21', 'fas_u1', 'fas_u2'],
            WARN_AQ_CO2:           ['aq_co2_u1', 'aq_co2_u2'],
            WARN_AQ_PM25:          ['aq_pm2_5_u1', 'aq_pm2_5_u2'],
            WARN_AQ_PM10:          ['aq_pm10_u1', 'aq_pm10_u2'],
            WARN_AQ_TVOC:          ['aq_tvoc_u1', 'aq_tvoc_u2'],
        };

        const params = paramsByWarnCode[warnCode] || ['ras_u1', 'fas_u1', 'tic'];
        const validParams = params.filter(p => TREND_PARAM_DEFS[p]);

        const selects: any[] = [
            sql.raw(this.timeCol).as('bucket'),
        ];
        for (const key of validParams) {
            const def = TREND_PARAM_DEFS[key];
            const colExpr = def.col.includes("->")
                ? `CAST(${def.col} AS FLOAT)`
                : `CAST(${def.col} AS FLOAT)`;
            selects.push(sql.raw(`ROUND(${colExpr} / ${def.scale}, 1)`).as(key));
        }

        const rows = await db
            .selectFrom('hvac.fact_raw')
            .select(selects)
            .where('train_id', '=', trainId)
            .where('carriage_id', '=', carriageId)
            .where(this.timeCol as any, '>=', startTime)
            .where(this.timeCol as any, '<=', triggerTime)
            .orderBy(this.timeCol as any, 'asc')
            .limit(500)
            .execute();

        return {
            params: validParams.map(k => ({ key: k, ...TREND_PARAM_DEFS[k] })),
            data: rows,
        };
    }

    static async getLatestStatus(deviceId: string) {
        return await db
            .selectFrom('hvac.fact_raw')
            .selectAll()
            .where('device_id', '=', deviceId)
            .orderBy(this.timeCol as any, 'desc')
            .limit(1)
            .executeTakeFirst();
    }
}
