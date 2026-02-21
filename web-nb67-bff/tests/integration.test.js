const axios = require('axios');
const { Pool } = require('pg');

// 配置测试环境
const BFF_URL = 'http://localhost:3000';
const DB_URL = process.env.DATABASE_URL || 'postgres://postgres:passw0rd@192.168.32.17:5432/postgres?sslmode=disable';

const pool = new Pool({ connectionString: DB_URL });

async function runTests() {
    console.log('🚀 Starting Integration Tests for Alarm Masking & Persistence...');

    const testTrainId = 9999;
    const testDeviceId = 'TEST-DEVICE-001';
    const testFaultCode = 'Bflt_TestFault';

    try {
        // --- 准备工作：清除旧测试数据 ---
        await pool.query("DELETE FROM hvac.fact_raw WHERE train_id = $1", [testTrainId]);
        await pool.query("DELETE FROM hvac.dim_alarm_mask WHERE device_id = $1", [testDeviceId]);

        // TEST 1: 模拟新故障产生
        console.log('\n[TEST 1] Simulating new physical fault...');
        const rawPayload = {
            train_id: testTrainId,
            device_id: testDeviceId,
            event_time: new Date(),
            ingest_time: new Date(),
            line_id: 1,
            carriage_id: 1,
            payload_json: {
                raw: {
                    [testFaultCode]: true, // 开启故障位
                    PresdiffU1: 500 // 正常压差
                }
            }
        };
        await pool.query(
            "INSERT INTO hvac.fact_raw (train_id, device_id, event_time, ingest_time, line_id, carriage_id, payload_json) VALUES ($1, $2, $3, $4, $5, $6, $7)",
            [rawPayload.train_id, rawPayload.device_id, rawPayload.event_time, rawPayload.ingest_time, 1, 1, rawPayload.payload_json]
        );

        // 验证 BFF 统计
        let res = await axios.get(`${BFF_URL}/api/rest/AirSystem`);
        let trainData = res.data.vw_train_alarm_count.find(t => t.train_no === testTrainId);
        console.log(`Initial count: Alarm=${trainData.alarm_count}, Warning=${trainData.warning_count}`);
        if (trainData.alarm_count !== 1) throw new Error('Alarm count should be 1');

        // TEST 2: 执行“删除告警” (Masking)
        console.log('\n[TEST 2] Performing "Delete Alarm" (Suppression)...');
        await axios.post(`${BFF_URL}/api/rest/v2/alarm/mask`, {
            deviceId: testDeviceId,
            faultCode: testFaultCode
        });

        // 验证统计是否减少
        res = await axios.get(`${BFF_URL}/api/rest/AirSystem`);
        trainData = res.data.vw_train_alarm_count.find(t => t.train_no === testTrainId);
        console.log(`After masking: Alarm=${trainData.alarm_count} (Expected: 0)`);
        if (trainData.alarm_count !== 0) throw new Error('Alarm count should be 0 after masking');

        // TEST 3: 模拟故障修好 (Auto-Reset)
        console.log('\n[TEST 3] Simulating fault cleared (Auto-Reset test)...');
        const clearPayload = JSON.parse(JSON.stringify(rawPayload.payload_json));
        clearPayload.raw[testFaultCode] = false; // 故障消失

        await pool.query(
            "INSERT INTO hvac.fact_raw (train_id, device_id, event_time, ingest_time, line_id, carriage_id, payload_json) VALUES ($1, $2, $3, $4, $5, $6, $7)",
            [testTrainId, testDeviceId, new Date(), new Date(), 1, 1, clearPayload]
        );

        // 调用统计接口触发 syncMasks
        await axios.get(`${BFF_URL}/api/rest/AirSystem`);

        // 验证屏蔽表是否自动清空
        const maskCheck = await pool.query("SELECT * FROM hvac.dim_alarm_mask WHERE device_id = $1", [testDeviceId]);
        console.log(`Mask table record count: ${maskCheck.rowCount} (Expected: 0)`);
        if (maskCheck.rowCount !== 0) throw new Error('Mask should be automatically removed when fault is cleared');

        console.log('\n✅ ALL TESTS PASSED SUCCESSFULLY!');

    } catch (err) {
        console.error('\n❌ TEST FAILED:', err.message);
        if (err.response) console.error('Response:', err.response.data);
    } finally {
        await pool.end();
    }
}

runTests();
