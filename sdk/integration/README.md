# Integration Tests (Manual)

These tests validate real NATS + JetStream flows for this SDK.

## Prerequisites

1. Start infrastructure manually (from `inbox` repo):

```bash
docker compose up -d --build
```

2. Ensure NATS is reachable (default: `nats://127.0.0.1:4222`).

## Run

```bash
go test -tags integration ./sdk/integration -v
```

Optional env vars:
- `NATS_URL` (default `nats://127.0.0.1:4222`)
- `SDK_NAMESPACE` (default `TW.XX`)
- `SDK_STREAM` (default `SDK_EVENTS`)
