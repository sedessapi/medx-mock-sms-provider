# medx-mock-sms-provider

Mock SMS Provider for **PSB-1** — implements **Provider Protocol v1**.

| Item | Value |
|------|--------|
| API | `http://localhost:8063` |
| Protocol | `/provider/v1/...` |
| Operation | `send-sms` |

## Quick start

```bash
cp .env.example .env   # or use committed defaults
go run ./cmd/api
```

## Endpoints

* `GET /healthz`, `GET /readyz`, `GET /openapi.json`
* `GET /provider/v1/manifest`
* `POST /provider/v1/test`
* `POST /provider/v1/invocations`

Simulation header: `X-Simulation-Mode: accepted|rejected|unavailable|timeout|malformed`

Contract: [`../medx-docs/psb-1/provider-protocol-v1.md`](../medx-docs/psb-1/provider-protocol-v1.md)
