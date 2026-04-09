# Runnable Examples (Subject-Only)

This SDK now uses subject-first APIs.
Examples use simplified bootstrap helpers:
- `twsp.NewClient(...)`

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

## Run consumer (event + command adapter)

```bash
go run ./sdk/examples/consumer
```

- Subscribes to events: `<namespace>.user.event.*`
- Handles request-reply commands: `<namespace>.user.command.create`

## Run producer

```bash
go run ./sdk/examples/producer
```

## Try request-reply manually

After consumer is running:

```bash
nats req "TW.XX.user.command.create" '{"name":"demo"}' -s nats://127.0.0.1:4222
```

You should receive a JSON reply from the adapter handler.

## Inbox bridge convention

No bridge configuration is required.

When published payload contains all fields below, SDK internally sends inbox `CreateMessage` command:
- `userId`
- `messageId`
- `title`
- `description`
- `category`
- `box`
