-- =============================================================================
-- Migration: 2026-06-03
-- 变更内容：
--   I1. 新增 hvac.warning_lifecycle 表（预警生命周期，一条预警一行）
--   I2. 部分唯一索引：UNIQUE (device_id, fault_code) WHERE end_time IS NULL
--       → 同一设备同一 fault_code 在"活跃中"状态下只能有 1 行，
--         从数据库层根除"同时刻同条预警重复"问题
--   I3. 辅助索引：实时查询（end_time IS NULL）+ 历史查询（按 train_id、start_time）
--
-- 背景：
--   现有 hvac.fact_event 主键是 (event_time, device_id, fault_code)，
--   nb67-connect 每帧重跑预警规则 → 同一条预警在持续期间会写出 N 行。
--   字段 recovery_time 在代码中从未被写入，BFF 端查询 `recovery_time IS NULL`
--   等价于死过滤。详见 RET-40 评论的根因分析。
--
--   本次 migration 引入 warning_lifecycle 作为预警的"权威状态表"，
--   后续 Phase 2/3 让 storage-adapter 改为状态机写入。
--   本次 migration **不动现有 fact_event 表**，保持向后兼容（双轨运行）。
--
-- 所有语句幂等，可重复执行。
-- =============================================================================

CREATE TABLE IF NOT EXISTS hvac.warning_lifecycle (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id       VARCHAR(64)  NOT NULL,
    line_id         INTEGER,
    train_id        INTEGER,
    carriage_id     INTEGER,
    unit_id         SMALLINT,                 -- 机组号 1 / 2 / NULL
    fault_code      VARCHAR(128) NOT NULL,    -- 例：HVAC112
    fault_name      TEXT,
    warn_code       VARCHAR(64),              -- 对应 hvac.warning_config.warn_code
    severity        SMALLINT,
    start_time      TIMESTAMPTZ  NOT NULL,    -- 首次命中时刻（取 event_time）
    end_time        TIMESTAMPTZ,              -- NULL 表示仍处于活动状态
    last_seen_time  TIMESTAMPTZ  NOT NULL,    -- 最近一次仍命中的时刻
    trigger_snapshot JSONB,                   -- 触发时 warning_config 快照（#11 已有约定）
    source          VARCHAR(32) DEFAULT 'connect-rule-v2',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE  hvac.warning_lifecycle              IS '预警/报警生命周期表：一条预警从开始到结束只有一行（end_time IS NULL 表示活跃中）';
COMMENT ON COLUMN hvac.warning_lifecycle.end_time     IS 'NULL 表示活跃中；UPDATE 时填写 = 预警消除时刻';
COMMENT ON COLUMN hvac.warning_lifecycle.last_seen_time IS '用于状态机判定：若连续 N 秒 last_seen_time 未刷新，可由清扫器兜底将 end_time = last_seen_time';
COMMENT ON COLUMN hvac.warning_lifecycle.trigger_snapshot IS '触发时刻 warning_config 配置快照（阈值/持续时间），避免后续配置修改影响历史记录';

-- I2: 部分唯一索引（核心约束）
-- 同一 (device_id, fault_code) 在 end_time IS NULL 时全局只能 1 行
CREATE UNIQUE INDEX IF NOT EXISTS uq_warning_lifecycle_active
    ON hvac.warning_lifecycle (device_id, fault_code)
    WHERE end_time IS NULL;

-- I3: 查询性能索引
CREATE INDEX IF NOT EXISTS ix_warning_lifecycle_active
    ON hvac.warning_lifecycle (train_id, carriage_id, fault_code)
    WHERE end_time IS NULL;

CREATE INDEX IF NOT EXISTS ix_warning_lifecycle_history
    ON hvac.warning_lifecycle (train_id, start_time DESC);

CREATE INDEX IF NOT EXISTS ix_warning_lifecycle_device_history
    ON hvac.warning_lifecycle (device_id, fault_code, start_time DESC);

-- updated_at 自动维护
CREATE OR REPLACE FUNCTION hvac.tg_warning_lifecycle_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS tr_warning_lifecycle_updated_at ON hvac.warning_lifecycle;
CREATE TRIGGER tr_warning_lifecycle_updated_at
    BEFORE UPDATE ON hvac.warning_lifecycle
    FOR EACH ROW
    EXECUTE FUNCTION hvac.tg_warning_lifecycle_updated_at();

-- 数据保留策略：历史 lifecycle 行（end_time 非 NULL）保留 365 天
-- 活跃行（end_time IS NULL）不受保留策略影响（无 end_time 时间锚点）
-- 这里不使用 hypertable，因为 lifecycle 行总量受"活跃预警数"约束，远小于 fact_event。
-- 清理任务由后台脚本/定时器在应用层触发，避免引入额外 timescaledb 依赖。
