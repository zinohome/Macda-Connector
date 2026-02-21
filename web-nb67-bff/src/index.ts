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
            const data = await StatusRepository.getTrainDetails(parseInt(trainNo?.[0] || '0'));
            // 这里的返回结构需要适配前端对 vw_train_alarm_info 的期望
            return {
                vw_train_alarm_info: data.map(d => ({
                    ...d,
                    alarm_time: d.event_time,
                    carriage_no: d.carriage_id
                }))
            };
        });

        fastify.get('/api/history/temperature', async (request: any, reply) => {
            const { deviceId, hours } = request.query;
            if (!deviceId) return reply.status(400).send({ error: 'deviceId is required' });

            const data = await HistoryRepository.getTemperatureTrend(deviceId, parseInt(hours || '24'));
            return { success: true, data };
        });

        fastify.get('/api/status/latest', async (request: any, reply) => {
            const { deviceId } = request.query;
            if (!deviceId) return reply.status(400).send({ error: 'deviceId is required' });

            const data = await HistoryRepository.getLatestStatus(deviceId);
            return { success: true, data };
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
        await fastify.listen({ port: config.port, host: config.host });
        console.log(`[BFF] Server is running at http://${config.host}:${config.port}`);
    } catch (err) {
        fastify.log.error(err);
        process.exit(1);
    }
}

bootstrap();
