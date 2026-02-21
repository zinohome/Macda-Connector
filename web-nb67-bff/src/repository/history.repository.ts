import { db } from '../config/db.js';
import { sql } from 'kysely';

/**
 * 历史数据访问层
 * 封装复杂的数据库查询逻辑，确保外部 Service 层不直接操作 SQL
 */
export class HistoryRepository {
    /**
     * 获取指定设备在过去 N 小时内的温度趋势
     * 使用 TimescaleDB 的分页或时间加权聚合查询
     */
    static async getTemperatureTrend(deviceId: string, hours: number = 24) {
        return await db
            .selectFrom('hvac.fact_raw')
            .select(['event_time', 'tveh_1', 'tveh_2'])
            .where('device_id', '=', deviceId)
            .where('event_time', '>', sql<Date>`NOW() - INTERVAL '${sql.raw(hours.toString())} hours'`)
            .orderBy('event_time', 'asc')
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
            .orderBy('event_time', 'desc')
            .limit(1)
            .executeTakeFirst();
    }
}
