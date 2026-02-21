import { Kafka } from 'kafkajs';
import type { Consumer, EachMessagePayload } from 'kafkajs';
import { config } from './index.js';

/**
 * Kafka 管理中心
 * 负责与 Redpanda 建立连接并管理消费者
 */
export class KafkaManager {
    private static instance: KafkaManager;
    private kafka: Kafka;
    private consumer: Consumer;

    private constructor() {
        this.kafka = new Kafka({
            clientId: config.kafka.clientId,
            brokers: config.kafka.brokers,
        });
        this.consumer = this.kafka.consumer({ groupId: config.kafka.consumerGroup });
    }

    public static getInstance(): KafkaManager {
        if (!KafkaManager.instance) {
            KafkaManager.instance = new KafkaManager();
        }
        return KafkaManager.instance;
    }

    /**
     * 启动消费者并订阅主题
     */
    async start(handler: (topic: string, message: any) => void) {
        try {
            await this.consumer.connect();
            await this.consumer.subscribe({
                topics: [config.kafka.topics.parsed, config.kafka.topics.alarm],
                fromBeginning: false
            });

            await this.consumer.run({
                eachMessage: async (payload: EachMessagePayload) => {
                    const { topic, message } = payload;
                    if (message.value) {
                        try {
                            const data = JSON.parse(message.value.toString());
                            handler(topic, data);
                        } catch (e) {
                            console.error(`[Kafka] Failed to parse message on ${topic}:`, e);
                        }
                    }
                },
            });
            console.log(`[Kafka] Consumer started on topics: ${Object.values(config.kafka.topics).join(', ')}`);
        } catch (err) {
            console.error('[Kafka] Error starting consumer:', err);
        }
    }

    async stop() {
        await this.consumer.disconnect();
    }
}
