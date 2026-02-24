import { db } from './src/config/db.js';

async function test() {
    const res = await db.selectFrom('hvac.fact_event').select(['train_id', db.fn.count('train_id').as('cnt')]).groupBy('train_id').execute();
    console.log("ALL TRAINS:", res);
    process.exit(0);
}
test().catch(console.error);
