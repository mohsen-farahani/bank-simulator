// Package transactions provides transaction service operations.
package transactions

import (
	"context"
	"fmt"
	"time"

	"bank-simulator/internal/storage"
)

// Service handles transaction business logic.
type Service struct {
	storage *storage.Client
}

// NewService creates a new transaction service.
func NewService(storage *storage.Client) *Service {
	return &Service{
		storage: storage,
	}
}

// key returns the Redis key for a transaction.
func (s *Service) key(refNum string) string {
	return fmt.Sprintf("sep:transaction:%s", refNum)
}

// Create stores a new transaction.
func (s *Service) Create(ctx context.Context, tx *Transaction) error {
	key := s.key(tx.RefNum)
	return s.storage.Set(ctx, key, tx, 24*3600*1000*1000*1000) // 24 hours
}

// Get retrieves a transaction by RefNum.
func (s *Service) Get(ctx context.Context, refNum string) (*Transaction, error) {
	key := s.key(refNum)
	var tx Transaction
	if err := s.storage.Get(ctx, key, &tx); err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}
	return &tx, nil
}

// Update updates an existing transaction.
func (s *Service) Update(ctx context.Context, tx *Transaction) error {
	key := s.key(tx.RefNum)
	return s.storage.Set(ctx, key, tx, 24*time.Hour)
}
