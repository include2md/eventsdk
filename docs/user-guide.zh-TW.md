# EventSDK 使用者指南（繁中）

Last updated: 2026-04-28

## 1. 文件定位
這份文件現在是「總覽與導覽頁」。
如果你是第一次使用 SDK，先看本頁，再依需求跳到分章文件。

## 2. EventSDK 提供什麼
EventSDK 提供 NATS + JetStream 之上的一致 API：
- `Request/Handle`：命令請求與回應
- `Emit/Listen`：事件發布與訂閱
- 事件封裝採用 CloudEvents（`specversion: 1.0`）

## 3. 建議閱讀順序（繁中）
1. 快速上手：[`README.zh-TW.md`](./README.zh-TW.md)
2. 概念與事件模型：[`concepts.zh-TW.md`](./concepts.zh-TW.md)
3. 設定與參數：[`configuration.zh-TW.md`](./configuration.zh-TW.md)
4. 實戰範例：[`examples.zh-TW.md`](./examples.zh-TW.md)

## 4. 常見任務入口
- 我想 3 分鐘跑起來：看 [`README.zh-TW.md`](./README.zh-TW.md)
- 我想確認 payload / CloudEvents 欄位：看 [`concepts.zh-TW.md`](./concepts.zh-TW.md)
- 我想設定 `AppID`、`ConnectorID`、`CloudEventSource`：看 [`configuration.zh-TW.md`](./configuration.zh-TW.md)
- 我想看完整呼叫與訂閱範例：看 [`examples.zh-TW.md`](./examples.zh-TW.md)

## 5. 其他相關文件
- 文件中心：[`README.md`](./README.md)
- 維護者指南（繁中）：[`maintainer-guide.zh-TW.md`](./maintainer-guide.zh-TW.md)
- SDK 範例執行說明：[`../sdk/examples/README.md`](../sdk/examples/README.md)
