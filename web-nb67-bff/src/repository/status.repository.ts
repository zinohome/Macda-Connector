import { db } from '../config/db.js';
import { sql } from 'kysely';
import { config } from '../config/index.js';

/**
 * 实时状态与聚合查询
 */
export class StatusRepository {
    /**
     * 根据当前 RUNTIME 环境返回用于时序分析的列名
     * DEV: 使用解析时间 (ingest_time)
     * PRD: 使用设备时间 (event_time)
     */
    private static get timeCol() {
        return config.runtime === 'DEV' ? 'ingest_time' : 'event_time';
    }
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
                ORDER BY device_id, ${sql.raw(this.timeCol)} DESC
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
                .orderBy(this.timeCol as any, 'desc')
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
     * 获取列车实时告警详情 (适配前端遍历逻辑)
     */
    static async getTrainDetails(trainIds: number[]) {
        let query = db
            .selectFrom('hvac.fact_event')
            .selectAll()
            // 实时故障去重：同一车厢同一编码只保留最新的
            .distinctOn(['train_id', 'carriage_id' as any, 'fault_code'])
            .where('event_type', '=', 'alarm')
            .orderBy('train_id')
            .orderBy('carriage_id' as any)
            .orderBy('fault_code')
            .orderBy(this.timeCol as any, 'desc')
            .limit(200);

        if (trainIds && trainIds.length > 0 && !isNaN(Number(trainIds[0]))) {
            const validIds = trainIds.filter(id => id !== undefined && id !== null) as number[];
            if (validIds.length > 0) {
                query = query.where('train_id', 'in', validIds as any);
            }
        }

        const data = await query.execute();

        return data.map(row => ({
            [row.fault_code as string]: 1,
            [`${row.fault_code}_name`]: row.fault_name,
            alarm_time: (row as any)[this.timeCol],
            carriage_no: row.carriage_id || 0,
            line_no: row.line_id || 0,
            train_no: row.train_id || 0
        }));
    }

    /**
     * 获取实时预警详情 (适配前端 Object.entries 逻辑)
     */
    static async getRealtimeWarnings(trainId?: number) {
        let query = db
            .selectFrom('hvac.fact_event')
            .selectAll()
            // 按列车、车厢、故障码去重，只保留最新的一条
            .distinctOn(['train_id', 'carriage_id' as any, 'fault_code'])
            .where('event_type', '=', 'predict')
            .orderBy('train_id')
            .orderBy('carriage_id' as any)
            .orderBy('fault_code')
            .orderBy(this.timeCol as any, 'desc')
            .limit(200);

        if (trainId && !isNaN(trainId)) {
            query = query.where('train_id', '=', trainId);
        }

        const data = await query.execute();

        const grouped: Record<string, any[]> = {};
        data.forEach(row => {
            const code = row.fault_code;
            if (code) {
                if (!grouped[code]) {
                    grouped[code] = [];
                }
                grouped[code].push({
                    name: row.fault_name,
                    warning_time: (row as any)[this.timeCol],
                    carriage_no: row.carriage_id || 0,
                    train_no: row.train_id || 0,
                    line_no: row.line_id || 0
                });
            }
        });
        return grouped;
    }

    /**
     * 获取列车车厢选择列表及其状态 (vw_carriage_status)
     */
    static async getTrainSelection(trainId: number) {
        const statuses = await db
            .with('latest_status', (db) => db
                .selectFrom('hvac.fact_raw')
                .select(['train_id', 'carriage_id', 'device_id', 'payload_json'])
                .where('train_id', '=', trainId)
                .distinctOn('device_id')
                .orderBy('device_id')
                .orderBy(this.timeCol as any, 'desc')
            )
            .selectFrom('latest_status')
            .select([
                'carriage_id as carriage_no',
                sql<number>`SUM(
                    (SELECT count(*) FROM jsonb_each_text(payload_json->'raw') 
                     WHERE (key ILIKE 'Bflt%' OR key ILIKE 'Blp%' OR key ILIKE 'Bsc%' OR key ILIKE 'Boc%') 
                     AND value = 'true'
                     AND NOT EXISTS (
                         SELECT 1 FROM hvac.dim_alarm_mask m 
                         WHERE m.device_id = latest_status.device_id AND m.fault_code = key
                     ))
                )`.as('alarm_count')
            ])
            .groupBy('carriage_id')
            .execute();

        return {
            vw_carriage_status: statuses.map(s => ({
                carriage_no: `${trainId}0${s.carriage_no}`,
                alarm_count: Number(s.alarm_count)
            })),
            vw_carriage_predict_status: [] // 预警先返回空
        };
    }

    /**
     * 获取全车各车厢最新的细节运行状态 (vw_running_state_info)
     */
    static async getTrainRunningState(trainId: number) {
        const records = await db
            .selectFrom('hvac.fact_raw')
            .selectAll()
            .where('train_id', '=', trainId)
            .distinctOn('carriage_id')
            .orderBy('carriage_id')
            .orderBy(this.timeCol as any, 'desc')
            .execute();

        return {
            vw_running_state_info: records.map(r => ({
                ...r,
                carriage_no: r.carriage_id,
                ...(r.payload_json?.raw || {}) // 将原始信号平铺
            }))
        };
    }

    /**
     * 获取车厢系统细节 (vw_system_info)
     */
    static async getCarriageSystemInfo(carriageId: string) {
        const trainId = parseInt(carriageId.slice(0, 4));
        const carNo = parseInt(carriageId.slice(-2));

        const record = await db
            .selectFrom('hvac.fact_raw')
            .selectAll()
            .where((eb) => eb.or([
                eb('carriage_id', '=', carNo),
                eb('carriage_id', '=', parseInt(carriageId))
            ]))
            .where('train_id', '=', trainId)
            .orderBy(this.timeCol as any, 'desc')
            .limit(1)
            .executeTakeFirst();

        return {
            vw_system_info: record ? [{
                ...record,
                ...(record.payload_json?.raw || {})
            }] : []
        };
    }

    /**
     * 获取车厢健康评估 (读取真实的设备寿命数据)
     * 根据 NB67 协议映射：风机/压缩机读取累计运行秒数，阀门读取开关次数
     */
    static async getCarriageHealthAssessment(carriageId: string) {
        const trainId = parseInt(carriageId.slice(0, 4));
        const carNo = parseInt(carriageId.slice(-2));

        // 1. 获取该车厢最新的原始报文快照
        const latestRaw = await db
            .selectFrom('hvac.fact_raw')
            .select(['payload_json'])
            .where('train_id', '=', trainId)
            .where((eb) => eb.or([
                eb('carriage_id', '=', carNo),
                eb('carriage_id', '=', parseInt(carriageId))
            ]))
            .orderBy(this.timeCol as any, 'desc')
            .limit(1)
            .executeTakeFirst();

        // 如果没有数据，返回空列表
        if (!latestRaw || !latestRaw.payload_json || !latestRaw.payload_json.raw) {
            return { vw_health_assessment: [] };
        }

        const raw = latestRaw.payload_json.raw;

        /**
         * 2. 定义部件寿命映射表 (与 nb67_event_processor.go 逻辑对齐)
         * 额定寿命参考：风机 25000h, 压缩机 50000h, 阀门 100万次
         */
        const componentSchema = [
            { name: '机组1通风机', timeField: 'DwefOpTmU11', ratedTime: 25000 },
            { name: '机组1冷凝风机', timeField: 'DwcfOpTmU11', ratedTime: 25000 },
            { name: '机组1压缩机1', timeField: 'DwcpOpTmU11', ratedTime: 50000 },
            { name: '机组1压缩机2', timeField: 'DwcpOpTmU12', ratedTime: 50000 },
            { name: '机组1新风阀', actionField: 'DwfadOpCntU1', ratedActions: 1000000 },
            { name: '机组1回风阀', actionField: 'DwradOpCntU1', ratedActions: 1000000 },

            { name: '机组2通风机', timeField: 'DwefOpTmU21', ratedTime: 25000 },
            { name: '机组2冷凝风机', timeField: 'DwcfOpTmU21', ratedTime: 25000 },
            { name: '机组2压缩机1', timeField: 'DwcpOpTmU21', ratedTime: 50000 },
            { name: '机组2压缩机2', timeField: 'DwcpOpTmU22', ratedTime: 50000 },
            { name: '机组2新风阀', actionField: 'DwfadOpCntU2', ratedActions: 1000000 },
            { name: '机组2回风阀', actionField: 'DwradOpCntU2', ratedActions: 1000000 },

            { name: '废排风机', timeField: 'DwexufanOpTm', ratedTime: 25000 },
            { name: '废排风阀', actionField: 'DwdmpexuOpCnt', ratedActions: 1000000 },
        ];

        // 3. 转换并计算健康度
        const assessment = componentSchema.map(c => {
            let actualTime = 0;
            let actualActions = 0;
            let ratio = 0;

            if (c.timeField) {
                // 原始数据为秒，转换为小时
                const seconds = Number(raw[c.timeField]) || 0;
                actualTime = Math.floor(seconds / 3600);
                ratio = actualTime / c.ratedTime;
            } else if (c.actionField) {
                // 原始数据即为次数
                actualActions = Number(raw[c.actionField]) || 0;
                ratio = actualActions / c.ratedActions;
            }

            // 判定逻辑：>90% 非健康，>75% 亚健康，否则健康
            let healthStatus = '健康';
            let suggestion = '设备运行良好，继续保持';

            if (ratio >= 0.9) {
                healthStatus = '非健康';
                suggestion = '设备寿命已耗尽，请立即安排更换';
            } else if (ratio >= 0.75) {
                healthStatus = '亚健康';
                suggestion = '部件进入磨损后期，请加强关注';
            }

            return {
                equipment_name: c.name,
                health_status: healthStatus,
                cumulative_running_time: actualTime,
                rated_running_time: c.ratedTime || 0,
                number_actions: actualActions,
                rated_actions: c.ratedActions || 0,
                suggestion: suggestion
            };
        });

        return {
            vw_health_assessment: assessment
        };
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
            .where(this.timeCol as any, '>=', new Date(params.startTime))
            .where(this.timeCol as any, '<=', new Date(params.endTime))
            .orderBy(this.timeCol as any, 'desc');

        if (params.trainId) {
            query = query.where('train_id', '=', params.trainId);
        }
        if (params.eventType) {
            query = query.where('event_type', '=', params.eventType);
        }

        return await query.limit(500).execute();
    }

    /**
     * 获取故障统计供 Echart 渲染 (返回 vw_alarm_info_all_date)
     */
    static async getFaultStatistics() {
        const data = await db
            .selectFrom('hvac.fact_event')
            .select(['fault_code'])
            .where('event_type', '=', 'alarm')
            .orderBy(this.timeCol as any, 'desc')
            .limit(5000)
            .execute();

        return {
            vw_alarm_info_all_date: data.map(r => ({
                [r.fault_code as string]: 1
            }))
        };
    }
}
