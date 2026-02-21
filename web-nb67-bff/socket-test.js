import net from 'net';

const client = net.createConnection({ host: '192.168.32.17', port: 5432 }, () => {
    console.log('Connected to server!');
    // Send a startup message or just wait for server to say something
    // Postgres usually waits for us.
});

client.on('data', (data) => {
    console.log('Received:', data.toString());
    client.end();
});

client.on('error', (err) => {
    console.error('Error:', err);
});

client.on('end', () => {
    console.log('Disconnected from server');
});

setTimeout(() => {
    console.log('Timeout - closing');
    client.end();
}, 5000);
