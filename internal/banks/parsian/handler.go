// Package parsian implements the Bank Parsian payment gateway simulator.
// This is a placeholder for future implementation.
package parsian

import (
	"bank-simulator/internal/banks"
	"bank-simulator/internal/storage"

	"github.com/go-chi/chi/v5"
)

// Handler implements the Bank Parsian payment gateway simulator.
type Handler struct {
	storage *storage.Client
}

// New creates a new Bank Parsian instance implementing the banks.Bank interface.
func New(storage *storage.Client) (banks.Bank, error) {
	return &Handler{
		storage: storage,
	}, nil
}

// Name implements banks.Bank interface.
func (h *Handler) Name() string {
	return "parsian"
}

// RegisterRoutes implements banks.Bank interface.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// TODO: Implement Parsian bank routes
	// Example:
	// r.Post("/parsian/request-token", h.RequestToken)
	// r.Get("/parsian/pay", h.Pay)
	// r.Post("/parsian/confirm", h.Confirm)
}
