# Bridge Pipeline Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Introduce a rule-based bridge pipeline around `Handle/Listen/Request/Emit` without breaking existing public APIs, and migrate inbox bridge behavior out of hard-coded client flow.

**Architecture:** Add a `BridgePipeline` abstraction into `SDKClient`, run stage-based hooks (`Before*`/`After*`) with deterministic ordering, and move inbox behavior into dedicated rules (`AfterEmit`, `AfterHandle`). Keep default behavior backward-compatible via policy defaults.

**Tech Stack:** Go 1.24, standard library, existing SDK transport/envelope modules, `go test`.

---

### Task 1: Add Bridge Core Types and Pipeline Skeleton

**Files:**
- Create: `sdk/bridge_pipeline.go`
- Test: `sdk/bridge_pipeline_test.go`

**Step 1: Write the failing test**

```go
func TestPipelineRunsMatchedRulesByPriority(t *testing.T) {
    // prepare rules with priority 10, 0, 5; expect execution order 0,5,10
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./sdk -run TestPipelineRunsMatchedRulesByPriority -v`
Expected: FAIL because pipeline types do not exist.

**Step 3: Write minimal implementation**

- Add `Stage`, `BridgeContext`, `BridgeRule`, `ErrorPolicy`, `BridgePipeline`.
- Implement `Apply(...)` with:
  - stage filtering
  - match filtering
  - sort by priority ascending
  - per-rule policy handling hooks (ignore/fail/retry skeleton)

**Step 4: Run test to verify it passes**

Run: `go test ./sdk -run TestPipelineRunsMatchedRulesByPriority -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/bridge_pipeline.go sdk/bridge_pipeline_test.go
git commit -m "feat: add bridge pipeline core abstractions"
```

### Task 2: Enforce Rewrite Constraints for Before/After Stages

**Files:**
- Modify: `sdk/bridge_pipeline.go`
- Modify: `sdk/bridge_pipeline_test.go`

**Step 1: Write the failing test**

```go
func TestPipelineRejectsSecondSubjectRewriteInSameStage(t *testing.T) {}
func TestPipelineDisallowsRewriteInAfterStage(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./sdk -run "TestPipelineRejectsSecondSubjectRewriteInSameStage|TestPipelineDisallowsRewriteInAfterStage" -v`
Expected: FAIL because constraints are not implemented.

**Step 3: Write minimal implementation**

- Add rewrite tracking in `Apply(...)`.
- Subject can be rewritten only once per `Apply` call.
- For `After*` stages, reject subject/payload rewrite attempts.

**Step 4: Run test to verify it passes**

Run: `go test ./sdk -run "TestPipelineRejectsSecondSubjectRewriteInSameStage|TestPipelineDisallowsRewriteInAfterStage" -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/bridge_pipeline.go sdk/bridge_pipeline_test.go
git commit -m "test: enforce deterministic bridge rewrite constraints"
```

### Task 3: Integrate Pipeline into SDKClient Lifecycle Hooks

**Files:**
- Modify: `sdk/client.go`
- Modify: `sdk/types.go` (if context metadata helpers are needed)
- Test: `sdk/client_test.go`
- Test: `sdk/service_test.go`

**Step 1: Write the failing test**

Add tests:
- `TestClientEmitRunsBeforeAndAfterBridgeStages`
- `TestServiceHandleRunsAfterHandleBridgeStage`
- `TestClientEmitBeforeStageCanRewritePayload`

**Step 2: Run test to verify it fails**

Run: `go test ./sdk -run "TestClientEmitRunsBeforeAndAfterBridgeStages|TestServiceHandleRunsAfterHandleBridgeStage|TestClientEmitBeforeStageCanRewritePayload" -v`
Expected: FAIL because client has no pipeline integration.

**Step 3: Write minimal implementation**

- Extend `SDKClient` with `pipeline *BridgePipeline`.
- Add constructor option or default pipeline injection without breaking `NewClient(...)` signature.
- In `Emit`:
  - run `BeforeEmit` pipeline before envelope creation/publish
  - run `AfterEmit` pipeline after publish
- In `Handle`:
  - run `BeforeHandle` prior to handler execution
  - run `AfterHandle` only on successful handler flow

**Step 4: Run test to verify it passes**

Run: `go test ./sdk -run "TestClientEmitRunsBeforeAndAfterBridgeStages|TestServiceHandleRunsAfterHandleBridgeStage|TestClientEmitBeforeStageCanRewritePayload" -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/client.go sdk/client_test.go sdk/service_test.go sdk/types.go
git commit -m "feat: hook bridge pipeline into client lifecycle"
```

### Task 4: Migrate Inbox Bridge to Rule Implementations

**Files:**
- Create: `sdk/bridge_inbox_rules.go`
- Modify: `sdk/inbox_bridge.go`
- Modify: `sdk/client.go`
- Test: `sdk/client_test.go`
- Test: `sdk/service_test.go`

**Step 1: Write the failing test**

Add tests:
- `TestInboxAfterEmitRuleMatchesAndRequestsCreate`
- `TestInboxAfterHandleRuleMatchesAndRequestsCreate`
- `TestInboxRuleSkipsWhenPayloadIncomplete`

**Step 2: Run test to verify it fails**

Run: `go test ./sdk -run "TestInboxAfterEmitRuleMatchesAndRequestsCreate|TestInboxAfterHandleRuleMatchesAndRequestsCreate|TestInboxRuleSkipsWhenPayloadIncomplete" -v`
Expected: FAIL because rules do not exist.

**Step 3: Write minimal implementation**

- Implement inbox rules as `BridgeRule` with `Stage`-specific matching.
- Reuse existing payload mapping from `mapToInboxCreatePayload`.
- Register rules in default pipeline.
- Remove direct calls to `bridgeInboxFromRequest/bridgeInboxFromPayload` from client flow.

**Step 4: Run test to verify it passes**

Run: `go test ./sdk -run "TestInboxAfterEmitRuleMatchesAndRequestsCreate|TestInboxAfterHandleRuleMatchesAndRequestsCreate|TestInboxRuleSkipsWhenPayloadIncomplete" -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/bridge_inbox_rules.go sdk/inbox_bridge.go sdk/client.go sdk/client_test.go sdk/service_test.go
git commit -m "refactor: move inbox bridge logic into bridge rules"
```

### Task 5: Add Error Policy Behavior Coverage

**Files:**
- Modify: `sdk/bridge_pipeline_test.go`
- Modify: `sdk/client_test.go`

**Step 1: Write the failing test**

Add tests:
- `TestPipelineIgnorePolicySwallowsRuleError`
- `TestPipelineFailPolicyStopsStage`
- `TestPipelineRetryPolicyRetriesRule`
- `TestEmitReturnsErrorWhenFailPolicyTriggered`

**Step 2: Run test to verify it fails**

Run: `go test ./sdk -run "TestPipelineIgnorePolicySwallowsRuleError|TestPipelineFailPolicyStopsStage|TestPipelineRetryPolicyRetriesRule|TestEmitReturnsErrorWhenFailPolicyTriggered" -v`
Expected: FAIL while policy behavior is incomplete.

**Step 3: Write minimal implementation**

- Complete policy handling in `Apply(...)`.
- Ensure `fail` returns error to caller.
- Ensure `ignore` continues.
- Ensure `retry` retries current rule only.

**Step 4: Run test to verify it passes**

Run: `go test ./sdk -run "TestPipelineIgnorePolicySwallowsRuleError|TestPipelineFailPolicyStopsStage|TestPipelineRetryPolicyRetriesRule|TestEmitReturnsErrorWhenFailPolicyTriggered" -v`
Expected: PASS.

**Step 5: Commit**

```bash
git add sdk/bridge_pipeline.go sdk/bridge_pipeline_test.go sdk/client_test.go
git commit -m "test: cover bridge rule error policy behaviors"
```

### Task 6: End-to-End Regression Sweep and Documentation

**Files:**
- Modify: `sdk/examples/README.md`
- Modify: `docs/plans/2026-04-26-bridge-pipeline-design.md` (status/update section)

**Step 1: Write/adjust docs first (no code changes)**

- Document that inbox bridge is now rule-based under bridge pipeline.
- Keep existing payload convention unchanged.

**Step 2: Run full relevant tests**

Run: `go test ./sdk/...`
Expected: PASS across sdk packages.

**Step 3: Run integration tests if infra is ready**

Run: `go test -tags=integration ./sdk/integration -v`
Expected: PASS when inbox stack is up, otherwise note as not-run/blocked.

**Step 4: Final commit**

```bash
git add sdk/examples/README.md docs/plans/2026-04-26-bridge-pipeline-design.md
git commit -m "docs: describe rule-based bridge pipeline behavior"
```

