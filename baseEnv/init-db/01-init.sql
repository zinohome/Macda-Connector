-- 1. 确保加载 TimescaleDB 扩展
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- 2. 设置当前数据库的全局默认时区为东八区
ALTER DATABASE postgres SET timezone TO 'Asia/Shanghai';

-- 3. 创建业务 Schema
CREATE SCHEMA IF NOT EXISTS hvac;

-- 4. 创建核心超表：实时传感器遥测与状态宽表 (Hypertable)
CREATE TABLE IF NOT EXISTS hvac.fact_raw (
    event_time TIMESTAMPTZ NOT NULL,              -- 报文产生时间 (带时序，将自动转换为上海时间展示)
    ingest_time TIMESTAMPTZ NOT NULL,             -- 网关接收时间
    process_time TIMESTAMPTZ,                     -- 解析处理时间
    event_time_valid BOOLEAN,                     -- event_time 是否有效合法(防漂移)
    
    line_id INTEGER,                              -- 线路编号
    train_id INTEGER,                             -- 列车编号
    carriage_id INTEGER,                          -- 车厢编号
    device_id VARCHAR(64) NOT NULL,               -- 设备唯一标识 (如 'HVAC-67-12345-01')
    frame_no BIGINT,                              -- 帧序号
    
    parser_version VARCHAR(32),                   -- 解析器版本
    quality_code SMALLINT,                        -- 数据质量码
    
    wmode_u1 SMALLINT, wmode_u2 SMALLINT,         -- 运行模式
    f_cp_u11 NUMERIC(8,2), f_cp_u12 NUMERIC(8,2), -- 压缩机频率 (Hz)
    f_cp_u21 NUMERIC(8,2), f_cp_u22 NUMERIC(8,2),
    i_cp_u11 NUMERIC(8,2), i_cp_u12 NUMERIC(8,2), -- 压缩机电流 (A)
    i_cp_u21 NUMERIC(8,2), i_cp_u22 NUMERIC(8,2),
    suckp_u11 NUMERIC(8,2), suckp_u12 NUMERIC(8,2), -- 吸气压力
    suckp_u21 NUMERIC(8,2), suckp_u22 NUMERIC(8,2),
    highpress_u11 NUMERIC(8,2), highpress_u12 NUMERIC(8,2), -- 高压
    highpress_u21 NUMERIC(8,2), highpress_u22 NUMERIC(8,2),
    
    fas_u1 NUMERIC(6,2), fas_u2 NUMERIC(6,2),     -- 综合新风温度
    ras_u1 NUMERIC(6,2), ras_u2 NUMERIC(6,2),     -- 综合回风温度
    presdiff_u1 NUMERIC(8,2), presdiff_u2 NUMERIC(8,2), -- 滤网压差
    aq_co2_u1 NUMERIC(8,2), aq_co2_u2 NUMERIC(8,2), -- CO2
    aq_tvoc_u1 NUMERIC(8,2), aq_tvoc_u2 NUMERIC(8,2), -- TVOC
    aq_pm2_5_u1 NUMERIC(8,2), aq_pm2_5_u2 NUMERIC(8,2), -- PM2.5
    aq_pm10_u1 NUMERIC(8,2), aq_pm10_u2 NUMERIC(8,2), -- PM10

    bocflt_ef_u11 BOOLEAN, bocflt_ef_u12 BOOLEAN,    -- 通风机过流等保护
    bocflt_ef_u21 BOOLEAN, bocflt_ef_u22 BOOLEAN,
    bocflt_cf_u11 BOOLEAN, bocflt_cf_u12 BOOLEAN,    -- 冷凝风机过流
    bocflt_cf_u21 BOOLEAN, bocflt_cf_u22 BOOLEAN,
    blpflt_comp_u11 BOOLEAN, blpflt_comp_u12 BOOLEAN,-- 压缩机低压故障
    blpflt_comp_u21 BOOLEAN, blpflt_comp_u22 BOOLEAN,
    bscflt_comp_u11 BOOLEAN, bscflt_comp_u12 BOOLEAN,-- 压缩机通讯故障
    bscflt_comp_u21 BOOLEAN, bscflt_comp_u22 BOOLEAN,
    bflt_tempover BOOLEAN,                           -- 整体车厢超级高温报警
    bflt_diffpres_u1 BOOLEAN, bflt_diffpres_u2 BOOLEAN, -- 滤网压差故障

    payload_json JSONB,                           -- 万能兜底全量解析 JSON
    UNIQUE (device_id, event_time, ingest_time)
);

COMMENT ON TABLE hvac.fact_raw IS '设备明细底层存根数据宽表';

-- 将表转化为 Hypertable，极速写入及查询。以 event_time 作为时序切片键，分块长度7天
SELECT create_hypertable('hvac.fact_raw', 'event_time', chunk_time_interval => INTERVAL '7 days');

-- 为经常进行 UI 端过滤的 device_id (车厢级别) 加速建立索引
CREATE INDEX IF NOT EXISTS ix_fact_raw_device_time ON hvac.fact_raw (device_id, event_time DESC);

-- 5. 创建告警屏蔽表 (用于业务上的“删除告警”/告警抑制)
CREATE TABLE IF NOT EXISTS hvac.dim_alarm_mask (
    device_id VARCHAR(64) NOT NULL,
    fault_code VARCHAR(128) NOT NULL,
    masked_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (device_id, fault_code)
);
COMMENT ON TABLE hvac.dim_alarm_mask IS '故障屏蔽/抑制状态表，用于处理用户在界面上删除告警的逻辑';

-- 6. 创建历史事件超表：存储计算后的告警、预警、寿命等事件
CREATE TABLE IF NOT EXISTS hvac.fact_event (
    event_time TIMESTAMPTZ NOT NULL,              -- 事件记录时间
    line_id INTEGER,                              -- 线路编号
    train_id INTEGER,                             -- 列车编号
    carriage_id INTEGER,                          -- 车厢编号
    device_id VARCHAR(64) NOT NULL,               -- 设备唯一标识
    event_type VARCHAR(32),                       -- 事件分类: alarm, predict, life
    fault_code VARCHAR(128) NOT NULL,             -- 故障/预警代码
    fault_name TEXT,                              -- 故障/预警名称
    severity SMALLINT,                            -- 严重等级 (0, 1, 2...)
    status VARCHAR(16) DEFAULT 'open',            -- 事件状态: open, resolved
    payload_json JSONB,                           -- 事件发生时的上下文详情
    PRIMARY KEY (event_time, device_id, fault_code)
);
COMMENT ON TABLE hvac.fact_event IS '经规则引擎处理后的事件事实表，支持历史追溯与报表分析';

-- 将事件表转化为 Hypertable
SELECT create_hypertable('hvac.fact_event', 'event_time', if_not_exists => TRUE);
CREATE INDEX IF NOT EXISTS ix_fact_event_searching ON hvac.fact_event (train_id, event_type, event_time DESC);

-- =============================================================================
-- 构建服务于前端 Web 展现的连续聚合层 (Continuous Aggregates)
-- =============================================================================

CREATE MATERIALIZED VIEW IF NOT EXISTS hvac.mv_temperature_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', event_time) AS bucket_time,
    train_id,
    carriage_id,
    AVG(ras_u1) AS avg_ras_u1,   
    AVG(ras_u2) AS avg_ras_u2,   
    MAX(fas_u1) AS max_fas_u1,   
    MAX(fas_u2) AS max_fas_u2    
FROM hvac.fact_raw
WHERE event_time_valid = true
GROUP BY bucket_time, train_id, carriage_id;

-- 定时拉起持续聚合刷新任务
SELECT add_continuous_aggregate_policy('hvac.mv_temperature_hourly',
  start_offset => INTERVAL '3 days',
  end_offset => INTERVAL '1 hour',
  schedule_interval => INTERVAL '1 hour');
