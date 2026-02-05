// Package merchants provides merchant registration HTTP handlers.
package merchants

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"bank-simulator/internal/storage"
	"github.com/google/uuid"
)

// Handler manages merchant registration endpoints.
type Handler struct {
	store  *Store
	client *storage.Client
}

// NewHandler creates a new merchant handler.
func NewHandler(client *storage.Client) *Handler {
	return &Handler{
		store:  NewStore(client),
		client: client,
	}
}

// RegisterRequest represents a merchant registration request.
type RegisterRequest struct {
	Name string `json:"name"`
}

// RegisterResponse represents a merchant registration response.
type RegisterResponse struct {
	MerchantID string `json:"merchant_id"`
	TerminalID string `json:"terminal_id"`
	APIToken   string `json:"api_token"`
}

// Register handles POST /merchants/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Generate unique merchant ID
	merchantID := fmt.Sprintf("MERCHANT-%s", uuid.New().String()[:8])

	// Generate terminal ID (8 digits)
	terminalID, err := generateTerminalID()
	if err != nil {
		http.Error(w, "Failed to generate terminal ID", http.StatusInternalServerError)
		return
	}

	// Generate secure API token
	apiToken, err := generateAPIToken()
	if err != nil {
		http.Error(w, "Failed to generate API token", http.StatusInternalServerError)
		return
	}

	merchant := &Merchant{
		MerchantID: merchantID,
		TerminalID: terminalID,
		APIKey:     apiToken,
		Name:       req.Name,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}

	ctx := context.Background()
	if err := h.store.Create(ctx, merchant); err != nil {
		http.Error(w, "Failed to register merchant", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(RegisterResponse{
		MerchantID: merchant.MerchantID,
		TerminalID: merchant.TerminalID,
		APIToken:   merchant.APIKey,
	})
}

// generateTerminalID generates a random 8-digit terminal ID.
func generateTerminalID() (string, error) {
	max := big.NewInt(99999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d", n.Int64()), nil
}

// generateAPIToken generates a secure random API token.
func generateAPIToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
