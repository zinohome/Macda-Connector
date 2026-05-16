import { db } from '../config/db.js';
import { sql } from 'kysely';
import { config } from '../config/index.js';
import { formatTime } from '../utils/time.js';

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
            // 实时故障去重：同一车厢同一编码只保留最早触发的（升序取第一条）
            .distinctOn(['train_id', 'carriage_id' as any, 'fault_code'])
            .where('event_type', '=', 'alarm')
            .orderBy('train_id')
            .orderBy('carriage_id' as any)
            .orderBy('fault_code')
            .orderBy(this.timeCol as any, 'asc')
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
            alarm_time: formatTime((row as any)[this.timeCol]),
            carriage_no: row.carriage_id || 0,
            line_no: row.line_id || 0,
            train_no: row.train_id || 0
        }));
    }

    /**
     * 获取实时预警详情 (适配前端 Object.entries 逻辑)
     */
    // HVAC 预警序号 → warning_config.warn_code 映射
    private static hvacSeqToWarnCode(seq: number): string | null {
        if (seq >= 1  && seq <= 4)  return 'WARN_REFRIGERANT_LEAK';
        if (seq >= 5  && seq <= 6)  return 'WARN_COOLING_SYSTEM';
        if (seq >= 7  && seq <= 8)  return 'WARN_TEMP_SENSOR';
        if (seq === 9)               return 'WARN_CABIN_OVERHEAT';
        if (seq >= 10 && seq <= 11) return 'WARN_FILTER_CLOG';
        if (seq >= 12 && seq <= 15) return 'WARN_EF_CURRENT';
        if (seq >= 16 && seq <= 19) return 'WARN_CF_CURRENT';
        if (seq === 20)              return 'WARN_EXUF_CURRENT';
        if (seq >= 21 && seq <= 24) return 'WARN_CP_CURRENT';
        if (seq === 25 || seq === 26) return 'WARN_AQ_CO2';
        return null;
    }

    static async getRealtimeWarnings(trainId?: number) {
        let query = db
            .selectFrom('hvac.fact_event')
            .selectAll()
            // 按列车、车厢、故障码去重，只保留最早触发的一条（升序取第一条）
            .distinctOn(['train_id', 'carriage_id' as any, 'fault_code'])
            .where('event_type', '=', 'predict')
            .orderBy('train_id')
            .orderBy('carriage_id' as any)
            .orderBy('fault_code')
            .orderBy(this.timeCol as any, 'asc')
            .limit(200);

        if (trainId && !isNaN(trainId)) {
            query = query.where('train_id', '=', trainId);
        }

        const [data, configs] = await Promise.all([
            query.execute(),
            db.selectFrom('hvac.warning_config' as any).selectAll().execute() as Promise<any[]>
        ]);

        // 构建 warn_code → strategy 查找表
        const strategyMap: Record<string, string> = {};
        configs.forEach((c: any) => {
            if (c.warn_code && c.strategy) strategyMap[c.warn_code] = c.strategy;
        });

        const grouped: Record<string, any[]> = {};
        data.forEach(row => {
            const code = row.fault_code;
            if (code) {
                // 从 HVAC 编码提取序号：HVAC307 → 307 % 100 = 7
                const num = parseInt(code.replace(/[^0-9]/g, ''), 10);
                const seq = isNaN(num) ? -1 : num % 100;
                const warnCode = this.hvacSeqToWarnCode(seq);
                const strategy = warnCode ? (strategyMap[warnCode] || '') : '';

                if (!grouped[code]) grouped[code] = [];
                grouped[code].push({
                    name: row.fault_name,
                    warning_time: formatTime((row as any)[this.timeCol]),
                    carriage_no: row.carriage_id || 0,
                    train_no: row.train_id || 0,
                    line_no: row.line_id || 0,
                    strategy
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

            // 判定逻辑：≥85% 非健康，≥70% 亚健康，否则健康 — PHM 2.2节
            let healthStatus = '健康';
            let suggestion = '设备运行良好，继续保持';

            if (ratio >= 0.85) {
                healthStatus = '非健康';
                suggestion = '设备寿命已耗尽，请立即安排更换';
            } else if (ratio >= 0.70) {
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

    // B1: 支持机组筛选（unitNames: ['机组一','机组二'] 或空=全部）
    static async getHistoricalEvents(params: {
        trainId?: number | undefined;
        carriageId?: number | undefined;
        carriageIds?: number[] | undefined;
        eventType?: string | undefined;
        unitNames?: string[] | undefined;
        startTime: string;
        endTime: string;
        page?: number;
        limit?: number;
    }) {
        // 分页参数处理，确保类型正确
        const page = Math.max(1, Number(params.page) || 1);
        const limit = Math.min(500, Math.max(1, Number(params.limit) || 500));
        const offset = (page - 1) * limit;

        // 【关键修复】count 查询和 list 查询必须分别独立构建
        // Kysely 的 QB 对象在调用 .select()/.selectAll() 后返回新对象，
        // 但共享同一个 QB 实例在某些版本中会导致 select 子句污染。
        // 这里通过工厂函数每次重新构建，彻底隔离两路查询。
        const buildBase = () => {
            let q = db.selectFrom('hvac.fact_event');

            // 只有当时间参数有效时才应用过滤器
            if (params.startTime) {
                const start = new Date(params.startTime);
                if (!isNaN(start.getTime())) {
                    q = q.where(this.timeCol as any, '>=', start);
                }
            }
            if (params.endTime) {
                const end = new Date(params.endTime);
                if (!isNaN(end.getTime())) {
                    q = q.where(this.timeCol as any, '<=', end);
                }
            }

            if (params.trainId) {
                q = q.where('train_id', '=', params.trainId);
            }
            if (params.carriageId) {
                q = q.where('carriage_id', '=', params.carriageId);
            }
            if (params.carriageIds && params.carriageIds.length > 0) {
                q = q.where('carriage_id', 'in', params.carriageIds as any);
            }
            if (params.eventType) {
                q = q.where('event_type', '=', params.eventType);
            }
            // 机组筛选：通过 fault_code 包含 U1/U2 判断机组
            if (params.unitNames && params.unitNames.length > 0 && params.unitNames.length < 2) {
                if (params.unitNames.includes('机组一')) {
                    q = q.where(sql`lower(fault_code) like '%u1%'` as any);
                } else if (params.unitNames.includes('机组二')) {
                    q = q.where(sql`lower(fault_code) like '%u2%'` as any);
                }
            }
            return q;
        };

        // 1. 独立执行 count 查询，获取总数
        const totalResult = await buildBase()
            .select(sql<number>`count(*)`.as('count'))
            .executeTakeFirst();
        const total = Number((totalResult as any)?.count || 0);

        // 2. 独立执行分页 list 查询
        // 增加 event_time 和 fault_code 作为 tie-breakers，确保结果集顺序在分页间绝对稳定
        const list = await buildBase()
            .selectAll()
            .orderBy(this.timeCol as any, 'desc')
            .orderBy('event_time', 'desc')
            .orderBy('fault_code', 'asc')
            .limit(limit)
            .offset(offset)
            .execute();

        console.log(`[Repository] Query Alarms: total=${total}, returned=${list.length}, page=${page}, offset=${offset}, range=[${params.startTime} - ${params.endTime}]`);

        return { list, total };
    }

    /**
     * 获取原始历史数据 (适配全量数据查询)
     */
    static async getRawHistory(params: {
        trainId?: number | undefined;
        carriageId?: number | undefined;
        startTime: string;
        endTime: string;
        page: number;
        limit: number;
    }) {
        const offset = (params.page - 1) * params.limit;

        let baseQuery = db
            .selectFrom('hvac.fact_raw')
            .where(this.timeCol as any, '>=', new Date(params.startTime))
            .where(this.timeCol as any, '<=', new Date(params.endTime));

        if (params.trainId) {
            baseQuery = baseQuery.where('train_id', '=', params.trainId);
        }
        if (params.carriageId) {
            baseQuery = baseQuery.where('carriage_id', '=', params.carriageId);
        }

        // 获取总数
        const totalResult = await baseQuery
            .select(sql`count(*)`.as('count'))
            .executeTakeFirst();
        const total = Number((totalResult as any)?.count || 0);

        // 获取列表
        const list = await baseQuery
            .selectAll()
            .orderBy(this.timeCol as any, 'desc')
            .limit(params.limit)
            .offset(offset)
            .execute();

        return { list, total };
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

    // B3: 历史预警查询 — LEFT JOIN warning_config 带出触发条件(strategy)
    static async getHistoricalWarnings(params: {
        trainId?: number;
        carriageIds?: number[];
        unitNames?: string[];
        startTime: string;
        endTime: string;
        page: number;
        limit: number;
    }) {
        const page = Math.max(1, params.page);
        const limit = Math.min(500, Math.max(1, params.limit));
        const offset = (page - 1) * limit;

        const buildBase = () => {
            let q = db
                .selectFrom('hvac.fact_event as e')
                .leftJoin('hvac.warning_config as wc', 'wc.warn_code', 'e.fault_code' as any)
                .where('e.event_type' as any, '=', 'predict');

            if (params.startTime) {
                const start = new Date(params.startTime);
                if (!isNaN(start.getTime())) q = q.where(`e.${this.timeCol}` as any, '>=', start);
            }
            if (params.endTime) {
                const end = new Date(params.endTime);
                if (!isNaN(end.getTime())) q = q.where(`e.${this.timeCol}` as any, '<=', end);
            }
            if (params.trainId) q = q.where('e.train_id' as any, '=', params.trainId);
            if (params.carriageIds && params.carriageIds.length > 0) {
                q = q.where('e.carriage_id' as any, 'in', params.carriageIds as any);
            }
            // 注意：unit 过滤在 JS 层处理（sql raw 在此版本 Kysely 的 where 中不生效）
            return q;
        };

        const totalResult = await buildBase()
            .select(sql<number>`count(*)`.as('count'))
            .executeTakeFirst();
        const total = Number((totalResult as any)?.count || 0);

        let list = await buildBase()
            .select([
                'e.event_time', 'e.ingest_time', 'e.train_id', 'e.carriage_id',
                'e.fault_code', 'e.fault_name', 'e.severity', 'e.status',
                'e.recovery_time' as any,
                'wc.strategy as trigger_condition' as any,
            ])
            .orderBy(`e.${this.timeCol}` as any, 'desc')
            .limit(1000)   // 宽松拉取，JS 过滤后再分页
            .execute();

        // JS 层机组过滤（sql raw 在此版本 Kysely where 中无法生效）
        if (params.unitNames && params.unitNames.length === 1) {
            const unit = params.unitNames[0];
            const u1Seqs = new Set([1,3,5,7,9,10,12,13,16,17,20,21,22,25]);
            const u2Seqs = new Set([2,4,6,8,11,14,15,18,19,23,24,26]);
            list = list.filter((row: any) => {
                const code: string = (row.fault_code || '').toLowerCase();
                if (code.startsWith('hvac')) {
                    const seq = parseInt(code.replace(/[^0-9]/g, '')) % 100;
                    return unit === '机组一' ? u1Seqs.has(seq) : u2Seqs.has(seq);
                }
                // u1/u2 风格编码（alarm 类）
                return unit === '机组一' ? code.includes('u1') : code.includes('u2');
            });
        }

        const filteredTotal = list.length;
        const paginated = list.slice(offset, offset + limit);

        return { list: paginated, total: filteredTotal };
    }

    // B7: 预警条件配置 CRUD
    static async getWarningConfigs() {
        return await db.selectFrom('hvac.warning_config').selectAll().orderBy('id', 'asc').execute();
    }

    static async getWarningConfig(id: number) {
        return await db.selectFrom('hvac.warning_config').selectAll().where('id', '=', id).executeTakeFirst();
    }

    static async updateWarningConfig(id: number, data: {
        threshold_good?: string;
        threshold_normal?: string;
        threshold_bad?: string;
        trigger_operator?: string;
        trigger_value?: number;
        clear_value?: number;
        duration_seconds?: number;
        unit?: string;
        strategy?: string;
        enabled?: boolean;
    }) {
        return await db
            .updateTable('hvac.warning_config' as any)
            .set(data as any)
            .where('id', '=', id)
            .execute();
    }

    static async resetWarningConfig(id: number) {
        return await db.executeQuery(
            sql`UPDATE hvac.warning_config
                SET trigger_value    = default_trigger_value,
                    clear_value      = default_clear_value,
                    duration_seconds = default_duration_seconds,
                    threshold_good   = default_threshold_good,
                    threshold_normal = default_threshold_normal,
                    threshold_bad    = default_threshold_bad,
                    strategy         = default_strategy
                WHERE id = ${sql.lit(id)}
                  AND default_trigger_value IS NOT NULL`.compile(db)
        );
    }

    // B8: 获取指定车厢最新数据时间（供前端离线检测）
    static async getLatestDataTime(trainId: number, carriageId: number): Promise<Date | null> {
        const result = await db
            .selectFrom('hvac.fact_raw')
            .select(sql<Date>`MAX(${sql.raw(this.timeCol)})`.as('latest'))
            .where('train_id', '=', trainId)
            .where('carriage_id', '=', carriageId)
            .executeTakeFirst();
        return (result as any)?.latest || null;
    }
}
