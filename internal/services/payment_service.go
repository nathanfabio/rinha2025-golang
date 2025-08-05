package services

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/nathanfabio/rinha2025-golang/internal/config"
	"github.com/nathanfabio/rinha2025-golang/internal/models"
	"github.com/nathanfabio/rinha2025-golang/internal/repository"
)

type PaymentService interface {
	ProcessPayment(payment models.PaymentProcessorRequest) bool 
	GetPaymentsSummary(ctx context.Context, from, to time.Time) (*models.PaymentsSummaryResponse, error)
}

type paymentService struct {
	repo repository.PaymentRepository
	cfg *config.Config
	healthService HealthService
}

func NewPaymentService(repo repository.PaymentRepository, cfg *config.Config, healthService HealthService) PaymentService {
	return &paymentService{
		repo: repo,
		cfg: cfg,
		healthService: healthService,
	}
}

func (s *paymentService) ProcessPayment(payment models.PaymentProcessorRequest) bool {
	jsonData, err := json.Marshal(payment)
	if err != nil {
		log.Printf("ERROR: error to serialize payment: %v", err)
		return false
	}

	client := &http.Client{
		Timeout: 10*time.Second,
	}

	s.healthService.CheckProcessorHealth()
	
	if s.healthService.GetTimeout("default") == 0 {
		s.tryProcessor(client, s.cfg.DefaultProcessorURL, jsonData, payment, false) 
		log.Printf("INFO: payment processed with DEFAULT processor")
		return true
	}
	
	if s.healthService.GetTimeout("fallback") == 0 {
		s.tryProcessor(client, s.cfg.FallbackProcessorURL, jsonData, payment, true)
		log.Printf("INFO: payment processed with FALLBACK processor")
		return true
	}

	return false
} 

func (s *paymentService) tryProcessor(client *http.Client, processorURL string, jsonData []byte, payload models.PaymentProcessorRequest, useFallback bool) bool {
	resp, err := client.Post(processorURL+"/payments", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}
	log.Printf("INFO: status code: %d", resp.StatusCode)

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if err := s.repo.StorePayment(context.Background(), payload, useFallback); err != nil {
			log.Printf("ERROR: error to store payment: %v", err)
		}
		return true
	}

	return false
}

func (s *paymentService) GetPaymentsSummary(ctx context.Context, from, to time.Time) (*models.PaymentsSummaryResponse, error) {
	payments, err := s.repo.GetPaymentRedis(ctx, from, to)
	if err != nil {
		return nil, err
	}


	var (
		countDefault int
		countFallback int
		totalDefault float64
		totalfallback float64
	)

	for _, payment := range payments {
		if payment.UseFallback {
			countFallback++
			totalfallback += payment.Amount
		} else {
			countDefault++
			totalDefault += payment.Amount
		}
	}

	return &models.PaymentsSummaryResponse{
		Default: models.PaymentSummary{
			TotalRequests: countDefault,
			TotalAmount: math.Round(totalDefault*100) / 100,
		},
		Fallback: models.PaymentSummary{
			TotalRequests: countFallback,
			TotalAmount: math.Round(totalfallback*100) / 100,
		},
	}, nil
}