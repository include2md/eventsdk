# Demo: Frontend -> NATS Request -> Adapter -> REST API -> Reply

This demo starts only:
- `adapter`
- `mock-api`

You run NATS yourself externally.

## 0) Start your own NATS with websocket enabled

Example ports expected by demo:
- NATS TCP: `4222`
- NATS WS: `9222`

## 1) Start adapter + mock-api

```bash
docker compose -f demo/docker-compose.yml up -d --build
```

If your NATS is not on host `4222`, edit `demo/docker-compose.yml`:
- `adapter.environment.NATS_URL`

## 2) Open frontend directly

Open file in browser:
- `demo/frontend/index.html`

Then:
1. Click `Connect`
2. Click `Send Request`
3. You should see reply JSON from adapter, which includes data from mock-api.
4. If payload includes `userId/messageId/title/description/category/box`, SDK in `Respond` will internally trigger inbox create.

Default request subject:
- `TW.XX.user.command.create`

Default payload:
```json
{"userId":"u-frontend","messageId":"m-frontend","title":"hello","description":"from frontend","category":"system","box":"primary"}
```

## 3) Inspect logs

```bash
docker compose -f demo/docker-compose.yml logs -f adapter mock-api
```
