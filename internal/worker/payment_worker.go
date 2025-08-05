package worker

import (
	"log"
	"sync"

	"github.com/nathanfabio/rinha2025-golang/internal/models"
	"github.com/nathanfabio/rinha2025-golang/internal/services"
)

type PaymentWorkerPool interface {
	Start()
	Stop()
	EnqueuePayment(request models.PaymentProcessorRequest)
}

type paymentWorkerPool struct {
	paymentService services.PaymentService
	paymentQueue   chan models.PaymentProcessorRequest
	wg             sync.WaitGroup
	maxWorkers     int
	quit           chan struct{}
}

func NewPaymentWorkerPool(paymentService services.PaymentService, maxWorkers int) PaymentWorkerPool {
	return &paymentWorkerPool{
		paymentService: paymentService,
		paymentQueue:   make(chan models.PaymentProcessorRequest, 10000),
		maxWorkers:     maxWorkers,
		quit:           make(chan struct{}),
	}
}

func (p *paymentWorkerPool) EnqueuePayment(request models.PaymentProcessorRequest) {
	select {
	case p.paymentQueue <- request:
	default:
		log.Println("INFO: queue is full")
	}
}

func (p *paymentWorkerPool) paymentWorker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.quit:
			return
		case payment, ok := <-p.paymentQueue:
			if !ok {
				return
			}
			p.processPayment(payment, id)
		}
	}
}

func (p *paymentWorkerPool) processPayment(payment models.PaymentProcessorRequest, workerID int) {
	success := p.paymentService.ProcessPayment(payment)

	if success {
		log.Printf("INFO: Worker %d: Payment successufully processed with CorrelationID: %s", workerID, payment.CorrelationID)
		return
	} else {
		log.Println("ERROR: payment failed!")
	}

}

func (p *paymentWorkerPool) Start() {
	for i := 1; i <= p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.paymentWorker(i)
	}
	log.Printf("INFO: worker pool started with %d workers", p.maxWorkers)
}

func (p *paymentWorkerPool) Stop() {
	close(p.quit)
	close(p.paymentQueue)
	p.wg.Wait()
	log.Println("INFO: worker pool ended")
}
