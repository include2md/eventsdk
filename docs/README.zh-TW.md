# EventSDK 快速上手（繁中）

Last updated: 2026-04-28

## 1. 這份文件是給誰看的
這份文件給「第一次導入 EventSDK」的開發者，目標是 3 分鐘內跑起最小流程。

## 2. 前置需求
- Go `1.23.0+`
- 可連線 NATS（預設 `nats://127.0.0.1:4222`）

安裝：
```bash
go get github.com/include2md/eventsdk
```

## 3. 建立 Client
```go
package main

import (
    "log"

    twsp "github.com/include2md/eventsdk/sdk/bootstrap"
)

func main() {
    client, err := twsp.NewClient(twsp.Options{
        NATSURL:     "nats://127.0.0.1:4222",
        AppID:       "billing",
        ConnectorID: "billing-consumer", // CloudEvents source => urn:connector:billing-consumer
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
}
```

## 4. 你最常用的 4 個 API
- `Request(ctx, subject, payload)`：送 command，等 reply。
- `Handle(ctx, subject, handler)`：註冊 command handler。
- `Emit(ctx, subject, payload)`：發 event（SDK 會封裝成 CloudEvents）。
- `Listen(ctx, subject, consumerName, handler)`：訂閱 event。

## 5. 最小流程範例
發 event：
```go
err := client.Emit(ctx, "evt.app.billing.invoice.created", map[string]any{
    "invoiceId": "inv-001",
    "amount":    1200,
})
```

收 event：
```go
err := client.Listen(ctx, "evt.app.billing.invoice.*", "billing-worker", func(ctx context.Context, msg sdk.Message) error {
    // msg.Payload 是 JSON bytes
    return nil
})
```

註冊 command handler：
```go
err := client.Handle(ctx, "cmd.app.billing.invoice.create", func(ctx context.Context, request []byte) ([]byte, error) {
    return []byte(`{"ok":true}`), nil
})
```

## 6. 下一份該看哪一個
- 觀念與事件格式：[`concepts.zh-TW.md`](./concepts.zh-TW.md)
- 設定細節：[`configuration.zh-TW.md`](./configuration.zh-TW.md)
- 實戰範例：[`examples.zh-TW.md`](./examples.zh-TW.md)
