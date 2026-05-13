package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// consumeTopic runs a single-topic Kafka consumer group in its own goroutine.
// handler is called once per message; marking is done after the handler returns.
// Reconnects automatically on error until ctx is cancelled.
func consumeTopic(
	ctx context.Context,
	wg *sync.WaitGroup,
	brokers []string,
	topic string,
	groupID string,
	handler func([]byte),
) {
	defer wg.Done()

	kcfg := sarama.NewConfig()
	kcfg.Version = sarama.V2_6_0_0
	kcfg.Consumer.Return.Errors = true
	kcfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	kcfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}

	h := &singleTopicHandler{handler: handler}

	for {
		group, err := sarama.NewConsumerGroup(brokers, groupID, kcfg)
		if err != nil {
			log.Printf("[ERROR] consumer group %s: create failed: %v – retrying in 5s", groupID, err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		go func() {
			for e := range group.Errors() {
				log.Printf("[WARN] consumer group %s error: %v", groupID, e)
			}
		}()

		for {
			if err := group.Consume(ctx, []string{topic}, h); err != nil {
				if ctx.Err() != nil {
					_ = group.Close()
					return
				}
				log.Printf("[ERROR] consumer group %s consume error: %v – retrying in 2s", groupID, err)
				time.Sleep(2 * time.Second)
			}
			if ctx.Err() != nil {
				_ = group.Close()
				return
			}
		}
	}
}

type singleTopicHandler struct {
	handler func([]byte)
}

func (h *singleTopicHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *singleTopicHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *singleTopicHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-sess.Context().Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			h.handler(msg.Value)
			sess.MarkMessage(msg, "")
		}
	}
}
