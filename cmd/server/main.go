// Package main provides the entry point for the bank simulator server.
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"bank-simulator/internal/banks"
	"bank-simulator/internal/banks/saman"
	"bank-simulator/internal/config"
	"bank-simulator/internal/merchants"
	"bank-simulator/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()

	// Initialize Redis storage
	storageClient, err := storage.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer storageClient.Close()

	homeTmpl, err := template.ParseFiles("web/templates/home.html")
	if err != nil {
		log.Fatalf("Failed to parse home template: %v", err)
	}

	// Initialize Chi router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Landing page for the public demo
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		scheme := "http"
		if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		baseURL := scheme + "://" + req.Host
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = homeTmpl.Execute(w, map[string]string{"BaseURL": baseURL})
	})

	// Register merchant routes
	merchantHandler := merchants.NewHandler(storageClient)
	r.Post("/merchants/register", merchantHandler.Register)

	// Initialize bank registry
	registry := banks.NewRegistry()

	// Load all available banks
	if err := loadBanks(registry, storageClient); err != nil {
		log.Fatalf("Failed to load banks: %v", err)
	}

	// Register all bank routes
	registry.RegisterAllRoutes(r)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := ":" + cfg.Port
	log.Printf("Server starting on %s", addr)
	log.Printf("Registered banks (%d): %v", registry.Count(), registry.List())
	log.Printf("SEP Endpoints (prefixed with /saman/sep.shaparak.ir):")
	log.Printf("  POST   /saman/sep.shaparak.ir/onlinepg/onlinepg")
	log.Printf("  GET    /saman/sep.shaparak.ir/OnlinePG/SendToken?token=...")
	log.Printf("  POST   /saman/sep.shaparak.ir/OnlinePG/OnlinePG")
	log.Printf("  POST   /saman/sep.shaparak.ir/OnlinePG/HandlePayment")
	log.Printf("  POST   /saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/VerifyTransaction")
	log.Printf("  POST   /saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/ReverseTransaction")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// loadBanks loads and registers all available banks into the registry.
func loadBanks(registry *banks.Registry, storage *storage.Client) error {
	// Register SEP (Saman)
	sepPlugin, err := saman.NewPlugin(storage)
	if err != nil {
		return fmt.Errorf("failed to load SEP plugin: %w", err)
	}
	if err := registry.Register(sepPlugin); err != nil {
		return fmt.Errorf("failed to register SEP plugin: %w", err)
	}

	// Future banks can be registered here:
	// mellatPlugin, err := mellat.NewPlugin(storage)
	// if err != nil {
	// 	return fmt.Errorf("failed to load Mellat plugin: %w", err)
	// }
	// if err := registry.Register(mellatPlugin); err != nil {
	// 	return fmt.Errorf("failed to register Mellat plugin: %w", err)
	// }

	return nil
}
