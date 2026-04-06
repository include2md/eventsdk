# Go Event SDK 設計文件（NATS + JetStream + Inbox Integration）

---

# 一、設計目標

本 SDK 用於建立統一的事件驅動（Event-Driven）開發模式。

系統基於：

* **Core NATS** → request-reply（Command）
* **JetStream** → durable event（Event）

---

# 二、核心設計原則（必讀）

## 原則 1：使用者不應感知 Inbox

以下屬於 internal 概念：

* inbox
* dispatcher
* retry
* DLQ
* subject naming

❗ **不得出現在 user-facing API / example**

---

## 原則 2：API 必須是業務語意

✅ 正確：

```go
CreateMessage(ctx, req)
PublishUserRegistered(ctx, event)
```

❌ 錯誤：

```go
CreateInboxMessage(...)
Send("TW.XX.inbox.command.create", ...)
```

---

## 原則 3：SDK 負責可靠性

SDK 必須內部處理：

* retry
* at-least-once
* dispatch
* envelope metadata

---

## 原則 4：分離 Producer / Consumer

| 角色 | SDK     |
| -- | ------- |
| 發送 | Client  |
| 接收 | Service |

---

# 三、整體架構

```mermaid
flowchart TD
    A[User Code] --> B[Adapter / Domain Layer]

    B --> C[SDK Client]

    C -->|Command| D[Core NATS]
    D --> E[External Service / Inbox]

    C -->|Event| F[JetStream]
    F --> G[Inbox (internal)]
    G --> H[Dispatcher (internal)]
    H --> I[Handler]

    I --> J{Result}
    J -->|Success| K[ACK]
    J -->|Fail| L[Retry]
```

---

# 四、Command vs Event

| 類型      | 技術        | 特性                      |
| ------- | --------- | ----------------------- |
| Command | Core NATS | request-reply           |
| Event   | JetStream | durable + at-least-once |

---

# 五、SDK API（Phase 1）

## Client（Producer）

```go
SendCommand(ctx, cmd)
PublishEvent(ctx, event)
```

---

## Service（Consumer）

```go
RegisterHandler(eventType, handler)
Run(ctx, config)
```

---

# 六、User-facing Example（標準）

## Producer

```go
adapter.CreateMessage(ctx, req)
publisher.PublishUserRegistered(ctx, event)
```

---

## Consumer

```go
service.RegisterHandler("UserRegistered", handler)
service.Run(ctx, cfg)
```

---

## ❌ 禁止

* inbox 出現在 method 名稱
* subject string 出現在 user code
* 手動建立 Envelope

---

# 七、Adapter 設計

```go
type MessageAdapter struct {
    client sdk.Client
}

func (a *MessageAdapter) CreateMessage(ctx context.Context, req CreateMessageRequest) error {
    return a.client.SendCommand(ctx, sdk.Command{
        Name: "CreateMessage",
        Payload: req,
    })
}
```

---

# 八、Event 發送

```go
func (p *UserPublisher) PublishUserRegistered(ctx context.Context, event UserRegistered) error {
    return p.client.PublishEvent(ctx, sdk.Event{
        Type: "UserRegistered",
        Payload: event,
    })
}
```

---

## SDK 必須負責

* Event ID
* Timestamp
* Correlation ID

---

# 九、Transport 抽象

```go
type Transport interface {
    Publish(...)
    Request(...)
    Subscribe(...)
}
```

---

# 十、Retry（Phase 1）

* max retry = 3
* 使用 republish + ack（簡化版）
* ❌ 不實作 DLQ（Phase 2）

---

# 十一、Inbox 微服務整合（重要）

## 📦 外部參考系統

本 SDK 需整合既有 Inbox 微服務：

👉 https://github.com/include2md/inbox

此 repo 提供：

* docker-compose（NATS / JetStream / inbox）
* inbox command API（透過 NATS）
* frontend push event

---

## ❗ 定位（非常重要）

Inbox 必須被視為：

> ✅ 外部微服務
> ❌ SDK internal
> ❌ SDK public API

---

## ❗ 設計規範

### 1. 不暴露 inbox 給使用者

❌ 不允許：

```go
CreateInboxMessage(...)
Send("TW.XX.inbox.command.create", ...)
```

---

### 2. Subject mapping 必須 internal 化

SDK internal 可做：

```go
CreateMessage → TW.XX.inbox.command.create
```

但 user 不應知道。

---

### 3. 建議實作 Subject Resolver（internal）

```go
type SubjectResolver interface {
    InboxCreate(namespace string) string
}
```

---

## Inbox Command Subjects（參考）

* TW.XX.inbox.command.list
* TW.XX.inbox.command.get
* TW.XX.inbox.command.create
* TW.XX.inbox.command.read
* TW.XX.inbox.command.delete

---

## Frontend Push

* TW.XX.user.<user_id>.inbox.>

---

# 十二、測試策略（重要）

## Unit Test

* 不依賴 NATS / docker-compose
* 測試 registry / dispatcher / retry

---

## Integration Test

需使用 inbox repo：

```bash
docker compose up -d --build
```

測試內容：

* Core NATS request-reply（command）
* JetStream publish / consume（event）
* inbox integration

---

## Smoke Test

驗證：

* command flow 可通
* event flow 可通

---

## ❗ Codex 測試要求

* integration test 可依賴 docker-compose
* unit test 不可依賴外部服務

---

# 十三、Phase 1 限制

為降低複雜度：

* ❌ 無 DLQ
* ❌ 無 backoff
* ❌ 無 idempotency
* ❌ 無 advanced consumer config

---

# 十四、設計總結

* Command → Core NATS
* Event → JetStream
* Inbox → internal abstraction
* Inbox service → external integration

---

## 一句話總結

👉 使用者只寫業務邏輯
👉 SDK 負責傳遞與可靠性
👉 Inbox 是外部服務，不是 SDK API

---

# 十五、Codex 實作要求

請生成：

1. Client（producer-only）
2. Service（consumer-only）
3. internal inbox / dispatcher
4. registry
5. transport（NATS + JetStream）
6. retry（簡化版）
7. subject resolver（internal）
8. integration test（使用 inbox repo）

---

## ❗ 最重要

生成的 example 必須：

* 不出現 inbox
* 不出現 subject string
* 不手動建立 Envelope
* 必須是業務語意 API

---
