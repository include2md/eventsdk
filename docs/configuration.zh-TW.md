# EventSDK 設定說明（繁中）

Last updated: 2026-04-28

## 1. `twsp.Options` 欄位

```go
type Options struct {
    NATSURL          string
    Timeout          time.Duration
    Username         string
    Password         string
    Token            string
    AppID            string
    ConnectorID      string
    CloudEventSource string
}
```

## 2. 欄位用途
- `NATSURL`: NATS 連線位址，預設 `nats://127.0.0.1:4222`
- `Timeout`: Request timeout，預設 3 秒
- `Username`/`Password`: 帳密驗證
- `Token`: Token 驗證（若提供，優先於帳密）
- `AppID`: CloudEvents extension `appid`
- `ConnectorID`: 用來推導 CloudEvents `source`
- `CloudEventSource`: 直接指定 CloudEvents `source`

## 3. CloudEvents source 優先順序
1. `CloudEventSource`（最高優先）
2. `ConnectorID`（轉成 `urn:connector:<connector_id>`）
3. 都沒給時使用 SDK 預設值 `urn:connector:unknown`

## 4. 建議設定

### 生產環境
- 明確指定 `AppID`
- 指定穩定的 `ConnectorID`（一個 connector 一個 durable 名稱）
- 視 SLA 調整 `Timeout`

### 本地開發
```go
client, err := twsp.NewClient(twsp.Options{
    NATSURL:     "nats://127.0.0.1:4222",
    AppID:       "demo",
    ConnectorID: "demo-worker",
})
```

## 5. 常見錯誤
- 收不到事件：先確認 Stream subject pattern 是否覆蓋發送 subject
- Request timeout：確認對應 `Handle` 是否啟動，subject 是否完全一致
- source 不符合預期：檢查是否同時設定 `CloudEventSource` 和 `ConnectorID`
