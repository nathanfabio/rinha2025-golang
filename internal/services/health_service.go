package services

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/nathanfabio/rinha2025-golang/internal/config"
	"github.com/nathanfabio/rinha2025-golang/internal/models"
)

type HealthService interface {
	CheckProcessorHealth(processorURL string)
	GetTimeout(processor string) int 
}

type processorStatus struct {
	ErrTimeout      int
	LastHealthCheck time.Time
}

type healthService struct {
	cfg              *config.Config
	status           map[string]*processorStatus
	healthCheckMutex sync.Mutex
}

func NewHealthService(cfg *config.Config) HealthService {
	return &healthService{
		cfg: cfg,
		status: map[string]*processorStatus{
			"default":  &processorStatus{},
			"fallback": &processorStatus{},
		},
	}
}

func (h *healthService) CheckProcessorHealth(processorURL string) {
	h.healthCheckMutex.Lock()
	defer h.healthCheckMutex.Unlock()

	processors := map[string]string{
		"default": h.cfg.DefaultProcessorURL,
		"fallback": h.cfg.FallbackProcessorURL,
	}

	for name, baseURL := range processors {
		status := h.status[name]

		if status.ErrTimeout > 0 {
			continue
		}

		if time.Since(status.LastHealthCheck) < h.cfg.HealthCheckInterval {
			continue
		}

		status.LastHealthCheck = time.Now()

		go h.checkProcessor(name, baseURL)
	}

}


func (h* healthService) checkProcessor(name, baseURL string) {
	url := baseURL + "/payments/service-health"

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()


	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var serviceHealth models.PaymentServiceHealth
	if err := json.Unmarshal(body, &serviceHealth); err != nil {
		return
	}

	h.healthCheckMutex.Lock()
	defer h.healthCheckMutex.Unlock()

	status := h.status[name]


	if serviceHealth.Failing {
		status.ErrTimeout = serviceHealth.MinResponseTime

		go func() {
			time.Sleep(10*time.Second)
			h.healthCheckMutex.Lock()
			status.ErrTimeout = 0
			h.healthCheckMutex.Unlock()
		}()
	} else {
		status.ErrTimeout = 0
	}
}

func (h *healthService) GetTimeout(processor string) int {
	h.healthCheckMutex.Lock()
	defer h.healthCheckMutex.Unlock()

	status, exists := h.status[processor]
	if !exists {
		return 0
	}

	return status.ErrTimeout
}
