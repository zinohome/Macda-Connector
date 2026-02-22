const axios = require('axios');
const { Pool } = require('pg');

// 配置测试环境
const BFF_URL = 'http://localhost:3000';
const DB_URL = process.env.DATABASE_URL || 'postgres://postgres:passw0rd@192.168.32.17:5432/postgres?sslmode=disable';

const pool = new Pool({ connectionString: DB_URL });

async function runTests() {
    console.log('🚀 Starting Integration Tests for Alarm Masking & Persistence...');

    const testTrainId = 7002;
    const testDeviceId = 'HVAC-TEST-SCRAPER-001';
    const testFaultCode = 'Bflt_TestFault';

    try {
        // --- 准备工作：清除旧测试数据 ---
        await pool.query("DELETE FROM hvac.fact_raw WHERE device_id = $1", [testDeviceId]);
        await pool.query("DELETE FROM hvac.dim_alarm_mask WHERE device_id = $1", [testDeviceId]);

        // TEST 1: 模拟新故障产生
        console.log(`\n[TEST 1] Simulating new physical fault for Train ${testTrainId}...`);
        const rawPayload = {
            train_id: testTrainId,
            device_id: testDeviceId,
            event_time: new Date(),
            ingest_time: new Date(),
            line_id: 1,
            carriage_id: 1,
            payload_json: {
                raw: {
                    [testFaultCode]: true,
                    PresdiffU1: 500
                }
            }
        };
        await pool.query(
            "INSERT INTO hvac.fact_raw (train_id, device_id, event_time, ingest_time, line_id, carriage_id, payload_json) VALUES ($1, $2, $3, $4, $5, $6, $7)",
            [rawPayload.train_id, rawPayload.device_id, rawPayload.event_time, rawPayload.ingest_time, 1, 1, rawPayload.payload_json]
        );

        // 验证 BFF 统计 (等待一秒确保数据库可见性)
        await new Promise(r => setTimeout(r, 1000));
        let res = await axios.get(`${BFF_URL}/api/rest/AirSystem`);
        let trainData = res.data.vw_train_alarm_count.find(t => t.train_no === testTrainId);
        
        console.log(`Current state for ${testTrainId}: Alarms=${trainData.alarm_count}`);
        // 验证我们的测试故障是否被计入 (原来是14个，现在应该是15个)
        if (trainData.alarm_count < 1) throw new Error('Alarm was not registered');

        // TEST 2: 执行“删除告警” (Masking)
        console.log('\n[TEST 2] Performing "Delete Alarm" (Suppression)...');
        await axios.post(`${BFF_URL}/api/rest/v2/alarm/mask`, {
            deviceId: testDeviceId,
            faultCode: testFaultCode
        });

        // 验证统计是否减少
        res = await axios.get(`${BFF_URL}/api/rest/AirSystem`);
        let trainDataAfter = res.data.vw_train_alarm_count.find(t => t.train_no === testTrainId);
        console.log(`After masking: Alarms=${trainDataAfter.alarm_count} (Reduced by 1)`);
        if (trainDataAfter.alarm_count !== trainData.alarm_count - 1) throw new Error('Alarm count did not decrease after masking');

        // TEST 3: 模拟故障修好 (Auto-Reset)
        console.log('\n[TEST 3] Simulating fault cleared (Auto-Reset test)...');
        const clearPayload = JSON.parse(JSON.stringify(rawPayload.payload_json));
        clearPayload.raw[testFaultCode] = false;

        await pool.query(
            "INSERT INTO hvac.fact_raw (train_id, device_id, event_time, ingest_time, line_id, carriage_id, payload_json) VALUES ($1, $2, $3, $4, $5, $6, $7)",
            [testTrainId, testDeviceId, new Date(), new Date(), 1, 1, clearPayload]
        );

        await axios.get(`${BFF_URL}/api/rest/AirSystem`); // 触发同步

        const maskCheck = await pool.query("SELECT * FROM hvac.dim_alarm_mask WHERE device_id = $1", [testDeviceId]);
        console.log(`Mask table record count: ${maskCheck.rowCount} (Expected: 0)`);
        if (maskCheck.rowCount !== 0) throw new Error('Mask was not auto-cleared');

        console.log('\n✅ ALL TESTS PASSED SUCCESSFULLY!');

    } catch (err) {
        console.error('\n❌ TEST FAILED:', err.message);
        if (err.response) console.error('Response:', err.response.data);
    } finally {
        // --- 战场打扫：删除所有测试产生的原始报文和屏蔽记录 ---
        await pool.query("DELETE FROM hvac.fact_raw WHERE device_id = $1", [testDeviceId]);
        await pool.query("DELETE FROM hvac.dim_alarm_mask WHERE device_id = $1", [testDeviceId]);
        console.log('🧹 Cleanup complete: Test data removed.');
        await pool.end();
    }
}

runTests();
