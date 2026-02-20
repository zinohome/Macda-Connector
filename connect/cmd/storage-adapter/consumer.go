package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type adapter struct {
	cfg  Config
	pool *pgxpool.Pool
}

type saramaMessage struct {
	inner *sarama.ConsumerMessage
}

func (m saramaMessage) GetTopic() string    { return m.inner.Topic }
func (m saramaMessage) GetPartition() int32 { return m.inner.Partition }
func (m saramaMessage) GetOffset() int64    { return m.inner.Offset }
func (m saramaMessage) GetValue() []byte    { return m.inner.Value }

const insertSQL = `
INSERT INTO hvac.fact_raw (
  event_time, ingest_time, process_time,
  line_id, train_id, carriage_id, device_id, frame_no,
  parser_version, quality_code, event_time_valid,
  wmode_u1, wmode_u2,
  f_cp_u11, f_cp_u12, f_cp_u21, f_cp_u22,
  i_cp_u11, i_cp_u12, i_cp_u21, i_cp_u22,
  suckp_u11, suckp_u12, suckp_u21, suckp_u22,
  highpress_u11, highpress_u12, highpress_u21, highpress_u22,
  fas_u1, fas_u2, ras_u1, ras_u2,
  presdiff_u1, presdiff_u2,
  aq_co2_u1, aq_co2_u2, aq_tvoc_u1, aq_tvoc_u2,
  aq_pm2_5_u1, aq_pm2_5_u2, aq_pm10_u1, aq_pm10_u2,
  bocflt_ef_u11, bocflt_ef_u12, bocflt_ef_u21, bocflt_ef_u22,
  bocflt_cf_u11, bocflt_cf_u12, bocflt_cf_u21, bocflt_cf_u22,
  blpflt_comp_u11, blpflt_comp_u12, blpflt_comp_u21, blpflt_comp_u22,
  bscflt_comp_u11, bscflt_comp_u12, bscflt_comp_u21, bscflt_comp_u22,
  bflt_tempover, bflt_diffpres_u1, bflt_diffpres_u2,
  payload_json
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,
  $12,$13,$14,$15,$16,$17,$18,$19,$20,$21,
  $22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,
  $36,$37,$38,$39,$40,$41,$42,$43,$44,$45,$46,$47,$48,$49,$50,
  $51,$52,$53,$54,$55,$56,$57,$58,$59,$60,$61,$62,$63
)
ON CONFLICT (device_id, event_time, ingest_time) DO NOTHING;
`

func (a *adapter) run(ctx context.Context) error {
	kcfg := sarama.NewConfig()
	kcfg.Version = sarama.V2_6_0_0
	kcfg.Consumer.Return.Errors = true
	kcfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	kcfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange

	group, err := sarama.NewConsumerGroup(a.cfg.KafkaBrokers, a.cfg.KafkaGroup, kcfg)
	if err != nil {
		return fmt.Errorf("create consumer group: %w", err)
	}
	defer group.Close()

	handler := &consumerGroupHandler{adapter: a}
	go func() {
		for err := range group.Errors() {
			log.Printf("[WARN] kafka consumer error: %v", err)
		}
	}()

	for {
		if err := group.Consume(ctx, []string{a.cfg.KafkaTopic}, handler); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			log.Printf("[ERROR] consume loop failed: %v", err)
			time.Sleep(2 * time.Second)
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

type consumerGroupHandler struct {
	adapter *adapter
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	batch := make([]pendingMessage, 0, h.adapter.cfg.BatchSize)
	ticker := time.NewTicker(h.adapter.cfg.FlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := h.adapter.flushBatch(sess.Context(), batch); err != nil {
			log.Printf("[ERROR] flush failed: %v", err)
			return
		}
		for _, item := range batch {
			sess.MarkMessage(item.msg.inner, "")
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-sess.Context().Done():
			flush()
			return nil
		case <-ticker.C:
			flush()
		case msg, ok := <-claim.Messages():
			if !ok {
				flush()
				return nil
			}
			var record StorageRecord
			if err := json.Unmarshal(msg.Value, &record); err != nil {
				log.Printf("[WARN] invalid json topic=%s partition=%d offset=%d err=%v", msg.Topic, msg.Partition, msg.Offset, err)
				sess.MarkMessage(msg, "invalid-json")
				continue
			}
			if len(record.PayloadJSON) == 0 {
				record.PayloadJSON = append(record.PayloadJSON, msg.Value...)
			}
			batch = append(batch, pendingMessage{record: record, msg: saramaMessage{inner: msg}})
			if len(batch) >= h.adapter.cfg.BatchSize {
				flush()
			}
		}
	}
}

func (a *adapter) flushBatch(ctx context.Context, batch []pendingMessage) error {
	txCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pgBatch := &pgx.Batch{}
	count := 0
	for _, item := range batch {
		args, err := item.record.toInsertArgs(item.msg.inner.Value)
		if err != nil {
			log.Printf("[WARN] skip record topic=%s partition=%d offset=%d err=%v", item.msg.inner.Topic, item.msg.inner.Partition, item.msg.inner.Offset, err)
			continue
		}
		pgBatch.Queue(insertSQL, args...)
		count++
	}
	if count == 0 {
		return nil
	}

	batchResult := a.pool.SendBatch(txCtx, pgBatch)
	defer batchResult.Close()
	for i := 0; i < count; i++ {
		if _, err := batchResult.Exec(); err != nil {
			return fmt.Errorf("exec batch insert: %w", err)
		}
	}
	return nil
}
