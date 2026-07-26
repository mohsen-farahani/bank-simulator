# Bank Simulator

**A local payment-gateway simulator for Iranian banks** — test your checkout flow without a real PSP, real cards, or waiting on bank sandboxes.

**Live demo:** [https://bank-simulator-50tb.onrender.com](https://bank-simulator-50tb.onrender.com) · Health: [/health](https://bank-simulator-50tb.onrender.com/health)

> اگر از داخل ایران به Render دسترسی ندارید، پروژه را لوکال با Docker Compose اجرا کنید (بخش Quick start).

If you are building an online shop, SaaS, or payment integration in Iran, this tool lets your QA and backend teams complete the full SEP (Saman) payment cycle on `localhost` in minutes.

> **QA / demo only.** This is not a real bank gateway and must not be used in production.

---

## Why this exists

Real Iranian payment gateways are slow to provision, rate-limited in sandbox, and awkward for automated tests. Bank Simulator gives you:

| Pain today | What you get |
|---|---|
| Waiting for bank sandbox access | Run a SEP-compatible API locally in seconds |
| Flaky or rate-limited test gateways | Deterministic success / cancel flows you control |
| Hard-to-reproduce payment bugs | Full token → pay page → callback → verify → reverse path |
| Coupling tests to real money / cards | Fake merchants, fake transactions, Redis-backed state |

**Current focus:** SEP (Saman Electronic Payment) with API paths that mirror the real gateway (`/saman/sep.shaparak.ir/...`). Mellat and Parsian hooks are reserved for later.

---

## Quick start

### Requirements

- Go **1.25+**
- Redis (or Docker)

### Option A — Docker Compose (recommended)

```bash
git clone https://github.com/mohsen-farahani/bank-simulator.git
cd bank-simulator
docker compose up --build
```

App: [http://localhost:8080](http://localhost:8080) · Health: [http://localhost:8080/health](http://localhost:8080/health)

### Option B — Run locally

```bash
git clone https://github.com/mohsen-farahani/bank-simulator.git
cd bank-simulator

# Start Redis only
docker compose up -d redis

cp .env.example .env   # optional; defaults work for local Redis
go run ./cmd/server
```

Server listens on `:8080` by default.

---

## Deploy for free on Render

**Public demo:** [https://bank-simulator-50tb.onrender.com](https://bank-simulator-50tb.onrender.com)

[Render](https://render.com) connects to your GitHub repo, gives you HTTPS on `*.onrender.com`, and includes a free Redis-compatible Key Value store. No credit card required for the free tier.

> **Iran access:** `*.onrender.com` is often filtered inside Iran. Run locally with Docker Compose instead.

### One-time setup

1. Push this repo to GitHub (if it is not already public).
2. Open [Render Dashboard → New → Blueprint](https://dashboard.render.com/select-repo?type=blueprint).
3. Connect the GitHub account and select `bank-simulator`.
4. Render reads `render.yaml` and creates:
   - **Web Service** (`bank-simulator`) — free plan
   - **Key Value** (`bank-simulator-redis`) — free Redis-compatible store
5. Deploy. Your public URL looks like:

```text
https://bank-simulator-xxxx.onrender.com
```

Open `/` for the demo landing page, or `/health` to confirm the service is up.

### Free-tier limits (important)

| Limit | What it means for users |
|---|---|
| Sleep after ~15 min idle | First request after sleep can take 30–60s (cold start) |
| Free Redis has no durable disk | Merchant/transaction data can disappear on restart |
| ~750 instance hours / month | Enough for one always-on free web service |

This is ideal for a shared demo / OSS playground, not for SLA-backed QA.

### After deploy

Point API calls at your Render URL instead of `localhost:8080`. Example:

```bash
curl -s -X POST https://bank-simulator-50tb.onrender.com/merchants/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Demo Shop"}'
```

---

## 5-minute walkthrough

### 1. Register a test merchant

```bash
curl -s -X POST http://localhost:8080/merchants/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Demo Shop"}'
```

Example response:

```json
{
  "merchant_id": "MERCHANT-a1b2c3d4",
  "terminal_id": "12345678",
  "api_token": "..."
}
```

Use `terminal_id` in the next step. (`api_token` is generated for future auth work; endpoints currently key off `TerminalId` only — see [Security](#security).)

### 2. Request a payment token

```bash
curl -s -X POST http://localhost:8080/saman/sep.shaparak.ir/onlinepg/onlinepg \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "token",
    "TerminalId": "12345678",
    "Amount": 100000,
    "ResNum": "ORDER-1001",
    "RedirectUrl": "http://localhost:3000/callback"
  }'
```

You get a `token` (UUID). Open the payment page:

```
http://localhost:8080/saman/sep.shaparak.ir/OnlinePG/SendToken?token=<TOKEN>
```

### 3. Pay or cancel

On the Persian payment page, choose **پرداخت موفق** or **انصراف**. The browser redirects to your `RedirectUrl` with SEP-style query params (`State`, `RefNum`, `ResNum`, …).

### 4. Verify (and optionally reverse)

```bash
# Verify
curl -s -X POST http://localhost:8080/saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/VerifyTransaction \
  -H 'Content-Type: application/json' \
  -d '{"RefNum":"<TOKEN>","TerminalNumber":12345678}'

# Reverse (only after verify)
curl -s -X POST http://localhost:8080/saman/sep.shaparak.ir/verifyTxnRandomSessionkey/ipg/ReverseTransaction \
  -H 'Content-Type: application/json' \
  -d '{"RefNum":"<TOKEN>","TerminalNumber":12345678}'
```

Full SEP field reference: [internal/banks/saman/README.md](internal/banks/saman/README.md)

---

## Payment flow

```text
Your app                    Bank Simulator                      Your callback
   |                              |                                    |
   |-- register merchant -------->|                                    |
   |-- request token ------------>|                                    |
   |<--------- token -------------|                                    |
   |-- redirect user to SendToken>|                                    |
   |                              |-- user pays / cancels ------------>|
   |-- verify / reverse --------->|                                    |
```

Statuses stored in Redis: `PENDING` → `PAID` → `VERIFIED` → `REVERSED`.

---

## Configuration

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port (Render sets this automatically too) |
| `REDIS_ADDR` | `localhost:6379` | Redis host:port (used when `REDIS_URL` is unset) |
| `REDIS_PASSWORD` | _(empty)_ | Redis password |
| `REDIS_DB` | `0` | Redis DB index |
| `REDIS_URL` | _(unset)_ | Full Redis URL; overrides addr/password/db (used on Render) |

Copy `.env.example` when you need a starting point. Keep real `.env` files out of git (already in `.gitignore`).

---

## Project layout

```text
cmd/server/           # process entrypoint
internal/
  banks/              # bank plugins (SEP + placeholders)
  merchants/          # merchant registration & store
  transactions/       # transaction model & service
  storage/            # Redis client
  config/             # env-based config
web/templates/        # payment page (RTL / Persian)
```

Banks plug in via a small interface (`Name` + `RegisterRoutes`). To add another PSP, create a package under `internal/banks/`, implement the interface, and register it in `cmd/server/main.go`.

---

## Security

This project is a **developer / QA simulator**, not a hardened public service.

| Topic | Status |
|---|---|
| Real bank credentials | None — no live PSP keys required |
| Secrets in repo | `.env` is gitignored; use `.env.example` only |
| API authentication | Merchant `api_token` is **not** enforced on SEP endpoints yet |
| Merchant registration | Open (`POST /merchants/register`) — fine on localhost, unsafe if public |
| Redis | No default password; bind Redis to localhost or set `REDIS_PASSWORD` |
| Docker image | Runs as non-root; `.dockerignore` keeps secrets out of build context |

**Do not** publish this service on the open internet without auth, network restrictions, and a strong Redis password.

---

## Who is this for?

- Backend engineers integrating SEP / Shaparak-style flows
- QA teams writing end-to-end payment tests
- Product teams demoing checkout without bank paperwork

---

## Roadmap

- [x] SEP token, pay page, verify, reverse
- [x] Dynamic merchant registration + Redis persistence
- [ ] Enforce `api_token` on sensitive endpoints
- [ ] Mellat / Parsian simulators
- [ ] Scripted scenarios (timeout, decline, double-verify)

Contributions and bank plugins are welcome.

---

## License

[MIT](LICENSE) © Mohsen Farahani
