import { db } from '../config/db.js';
import { sql } from 'kysely';
import { config } from '../config/index.js';

/**
 * 历史数据访问层
 * 封装复杂的数据库查询逻辑，确保外部 Service 层不直接操作 SQL
 */
export class HistoryRepository {
    /**
     * 根据当前 RUNTIME 环境返回用于时序分析的列名
     */
    private static get timeCol() {
        return config.runtime === 'DEV' ? 'ingest_time' : 'event_time';
    }

    /**
     * 获取指定设备在过去 N 小时内的温度趋势
     * 使用 TimescaleDB 的分页或时间加权聚合查询
     */
    static async getTemperatureTrend(deviceId: string, hours: number = 24) {
        return await db
            .selectFrom('hvac.fact_raw')
            .select([this.timeCol as any, 'tveh_1', 'tveh_2' as any])
            .where('device_id', '=', deviceId)
            .where(this.timeCol as any, '>', sql<Date>`NOW() - INTERVAL '${sql.raw(hours.toString())} hours'`)
            .orderBy(this.timeCol as any, 'asc')
            .execute();
    }

    /**
     * 获取设备最新的实时状态
     */
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
