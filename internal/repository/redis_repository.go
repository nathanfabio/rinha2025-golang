package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nathanfabio/rinha2025-golang/internal/models"
	"github.com/redis/go-redis/v9"
)

type PaymentRepository interface {
}

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) PaymentRepository {
	return &RedisRepository{
		client: client,
	}
}

func (r *RedisRepository) StorePayment(ctx context.Context, payment models.PaymentProcessorRequest, useFallback bool) error {
	processor := models.ProcessorType{
		PaymentProcessorRequest: payment,
		UseFallback:             useFallback,
	}

	//requestedAtStr := payment.RequestedAt.Format(time.RFC3339)

	jsonData, err := json.Marshal(processor)
	if err != nil {
		return fmt.Errorf("error to serialize payment: %v", err)
	}

	err = r.client.ZAdd(ctx, "payments", redis.Z{
		Score: float64(payment.RequestedAt.Unix()),
		Member: string(jsonData),
	}).Err()

	if err != nil {
		return fmt.Errorf("error to store payment: %v", err)
	}

	return nil
}
