# EventSDK 概念說明（繁中）

Last updated: 2026-04-28

## 1. 兩種互動模型

### Request/Reply（命令）
- 使用 `Request` 呼叫命令。
- 使用 `Handle` 接收命令並回應。
- 適合「需要立即結果」的流程。

### Publish/Subscribe（事件）
- 使用 `Emit` 發事件。
- 使用 `Listen` 訂閱事件。
- 適合「解耦、非同步」的流程。

## 2. Subject 命名慣例
- Command：`cmd.app.<app_id>.<domain>.<action>`
- Event：`evt.app.<app_id>.<domain>.<event>`

建議使用 helper：
- `subject.CmdApp(appID, domain, action)`
- `subject.EvtApp(appID, domain, event)`

## 3. CloudEvents 映射（SDK Emit）
`Emit` 會把 payload 包裝成 CloudEvents JSON。

### 3.1 標準欄位
- `specversion`: 固定 `1.0`
- `type`: 事件型別（目前預設使用 `Emit` 的 subject）
- `source`: 事件來源（預設 `urn:connector:unknown`，可由 options 設定）
- `subject`: 資源識別（可選）
- `id`: 事件 ID（UUID）
- `time`: 事件時間（UTC）
- `datacontenttype`: `application/json`
- `data`: 事件內容

### 3.2 data 欄位
- `data.request`: 你的業務 payload
- `data.subject`: 來源 command subject（可選）

### 3.3 Extension 欄位
- `correlationid`: 關聯 ID
- `appid`: 應用 ID
- `attempt`: 重試次數
- `natssubject`: NATS subject（目前與 `type` 一致）

## 4. 事件資料範例
```json
{
  "specversion": "1.0",
  "type": "evt.app.billing.invoice.created",
  "source": "urn:connector:billing-consumer",
  "id": "e1e0c7d6e2cd4b8f9f37f6b53289b4a8",
  "time": "2026-04-28T08:00:00Z",
  "datacontenttype": "application/json",
  "data": {
    "request": {
      "invoiceId": "inv-001",
      "amount": 1200
    },
    "subject": "cmd.app.billing.invoice.create"
  },
  "correlationid": "c-123",
  "appid": "billing",
  "attempt": 1,
  "natssubject": "evt.app.billing.invoice.created"
}
```
