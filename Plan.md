# Bank Simulator (Open Source)

## Goal
An open-source simulator for Iranian payment gateways (PSPs),
used for QA/testing/staging environments without connecting to real banks.

## MVP Phase (Saman Bank)
- Token request endpoint
- Payment page simulation
- Confirm endpoint
- Merchant callback redirect
- Redis-based transaction store

## Future Phase
- Multiple banks (Mellat, Parsian, Pasargad)
- Scenario simulation (success/fail/timeout)
- Admin API
- Web UI improvements

## Tech Stack
- Golang
- Redis
- Docker Compose
