# EventSDK 實戰範例（繁中）

Last updated: 2026-04-28

## 1. 建立 Client
```go
service, err := twsp.NewClient(twsp.Options{
    NATSURL:     "nats://127.0.0.1:4222",
    AppID:       "billing",
    ConnectorID: "billing-consumer",
})
if err != nil {
    return err
}
defer service.Close()
```

## 2. Emit：發布事件
```go
err := service.Emit(ctx, "evt.app.billing.invoice.created", map[string]any{
    "invoiceId": "inv-001",
    "amount":    1200,
})
if err != nil {
    return err
}
```

## 3. Listen：消費事件
```go
err := service.Listen(ctx, "evt.app.billing.invoice.*", "billing-projection", func(ctx context.Context, msg sdk.Message) error {
    log.Printf("subject=%s event_id=%s correlation_id=%s payload=%s",
        msg.Subject, msg.EventID, msg.CorrelationID, string(msg.Payload))
    return nil
})
if err != nil {
    return err
}
```

## 4. Handle：處理命令
```go
err := service.Handle(ctx, "cmd.app.billing.invoice.create", func(ctx context.Context, request []byte) ([]byte, error) {
    // TODO: 驗證並執行業務邏輯
    return []byte(`{"ok":true,"invoiceId":"inv-001"}`), nil
})
if err != nil {
    return err
}
```

## 5. Request：呼叫命令
```go
resp, err := service.Request(ctx, "cmd.app.billing.invoice.create", map[string]any{
    "customerId": "c-100",
    "amount":     1200,
})
if err != nil {
    return err
}
log.Printf("response=%s", string(resp))
```

## 6. 端到端流程（簡化）
1. API 或上游系統送出 command（`Request`）
2. 服務側 `Handle` 執行邏輯後回應
3. 服務側再 `Emit` domain event
4. 下游服務用 `Listen` 訂閱事件並更新自己的資料模型

## 7. 可直接執行的 repo 範例
- `go run ./sdk/examples/consumer`
- `go run ./sdk/examples/producer`
- 詳細說明：[`../sdk/examples/README.md`](../sdk/examples/README.md)
