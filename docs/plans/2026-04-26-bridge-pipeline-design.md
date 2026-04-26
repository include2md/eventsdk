# Bridge Pipeline Design (EventSDK)

**Date:** 2026-04-26  
**Status:** Draft approved by user conversation

## 1. Problem Statement

Current bridge behavior is hard-coded in [`sdk/client.go`](../../sdk/client.go):
- `Handle(...)` calls `bridgeInboxFromRequest(...)` after handler success.
- `Emit(...)` calls `bridgeInboxFromPayload(...)` after publish success.

This creates tight coupling to inbox-specific mapping (`TW.XX.inbox.command.create`) and makes future post/pre actions hard to scale.

## 2. Goals

1. Keep public API (`Handle/Listen/Request/Emit`) stable.
2. Introduce extensible hook points around these APIs.
3. Allow rule-based bridge actions with per-rule error policy.
4. Allow `before` stage to rewrite `subject` and `payload`.
5. Keep existing inbox bridge behavior as first migrated rule set.

## 3. Non-Goals

1. No breaking changes to existing SDK constructor usage in V1.
2. No global rule DSL/parser in V1 (code-defined rules only).
3. No distributed transaction semantics between main flow and bridge actions.

## 4. Proposed Architecture

### 4.1 New Core Types

- `Stage` enum:
  - `BeforeRequest`, `AfterRequest`
  - `BeforeEmit`, `AfterEmit`
  - `BeforeHandle`, `AfterHandle`
  - `BeforeListen`, `AfterListen`

- `BridgeContext`:
  - `Stage Stage`
  - `Subject string`
  - `Payload any`
  - `Reply []byte`
  - `Err error`
  - `Metadata map[string]string` (trace/tenant/app-id extensibility)

- `BridgeRule` interface:
  - `Name() string`
  - `Priority() int`
  - `Match(*BridgeContext) bool`
  - `Run(context.Context, *BridgeContext) error`
  - `Policy() ErrorPolicy`

- `ErrorPolicy`:
  - `Mode` = `ignore | fail | retry`
  - `MaxRetry int`
  - `Backoff time.Duration`

### 4.2 Pipeline Execution

Introduce `BridgePipeline` in `SDKClient`:
- Holds registered `BridgeRule` list.
- Executes rules by `Stage` and `Priority`.
- Exposes one method: `Apply(ctx context.Context, bc *BridgeContext) error`.

### 4.3 Deterministic Conflict Rules

1. Rules run in ascending `Priority` (`0` is highest).
2. `before` stages support chained rewrite:
   - Rule N output becomes Rule N+1 input.
3. `Subject` rewrite allowed only once in a stage.
4. `Payload` rewrite can happen multiple times but must replace full value.
5. `after` stages must not rewrite subject/payload (side effects only).
6. `retry` retries only the failing rule, not previous successful rules.

## 5. Data Flow by API

### 5.1 Emit

1. Build `BridgeContext{Stage: BeforeEmit, Subject, Payload}`.
2. Apply pipeline. If fail policy triggers -> return error.
3. Build envelope from (possibly rewritten) subject/payload.
4. Publish to transport.
5. Build `BridgeContext{Stage: AfterEmit, ...}` with result/error.
6. Apply pipeline.

### 5.2 Handle

1. On incoming request bytes, decode payload for bridge context.
2. Apply `BeforeHandle`.
3. Execute user handler.
4. Apply `AfterHandle` (existing inbox behavior moved here as rule).
5. Return user handler response unchanged.

### 5.3 Request / Listen

- Same pattern with stage-specific context and hooks.
- V1 may start with `Emit` + `Handle` hooks enabled first, then expand.

## 6. Error Handling Model

Per-rule policy:
- `ignore`: swallow rule error, continue.
- `fail`: stop stage and return error to main flow.
- `retry`: retry this rule with backoff, then fallback to fail/ignore according to mode design in code.

Recommendation for initial default:
- Backward compatibility: migrated inbox rules default `ignore`.

## 7. Observability Requirements

For each rule execution log/metric fields:
- `rule_name`
- `stage`
- `matched`
- `duration_ms`
- `result` (`success|error|skipped`)
- `error_mode`

V1 can use structured logs; metric integration can be phase 2.

## 8. Migration Plan

1. Add pipeline with no rules: behavior unchanged.
2. Add inbox create mapping as `AfterEmit` rule.
3. Add inbox create mapping from request as `AfterHandle` rule.
4. Remove `bridgeInboxFromRequest` / `bridgeInboxFromPayload` hard-coded calls.
5. Keep mapping logic from `sdk/inbox_bridge.go` as reusable mapper utility.

## 9. Testing Strategy

1. Pipeline unit tests:
  - priority ordering
  - single subject rewrite constraint
  - policy behaviors (`ignore/fail/retry`)
2. Client tests:
  - `Emit` executes before/after stages
  - `Handle` executes after success only
  - bridge failure with `fail` stops main flow
  - bridge failure with `ignore` does not stop main flow
3. Backward compatibility:
  - existing inbox auto-bridge tests still pass after migration.

## 10. Risks and Mitigations

1. Risk: hidden behavior changes from rewrite support.
  - Mitigation: start rewrite only in `BeforeEmit`/`BeforeRequest`; leave others read-only until needed.
2. Risk: rule side effects duplicate during retries.
  - Mitigation: retry scope only per rule, require idempotent external commands.
3. Risk: complexity growth.
  - Mitigation: strict rule interface and deterministic execution constraints.

## 11. Decision Summary

Adopt **Hook Pipeline + Rule** hybrid architecture with:
- before-stage rewrite enabled,
- after-stage side effects only,
- per-rule error policy,
- deterministic ordering and conflict control.

This replaces current hard-coded inbox bridge and prepares SDK for multiple future bridge actions.
