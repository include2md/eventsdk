# EventSDK 維護者指南（繁中）

Last updated: 2026-04-26

## 1. 文件目的與適用對象
本文件給需要修改 SDK 核心邏輯、測試與發版流程的維護者。

## 2. 架構總覽
EventSDK 的對外核心位於 `sdk/`：
- `SDKClient` 對外 API：`Request` / `Handle` / `Emit` / `Listen`
- `Transport` 介面抽象實際 NATS/JetStream 實作
- `Message` 統一接收格式（event metadata + payload）

整體資料流：
1. Producer 呼叫 `Emit` 或 `Request`
2. Transport 對接 NATS/JetStream
3. Consumer 透過 `Listen`/`Handle` 收取資料
4. Event 經 envelope 還原為 `sdk.Message`

## 3. 模組責任
- `sdk/client.go`：SDKClient 行為與 hook 套用點
- `sdk/interfaces.go`：Transport 抽象契約
- `sdk/types.go`：公開型別（`Message`、handler 型別）
- `sdk/bootstrap/`：`twsp.NewClient` 與連線選項
- `sdk/subject/`：subject 建構 helper
- `sdk/internal/`：內部實作（不應直接被外部依賴）

## 4. Public API 相容性政策
- `sdk/` 對外可見 API 視為契約
- 變更方法簽名、型別欄位、行為語意屬 breaking change
- `sdk/internal/*` 可重構，但不得破壞公開 API 既有行為
- breaking change 必須附 migration 說明

## 5. 測試策略

### 5.1 單元與契約測試
```bash
go test ./sdk/...
```

重點檔案：
- `sdk/client_test.go`
- `sdk/public_contract_test.go`
- `sdk/bootstrap/bootstrap_options_test.go`
- `sdk/bootstrap/bootstrap_signature_test.go`

### 5.2 整合測試（需外部 NATS/JetStream）
```bash
go test -tags integration ./sdk/integration -v
```

參考：`sdk/integration/README.md`

## 6. 變更流程（建議）
1. 先寫/補測試，重現目標行為
2. 以最小變更修改 `sdk/` 或對應 internal 模組
3. 跑 unit + integration（可用環境下）
4. 更新 user/maintainer 中英文件
5. PR 附上相容性與風險說明

## 7. 發版與變更記錄
建議採用 semver 並維持 changelog：
- Breaking changes
- Features
- Fixes
- Migration notes

## 8. 常見維護風險
- subject 規則改動導致路由不一致
- envelope 欄位或編碼改動影響既有 consumer
- handle hook 規則變更造成副作用流程改變
- durable consumer 命名策略變更造成重複/遺漏消費

## 9. 文件同步規範
- 先更新 `maintainer-guide.zh-TW.md`
- 同步更新 `maintainer-guide.en.md`
- 更新 `docs/README.md` 的日期與導覽敘述
