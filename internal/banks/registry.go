// Package banks provides the bank registry system.
package banks

import (
	"fmt"

	"github.com/go-chi/chi/v5"
)

// Bank defines the interface that all bank plugins must implement.
type Bank interface {
	// Name returns the unique identifier for this bank.
	Name() string

	// RegisterRoutes registers all HTTP routes for the bank on the given router.
	RegisterRoutes(r chi.Router)
}

// Registry holds all registered banks.
type Registry struct {
	banks map[string]Bank
}

// NewRegistry creates a new bank registry.
func NewRegistry() *Registry {
	return &Registry{
		banks: make(map[string]Bank),
	}
}

// Register adds a bank to the registry.
func (r *Registry) Register(bank Bank) error {
	if bank == nil {
		return fmt.Errorf("cannot register nil bank")
	}

	name := bank.Name()
	if name == "" {
		return fmt.Errorf("bank name cannot be empty")
	}

	r.banks[name] = bank
	return nil
}

// Get retrieves a bank from the registry by name.
func (r *Registry) Get(name string) (Bank, bool) {
	bank, ok := r.banks[name]
	return bank, ok
}

// RegisterAllRoutes registers routes for all banks in the registry.
func (r *Registry) RegisterAllRoutes(router chi.Router) {
	for _, bank := range r.banks {
		bank.RegisterRoutes(router)
	}
}

// List returns all registered bank names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.banks))
	for name := range r.banks {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered banks.
func (r *Registry) Count() int {
	return len(r.banks)
}
