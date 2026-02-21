import pkg from 'pg';
const { Client } = pkg;

async function test() {
    const connectionString = 'postgres://postgres:passw0rd@192.168.32.17:5432/postgres';

    console.log('Testing connection to:', connectionString);

    const client = new Client({
        host: '192.168.32.17',
        port: 5432,
        user: 'postgres',
        password: 'passw0rd',
        database: 'postgres',
        ssl: false,
        connectionTimeoutMillis: 5000,
    });

    try {
        await client.connect();
        console.log('Connection successful!');
        const res = await client.query('SELECT NOW()');
        console.log('Query result:', res.rows[0]);
        await client.end();
    } catch (err) {
        console.error('Connection failed:', err);
    }
}

test();
