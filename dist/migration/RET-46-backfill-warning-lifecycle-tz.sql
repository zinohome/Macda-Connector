-- =============================================================================
-- RET-46 / GitHub #24 修复后的脏数据回填脚本
--
-- 背景：storage-adapter v2.5.26 及更早版本的 parseTimeText 把上游裸时间戳
--      "2026-05-28 16:28:46"（Asia/Shanghai 本地时间）当成 UTC 解析，
--      导致 hvac.warning_lifecycle 的 start_time / last_seen_time / end_time
--      比真实时刻多 8 小时。
-- 修复版本：storage-adapter:v2.5.30（含 parseTimeText 走 Asia/Shanghai 的修复）
--
-- 使用方式：
--   1. 部署 storage-adapter:v2.5.30 之后，记录精确的容器启动时间 $CUTOFF
--      （注意：是新版镜像替换旧容器的精确时刻，而不是脚本执行时刻）
--   2. 把脚本里的 :cutoff 占位换成 $CUTOFF
--   3. 在事务里跑一次（脚本只能跑一次！再跑会重复 -8h）
-- =============================================================================

\set cutoff '2026-06-12 00:00:00+08'   -- ← 部署前必须改成真实 cutoff

BEGIN;

-- ① 先看一眼会受影响多少行（不改数据）
SELECT
    count(*)                                AS rows_total,
    count(*) FILTER (WHERE end_time IS NULL) AS rows_active,
    count(*) FILTER (WHERE end_time IS NOT NULL) AS rows_closed,
    min(start_time) AS oldest_start,
    max(start_time) AS newest_start
  FROM hvac.warning_lifecycle
 WHERE created_at <  :'cutoff' ::timestamptz
   AND start_time  >= '2026-06-03'::timestamptz;

-- ② 回填：三列各减 8 小时（end_time 仅在非空时减）
UPDATE hvac.warning_lifecycle
   SET start_time     = start_time     - INTERVAL '8 hours',
       last_seen_time = last_seen_time - INTERVAL '8 hours',
       end_time       = CASE WHEN end_time IS NOT NULL
                             THEN end_time - INTERVAL '8 hours'
                             ELSE NULL END
 WHERE created_at <  :'cutoff' ::timestamptz
   AND start_time  >= '2026-06-03'::timestamptz;

-- ③ 二次抽查：随机看 5 行确认时间合理
SELECT id, device_id, fault_code, start_time, last_seen_time, end_time, created_at
  FROM hvac.warning_lifecycle
 ORDER BY start_time DESC
 LIMIT 5;

-- ④ 如果第 ① 步行数和第 ③ 步抽样都合理，再 COMMIT；否则 ROLLBACK；
--    脚本默认 COMMIT，需要确认后手动调成 ROLLBACK 取消。
COMMIT;
