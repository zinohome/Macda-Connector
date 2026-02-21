import net from 'net';

const client = new net.Socket();
client.connect(5432, '192.168.32.17', () => {
    console.log('Connected');

    // StartupMessage (Protocol 3.0)
    // Length (4 bytes) | Version (4 bytes, 196608) | user\0postgres\0database\0postgres\0\0
    const user = 'postgres';
    const db = 'postgres';
    const startup = Buffer.alloc(1024);
    let pos = 4;
    startup.writeInt32BE(196608, pos); pos += 4;
    startup.write('user\0', pos); pos += 5;
    startup.write(user + '\0', pos); pos += user.length + 1;
    startup.write('database\0', pos); pos += 9;
    startup.write(db + '\0', pos); pos += db.length + 1;
    startup.write('\0', pos); pos += 1;
    startup.writeInt32BE(pos, 0);

    client.write(startup.slice(0, pos));
});

client.on('data', (data) => {
    console.log('Received data (hex):', data.toString('hex'));
    console.log('Received data (string):', data.toString());
});

client.on('error', (err) => {
    console.error('Socket Error:', err);
});

client.on('close', () => {
    console.log('Connection closed');
});
