package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/nathanfabio/rinha2025-golang/internal/config"
	"github.com/nathanfabio/rinha2025-golang/internal/handlers"
	"github.com/nathanfabio/rinha2025-golang/internal/repository"
	"github.com/nathanfabio/rinha2025-golang/internal/services"
	"github.com/nathanfabio/rinha2025-golang/internal/worker"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisURL,
		PoolSize:     50,
		MinIdleConns: 10,
		MaxRetries:   3,
		PoolTimeout:  30 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("ERROR: error to conect to redis: %v", err)
	}

	paymentRepo := repository.NewRedisRepository(redisClient)

	healthService := services.NewHealthService(cfg)
	paymentService := services.NewPaymentService(paymentRepo, cfg, healthService)

	workerPool := worker.NewPaymentWorkerPool(paymentService, cfg.MaxWorkers)
	workerPool.Start()

	paymentHandler := handlers.NewPaymentHandler(paymentService, workerPool)

	r := chi.NewRouter()
	r.Use(middleware.Timeout(30*time.Second))

	r.Post("/payments", paymentHandler.CreatePayment)
	r.Get("/payments-summary", paymentHandler.GetPaymentsSummary)

	server := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: r,
	}


	go func(){
		log.Printf("INFO: starting server on port: %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ERROR: error to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<- quit

	log.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	workerPool.Stop()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("ERROR: error to shutting down server: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("ERROR: error to close redis connection: %v", err)
	}

	log.Println("server ended")
	
}
