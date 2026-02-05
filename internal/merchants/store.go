// Package merchants provides merchant data storage operations.
package merchants

import (
	"context"
	"fmt"

	"bank-simulator/internal/storage"
)

// Store handles merchant data persistence in Redis.
type Store struct {
	client *storage.Client
}

// NewStore creates a new merchant store.
func NewStore(client *storage.Client) *Store {
	return &Store{
		client: client,
	}
}

// Merchant represents a registered merchant.
type Merchant struct {
	MerchantID string `json:"merchant_id"`
	TerminalID string `json:"terminal_id"`
	APIKey     string `json:"api_key"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"` // ISO 8601 timestamp
}

// key returns the Redis key for a merchant.
func (s *Store) key(merchantID string) string {
	return fmt.Sprintf("merchant:%s", merchantID)
}

// terminalKey returns the Redis key for terminal mapping.
func (s *Store) terminalKey(terminalID string) string {
	return fmt.Sprintf("terminal:%s", terminalID)
}

// Create stores a new merchant.
func (s *Store) Create(ctx context.Context, merchant *Merchant) error {
	key := s.key(merchant.MerchantID)
	if err := s.client.Set(ctx, key, merchant, 0); err != nil {
		return err
	}

	// Store terminal ID mapping
	terminalKey := s.terminalKey(merchant.TerminalID)
	return s.client.Set(ctx, terminalKey, merchant.MerchantID, 0)
}

// Get retrieves a merchant by ID.
func (s *Store) Get(ctx context.Context, merchantID string) (*Merchant, error) {
	key := s.key(merchantID)
	var merchant Merchant
	if err := s.client.Get(ctx, key, &merchant); err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}
	return &merchant, nil
}

// GetByTerminalID retrieves a merchant by terminal ID.
func (s *Store) GetByTerminalID(ctx context.Context, terminalID string) (*Merchant, error) {
	terminalKey := s.terminalKey(terminalID)
	var merchantID string
	if err := s.client.Get(ctx, terminalKey, &merchantID); err != nil {
		return nil, fmt.Errorf("terminal not found: %w", err)
	}
	return s.Get(ctx, merchantID)
}
