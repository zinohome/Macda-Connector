-- Connect Reimplementation TimescaleDB DDL V1
-- Date: 2026-02-20

CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS timescaledb_toolkit;

CREATE SCHEMA IF NOT EXISTS hvac;

-- 0) Runtime config: control dev/prod time axis behavior
CREATE TABLE IF NOT EXISTS hvac.runtime_config (
  id SMALLINT PRIMARY KEY DEFAULT 1,
  env_mode TEXT NOT NULL DEFAULT 'dev',         -- dev | prod
  time_axis TEXT NOT NULL DEFAULT 'ingest_time',-- ingest_time | event_time | smart
  replay_offset_ms INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT runtime_config_one_row CHECK (id = 1),
  CONSTRAINT runtime_config_env_mode CHECK (env_mode IN ('dev', 'prod')),
  CONSTRAINT runtime_config_time_axis CHECK (time_axis IN ('ingest_time', 'event_time', 'smart'))
);

INSERT INTO hvac.runtime_config(id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;

-- 1) Single Source of Truth raw table
CREATE TABLE IF NOT EXISTS hvac.fact_raw (
  event_time TIMESTAMPTZ NOT NULL,
  ingest_time TIMESTAMPTZ NOT NULL DEFAULT now(),
  process_time TIMESTAMPTZ,

  line_id INTEGER NOT NULL,
  train_id INTEGER NOT NULL,
  carriage_id INTEGER NOT NULL,
  device_id TEXT NOT NULL,
  frame_no BIGINT,

  parser_version TEXT,
  quality_code SMALLINT NOT NULL DEFAULT 0,
  event_time_valid BOOLEAN NOT NULL DEFAULT TRUE,

  wmode_u1 SMALLINT,
  wmode_u2 SMALLINT,

  f_cp_u11 DOUBLE PRECISION,
  f_cp_u12 DOUBLE PRECISION,
  f_cp_u21 DOUBLE PRECISION,
  f_cp_u22 DOUBLE PRECISION,

  i_cp_u11 DOUBLE PRECISION,
  i_cp_u12 DOUBLE PRECISION,
  i_cp_u21 DOUBLE PRECISION,
  i_cp_u22 DOUBLE PRECISION,

  suckp_u11 DOUBLE PRECISION,
  suckp_u12 DOUBLE PRECISION,
  suckp_u21 DOUBLE PRECISION,
  suckp_u22 DOUBLE PRECISION,

  highpress_u11 DOUBLE PRECISION,
  highpress_u12 DOUBLE PRECISION,
  highpress_u21 DOUBLE PRECISION,
  highpress_u22 DOUBLE PRECISION,

  fas_u1 DOUBLE PRECISION,
  fas_u2 DOUBLE PRECISION,
  ras_u1 DOUBLE PRECISION,
  ras_u2 DOUBLE PRECISION,

  presdiff_u1 DOUBLE PRECISION,
  presdiff_u2 DOUBLE PRECISION,

  aq_co2_u1 DOUBLE PRECISION,
  aq_co2_u2 DOUBLE PRECISION,
  aq_tvoc_u1 DOUBLE PRECISION,
  aq_tvoc_u2 DOUBLE PRECISION,
  aq_pm2_5_u1 DOUBLE PRECISION,
  aq_pm2_5_u2 DOUBLE PRECISION,
  aq_pm10_u1 DOUBLE PRECISION,
  aq_pm10_u2 DOUBLE PRECISION,

  bocflt_ef_u11 BOOLEAN,
  bocflt_ef_u12 BOOLEAN,
  bocflt_ef_u21 BOOLEAN,
  bocflt_ef_u22 BOOLEAN,
  bocflt_cf_u11 BOOLEAN,
  bocflt_cf_u12 BOOLEAN,
  bocflt_cf_u21 BOOLEAN,
  bocflt_cf_u22 BOOLEAN,

  blpflt_comp_u11 BOOLEAN,
  blpflt_comp_u12 BOOLEAN,
  blpflt_comp_u21 BOOLEAN,
  blpflt_comp_u22 BOOLEAN,

  bscflt_comp_u11 BOOLEAN,
  bscflt_comp_u12 BOOLEAN,
  bscflt_comp_u21 BOOLEAN,
  bscflt_comp_u22 BOOLEAN,

  bflt_tempover BOOLEAN,
  bflt_diffpres_u1 BOOLEAN,
  bflt_diffpres_u2 BOOLEAN,

  payload_json JSONB,

  PRIMARY KEY (device_id, event_time, ingest_time)
);

SELECT create_hypertable('hvac.fact_raw', 'event_time', if_not_exists => TRUE, migrate_data => TRUE);

CREATE INDEX IF NOT EXISTS idx_fact_raw_device_event_desc ON hvac.fact_raw (device_id, event_time DESC);
CREATE INDEX IF NOT EXISTS idx_fact_raw_train_event_desc ON hvac.fact_raw (line_id, train_id, event_time DESC);
CREATE INDEX IF NOT EXISTS idx_fact_raw_ingest_desc ON hvac.fact_raw (ingest_time DESC);
CREATE INDEX IF NOT EXISTS idx_fact_raw_quality ON hvac.fact_raw (quality_code, event_time DESC);

ALTER TABLE hvac.fact_raw SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'device_id',
  timescaledb.compress_orderby = 'event_time DESC'
);

SELECT add_compression_policy('hvac.fact_raw', INTERVAL '3 days', if_not_exists => TRUE);
SELECT add_retention_policy('hvac.fact_raw', INTERVAL '180 days', if_not_exists => TRUE);

-- 2) Continuous features
CREATE MATERIALIZED VIEW IF NOT EXISTS hvac.feature_5m
WITH (timescaledb.continuous) AS
SELECT
  time_bucket(INTERVAL '5 minutes', event_time) AS bucket_time,
  device_id,
  line_id,
  train_id,
  carriage_id,
  avg(f_cp_u11) AS avg_f_cp_u11,
  avg(f_cp_u12) AS avg_f_cp_u12,
  avg(f_cp_u21) AS avg_f_cp_u21,
  avg(f_cp_u22) AS avg_f_cp_u22,
  percentile_agg(suckp_u11) AS p_suckp_u11,
  percentile_agg(suckp_u12) AS p_suckp_u12,
  percentile_agg(suckp_u21) AS p_suckp_u21,
  percentile_agg(suckp_u22) AS p_suckp_u22,
  avg(abs(fas_u1 - fas_u2)) AS avg_fas_sub,
  avg(abs(ras_u1 - ras_u2)) AS avg_ras_sub,
  max(CASE WHEN bflt_tempover THEN 1 ELSE 0 END) AS max_cabin_overtemp,
  max(CASE WHEN bocflt_ef_u11 THEN 1 ELSE 0 END) AS max_bocflt_ef_u11,
  max(CASE WHEN bocflt_ef_u21 THEN 1 ELSE 0 END) AS max_bocflt_ef_u21,
  avg(presdiff_u1) AS avg_presdiff_u1,
  avg(presdiff_u2) AS avg_presdiff_u2,
  max(ingest_time) AS last_ingest_time
FROM hvac.fact_raw
GROUP BY bucket_time, device_id, line_id, train_id, carriage_id;

CREATE INDEX IF NOT EXISTS idx_feature_5m_device_bucket ON hvac.feature_5m (device_id, bucket_time DESC);
CREATE INDEX IF NOT EXISTS idx_feature_5m_train_bucket ON hvac.feature_5m (line_id, train_id, bucket_time DESC);

SELECT add_continuous_aggregate_policy('hvac.feature_5m',
  start_offset => INTERVAL '2 days',
  end_offset => INTERVAL '1 minute',
  schedule_interval => INTERVAL '1 minute',
  if_not_exists => TRUE
);

CREATE MATERIALIZED VIEW IF NOT EXISTS hvac.feature_30m
WITH (timescaledb.continuous) AS
SELECT
  time_bucket(INTERVAL '30 minutes', event_time) AS bucket_time,
  device_id,
  line_id,
  train_id,
  carriage_id,
  avg(presdiff_u1) AS avg_presdiff_u1,
  avg(presdiff_u2) AS avg_presdiff_u2,
  max(CASE WHEN bocflt_ef_u11 THEN 1 ELSE 0 END) AS max_bocflt_ef_u11,
  max(CASE WHEN bocflt_ef_u21 THEN 1 ELSE 0 END) AS max_bocflt_ef_u21,
  max(ingest_time) AS last_ingest_time
FROM hvac.fact_raw
GROUP BY bucket_time, device_id, line_id, train_id, carriage_id;

CREATE INDEX IF NOT EXISTS idx_feature_30m_device_bucket ON hvac.feature_30m (device_id, bucket_time DESC);

SELECT add_continuous_aggregate_policy('hvac.feature_30m',
  start_offset => INTERVAL '7 days',
  end_offset => INTERVAL '1 minute',
  schedule_interval => INTERVAL '5 minutes',
  if_not_exists => TRUE
);

-- 3) Event tables: alarm and predict lifecycle
CREATE TABLE IF NOT EXISTS hvac.alarm_event (
  event_id BIGSERIAL PRIMARY KEY,
  event_key TEXT NOT NULL,
  event_type TEXT NOT NULL DEFAULT 'alarm',
  level SMALLINT,
  fault_code TEXT NOT NULL,
  fault_name TEXT,

  line_id INTEGER NOT NULL,
  train_id INTEGER NOT NULL,
  carriage_id INTEGER NOT NULL,
  device_id TEXT NOT NULL,

  start_time TIMESTAMPTZ NOT NULL,
  end_time TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'open',

  source TEXT NOT NULL DEFAULT 'raw_fault',
  signal_snapshot JSONB,
  extra JSONB,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT alarm_event_status CHECK (status IN ('open', 'closed'))
);

CREATE INDEX IF NOT EXISTS idx_alarm_event_device_status ON hvac.alarm_event (device_id, status, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_alarm_event_train_range ON hvac.alarm_event (line_id, train_id, start_time DESC);

CREATE TABLE IF NOT EXISTS hvac.predict_event (
  event_id BIGSERIAL PRIMARY KEY,
  event_key TEXT NOT NULL,
  event_type TEXT NOT NULL DEFAULT 'predict',
  severity SMALLINT,
  predict_code TEXT NOT NULL,
  predict_name TEXT,
  score DOUBLE PRECISION,

  line_id INTEGER NOT NULL,
  train_id INTEGER NOT NULL,
  carriage_id INTEGER NOT NULL,
  device_id TEXT NOT NULL,

  start_time TIMESTAMPTZ NOT NULL,
  end_time TIMESTAMPTZ,
  status TEXT NOT NULL DEFAULT 'open',

  source TEXT NOT NULL DEFAULT 'rule_engine',
  feature_snapshot JSONB,
  extra JSONB,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT predict_event_status CHECK (status IN ('open', 'closed'))
);

CREATE INDEX IF NOT EXISTS idx_predict_event_device_status ON hvac.predict_event (device_id, status, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_predict_event_train_range ON hvac.predict_event (line_id, train_id, start_time DESC);

SELECT add_retention_policy('hvac.alarm_event', INTERVAL '365 days', if_not_exists => TRUE);
SELECT add_retention_policy('hvac.predict_event', INTERVAL '365 days', if_not_exists => TRUE);

-- 4) Rule config tables
CREATE TABLE IF NOT EXISTS hvac.predict_rule_config (
  rule_id TEXT PRIMARY KEY,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  feature_layer TEXT NOT NULL DEFAULT 'feature_5m',
  expression TEXT NOT NULL,
  severity SMALLINT,
  cooldown_sec INTEGER NOT NULL DEFAULT 300,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS hvac.alert_rule_config (
  rule_id TEXT PRIMARY KEY,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  source_type TEXT NOT NULL DEFAULT 'raw',
  expression TEXT NOT NULL,
  level SMALLINT,
  debounce_open_sec INTEGER NOT NULL DEFAULT 30,
  debounce_close_sec INTEGER NOT NULL DEFAULT 60,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 5) Query time helper view for dev/prod switching
CREATE OR REPLACE VIEW hvac.vw_fact_raw_query_time AS
SELECT
  r.*,
  CASE
    WHEN c.time_axis = 'ingest_time' THEN r.ingest_time
    WHEN c.time_axis = 'event_time' THEN r.event_time
    ELSE CASE WHEN r.event_time_valid THEN r.event_time ELSE r.ingest_time END
  END
  + make_interval(secs => c.replay_offset_ms / 1000.0) AS query_time
FROM hvac.fact_raw r
CROSS JOIN hvac.runtime_config c;

-- 6) Service layer views for dashboard/api
CREATE OR REPLACE VIEW hvac.vw_train_overview_rt AS
SELECT
  line_id,
  train_id,
  count(DISTINCT carriage_id) AS carriage_count,
  max(query_time) AS last_time
FROM hvac.vw_fact_raw_query_time
WHERE query_time > now() - INTERVAL '10 minutes'
GROUP BY line_id, train_id;

CREATE OR REPLACE VIEW hvac.vw_carriage_health_rt AS
SELECT
  line_id,
  train_id,
  carriage_id,
  device_id,
  max(query_time) AS last_time,
  max(CASE WHEN bflt_tempover THEN 1 ELSE 0 END) AS tempover_flag,
  max(CASE WHEN bocflt_ef_u11 OR bocflt_ef_u12 OR bocflt_ef_u21 OR bocflt_ef_u22 THEN 1 ELSE 0 END) AS fan_fault_flag
FROM hvac.vw_fact_raw_query_time
WHERE query_time > now() - INTERVAL '10 minutes'
GROUP BY line_id, train_id, carriage_id, device_id;

CREATE OR REPLACE VIEW hvac.vw_alarm_list_open AS
SELECT
  event_id,
  line_id,
  train_id,
  carriage_id,
  device_id,
  fault_code,
  fault_name,
  level,
  start_time,
  end_time,
  status,
  source,
  signal_snapshot
FROM hvac.alarm_event
WHERE status = 'open';

CREATE OR REPLACE VIEW hvac.vw_predict_list_open AS
SELECT
  event_id,
  line_id,
  train_id,
  carriage_id,
  device_id,
  predict_code,
  predict_name,
  severity,
  score,
  start_time,
  end_time,
  status,
  source,
  feature_snapshot
FROM hvac.predict_event
WHERE status = 'open';

-- 7) Recommended grants placeholder
-- GRANT USAGE ON SCHEMA hvac TO app_reader;
-- GRANT SELECT ON ALL TABLES IN SCHEMA hvac TO app_reader;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA hvac GRANT SELECT ON TABLES TO app_reader;
