# Go Event SDK Phase 1 Design

## Goal
建立一個 Go Event SDK，對外提供業務語意導向 API，內部整合 Core NATS（Command）與 JetStream（Event），並對 inbox 微服務做 internal subject mapping，不暴露 inbox/subject/envelope 細節給使用者。

## Scope
- Producer-only `Client`
- Consumer-only `Service`
- Internal: inbox mapping, dispatcher, registry, retry, envelope metadata
- Transport: NATS + JetStream
- Tests: unit tests + manual integration tests

## Non-Goals (Phase 1)
- DLQ
- backoff
- idempotency
- advanced consumer config

## Architecture
- `client`: `SendCommand(ctx, cmd)`、`PublishEvent(ctx, event)`
- `service`: `RegisterHandler(eventType, handler)`、`Run(ctx, cfg)`
- `internal/transport/nats`: 封裝 Core NATS request/reply 與 JetStream publish/subscribe
- `internal/subject`: subject resolver（含 inbox command mapping）
- `internal/registry`: event handler registry
- `internal/dispatcher`: envelope decode + handler dispatch + retry routing
- `internal/retry`: max retry = 3（republish + ack）
- `internal/envelope`: event metadata（event id/timestamp/correlation id）

## Data Flow
### Command
1. adapter/publisher 呼叫 `Client.SendCommand`
2. resolver 將 command name 映射到 internal subject
3. transport 透過 Core NATS request/reply 發送
4. error/success 原樣回傳

### Event Publish
1. user 呼叫 `Client.PublishEvent`
2. SDK 自動建立 envelope metadata
3. 透過 JetStream publish

### Event Consume
1. `Service.RegisterHandler(eventType, handler)`
2. `Service.Run` 啟動 consumer loop
3. dispatcher 解 envelope 並依 `eventType` 分派
4. success: ACK
5. fail: retry republish（attempt+1）並 ACK 原訊息
6. 超過 3 次：記錄錯誤後丟棄（無 DLQ）

## Error Handling
- Command: request timeout/transport/decode 直接回傳
- Publish: publish failure 直接回傳
- Consume: handler error 走 retry policy
- Correlation ID: 若 command/event 傳入則沿用，否則 SDK 產生

## Testing Strategy
### Unit tests（不依賴外部服務）
- registry: register/get
- subject resolver: command→subject mapping
- dispatcher: no handler / decode fail / handler fail / handler success
- retry: attempt increment + max retries

### Integration tests（手動執行）
- 前置：使用者先在 inbox repo 啟動 docker compose
- 驗證 command request/reply
- 驗證 event publish/consume
- 驗證 inbox command mapping 可用

### Smoke
- 最小 happy path：command flow + event flow

## User-facing API Constraints
- example 不出現 `inbox`
- example 不出現 subject string
- example 不手動建立 envelope
- API 命名必須是業務語意
