import Fastify from 'fastify';
import swagger from '@fastify/swagger';
import swaggerUi from '@fastify/swagger-ui';
import websocket from '@fastify/websocket';
import cors from '@fastify/cors';
import { config } from './config/index.js';
import { db } from './config/db.js';
import { checkDbConnection } from './config/db.js';
import { sql } from 'kysely';
import { KafkaManager } from './config/kafka.js';
import { HistoryRepository } from './repository/history.repository.js';
import { StatusRepository } from './repository/status.repository.js';
import { formatTime } from './utils/time.js';

const fastify = Fastify({
    logger: {
        level: config.logLevel,
        timestamp: () => `,"time":"${formatTime(new Date())}"`
    },
});

/**
 * 启动服务
 */
async function bootstrap() {
    try {
        // 1. 注册插件
        await fastify.register(cors, { origin: '*' });
        await fastify.register(websocket);

        // 注册 Swagger
        await fastify.register(swagger, {
            openapi: {
                info: {
                    title: 'Macda-Connector BFF API',
                    description: '车载空调网关 BFF 接口文档',
                    version: '1.0.0'
                },
                servers: [{ url: `http://localhost:${config.port}` }]
            }
        });

        await fastify.register(swaggerUi, {
            routePrefix: '/docs',
            uiConfig: {
                docExpansion: 'list',
                deepLinking: false
            }
        });

        // 2. 数据库健康检查
        await checkDbConnection();

        // 3. 注册 WebSocket 路由
        fastify.register(async (instance) => {
            instance.get('/ws/signals', { websocket: true }, (connection: any, req) => {
                instance.log.info('[WebSocket] Client connected');

                connection.socket.on('message', (message: Buffer) => {
                    instance.log.info(`[WebSocket] Received: ${message.toString()}`);
                });

                connection.socket.on('close', () => {
                    instance.log.info('[WebSocket] Client disconnected');
                });
            });
        });

        // 4. 注册 HTTP 路由 (REST)
        fastify.get('/api/rest/AirSystem', {
            schema: {
                description: '获取在线列车空调运行状态汇总',
                tags: ['Status'],
                response: {
                    200: {
                        type: 'object',
                        properties: {
                            vw_train_alarm_count: {
                                type: 'array',
                                items: {
                                    type: 'object',
                                    properties: {
                                        train_no: { type: 'number' },
                                        alarm_count: { type: 'number' },
                                        warning_count: { type: 'number' },
                                        total_number: { type: 'number' }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }, async () => {
            return await StatusRepository.getTrainAlarmSummary();
        });

        fastify.post('/api/rest/v2/train/RealtimeAlarm', {
            schema: {
                description: '获取指定列车的详细实时信号',
                tags: ['Status'],
                body: {
                    type: 'object',
                    properties: {
                        trainNo: { type: 'array', items: { type: 'string' } }
                    }
                }
            }
        }, async (request: any) => {
            const { trainNo } = request.body;
            // parse to array of numbers
            const trainIds = Array.isArray(trainNo) ? trainNo.map(n => parseInt(n)) : [];
            const data = await StatusRepository.getTrainDetails(trainIds);

            return {
                vw_train_alarm_info: data
            };
        });

        fastify.get('/api/history/events', {
            // ... (keep existing historical events route)
        }, async (request: any) => {
            return await StatusRepository.getHistoricalEvents(request.query);
        });

        /**
         * 业务接口：删除（屏蔽）具体告警
         */
        fastify.post('/api/rest/v2/alarm/mask', {
            schema: {
                description: '删除(屏蔽)指定的故障位。被屏蔽的故障将不再计入实时统计，直到该故障物理消失后再次发生。',
                tags: ['Alarm'],
                body: {
                    type: 'object',
                    required: ['deviceId', 'faultCode'],
                    properties: {
                        deviceId: { type: 'string', description: '设备唯一标识 (如 HVAC-67-07002-01)' },
                        faultCode: { type: 'string', description: '故障代码位名称 (如 Bflt_FadU11)' }
                    }
                },
                response: {
                    200: {
                        type: 'object',
                        properties: {
                            success: { type: 'boolean' },
                            message: { type: 'string' }
                        }
                    }
                }
            }
        }, async (request: any) => {
            const { deviceId, faultCode } = request.body;
            await StatusRepository.maskAlarm(deviceId, faultCode);
            return { success: true, message: `Alarm ${faultCode} for ${deviceId} masked.` };
        });

        fastify.get('/api/history/temperature', async (request: any, reply) => {
            const { deviceId, type = 'hour' } = request.query;
            if (!deviceId) return reply.status(400).send({ error: 'deviceId is required' });

            // 假设 deviceId 格式为 700203
            const trainId = parseInt(String(deviceId).slice(0, 4));
            const carriageId = parseInt(String(deviceId).slice(-2));

            const data = await HistoryRepository.getTemperatureTrend(trainId, carriageId, type as string);
            return { success: true, data };
        });

        // ==========================
        // 兼容前端页面的统计与预警接口
        // ==========================
        fastify.post('/api/rest/v2/FaultStatistics', {
            schema: { tags: ['Status'] }
        }, async () => {
            return await StatusRepository.getFaultStatistics();
        });

        fastify.get('/api/rest/RealtimeWarning', {
            schema: { tags: ['Status'] }
        }, async () => {
            return await StatusRepository.getRealtimeWarnings();
        });

        fastify.get('/api/status/latest', async (request: any, reply) => {
            const { deviceId } = request.query;
            if (!deviceId) return reply.status(400).send({ error: 'deviceId is required' });

            const data = await HistoryRepository.getLatestStatus(deviceId);
            return { success: true, data };
        });

        // ── 健康检查端点（Docker healthcheck 专用）────────────────
        // 此端点不访问数据库，保证容器启动阶段能快速通过健康检查
        fastify.get('/health', async () => {
            return { status: 'ok', time: formatTime(new Date()) };
        });

        fastify.get('/api/test', async () => {
            return { status: 'ok', time: formatTime(new Date()) };
        });

        // 调试接口：获取最新一条原始报文的所有字段名和示例值，方便前端对齐字段名
        fastify.get('/api/debug/raw-fields', async (request: any) => {
            const { trainId, carriageId } = request.query;
            const row = await db
                .selectFrom('hvac.fact_raw')
                .selectAll()
                .where('train_id', '=', parseInt(trainId || '7002'))
                .where('carriage_id', '=', parseInt(carriageId || '1'))
                .orderBy('ingest_time', 'desc')
                .limit(1)
                .executeTakeFirst();
            if (!row) return { error: 'no data' };
            const raw = (row as any).payload_json?.raw || {};
            return {
                // 顶层 DB 列（排除大字段）
                db_columns: Object.keys(row).filter(k => k !== 'payload_json'),
                db_sample: Object.fromEntries(Object.entries(row).filter(([k]) => k !== 'payload_json')),
                // payload_json.raw 内的信号字段
                raw_keys: Object.keys(raw),
                raw_sample: raw
            };
        });

        // ==========================
        // 列车/车厢 详细 REST 接口 (适配原有前端)
        // ==========================
        fastify.get('/api/rest/train/StatusAlert/:trainId', async (request: any) => {
            const trainId = parseInt(request.params.trainId);
            return await StatusRepository.getRealtimeWarnings(trainId);
        });

        fastify.get('/api/rest/train/TrainSelection/:trainId', async (request: any) => {
            return await StatusRepository.getTrainSelection(parseInt(request.params.trainId));
        });

        fastify.get('/api/rest/train/RunningState/:trainId', async (request: any) => {
            return await StatusRepository.getTrainRunningState(parseInt(request.params.trainId));
        });

        // 历史报警接口 (适配原有前端)
        fastify.post('/api/rest/train/AlarmInformation', async (request: any) => {
            const { state, startTime, endTime, page, limit } = request.body;
            fastify.log.info(`[BFF] AlarmInformation Request Body: ${JSON.stringify(request.body)}`);

            const trainId = state ? parseInt(state.slice(0, 4)) : undefined;
            const carriageId = state ? parseInt(state.slice(4, 6)) : undefined;

            // 直接使用前端传递的时间戳，严禁在后端动态生成时间（防止翻页漂移）
            const safeStartTime = startTime || '';
            const safeEndTime = endTime || '';

            // 强制类型转换，杜绝 NaN 或字符串导致的计算错误
            const safePage = Math.max(1, parseInt(page as string) || 1);
            const safeLimit = Math.max(1, parseInt(limit as string) || 10);

            fastify.log.info(`[BFF] AlarmInformation Parsed: trainId=${trainId}, carriageId=${carriageId}, range=[${safeStartTime} - ${safeEndTime}], page=${safePage}, limit=${safeLimit}`);

            try {
                const result = await StatusRepository.getHistoricalEvents({
                    trainId,
                    carriageId,
                    startTime: safeStartTime,
                    endTime: safeEndTime,
                    eventType: 'alarm',
                    page: safePage,
                    limit: safeLimit
                });

                const list = (result.list || []).map((row: any) => {
                    let level = '一般';
                    if (row.severity === 3) level = '严重';
                    else if (row.severity === 2) level = '一般';
                    else if (row.severity === 1) level = '轻微';

                    return {
                        train_id: row.train_id,
                        carriage_no: row.carriage_id,
                        fault_code: row.fault_code,
                        fault_desc: row.fault_name,
                        fault_level: level,
                        occurrence_time: formatTime(row[config.runtime === 'DEV' ? 'ingest_time' : 'event_time']),
                        recovery_time: null,
                        status: '活动'
                    };
                });

                const timeline = (result.list || []).map((row: any) => ({
                    train_id: row.train_id,
                    carriage_no: row.carriage_id,
                    fault_code: row.fault_code,
                    name: row.fault_name,
                    start_time: formatTime(row[config.runtime === 'DEV' ? 'ingest_time' : 'event_time']),
                    end_time: null,
                    state: `${row.train_id}${String(row.carriage_id).padStart(2, '0')}1`
                }));

                return {
                    code: 200,
                    data: {
                        list: list,
                        total: result.total,
                        alarm_timeline: timeline
                    }
                };
            } catch (err: any) {
                fastify.log.error(err);
                return {
                    code: 500,
                    message: err.message,
                    data: { list: [], total: 0, alarm_timeline: [] }
                };
            }
        });

        // 历史数据接口 (适配 HistoryData 页面提取存在数据的日期)
        fastify.get('/api/getTrainDataDates', async (request: any) => {
            const { state, startTime, endTime } = request.query;
            const trainId = state ? parseInt(String(state).slice(0, 4)) : undefined;
            const carriageId = state ? parseInt(String(state).slice(4, 6)) : undefined;

            try {
                const queryStartTime = new Date(String(startTime || ''));
                const queryEndTime = new Date(String(endTime || ''));
                const timeCol = config.runtime === 'DEV' ? 'ingest_time' : 'event_time';

                let q = db.selectFrom('hvac.fact_raw')
                    .select(sql<string>`TO_CHAR(${sql.raw(timeCol)}, 'YYYY-MM-DD')`.as('date'))
                    .distinct()
                    .where(timeCol as any, '>=', queryStartTime)
                    .where(timeCol as any, '<=', queryEndTime);

                if (trainId) q = q.where('train_id', '=', trainId);
                if (carriageId) q = q.where('carriage_id', '=', carriageId);

                const records = await q.execute();
                const dates = records.map((r: any) => r.date).filter(Boolean).sort((a: string, b: string) => b.localeCompare(a));
                return { code: 200, data: dates };
            } catch (err: any) {
                fastify.log.error(err);
                return { code: 500, message: err.message, data: [] };
            }
        });

        // 历史数据接口 (适配 HistoryData 页面)
        // 每条 DB 记录按机组一/机组二拆成两行输出
        fastify.get('/api/getTrainData', async (request: any) => {
            const { state, startTime, endTime, page = 1, limit = 50 } = request.query;
            const trainId = state ? parseInt(String(state).slice(0, 4)) : undefined;
            const carriageId = state ? parseInt(String(state).slice(4, 6)) : undefined;

            const result = await StatusRepository.getRawHistory({
                trainId,
                carriageId,
                startTime: String(startTime || ''),
                endTime: String(endTime || ''),
                page: parseInt(page as string),
                limit: parseInt(limit as string)
            });

            // ms 转小时，保留 1 位小数
            const ms2h = (ms: any) =>
                ms !== undefined && ms !== null && ms !== ''
                    ? (Number(ms) / 3600000).toFixed(1) + 'h'
                    : '';

            return {
                code: 200,
                data: {
                    list: result.list.flatMap((row: any) => {
                        const raw = row.payload_json?.raw || {};

                        // ── 机组一 (U1x) ─────────────────────────────────
                        const unit1 = {
                            train_id: row.train_id,
                            carriage_id: row.carriage_id,
                            ingest_time: formatTime(row.ingest_time),
                            event_time: formatTime(row.event_time),
                            unit_name: '机组一',

                            i_inner_temp: raw.Tic !== undefined ? (Number(raw.Tic) / 10).toFixed(1) : '',
                            i_outer_temp: raw.FasU1 !== undefined ? (Number(raw.FasU1) / 10).toFixed(1) : '',
                            i_set_temp: raw.SpU11 !== undefined ? (Number(raw.SpU11) / 10).toFixed(1) : '',
                            i_hum: raw.Humdity1 ?? '',
                            i_co2: raw.AqCo2U1 ?? '',
                            work_status: raw.WmodeU1 ?? '',

                            dwef_op_tm: raw.DwefOpTmU11 ?? '',
                            w_crnt_cf1: raw.ICfU11 ?? '',
                            w_crnt_cf2: raw.ICfU12 ?? '',
                            w_crnt_cp1: raw.ICpU11 ?? '',
                            w_crnt_cp2: raw.ICpU12 ?? '',
                            w_freq_cp1: raw.FCpU11 ?? '',
                            w_freq_cp2: raw.FCpU12 ?? '',
                            w_crnt_ef1: raw.IEfU11 ?? '',
                            w_crnt_ef2: raw.IEfU12 ?? '',

                            i_high_pres1: raw.HighpressU11 ?? '',
                            i_low_pres1: raw.SuckpU11 ?? '',
                            i_high_pres2: raw.HighpressU12 ?? '',
                            i_low_pres2: raw.SuckpU12 ?? '',

                            i_sat1: raw.SasU11 !== undefined ? (Number(raw.SasU11) / 10).toFixed(1) : '',
                            i_sat2: raw.SasU12 !== undefined ? (Number(raw.SasU12) / 10).toFixed(1) : '',

                            dwcf_op_tm1: raw.DwcfOpTmU11 ?? '',
                            dwcf_op_tm2: raw.DwcfOpTmU12 ?? '',
                            dwcp_op_tm1: raw.DwcpOpTmU11 ?? '',
                            dwcp_op_tm2: raw.DwcpOpTmU12 ?? '',
                            dwexufan_op_tm: raw.DwexufanOpTm ?? '',

                            dwfad_op_cnt: raw.FadposU1 ?? '',
                            dwrad_op_cnt: raw.RadposU1 ?? '',
                        };

                        // ── 机组二 (U2x) ─────────────────────────────────
                        const unit2 = {
                            train_id: row.train_id,
                            carriage_id: row.carriage_id,
                            ingest_time: formatTime(row.ingest_time),
                            event_time: formatTime(row.event_time),
                            unit_name: '机组二',

                            i_inner_temp: raw.Tic !== undefined ? (Number(raw.Tic) / 10).toFixed(1) : '',
                            i_outer_temp: raw.FasU2 !== undefined ? (Number(raw.FasU2) / 10).toFixed(1) : '',
                            i_set_temp: raw.SpU21 !== undefined ? (Number(raw.SpU21) / 10).toFixed(1) : '',
                            i_hum: raw.Humdity2 ?? '',
                            i_co2: raw.AqCo2U2 ?? '',
                            work_status: raw.WmodeU2 ?? '',

                            dwef_op_tm: raw.DwefOpTmU21 ?? '',
                            w_crnt_cf1: raw.ICfU21 ?? '',
                            w_crnt_cf2: raw.ICfU22 ?? '',
                            w_crnt_cp1: raw.ICpU21 ?? '',
                            w_crnt_cp2: raw.ICpU22 ?? '',
                            w_freq_cp1: raw.FCpU21 ?? '',
                            w_freq_cp2: raw.FCpU22 ?? '',
                            w_crnt_ef1: raw.IEfU21 ?? '',
                            w_crnt_ef2: raw.IEfU22 ?? '',

                            i_high_pres1: raw.HighpressU21 ?? '',
                            i_low_pres1: raw.SuckpU21 ?? '',
                            i_high_pres2: raw.HighpressU22 ?? '',
                            i_low_pres2: raw.SuckpU22 ?? '',

                            i_sat1: raw.SasU21 !== undefined ? (Number(raw.SasU21) / 10).toFixed(1) : '',
                            i_sat2: raw.SasU22 !== undefined ? (Number(raw.SasU22) / 10).toFixed(1) : '',

                            dwcf_op_tm1: raw.DwcfOpTmU21 ?? '',
                            dwcf_op_tm2: raw.DwcfOpTmU22 ?? '',
                            dwcp_op_tm1: raw.DwcpOpTmU21 ?? '',
                            dwcp_op_tm2: raw.DwcpOpTmU22 ?? '',
                            dwexufan_op_tm: '', // 仅在机组一展示，避免重复

                            dwfad_op_cnt: raw.FadposU2 ?? '',
                            dwrad_op_cnt: raw.RadposU2 ?? '',
                        };

                        return [unit1, unit2];
                    }),
                    // 每条 DB 记录拆成 2 行，总数同步 ×2
                    total: result.total * 2
                }
            };
        });

        fastify.get('/api/rest/carriage/SystemInfo/:carriageId', async (request: any) => {
            return await StatusRepository.getCarriageSystemInfo(request.params.carriageId);
        });

        fastify.get('/api/rest/carriage/HealthAssessment/:carriageId', async (request: any) => {
            return await StatusRepository.getCarriageHealthAssessment(request.params.carriageId);
        });

        fastify.post('/api/rest/carriage/TemperatureTrend', async (request: any) => {
            const { carriageNo, type = 'hour' } = request.body; // 假设传入 "700203" 和 "day"
            const trainId = parseInt(String(carriageNo).slice(0, 4));
            const carriageId = parseInt(String(carriageNo).slice(-2));

            // 根据传入的维度（时、天、周、月）调用聚合查询
            const data = await HistoryRepository.getTemperatureTrend(trainId, carriageId, type as string);

            // 返回格式适配前端： { [type]: data }
            return { [type]: data };
        });

        // 5. 启动 Kafka 消费者并广播
        const kafka = KafkaManager.getInstance();
        await kafka.start((topic, data) => {
            // 广播给所有 WebSocket 客户端
            const payload = JSON.stringify({ topic, data });
            fastify.websocketServer.clients.forEach((client) => {
                if (client.readyState === 1) { // OPEN
                    client.send(payload);
                }
            });
        });

        // 6. 运行监听
        await fastify.ready();
        console.log('[BFF] All Routes registered:');
        console.log(fastify.printRoutes());

        await fastify.listen({ port: config.port, host: config.host });
        console.log(`[BFF] Server V2-DISTINCT is running at http://${config.host}:${config.port}`);
    } catch (err) {
        fastify.log.error(err);
        process.exit(1);
    }
}

bootstrap();
