# SEP (Saman Electronic Payment) Gateway Simulator

This package implements a complete simulator for the SEP (Saman Electronic Payment) gateway, designed for QA and testing purposes.

## Overview

The SEP simulator provides a full implementation of the Saman bank payment gateway API, matching the official SEP documentation endpoints exactly. It allows developers to test payment flows without connecting to real bank systems.

## Features

- ✅ **Exact SEP API Compatibility**: All endpoints match official SEP documentation
- ✅ **Token Generation**: Secure token generation for payment requests
- ✅ **Payment Page**: Beautiful Persian payment interface
- ✅ **Transaction Verification**: Verify completed transactions
- ✅ **Transaction Reversal**: Reverse verified transactions
- ✅ **Merchant Integration**: Dynamic merchant registration and authentication

## API Endpoints

### 1. Token Request

**Endpoint:** `POST /saman/sep.shaparak.ir/onlinepg/onlinepg`

Request a payment token to initiate a transaction.

**Request Body:**
```json
{
  "action": "token",
  "TerminalId": "12345678",
  "Amount": 100000,
  "ResNum": "ORDER-123",
  "RedirectUrl": "http://localhost:3000/callback",
  "OrderId": "ORDER-12345",
  "CellNumber": "09123456789",
  "Description": "Payment for order #12345"
}
```

**Request Fields:**
- `action` (required): Must be `"token"`
- `TerminalId` (required): Terminal ID from merchant registration
- `Amount` (required): Transaction amount in Rials
- `ResNum` (required): Merchant's reservation/order number
- `RedirectUrl` (required): Callback URL where user will be redirected after payment
- `OrderId` (optional): Order identifier
- `CellNumber` (optional): Customer's cell phone number
- `Description` (optional): Transaction description

**Response (Success):**
```json
{
  "status": 1,
  "token": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response (Error):**
```json
{
  "status": -1,
  "errorCode": "5",
  "errorDesc": "Error description"
}
```

**Error Codes:**
- `5`: Invalid request parameters
- `-1`: General error

### 2. Payment Page

**Endpoint:** `GET /saman/sep.shaparak.ir/OnlinePG/SendToken?token=XXXX`

Display the payment page for the given token. Users can choose to pay successfully or cancel the payment.

**Query Parameters:**
- `token` (required): The payment token received from token request

### 3. Payment Form Redirect

**Endpoint:** `POST /saman/sep.shaparak.ir/OnlinePG/OnlinePG`

Redirects to the payment page. This endpoint accepts form submissions and redirects to `/OnlinePG/SendToken`.

**Form Fields:**
- `Token` (required): Payment token

### 4. Payment Confirmation

**Endpoint:** `POST /saman/sep.shaparak.ir/OnlinePG/HandlePayment`

Handle payment confirmation from the payment page.

**Form Fields:**
- `RefNum` (required): Transaction reference number
- `action` (required): `pay` or `cancel`

**Behavior:**
- If `action=pay`: Transaction status is set to `PAID` and user is redirected to merchant's `RedirectUrl` with success parameters
- If `action=cancel`: User is redirected to merchant's `RedirectUrl` with `State=CanceledByUser`

**Success Redirect Parameters:**
```
?State=OK&Status=2&RefNum=XXXX&ResNum=YYYY&TerminalId=ZZZZ
```

### 5. Verify Transaction

**Endpoint:** `POST /saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/VerifyTransaction`

Verify a completed transaction. Only transactions with status `PAID` can be verified.

**Request Body:**
```json
{
  "RefNum": "550e8400-e29b-41d4-a716-446655440000",
  "TerminalNumber": 12345678
}
```

**Response (Success):**
```json
{
  "Success": true,
  "ResultCode": 0,
  "ResultDescription": "عملیات با موفقیت انجام شد",
  "TransactionDetail": {
    "RefNum": "550e8400-e29b-41d4-a716-446655440000",
    "ResNum": "ORDER-123",
    "TerminalId": "12345678",
    "Amount": 100000,
    "Status": "VERIFIED",
    "CreatedAt": "2024-01-01T12:00:00Z",
    "UpdatedAt": "2024-01-01T12:05:00Z"
  }
}
```

**Response (Error):**
```json
{
  "Success": false,
  "ResultCode": -1,
  "ResultDescription": "Error message"
}
```

**Result Codes:**
- `0`: Success
- `-1`: Error (transaction not found, not paid, etc.)

### 6. Reverse Transaction

**Endpoint:** `POST /saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/ReverseTransaction`

Reverse a verified transaction. Only transactions with status `VERIFIED` can be reversed.

**Request Body:**
```json
{
  "RefNum": "550e8400-e29b-41d4-a716-446655440000",
  "TerminalNumber": 12345678
}
```

**Response:** Same format as Verify Transaction

## Payment Flow

```
┌─────────────┐
│   Merchant  │
└──────┬──────┘
       │ 1. POST /saman/sep.shaparak.ir/onlinepg/onlinepg
       │    {TerminalId, Amount, ResNum, RedirectUrl}
       ▼
┌──────────────────┐
│   SEP Simulator  │
│  Generate Token  │
└──────┬───────────┘
       │ 2. Return {status: 1, token: "..."}
       ▼
┌─────────────┐
│   Merchant  │
└──────┬──────┘
       │ 3. Redirect to /saman/sep.shaparak.ir/OnlinePG/SendToken?token=...
       ▼
┌──────────────────┐
│  Payment Page    │
│  (User chooses)  │
└──────┬───────────┘
       │ 4. POST /saman/sep.shaparak.ir/OnlinePG/HandlePayment
       │    {RefNum, action: "pay"|"cancel"}
       ▼
┌──────────────────┐
│   SEP Simulator  │
│  Update Status   │
└──────┬───────────┘
       │ 5. Redirect to RedirectUrl (from token request)
       ▼
┌─────────────┐
│   Merchant  │
│  Callback   │
└──────┬──────┘
       │ 6. POST /saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/VerifyTransaction
       │    {RefNum, TerminalNumber}
       ▼
┌──────────────────┐
│   SEP Simulator  │
│  Verify & Return │
└──────────────────┘
```

## Transaction Status Flow

```
PENDING → PAID → VERIFIED → REVERSED
   │        │
   └────────┴→ FAILED (if canceled)
```

- **PENDING**: Initial state after token generation
- **PAID**: User clicked "Pay Successfully" on payment page
- **VERIFIED**: Merchant verified the transaction
- **REVERSED**: Merchant reversed a verified transaction
- **FAILED**: User canceled the payment

## Merchant Integration

Before using SEP endpoints, merchants must be registered:

1. **Register Merchant:**
   ```bash
   POST /merchants/register
   {
     "name": "My Shop"
   }
   ```

2. **Receive Credentials:**
   ```json
   {
     "merchant_id": "MERCHANT-abc123",
     "terminal_id": "12345678",
     "api_token": "secure-token"
   }
   ```

3. **Use TerminalId** in token requests

**Note:** The `callback_url` (RedirectUrl) is now provided in each token request, not during merchant registration. This allows merchants to use different callback URLs for different transactions.

## Example Usage

### Complete Payment Flow (cURL)

```bash
# 1. Register merchant
curl -X POST http://localhost:8080/merchants/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Shop"
  }'

# Response: {"merchant_id": "...", "terminal_id": "12345678", "api_token": "..."}

# 2. Request token
curl -X POST http://localhost:8080/saman/sep.shaparak.ir/onlinepg/onlinepg \
  -H "Content-Type: application/json" \
  -d '{
    "action": "token",
    "TerminalId": "12345678",
    "Amount": 100000,
    "ResNum": "ORDER-123",
    "RedirectUrl": "http://localhost:3000/callback",
    "OrderId": "ORDER-12345",
    "CellNumber": "09123456789",
    "Description": "Payment for order #12345"
  }'

# Response: {"status": 1, "token": "..."}

# 3. Open payment page in browser
# http://localhost:8080/saman/sep.shaparak.ir/OnlinePG/SendToken?token=...

# 4. After payment, verify transaction
curl -X POST http://localhost:8080/saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/VerifyTransaction \
  -H "Content-Type: application/json" \
  -d '{
    "RefNum": "transaction-refnum",
    "TerminalNumber": 12345678
  }'
```

### Go Example

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

func main() {
    // Request token
    tokenReq := map[string]interface{}{
        "action":      "token",
        "TerminalId":  "12345678",
        "Amount":      100000,
        "ResNum":      "ORDER-123",
        "RedirectUrl": "http://localhost:3000/callback",
        "OrderId":     "ORDER-12345",
        "CellNumber":  "09123456789",
        "Description": "Payment for order #12345",
    }
    
    jsonData, _ := json.Marshal(tokenReq)
    resp, _ := http.Post("http://localhost:8080/saman/sep.shaparak.ir/onlinepg/onlinepg", 
        "application/json", bytes.NewBuffer(jsonData))
    
    var tokenResp map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&tokenResp)
    
    token := tokenResp["token"].(string)
    fmt.Printf("Payment URL: http://localhost:8080/saman/sep.shaparak.ir/OnlinePG/SendToken?token=%s\n", token)
}
```

## Testing Scenarios

### Success Flow
1. Request token → Receive token
2. Open payment page → Click "Pay Successfully"
3. Receive callback with `State=OK&Status=2`
4. Verify transaction → Status becomes `VERIFIED`

### Cancel Flow
1. Request token → Receive token
2. Open payment page → Click "Cancel Payment"
3. Receive callback with `State=CanceledByUser`
4. Transaction status remains `PENDING` or becomes `FAILED`

### Verification Flow
1. Complete payment (status = `PAID`)
2. Call verify endpoint
3. Status changes to `VERIFIED`
4. Can now reverse if needed

### Reversal Flow
1. Verify transaction (status = `VERIFIED`)
2. Call reverse endpoint
3. Status changes to `REVERSED`

## Error Handling

All endpoints return appropriate HTTP status codes:
- `200 OK`: Request processed (check response body for success/error)
- `400 Bad Request`: Invalid request parameters
- `404 Not Found`: Transaction not found
- `500 Internal Server Error`: Server error

Error responses follow SEP format:
```json
{
  "status": -1,
  "errorCode": "5",
  "errorDesc": "Error description"
}
```

## Storage

Transactions are stored in Redis with the following key pattern:
```
sep:transaction:{RefNum}
```

Transaction data includes:
- `RefNum`: Transaction reference number (token)
- `ResNum`: Merchant reservation number
- `TerminalId`: Terminal ID
- `Amount`: Transaction amount in Rials
- `Status`: Transaction status (PENDING, PAID, VERIFIED, REVERSED, FAILED)
- `RedirectUrl`: Merchant callback URL
- `MerchantId`: Merchant ID
- `CreatedAt`: Creation timestamp
- `UpdatedAt`: Last update timestamp

## Security Notes

⚠️ **This is a simulator for testing purposes only.**
- No real money is transferred
- No actual bank connections
- Suitable for development, QA, and staging environments
- **DO NOT** use in production with real payment processing

## Related Documentation

- [Main README](../../../../README.md) - Project overview and setup
- [Bank Registry](../registry.go) - How banks are registered
- [Transaction Models](../../../transactions/model.go) - Transaction data structures

## Support

For issues, questions, or contributions, please refer to the main project repository.
