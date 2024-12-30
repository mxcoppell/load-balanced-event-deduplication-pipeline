package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mxie/load-balanced-event-deduplication-pipeline/pkg/redis"
)

type TestConfig struct {
	NumKeys     int64 `json:"num_keys"`
	KeyDelay    int64 `json:"key_delay"`    // milliseconds
	KeyTTL      int64 `json:"key_ttl"`      // milliseconds
	DedupWindow int64 `json:"dedup_window"` // milliseconds
}

type TestStatus struct {
	IsRunning bool  `json:"is_running"`
	Generated int64 `json:"generated"`
	Consumed  int64 `json:"consumed"`
}

// TestMetrics represents the metrics for a key generation test
type TestMetrics struct {
	Generated int64            `json:"generated"`
	Consumed  int64            `json:"consumed"`
	Consumers map[string]int64 `json:"consumers"`
}

type server struct {
	redis     *redis.Client
	isRunning bool
	mu        sync.Mutex
}

const (
	defaultRedisTimeout = 2 * time.Second
)

func startWebServer(ctx context.Context, redisClient *redis.Client) error {
	s := &server{
		redis: redisClient,
	}

	// API endpoints
	http.HandleFunc("/api/start", s.handleStart)
	http.HandleFunc("/api/stop", s.handleStop)
	http.HandleFunc("/api/status", s.handleStatus)
	http.HandleFunc("/api/metrics", s.getTestMetrics)

	// Serve static files for the UI
	http.Handle("/", http.FileServer(http.Dir("web/dist")))

	server := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	// Run server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return server.Shutdown(context.Background())
}

func (s *server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		http.Error(w, "Test is already running", http.StatusConflict)
		return
	}

	var config TestConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		s.mu.Unlock()
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	log.Printf("Starting test with config: NumKeys=%d, KeyDelay=%dms, KeyTTL=%dms, DedupWindow=%dms",
		config.NumKeys, config.KeyDelay, config.KeyTTL, config.DedupWindow)

	// Create a new background context for the key generation
	genCtx, genCancel := context.WithCancel(context.Background())

	// Reset metrics before starting
	if err := s.redis.ResetMetrics(r.Context()); err != nil {
		genCancel()
		s.mu.Unlock()
		http.Error(w, fmt.Sprintf("Failed to reset metrics: %v", err), http.StatusInternalServerError)
		return
	}

	s.isRunning = true
	s.mu.Unlock()

	// Start generating keys in background
	go func() {
		s.generateKeys(genCtx, config)
		genCancel() // Clean up when done
	}()

	w.WriteHeader(http.StatusOK)
}

func (s *server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	isRunning := s.isRunning
	s.mu.Unlock()

	generated, consumed, err := s.redis.GetMetrics(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metrics: %v", err), http.StatusInternalServerError)
		return
	}

	status := TestStatus{
		IsRunning: isRunning,
		Generated: generated,
		Consumed:  consumed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *server) generateKeys(ctx context.Context, config TestConfig) {
	log.Printf("Starting key generation: total keys=%d", config.NumKeys)
	var i int64
	for i = 0; i < config.NumKeys; i++ {
		// Check if we should stop
		s.mu.Lock()
		if !s.isRunning {
			log.Printf("Test stopped at key %d", i)
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()

		// Create timeout context for Redis operation
		opCtx, cancel := context.WithTimeout(ctx, defaultRedisTimeout)

		// Generate key with TTL
		if err := s.redis.GenerateKey(opCtx, i, time.Duration(config.KeyTTL)*time.Millisecond); err != nil {
			log.Printf("Failed to generate key %d: %v", i, err)
			cancel()
			continue
		}
		cancel()

		if i > 0 && i%100 == 0 {
			log.Printf("Generated %d keys", i)
		}

		// Wait for configured delay
		time.Sleep(time.Duration(config.KeyDelay) * time.Millisecond)
	}

	log.Printf("Key generation complete: generated %d keys", i)

	// Test complete
	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()
}

func (s *server) getTestMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get generated and consumed counts
	generated, consumed, err := s.redis.GetMetrics(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metrics: %v", err), http.StatusInternalServerError)
		return
	}

	// Get per-consumer metrics
	consumers, err := s.redis.GetConsumerMetrics(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get consumer metrics: %v", err), http.StatusInternalServerError)
		return
	}

	metrics := TestMetrics{
		Generated: generated,
		Consumed:  consumed,
		Consumers: consumers,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode metrics: %v", err), http.StatusInternalServerError)
		return
	}
}
