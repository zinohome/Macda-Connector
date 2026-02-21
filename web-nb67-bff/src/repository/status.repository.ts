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
    /**
     * 实现客户要求的“删掉告警，直到修好再弹出来”的自动复位逻辑
     */
    private static async syncMasks() {
        // [自动复位逻辑]: 寻找那些已被屏蔽、但当前物理状态已变回 false (即修好了) 的位
        // 只有修好了，我们才把屏蔽记作删除，确保下次再坏时能重新提醒
        await sql`
            DELETE FROM hvac.dim_alarm_mask m
            USING (
                SELECT DISTINCT ON (device_id) device_id, payload_json->'raw' as raw
                FROM hvac.fact_raw
                ORDER BY device_id, event_time DESC
            ) latest
            WHERE m.device_id = latest.device_id
            AND (latest.raw->>m.fault_code)::boolean = false
        `.execute(db);
    }

    static async getTrainAlarmSummary() {
        // 在聚合前同步屏蔽状态
        await this.syncMasks();

        const records = await db
            .with('latest_status', (db) => db
                .selectFrom('hvac.fact_raw')
                .select(['train_id', 'device_id', 'payload_json'])
                .distinctOn('device_id')
                .orderBy('device_id')
                .orderBy('event_time', 'desc')
            )
            .selectFrom('latest_status')
            .select([
                'train_id as train_no',

                // 1. 告警 (Alarm): 统计物理比特位，但排除掉“已被删除(屏蔽)”的位点
                sql<number>`SUM(
                    (SELECT count(*) FROM jsonb_each_text(payload_json->'raw') 
                     WHERE (key ILIKE 'Bflt%' OR key ILIKE 'Blp%' OR key ILIKE 'Bsc%' OR key ILIKE 'Boc%') 
                     AND value = 'true'
                     AND NOT EXISTS (
                         SELECT 1 FROM hvac.dim_alarm_mask m 
                         WHERE m.device_id = latest_status.device_id AND m.fault_code = key
                     ))
                )`.as('alarm_count'),

                // 2. 预警 (Warning)
                sql<number>`SUM(
                    (CASE WHEN (payload_json->'raw'->>'PresdiffU1')::int > 3000 OR (payload_json->'raw'->>'PresdiffU2')::int > 3000 THEN 1 ELSE 0 END) +
                    (CASE WHEN (payload_json->'raw'->>'DwefOpTmU11')::bigint >= 67500000 THEN 1 ELSE 0 END)
                )`.as('warning_count')
            ])
            .groupBy('train_id')
            .execute();

        return {
            vw_train_alarm_count: records.map(r => {
                const a = Number(r.alarm_count);
                const w = Number(r.warning_count);
                return {
                    train_no: r.train_no,
                    alarm_count: a,
                    warning_count: w,
                    total_number: a + w
                };
            })
        };
    }

    /**
     * 屏蔽告警 API
     */
    static async maskAlarm(deviceId: string, faultCode: string) {
        return await db
            .insertInto('hvac.dim_alarm_mask')
            .values({ device_id: deviceId, fault_code: faultCode })
            .onConflict((oc) => oc.doNothing())
            .execute();
    }

    /**
     * 获取列车详情
     * 为了适配前端 Object.keys(item) 的遍历逻辑，我们将 payload_json 中的数据平铺
     */
    static async getTrainDetails(trainId: number) {
        const data = await db
            .selectFrom('hvac.fact_raw')
            .selectAll()
            .where('train_id', '=', trainId)
            .orderBy('event_time', 'desc')
            .limit(8)
            .execute();

        return data.map(row => {
            const { payload_json, ...rest } = row;
            // 将 payload_json 里的字段解构到一级，这样前端 item[value] 就能拿到了
            return {
                ...rest,
                ...(payload_json as object || {})
            };
        });
    }
    /**
     * 获取历史事件列表 (包含告警、预警、寿命等)
     */
    static async getHistoricalEvents(params: {
        trainId?: number;
        eventType?: string;
        startTime: string;
        endTime: string;
    }) {
        let query = db
            .selectFrom('hvac.fact_event')
            .selectAll()
            .where('event_time', '>=', new Date(params.startTime))
            .where('event_time', '<=', new Date(params.endTime))
            .orderBy('event_time', 'desc');

        if (params.trainId) {
            query = query.where('train_id', '=', params.trainId);
        }
        if (params.eventType) {
            query = query.where('event_type', '=', params.eventType);
        }

        return await query.limit(500).execute();
    }
}
