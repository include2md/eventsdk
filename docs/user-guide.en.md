# EventSDK User Guide (English)

Last updated: 2026-04-26

## 1. What this SDK does
EventSDK provides a unified API for command and event workflows on top of NATS + JetStream.

It supports:
- Request/reply command calls
- Event publish/subscribe
- Envelope metadata management (`eventId`, `correlationId`, `timestamp`)

## 2. Quick Start (5 minutes)

### 2.1 Prerequisites
- Go `1.23.0+`
- Reachable NATS endpoint (default `nats://127.0.0.1:4222`)

### 2.2 Install
```bash
go get github.com/include2md/eventsdk
```

### 2.3 Run minimal examples
Start consumer first:
```bash
go run ./sdk/examples/consumer
```

Then run producer in another terminal:
```bash
go run ./sdk/examples/producer
```

Default env vars:
- `NATS_URL`: `nats://127.0.0.1:4222`
- `SDK_APP_ID`: `tdemo`
- `SDK_STREAM`: `SDK_EVENTS` (producer)
- `SDK_CONSUMER_NAME`: consumer durable name

## 3. Core APIs (user view)
- `Request(ctx, subject, payload)`: send command and wait for reply
- `Handle(ctx, subject, handler)`: register request/reply handler
- `Emit(ctx, subject, payload)`: publish event (wrapped by SDK envelope)
- `Listen(ctx, subject, consumerName, handler)`: consume events

## 4. Common scenarios

### 4.1 Create client
```go
service, err := twsp.NewClient(twsp.Options{NATSURL: "nats://127.0.0.1:4222"})
if err != nil {
    return err
}
defer service.Close()
```

### 4.2 Send command
```go
resp, err := service.Request(ctx, "cmd.app.tdemo.user.create", map[string]any{"name": "demo"})
```

### 4.3 Register handler
```go
err := service.Handle(ctx, "cmd.app.tdemo.user.create", func(ctx context.Context, request []byte) ([]byte, error) {
    return []byte(`{"ok":true}`), nil
})
```

### 4.4 Emit event
```go
err := service.Emit(ctx, "evt.app.tdemo.user.registered", map[string]any{"userId": "u-1"})
```

### 4.5 Listen event
```go
err := service.Listen(ctx, "evt.app.tdemo.user.*", "demo-consumer", func(ctx context.Context, msg sdk.Message) error {
    return nil
})
```

## 5. Subject naming convention
Prefer helper functions over manual strings:
- `subject.CmdApp(appID, domain, action)`
- `subject.EvtApp(appID, domain, event)`

Conventions:
- Command: `cmd.app.<app_id>.<domain>.<action>`
- Event: `evt.app.<app_id>.<domain>.<event>`

## 6. Error handling notes
- Tune request timeout based on your SLA (default client timeout is 3 seconds)
- Return explicit errors in handlers
- Keep stable, meaningful consumer names for operations

## 7. FAQ
Q: Why no event is received?
- Ensure stream covers your event subject pattern
- Ensure publish and listen subjects match

Q: Why request timeout happens?
- Ensure `Handle` is running for that command subject
- Ensure command subject names are identical

Q: How to verify request/reply manually?
```bash
nats req "cmd.app.tdemo.user.create" '{"name":"demo"}' -s nats://127.0.0.1:4222
```

## 8. Troubleshooting checklist
- Is NATS running?
- Are subject names consistent?
- Is `SDK_CONSUMER_NAME` conflicting?
- Is payload valid JSON?
- Check runnable docs: `sdk/examples/README.md`, `demo/README.md`

## 9. Next steps
- `sdk/examples/README.md`
- `sdk/integration/README.md`
- `demo/README.md`
- `maintainer-guide.en.md`
