# Examples

Runnable examples are placed under `sdk/examples` to match the current SDK shape.

## Run consumer

```bash
go run ./sdk/examples/consumer
```

## Run producer

```bash
go run ./sdk/examples/producer
```

Environment variables:
- `NATS_URL` (default `nats://127.0.0.1:4222`)
- `SDK_NAMESPACE` (default `TW.XX`)
- `SDK_STREAM` (producer)
- `SDK_CONSUMER_NAME` (consumer)

Inbox bridge is convention-based and automatic. If event payload includes
`userId/messageId/title/description/category/box`, SDK will internally send inbox `CreateMessage`.
