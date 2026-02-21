import net from 'net';

const client = new net.Socket();
client.connect(5432, '192.168.32.17', () => {
    console.log('Connected, sending SSLRequest...');

    const sslReq = Buffer.alloc(8);
    sslReq.writeInt32BE(8, 0);
    sslReq.writeInt32BE(80877103, 4);

    client.write(sslReq);
});

client.on('data', (data) => {
    // Expect 'S' (SSL supported) or 'N' (not supported)
    console.log('Received response:', data.toString());
    client.end();
});

client.on('error', (err) => {
    console.error('Socket Error:', err);
});

client.on('close', () => {
    console.log('Connection closed');
});
