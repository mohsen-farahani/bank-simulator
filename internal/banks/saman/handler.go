// Package saman implements the SEP (Saman Electronic Payment) gateway simulator.
package saman

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bank-simulator/internal/merchants"
	"bank-simulator/internal/storage"
	"bank-simulator/internal/transactions"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler handles SEP HTTP requests.
type Handler struct {
	storage       *storage.Client
	merchantStore *merchants.Store
	txService     *transactions.Service
	tmpl          *template.Template
}

// NewHandler creates a new SEP handler.
func NewHandler(storage *storage.Client) (*Handler, error) {
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Handler{
		storage:       storage,
		merchantStore: merchants.NewStore(storage),
		txService:     transactions.NewService(storage),
		tmpl:          tmpl,
	}, nil
}

// RegisterRoutes registers all SEP routes on the router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// All SEP routes are prefixed with /saman/sep.shaparak.ir
	r.Route("/saman/sep.shaparak.ir", func(r chi.Router) {
		// Exact SEP endpoints matching official documentation
		r.Post("/onlinepg/onlinepg", h.TokenRequest)
		r.Get("/OnlinePG/SendToken", h.SendToken)
		r.Post("/OnlinePG/OnlinePG", h.PaymentForm)
		r.Post("/OnlinePG/HandlePayment", h.HandlePayment)
		r.Post("/verifyTxnRandomSessionkey/ipg/VerifyTransaction", h.VerifyTransaction)
		r.Post("/verifyTxnRandomSessionkey/ipg/ReverseTransaction", h.ReverseTransaction)
	})
}

// TokenRequest handles POST /onlinepg/onlinepg
func (h *Handler) TokenRequest(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "5", "Invalid request body")
		return
	}

	// Validate action
	if req.Action != "token" {
		h.sendErrorResponse(w, "5", "Invalid action")
		return
	}

	// Validate and find merchant by TerminalId
	ctx := context.Background()
	merchant, err := h.merchantStore.GetByTerminalID(ctx, req.TerminalId)
	if err != nil {
		h.sendErrorResponse(w, "5", "Invalid TerminalId")
		return
	}

	// Validate amount
	if req.Amount <= 0 {
		h.sendErrorResponse(w, "5", "Amount must be greater than 0")
		return
	}

	// Validate ResNum
	if req.ResNum == "" {
		h.sendErrorResponse(w, "5", "ResNum is required")
		return
	}

	// Validate RedirectUrl
	if req.RedirectUrl == "" {
		h.sendErrorResponse(w, "5", "RedirectUrl is required")
		return
	}

	// Generate RefNum (token)
	refNum := uuid.New().String()

	// Create transaction
	tx := &transactions.Transaction{
		RefNum:      refNum,
		ResNum:      req.ResNum,
		TerminalId:  req.TerminalId,
		Amount:      req.Amount,
		Status:      transactions.StatusPending,
		RedirectUrl: req.RedirectUrl,
		MerchantId:  merchant.MerchantID,
		OrderId:     req.OrderId,
		CellNumber:  req.CellNumber,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store in Redis
	if err := h.txService.Create(ctx, tx); err != nil {
		h.sendErrorResponse(w, "5", "Failed to create transaction")
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{
		Status: 1,
		Token:  refNum,
	})
}

// SendToken handles GET /OnlinePG/SendToken?token=XXXX
func (h *Handler) SendToken(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	tx, err := h.txService.Get(ctx, token)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	// Format amount with thousand separators
	amountStr := formatAmount(tx.Amount)

	w.Header().Set("Content-Type", "text/html")
	h.tmpl.ExecuteTemplate(w, "payment.html", map[string]interface{}{
		"RefNum":     tx.RefNum,
		"Amount":     amountStr,
		"TerminalId": tx.TerminalId,
		"ResNum":     tx.ResNum,
	})
}

// PaymentForm handles POST /OnlinePG/OnlinePG
func (h *Handler) PaymentForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	token := r.FormValue("Token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Redirect to SendToken with prefix
	redirectURL := fmt.Sprintf("/saman/sep.shaparak.ir/OnlinePG/SendToken?token=%s", url.QueryEscape(token))
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// HandlePayment handles payment confirmation from the payment page
func (h *Handler) HandlePayment(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	refNum := r.FormValue("RefNum")
	action := r.FormValue("action") // "pay" or "cancel"

	if refNum == "" {
		http.Error(w, "RefNum is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	tx, err := h.txService.Get(ctx, refNum)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if action == "pay" {
		// Update status to PAID
		tx.Status = transactions.StatusPaid
		tx.UpdatedAt = time.Now()
		h.txService.Update(ctx, tx)

		// Redirect to merchant's RedirectUrl with SEP parameters
		redirectURL := fmt.Sprintf("%s?State=OK&Status=2&RefNum=%s&ResNum=%s&TerminalId=%s",
			tx.RedirectUrl,
			url.QueryEscape(tx.RefNum),
			url.QueryEscape(tx.ResNum),
			url.QueryEscape(tx.TerminalId),
		)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	} else {
		// Cancel - redirect with State=CanceledByUser
		redirectURL := fmt.Sprintf("%s?State=CanceledByUser", tx.RedirectUrl)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}

// VerifyTransaction handles POST /verifyTxnRandomSessionkey/ipg/VerifyTransaction
func (h *Handler) VerifyTransaction(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendVerifyError(w, "Invalid request body")
		return
	}

	if req.RefNum == "" {
		h.sendVerifyError(w, "RefNum is required")
		return
	}

	ctx := context.Background()
	tx, err := h.txService.Get(ctx, req.RefNum)
	if err != nil {
		h.sendVerifyError(w, "Transaction not found")
		return
	}

	// Only allow verify if transaction is PAID
	if tx.Status != transactions.StatusPaid {
		h.sendVerifyError(w, "Transaction is not paid")
		return
	}

	// Update status to VERIFIED
	tx.Status = transactions.StatusVerified
	tx.UpdatedAt = time.Now()

	if err := h.txService.Update(ctx, tx); err != nil {
		h.sendVerifyError(w, "Failed to update transaction")
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VerifyResponse{
		TransactionDetail: map[string]interface{}{
			"RefNum":     tx.RefNum,
			"ResNum":     tx.ResNum,
			"TerminalId": tx.TerminalId,
			"Amount":     tx.Amount,
			"Status":     string(tx.Status),
			"CreatedAt":  tx.CreatedAt.Format(time.RFC3339),
			"UpdatedAt":  tx.UpdatedAt.Format(time.RFC3339),
		},
		ResultCode:        0,
		ResultDescription: "عملیات با موفقیت انجام شد",
		Success:           true,
	})
}

// ReverseTransaction handles POST /verifyTxnRandomSessionkey/ipg/ReverseTransaction
func (h *Handler) ReverseTransaction(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendVerifyError(w, "Invalid request body")
		return
	}

	if req.RefNum == "" {
		h.sendVerifyError(w, "RefNum is required")
		return
	}

	ctx := context.Background()
	tx, err := h.txService.Get(ctx, req.RefNum)
	if err != nil {
		h.sendVerifyError(w, "Transaction not found")
		return
	}

	// Only allow reverse if VERIFIED
	if tx.Status != transactions.StatusVerified {
		h.sendVerifyError(w, "Transaction must be verified before reversal")
		return
	}

	// Update status to REVERSED
	tx.Status = transactions.StatusReversed
	tx.UpdatedAt = time.Now()

	if err := h.txService.Update(ctx, tx); err != nil {
		h.sendVerifyError(w, "Failed to update transaction")
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VerifyResponse{
		TransactionDetail: map[string]interface{}{
			"RefNum":     tx.RefNum,
			"ResNum":     tx.ResNum,
			"TerminalId": tx.TerminalId,
			"Amount":     tx.Amount,
			"Status":     string(tx.Status),
			"CreatedAt":  tx.CreatedAt.Format(time.RFC3339),
			"UpdatedAt":  tx.UpdatedAt.Format(time.RFC3339),
		},
		ResultCode:        0,
		ResultDescription: "عملیات با موفقیت انجام شد",
		Success:           true,
	})
}

// Helper functions

func (h *Handler) sendErrorResponse(w http.ResponseWriter, errorCode, errorDesc string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TokenResponse{
		Status:    -1,
		ErrorCode: errorCode,
		ErrorDesc: errorDesc,
	})
}

func (h *Handler) sendVerifyError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(VerifyResponse{
		ResultCode:        -1,
		ResultDescription: message,
		Success:           false,
	})
}

// formatAmount formats an amount with thousand separators (commas).
func formatAmount(amount int64) string {
	amountStr := strconv.FormatInt(amount, 10)
	if len(amountStr) <= 3 {
		return amountStr
	}

	var result strings.Builder
	for i, digit := range amountStr {
		if i > 0 && (len(amountStr)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}
