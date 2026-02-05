// Package saman defines SEP request and response models.
package saman

// TokenRequest represents the SEP token request.
type TokenRequest struct {
	Action      string `json:"action"`
	TerminalId  string `json:"TerminalId"`
	Amount      int64  `json:"Amount"`
	ResNum      string `json:"ResNum"`
	RedirectUrl string `json:"RedirectUrl"`
	OrderId     string `json:"OrderId,omitempty"`
	CellNumber  string `json:"CellNumber,omitempty"`
	Description string `json:"Description,omitempty"`
}

// TokenResponse represents the SEP token response.
type TokenResponse struct {
	Status    int    `json:"status"`
	Token     string `json:"token,omitempty"`
	ErrorCode string `json:"errorCode,omitempty"`
	ErrorDesc string `json:"errorDesc,omitempty"`
}

// VerifyRequest represents the SEP verify/reverse request.
type VerifyRequest struct {
	RefNum         string `json:"RefNum"`
	TerminalNumber int    `json:"TerminalNumber"`
}

// VerifyResponse represents the SEP verify/reverse response.
type VerifyResponse struct {
	TransactionDetail map[string]interface{} `json:"TransactionDetail,omitempty"`
	ResultCode        int                    `json:"ResultCode"`
	ResultDescription string                 `json:"ResultDescription"`
	Success           bool                   `json:"Success"`
}
