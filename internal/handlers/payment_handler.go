package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nathanfabio/rinha2025-golang/internal/models"
	"github.com/nathanfabio/rinha2025-golang/internal/services"
	"github.com/nathanfabio/rinha2025-golang/internal/worker"
)

type PaymentHandler struct {
	paymentService services.PaymentService
	workerPool worker.PaymentWorkerPool
}

func NewPaymentHandler(paymentService services.PaymentService, workerPool worker.PaymentWorkerPool) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		workerPool: workerPool,
	}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var payment models.PaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, "invalid data", http.StatusUnprocessableEntity)
		return
	}

	if payment.CorrelationID == uuid.Nil || payment.Amount <= 0 {
		http.Error(w, "correlationID and amount are required", http.StatusBadRequest)
		return
	}

	paymentProcessorRequest := models.PaymentProcessorRequest{
		Amount: payment.Amount,
		CorrelationID: payment.CorrelationID,
		RequestedAt: time.Now().UTC(),
	}

	h.workerPool.EnqueuePayment(paymentProcessorRequest)

	w.WriteHeader(http.StatusNoContent)
}

func (h *PaymentHandler) GetPaymentsSummary(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var (
		from time.Time
		to time.Time
		err error
	)
	
	if fromStr == "" {
		from = time.Time{}
	} else {
		from, err = time.Parse("2006-01-02T15:04:05.000Z", fromStr)
		if err != nil {
			http.Error(w, "'from' parameter invalid", http.StatusBadRequest)
			return
		}
	}

	if toStr == "" {
		to = time.Time{}
	} else {
		to, err = time.Parse("2006-01-02T15:04:05.000Z", toStr)
		if err != nil {
			http.Error(w, "'to' paramter invalid", http.StatusBadRequest)
			return
		}
	}

	if from.After(to) {
		http.Error(w, "'from' should be before 'to'", http.StatusBadRequest)
		return
	}

	summary, err := h.paymentService.GetPaymentsSummary(r.Context(), from, to)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}