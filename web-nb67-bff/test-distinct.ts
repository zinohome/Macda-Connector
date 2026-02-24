import { db } from './src/config/db.js';
import { sql } from 'kysely';

async function test() {
    try {
        const timeCol = 'ingest_time';
        let q = db.selectFrom('hvac.fact_raw')
            .select(sql<string>`TO_CHAR(${sql.raw(timeCol)}, 'YYYY-MM-DD')`.as('date'))
            .distinct()
            .limit(5);
        console.log(q.compile().sql);
        const records = await q.execute();
        console.log("SUCCESS:", records);
        process.exit(0);
    } catch (err) {
        console.error("DB ERROR", err);
        process.exit(1);
    }
}
test();
