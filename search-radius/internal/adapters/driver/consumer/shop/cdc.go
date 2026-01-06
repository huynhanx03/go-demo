package shop

import (
	"context"
	"time"

	"go.uber.org/zap"

	"search-radius/global"
	"search-radius/internal/constant"
	"search-radius/internal/core/entity"
	"search-radius/internal/ports"
	"search-radius/pkg/cdc"
	"search-radius/pkg/common/workerpool"
	"search-radius/pkg/mq/kafka"
	"search-radius/pkg/utils"
)

const (
	poolSize = 10
)

type CDCConsumer struct {
	consumer    kafka.ConsumerGroup
	shopService ports.ShopService
	workerPool  *workerpool.GenericPool[[]*cdc.DebeziumPayload[entity.Shop]]
}

func NewCDCConsumer(cfg *kafka.Config, shopService ports.ShopService) (ports.ShopConsumer, error) {
	c, err := kafka.NewConsumer(cfg, constant.ConsumerGroupShopCDC, kafka.Recovery)
	if err != nil {
		return nil, err
	}

	taskFunc := func(batch []*cdc.DebeziumPayload[entity.Shop]) {
		if err := shopService.HandleShopBatchChange(context.Background(), batch); err != nil {
			global.Logger.Error("Failed to process shop batch", zap.Error(err))
		}
	}

	pool, err := workerpool.NewGenericPool(poolSize, taskFunc)
	if err != nil {
		return nil, err
	}

	return &CDCConsumer{
		consumer:    c,
		shopService: shopService,
		workerPool:  pool,
	}, nil
}

// Start starts the consumer
func (c *CDCConsumer) Start(ctx context.Context) error {
	batchSize := global.Config.Kafka.ConsumerBatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	batchInterval := utils.ToDurationMs(global.Config.Kafka.ConsumerBatchInterval)
	if batchInterval <= 0 {
		batchInterval = 500 * time.Millisecond
	}

	batchChan := make(chan *cdc.DebeziumPayload[entity.Shop], batchSize*2)

	go c.processBatchLoop(ctx, batchChan, batchSize, batchInterval)

	handler := func(ctx context.Context, key, value []byte) error {
		_, payload, err := c.extractPayload(key, value)
		if err != nil {
			global.Logger.Warn("Skipping malformed CDC message", zap.Error(err))
			return nil
		}

		if payload == nil {
			return nil
		}

		select {
		case batchChan <- payload:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	}

	errHandler := func(err error) {
		global.Logger.Error("ShopCDCConsumer error", zap.Error(err))
	}

	global.Logger.Info("Starting Shop CDC Consumer (Batch Mode)", zap.String("topic", constant.TopicShopCDC))
	return c.consumer.Start(ctx, []string{constant.TopicShopCDC}, handler, errHandler)
}

func (c *CDCConsumer) Stop() error {
	return c.consumer.Close()
}

func (c *CDCConsumer) processBatchLoop(
	ctx context.Context,
	batchChan <-chan *cdc.DebeziumPayload[entity.Shop],
	batchSize int,
	batchInterval time.Duration,
) {
	batch := make([]*cdc.DebeziumPayload[entity.Shop], 0, batchSize)
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	flush := func() {
		size := len(batch)
		if size == 0 {
			return
		}

		finalBatch := make([]*cdc.DebeziumPayload[entity.Shop], size)
		copy(finalBatch, batch)

		if err := c.workerPool.Invoke(finalBatch); err != nil {
			global.Logger.Error("Failed to submit batch to worker pool", zap.Error(err))
		}

		batch = batch[:0]
	}

	for {
		select {
		case msg := <-batchChan:
			batch = append(batch, msg)
			if len(batch) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-ctx.Done():
			flush()
			return
		}
	}
}

func (c *CDCConsumer) extractPayload(key, value []byte) (int, *cdc.DebeziumPayload[entity.Shop], error) {
	if len(value) == 0 {
		return 0, nil, nil
	}

	cdcPayload, err := cdc.ParseDebeziumMessage[CDCShop](value)
	if err != nil {
		return 0, nil, err
	}

	payload := &cdc.DebeziumPayload[entity.Shop]{
		Source: cdcPayload.Source,
		Op:     cdcPayload.Op,
		TsMs:   cdcPayload.TsMs,
	}

	if cdcPayload.After != nil {
		payload.After = cdcPayload.After.ToEntity()
	}

	if cdcPayload.Before != nil {
		payload.Before = cdcPayload.Before.ToEntity()
	}

	// For Delete, if Before is present, we have ID.
	var id int
	if payload.Before != nil {
		id = payload.Before.ID
	} else if payload.After != nil {
		id = payload.After.ID
	}

	return id, payload, nil
}
