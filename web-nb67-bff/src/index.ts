import Fastify from 'fastify';
import swagger from '@fastify/swagger';
import swaggerUi from '@fastify/swagger-ui';
import websocket from '@fastify/websocket';
import cors from '@fastify/cors';
import { config } from './config/index.js';
import { checkDbConnection } from './config/db.js';
import { KafkaManager } from './config/kafka.js';
import { HistoryRepository } from './repository/history.repository.js';
import { StatusRepository } from './repository/status.repository.js';

const fastify = Fastify({
    logger: {
        level: config.logLevel,
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
            const { deviceId, hours } = request.query;
            if (!deviceId) return reply.status(400).send({ error: 'deviceId is required' });

            const data = await HistoryRepository.getTemperatureTrend(deviceId, parseInt(hours || '24'));
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

        fastify.get('/api/test', async () => {
            return { status: 'ok', time: new Date().toISOString() };
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

            // 【防御修复】前端 DatePicker 在某些交互后会传来空时间，自动补充默认范围（最近24小时）
            const now = new Date();
            const safeEndTime = endTime || new Date(now).toISOString();
            const safeStartTime = startTime || new Date(now.getTime() - 24 * 60 * 60 * 1000).toISOString();

            try {
                const result = await StatusRepository.getHistoricalEvents({
                    trainId,
                    carriageId,
                    startTime: safeStartTime,
                    endTime: safeEndTime,
                    eventType: 'alarm',
                    page: page ? Number(page) : 1,
                    limit: limit ? Number(limit) : 10
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
                        occurrence_time: row[config.runtime === 'DEV' ? 'ingest_time' : 'event_time'],
                        recovery_time: null,
                        status: '活动'
                    };
                });

                const timeline = (result.list || []).map((row: any) => ({
                    train_id: row.train_id,
                    carriage_no: row.carriage_id,
                    fault_code: row.fault_code,
                    name: row.fault_name,
                    start_time: row[config.runtime === 'DEV' ? 'ingest_time' : 'event_time'],
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

        // 历史数据接口 (适配 HistoryData 页面)
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

            return {
                code: 200,
                data: {
                    list: result.list.map(row => ({
                        ...row,
                        ...(row.payload_json?.raw || {}),
                        event_time: (row as any)[config.runtime === 'DEV' ? 'ingest_time' : 'event_time']
                    })),
                    total: result.total
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
