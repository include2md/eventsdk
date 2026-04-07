# Runnable Examples

These examples follow the current SDK structure and can run directly.

## Prerequisites

1. Start NATS/JetStream (for example from inbox repo):

```bash
docker compose up -d --build
```

2. Optional env vars:
- `NATS_URL` (default `nats://127.0.0.1:4222`)
- `SDK_NAMESPACE` (default `TW.XX`)
- `SDK_STREAM` (default `SDK_EVENTS`)
- `SDK_CONSUMER_NAME` (consumer only)

## Run consumer

```bash
go run ./sdk/examples/consumer
```

## Run producer

```bash
go run ./sdk/examples/producer
```

Producer publishes `UserRegistered`. Consumer receives and prints payload.

## Inbox bridge behavior

No bridge rule configuration is required.

When `PublishEvent` payload contains all inbox create required fields below, SDK internally calls inbox command `CreateMessage`:
- `userId`
- `messageId`
- `title`
- `description`
- `category`
- `box`
