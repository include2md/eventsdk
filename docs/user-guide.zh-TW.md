# EventSDK 使用者指南（繁中）

Last updated: 2026-04-26

## 1. 這份 SDK 在做什麼？
EventSDK 提供一致的命令與事件通訊介面，底層使用 NATS + JetStream。
你可以用它做：
- Request/Reply 命令呼叫
- Event 發布與訂閱
- 統一 envelope 與 metadata（`eventId` / `correlationId` / `timestamp`）

## 2. 快速開始（5 分鐘）

### 2.1 前置需求
- Go `1.23.0+`
- 可連線 NATS（預設 `nats://127.0.0.1:4222`）

### 2.2 安裝
```bash
go get github.com/include2md/eventsdk
```

### 2.3 直接跑範例
先開一個 consumer：
```bash
go run ./sdk/examples/consumer
```

再開另一個 terminal 跑 producer：
```bash
go run ./sdk/examples/producer
```

預設環境變數：
- `NATS_URL`: `nats://127.0.0.1:4222`
- `SDK_APP_ID`: `tdemo`
- `SDK_STREAM`: `SDK_EVENTS`（producer）
- `SDK_CONSUMER_NAME`: consumer durable 名稱

## 3. 核心 API（使用者視角）
- `Request(ctx, subject, payload)`：發送命令並等待回應
- `Handle(ctx, subject, handler)`：註冊 request-reply handler
- `Emit(ctx, subject, payload)`：發布事件（SDK 會封裝 envelope）
- `Listen(ctx, subject, consumerName, handler)`：訂閱事件

> 使用者通常不需要了解 internal bridge/inbox 細節即可上手。

## 4. 常見使用情境

### 4.1 建立 client
```go
service, err := twsp.NewClient(twsp.Options{NATSURL: "nats://127.0.0.1:4222"})
if err != nil {
    return err
}
defer service.Close()
```

### 4.2 發送命令（Request/Reply）
```go
resp, err := service.Request(ctx, "cmd.app.tdemo.user.create", map[string]any{"name": "demo"})
```

### 4.3 註冊命令處理器
```go
err := service.Handle(ctx, "cmd.app.tdemo.user.create", func(ctx context.Context, request []byte) ([]byte, error) {
    return []byte(`{"ok":true}`), nil
})
```

### 4.4 發送事件
```go
err := service.Emit(ctx, "evt.app.tdemo.user.registered", map[string]any{"userId": "u-1"})
```

### 4.5 監聽事件
```go
err := service.Listen(ctx, "evt.app.tdemo.user.*", "demo-consumer", func(ctx context.Context, msg sdk.Message) error {
    // handle msg.Payload
    return nil
})
```

## 5. Subject 命名建議
建議透過 helper 建立 subject，避免手刻字串：
- `subject.CmdApp(appID, domain, action)`
- `subject.EvtApp(appID, domain, event)`

命名慣例：
- Command：`cmd.app.<app_id>.<domain>.<action>`
- Event：`evt.app.<app_id>.<domain>.<event>`

## 6. 錯誤處理建議
- `Request` timeout 依情境調整（預設 client timeout 為 3 秒）
- handler 回傳錯誤時，讓呼叫端可觀察失敗
- consumer name 盡量固定且具語意，便於追蹤與維運

## 7. 常見問題（FAQ）
Q: 為什麼收不到事件？
- 檢查是否有先 `EnsureStream` 並包含對應 subject pattern
- 檢查 `Listen` 的 subject 是否與發送端一致

Q: 為什麼 request timeout？
- 檢查對應 `Handle` 是否已啟動
- 檢查 command subject 是否一致

Q: 如何手動驗證 request/reply？
- 啟動 consumer 後可用：
```bash
nats req "cmd.app.tdemo.user.create" '{"name":"demo"}' -s nats://127.0.0.1:4222
```

## 8. 疑難排解（Checklist）
- NATS 是否啟動
- subject 是否一致（建議用 `sdk/subject` helper）
- `SDK_CONSUMER_NAME` 是否衝突
- payload JSON 是否可被處理
- 參考執行紀錄：`sdk/examples/README.md`、`demo/README.md`

## 9. 下一步
- Runnable examples：`sdk/examples/README.md`
- Integration tests：`sdk/integration/README.md`
- Demo（frontend + adapter）：`demo/README.md`
- 維護者文件：`maintainer-guide.zh-TW.md`
