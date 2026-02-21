import pkg from 'pg';
const { Client } = pkg;

async function checkData() {
    const client = new Client({
        connectionString: 'postgres://postgres:passw0rd@192.168.32.17:5432/postgres?sslmode=disable',
    });

    try {
        await client.connect();
        console.log('--- Database Summary ---');

        // Check total count and time range
        const range = await client.query('SELECT MIN(event_time) as start, MAX(event_time) as end, COUNT(*) as total FROM hvac.fact_raw');
        console.log('Total Records:', range.rows[0].total);
        console.log('Time Range:', range.rows[0].start, 'to', range.rows[0].end);

        // Check records per train
        const trains = await client.query('SELECT train_id, COUNT(*) as count FROM hvac.fact_raw GROUP BY train_id ORDER BY count DESC LIMIT 5');
        console.log('\n--- Top 5 Trains ---');
        console.table(trains.rows);

        // Check if there are any binary flags set (alarms)
        const alarms = await client.query(`
      SELECT 'Alarms' as type, COUNT(*) as count FROM hvac.fact_raw 
      WHERE blpflt_comp_u11 = true OR blpflt_comp_u12 = true OR 
            blpflt_comp_u21 = true OR blpflt_comp_u22 = true OR
            bocflt_ef_u11 = true OR bocflt_ef_u12 = true OR
            bflt_tempover = true
    `);
        console.log('Records with Alarms:', alarms.rows[0].count);

        await client.end();
    } catch (err) {
        console.error('Failed to query DB:', err);
    }
}

checkData();
