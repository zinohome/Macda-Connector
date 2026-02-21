import { Pool } from 'pg';
import { Kysely, PostgresDialect } from 'kysely';
import type { Database } from '../repository/types.js';
import { config } from './index.js';

/**
 * 初始化 TimescaleDB 连接池
 */
const pool = new Pool(config.db);

/**
 * Kysely 强类型查询构造器实例
 */
export const db = new Kysely<Database>({
    dialect: new PostgresDialect({
        pool,
    }),
});

/**
 * 测试数据库连接
 */
export async function checkDbConnection() {
    try {
        const result = await db.selectFrom('hvac.fact_raw' as any)
            .select(({ fn }) => [fn.countAll().as('total')])
            .executeTakeFirst();
        console.log(`[Database] Connected to TimescaleDB. Sample count check: ${JSON.stringify(result)}`);
    } catch (err) {
        console.error('[Database] Failed to connect to TimescaleDB:', err);
    }
}
