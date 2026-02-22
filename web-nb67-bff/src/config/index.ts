import dotenv from 'dotenv';

dotenv.config();

/**
 * 全局配置中心
 */
export const config = {
    port: parseInt(process.env.PORT || '3000', 10),
    host: process.env.HOST || '0.0.0.0',

    // 数据库配置 (TimescaleDB)
    db: {
        connectionString: process.env.DATABASE_URL || 'postgres://postgres:passw0rd@timescaledb:5432/postgres?sslmode=disable',
        max: 20, // 最大连接池
    },

    // 消息队列配置 (Redpanda)
    kafka: {
        brokers: (process.env.KAFKA_BROKERS || 'redpanda-1:9092,redpanda-2:9092,redpanda-3:9092').split(','),
        clientId: 'macda-web-bff',
        consumerGroup: 'macda-web-group',
        topics: {
            parsed: 'signal-parsed',
            alarm: 'signal-alarm',
        }
    },

    // 运行环境与时间分析逻辑 (DEV=解析时间分析, PRD=设备时间分析)
    runtime: process.env.RUNTIME || 'PRD',

    // 日志等级
    logLevel: process.env.LOG_LEVEL || 'info',
};
