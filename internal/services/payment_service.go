package services

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nathanfabio/rinha2025-golang/internal/config"
	"github.com/nathanfabio/rinha2025-golang/internal/models"
	"github.com/nathanfabio/rinha2025-golang/internal/repository"
)

type PaymentService interface {
}

type paymentService struct {
	repo repository.PaymentRepository
	cfg *config.Config
}

func NewPaymentService(repo repository.PaymentRepository, cfg *config.Config) PaymentService {
	return &paymentService{
		repo: repo,
		cfg: cfg,
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
	
	if s.tryProcessor(client, s.cfg.DefaultProcessorURL, jsonData, payment, false) {
		return true
	}

	if s.tryProcessor(client, s.cfg.FallbackProcessorURL, jsonData, payment, true) {
		return true
	}

	return false
} 

func (s *paymentService) tryProcessor(client *http.Client, processorURL string, jsonData []byte, payload models.PaymentProcessorRequest, useFallback bool) bool {
	resp, err := client.Post(processorURL+"/payments", "aplication/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if err := s.repo.StorePayment(context.Background(), payload, useFallback); err != nil {
			log.Printf("ERROR: error to store payment: %v", err)
		}
		return true
	}

	return false
}