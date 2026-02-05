// Package mellat implements the Bank Mellat payment gateway simulator.
// This is a placeholder for future implementation.
package mellat

import (
	"bank-simulator/internal/banks"
	"bank-simulator/internal/storage"

	"github.com/go-chi/chi/v5"
)

// Handler implements the Bank Mellat payment gateway simulator.
type Handler struct {
	storage *storage.Client
}

// New creates a new Bank Mellat instance implementing the banks.Bank interface.
func New(storage *storage.Client) (banks.Bank, error) {
	return &Handler{
		storage: storage,
	}, nil
}

// Name implements banks.Bank interface.
func (h *Handler) Name() string {
	return "mellat"
}

// RegisterRoutes implements banks.Bank interface.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// TODO: Implement Mellat bank routes
	// Example:
	// r.Post("/mellat/request-token", h.RequestToken)
	// r.Get("/mellat/pay", h.Pay)
	// r.Post("/mellat/confirm", h.Confirm)
}
