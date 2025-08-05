package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

func (r *RedisRepository) GetPaymentRedis(ctx context.Context, from, to time.Time) ([]models.ProcessorType, error) {
	fromScore := float64(from.Unix())
	toScore := float64(to.Unix())

	min := fmt.Sprintf("%f", fromScore)
	max := fmt.Sprintf("%f", toScore)

	results, err := r.client.ZRangeByScore(ctx, "payments", &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()


	if err != nil {
		return nil, fmt.Errorf("error to get payments from redis %v", err)
	}

	var payments []models.ProcessorType
	for _, result := range results {
		var payment models.ProcessorType
		if err := json.Unmarshal([]byte(result), &payment); err != nil {
			log.Printf("ERROR: error to deserialise payment: %v", err)
			continue
		}
		payments = append(payments, payment)
	}

	return payments, nil
}
