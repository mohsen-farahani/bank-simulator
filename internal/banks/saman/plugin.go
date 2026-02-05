// Package saman provides the SEP bank plugin implementation.
package saman

import (
	"bank-simulator/internal/storage"

	"github.com/go-chi/chi/v5"
)

// Plugin represents the SEP bank plugin.
type Plugin struct {
	handler *Handler
}

// NewPlugin creates a new SEP plugin instance.
func NewPlugin(storage *storage.Client) (*Plugin, error) {
	handler, err := NewHandler(storage)
	if err != nil {
		return nil, err
	}
	return &Plugin{handler: handler}, nil
}

// Name implements banks.Bank interface.
func (p *Plugin) Name() string {
	return "saman"
}

// RegisterRoutes implements banks.Bank interface.
func (p *Plugin) RegisterRoutes(r chi.Router) {
	p.handler.RegisterRoutes(r)
}
