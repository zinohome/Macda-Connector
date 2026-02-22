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
     * 获取多维度的温度趋势
     * @param type 'hour' | 'day' | 'week' | 'month'
     */
    static async getTemperatureTrend(trainId: number, carriageId: number, type: string = 'hour') {
        let bucketSize = '1 second';
        let lookback = '1 hour';

        // 设置各维度的颗粒度与回顾时间
        switch (type) {
            case 'hour':
                bucketSize = '1 second';
                lookback = '1 hour';
                break;
            case 'day':
                bucketSize = '1 minute';
                lookback = '24 hours';
                break;
            case 'week':
                bucketSize = '15 minutes';
                lookback = '168 hours'; // 7 days
                break;
            case 'month':
                bucketSize = '30 minutes';
                lookback = '720 hours'; // 30 days
                break;
        }

        return await db
            .selectFrom('hvac.fact_raw')
            .select([
                // 使用 TimescaleDB 的 time_bucket 进行聚合
                sql.raw(`time_bucket('${bucketSize}', ${this.timeCol})`).as('bucket'),
                sql<number>`AVG(CAST(ras_u1 AS FLOAT) / 10.0)`.as('ras_sys'),
                sql<number>`AVG(CAST(fas_u1 AS FLOAT) / 10.0)`.as('fas_sys'),
                sql<number>`AVG(CAST(payload_json->'raw'->>'Tic' AS FLOAT) / 10.0)`.as('tic')
            ])
            .where('train_id', '=', trainId)
            .where('carriage_id', '=', carriageId)
            .where(sql.raw(this.timeCol), '>', sql`NOW() - INTERVAL '${sql.raw(lookback)}'`)
            .groupBy('bucket')
            .orderBy('bucket', 'asc')
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
