import { db } from '../config/db.js';
import { sql } from 'kysely';

/**
 * 实时状态与聚合查询
 */
export class StatusRepository {
    /**
     * 模拟原有的 /api/rest/AirSystem 结构
     * 返回各列车的报警统计摘要
     */
    static async getTrainAlarmSummary() {
        const records = await db
            .selectFrom('hvac.fact_raw')
            .select([
                'train_id as train_no',
                // 统计该列车过去 1 小时内是否存在任何报警标志位
                sql`COUNT(*) FILTER (WHERE 
           blpflt_comp_u11 = true OR blpflt_comp_u12 = true OR 
           blpflt_comp_u21 = true OR blpflt_comp_u22 = true OR
           bocflt_ef_u11 = true OR bocflt_ef_u12 = true OR
           bflt_tempover = true
        )`.as('alarm_count'),
                sql`COUNT(*) FILTER (WHERE 
           bflt_diffpres_u1 = true OR bflt_diffpres_u2 = true
        )`.as('warning_count')
            ])
            // .where('event_time', '>', sql<Date>`NOW() - INTERVAL '1 hour'`)
            .groupBy('train_id')
            .execute();

        console.log('[BFF Debug] Records from DB:', records);

        return {
            vw_train_alarm_count: records.map(r => ({
                train_no: r.train_no,
                alarm_count: Number(r.alarm_count),
                warning_count: Number(r.warning_count),
                total_number: records.length
            }))
        };
    }

    /**
     * 获取列车详情
     */
    static async getTrainDetails(trainId: number) {
        return await db
            .selectFrom('hvac.fact_raw')
            .selectAll()
            .where('train_id', '=', trainId)
            .orderBy('event_time', 'desc')
            .limit(8) // 通常一列车8节
            .execute();
    }
}
