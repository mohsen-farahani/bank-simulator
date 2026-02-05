# SEP Gateway Simulator

An open-source simulator for Iranian payment gateways (PSPs) designed for QA and testing purposes. This project simulates the SEP (Saman Electronic Payment) gateway without connecting to real banks.

## Features

- **SEP Gateway Simulation**: Full implementation of SEP payment flow with exact API compatibility
- **Chi Router**: Modern HTTP routing with Chi
- **Redis Storage**: Transaction and merchant data persistence
- **Dynamic Merchant Registration**: Register merchants and get unique credentials
- **Plugin Architecture**: Easy to add new bank gateways
- **Docker Support**: Ready-to-use Docker Compose setup
- **Bank-Specific Documentation**: Detailed README for each bank implementation

## Project Structure

```
sep-gateway-simulator/
├── cmd/
│   └── server/
│       └── main.go          # Server entry point
├── internal/
│   ├── banks/
│   │   ├── registry.go      # Bank registry system
│   │   └── saman/
│   │       ├── handler.go   # SEP HTTP handlers
│   │       ├── models.go    # Request/response models
│   │       └── plugin.go    # Plugin implementation
│   ├── merchants/
│   │   ├── register.go      # Merchant registration handler
│   │   └── store.go         # Merchant storage
│   ├── storage/
│   │   └── redis.go         # Redis client wrapper
│   └── transactions/
│       ├── model.go         # Transaction data model
│       └── service.go       # Transaction service
├── web/
│   └── templates/
│       └── payment.html     # Payment page template
├── docker-compose.yml
├── go.mod
└── README.md
```

## Prerequisites

- Go 1.23.5 or higher
- Redis (via Docker Compose or standalone)
- Docker and Docker Compose (optional)

## Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd BankSimulator
   ```

2. **Start Redis with Docker Compose**
   ```bash
   docker-compose up -d
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Run the server**
   ```bash
   go run cmd/server/main.go
   ```

   Or use Make:
   ```bash
   make run
   ```

The server will start on `http://localhost:8080` by default.

## API Endpoints

### Merchant Registration

**POST** `/merchants/register`

Register a new merchant and receive unique credentials.

Request:
```json
{
  "name": "Test Shop"
}
```

Response:
```json
{
  "merchant_id": "MERCHANT-abc123",
  "terminal_id": "12345678",
  "api_token": "secure-random-token"
}
```

### Bank Gateway Endpoints

#### SEP (Saman Electronic Payment)

All SEP endpoints match the official SEP gateway specification. For detailed documentation, see [SEP Bank README](internal/banks/saman/README.md).

**Quick Reference (all routes prefixed with `/saman/sep.shaparak.ir`):**
- **POST** `/saman/sep.shaparak.ir/onlinepg/onlinepg` - Request payment token
- **GET** `/saman/sep.shaparak.ir/OnlinePG/SendToken?token=XXXX` - Display payment page
- **POST** `/saman/sep.shaparak.ir/OnlinePG/OnlinePG` - Payment form redirect
- **POST** `/saman/sep.shaparak.ir/OnlinePG/HandlePayment` - Payment confirmation
- **POST** `/saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/VerifyTransaction` - Verify transaction
- **POST** `/saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/ReverseTransaction` - Reverse transaction

📖 **Full Documentation:** [SEP Bank Implementation Guide](internal/banks/saman/README.md)

## Payment Flow

### SEP Payment Flow

1. **Register Merchant**: Create a merchant account and receive `terminal_id` and `api_token` (no callback URL needed)
2. **Request Token**: Call `/saman/sep.shaparak.ir/onlinepg/onlinepg` with merchant credentials and `RedirectUrl` in request
3. **Redirect to Payment**: Use the returned token to redirect to `/saman/sep.shaparak.ir/OnlinePG/SendToken?token=XXXX`
4. **Complete Payment**: User clicks "Pay Successfully" or "Cancel Payment" on the payment page
5. **Callback**: Merchant receives callback at `RedirectUrl` (from token request) with transaction status
6. **Verify Transaction**: Call `/saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/VerifyTransaction` to verify
7. **Reverse (Optional)**: Call `/saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/ReverseTransaction` to reverse

📖 **Detailed Flow Diagram:** See [SEP Bank README](internal/banks/saman/README.md#payment-flow)

## Configuration

Environment variables:

- `PORT`: Server port (default: `80`)
- `REDIS_ADDR`: Redis address (default: `localhost:6379`)
- `REDIS_PASSWORD`: Redis password (default: `1234`)
- `REDIS_DB`: Redis database number (default: `0`)

## Bank Implementations

### SEP (Saman Electronic Payment)

Full implementation of the SEP gateway with all official endpoints.

📖 **[SEP Bank Documentation →](internal/banks/saman/README.md)**

### Other Banks

- **Mellat** - Placeholder (coming soon)
- **Parsian** - Placeholder (coming soon)

## Development

### Make Commands

- `make run` - Run the server
- `make test` - Run tests
- `make docker-up` - Start Docker services
- `make docker-down` - Stop Docker services
- `make build` - Build the binary
- `make clean` - Clean build artifacts

### Adding a New Bank

1. Create a new package under `internal/banks/` (e.g., `internal/banks/mellat/`)
2. Implement the `banks.Bank` interface:
   ```go
   type Bank interface {
       Name() string
       RegisterRoutes(r chi.Router)
   }
   ```
3. Register the bank in `cmd/server/main.go`:
   ```go
   bankPlugin, err := mellat.NewPlugin(storage)
   if err != nil {
       return err
   }
   registry.Register(bankPlugin)
   ```
4. Create a `README.md` in your bank package documenting the API endpoints and usage

## Docker

Start all services:
```bash
docker-compose up -d
```

Stop all services:
```bash
docker-compose down
```

## License

MIT License - see LICENSE file for details.
