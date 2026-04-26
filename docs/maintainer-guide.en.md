# EventSDK Maintainer Guide (English)

Last updated: 2026-04-26

## 1. Purpose and audience
This guide is for maintainers who modify SDK internals, tests, and release workflows.

## 2. Architecture overview
Core public surface is under `sdk/`:
- `SDKClient` APIs: `Request`, `Handle`, `Emit`, `Listen`
- `Transport` interface for NATS/JetStream implementations
- `Message` as normalized event payload + metadata

High-level flow:
1. Producer calls `Emit` or `Request`
2. Transport layer publishes/requests through NATS/JetStream
3. Consumer handles data via `Listen`/`Handle`
4. Event envelope is decoded into `sdk.Message`

## 3. Module ownership
- `sdk/client.go`: SDKClient behavior and hook integration points
- `sdk/interfaces.go`: transport contracts
- `sdk/types.go`: public types (`Message`, handler types)
- `sdk/bootstrap/`: `twsp.NewClient` and connection options
- `sdk/subject/`: subject builder helpers
- `sdk/internal/`: internal implementations (not for external dependency)

## 4. Public API compatibility policy
- Public APIs in `sdk/` are treated as contracts
- Signature, field, and semantic changes are breaking changes
- `sdk/internal/*` may be refactored, but must not break public behavior
- Breaking changes must include migration notes

## 5. Testing strategy

### 5.1 Unit and contract tests
```bash
go test ./sdk/...
```

Key files:
- `sdk/client_test.go`
- `sdk/public_contract_test.go`
- `sdk/bootstrap/bootstrap_options_test.go`
- `sdk/bootstrap/bootstrap_signature_test.go`

### 5.2 Integration tests (requires external NATS/JetStream)
```bash
go test -tags integration ./sdk/integration -v
```

Reference: `sdk/integration/README.md`

## 6. Suggested change workflow
1. Add/update tests first
2. Implement minimal changes in `sdk/` and related internals
3. Run unit + integration tests when environment is ready
4. Update user/maintainer docs in zh-TW and English
5. Include compatibility and risk notes in PR

## 7. Release and changelog policy
Use semver and keep release notes with:
- Breaking changes
- Features
- Fixes
- Migration notes

## 8. Common maintenance risks
- Subject naming changes causing routing mismatches
- Envelope format changes breaking existing consumers
- Hook lifecycle updates introducing side effects
- Durable consumer naming changes causing duplicate or missing processing

## 9. Documentation sync rules
- Update `maintainer-guide.zh-TW.md` first
- Sync `maintainer-guide.en.md`
- Refresh dates and links in `docs/README.md`
