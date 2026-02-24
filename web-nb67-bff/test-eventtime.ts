import { db } from './src/config/db.js';

async function test() {
    try {
        const res = await db.selectFrom('hvac.fact_event').selectAll().limit(5).execute();
        console.log("EVENTS:", res.map(r => ({ id: r.line_id, event_time: r.event_time, ingest: r.ingest_time })));
        process.exit(0);
    } catch (e) { console.error(e); process.exit(1); }
}
test();
