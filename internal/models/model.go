package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentRequest struct {
	CorrelationID uuid.UUID `json:"correlationId"`
	Amount        float64   `json:"amount"`
}

type PaymentProcessorRequest struct {
	CorrelationID uuid.UUID `json:"correlationId"`
	Amount        float64   `json:"amount"`
	RequestedAt   time.Time `json:"requestedAt"`
}

type PaymentSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type PaymentsSummaryResponse struct {
	Default  PaymentSummary `json:"default"`
	Fallback PaymentSummary `json:"fallback"`
}

type PaymentServiceHealth struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}
