import { Kafka } from 'kafkajs';

async function testKafka() {
    const kafka = new Kafka({
        clientId: 'test-client',
        brokers: ['192.168.32.17:19092'],
        connectionTimeout: 5000,
    });

    const admin = kafka.admin();
    try {
        console.log('Connecting to Kafka...');
        await admin.connect();
        console.log('Kafka connection successful!');
        const topics = await admin.listTopics();
        console.log('Topics:', topics);
        await admin.disconnect();
    } catch (err) {
        console.error('Kafka connection failed:', err);
    }
}

testKafka();
