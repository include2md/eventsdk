# Handle Bridge Hooks Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add configurable `BeforeHandle` and `AfterHandle` bridge hooks with per-rule `ignore|fail` policy and global default policy `fail`.

**Architecture:** Introduce a small bridge pipeline dedicated to `Handle` flow only. Keep `NewClient` behavior available while adding an options-based constructor for bridge rules and policy. Execute before-rules pre-handler and after-rules post-handler-success, but never override a successful handler response.

**Tech Stack:** Go 1.24, standard library, existing SDK transport abstraction, `go test`.

---

### Task 1: Add Bridge Core Types for Handle Stages

**Files:**
- Create: `sdk/bridge_hooks.go`
- Test: `sdk/bridge_hooks_test.go`

**Step 1: Write failing tests**
- Rule order by priority
- Default policy fallback
- `fail` stops stage
- `ignore` continues stage

**Step 2: Run tests to verify fail**
Run: `go test ./sdk -run "TestBridgeHooks" -v`
Expected: FAIL (types/pipeline missing)

**Step 3: Implement minimal code**
- Add `BridgeStage`, `BridgeContext`, `BridgePolicyMode`, `BridgeRule`, `BridgeHooks`
- Add stage execution with policy handling

**Step 4: Run tests to verify pass**
Run: `go test ./sdk -run "TestBridgeHooks" -v`
Expected: PASS

### Task 2: Integrate Hooks into Handle Flow

**Files:**
- Modify: `sdk/client.go`
- Modify: `sdk/types.go`
- Test: `sdk/service_test.go`

**Step 1: Write failing tests**
- BeforeHandle fail blocks handler
- BeforeHandle ignore allows handler
- AfterHandle executes only after handler success
- AfterHandle fail does not override successful response

**Step 2: Run tests to verify fail**
Run: `go test ./sdk -run "TestServiceHandle" -v`
Expected: FAIL (hooks not integrated)

**Step 3: Implement minimal code**
- Add `BridgeOptions` + options constructor
- Wire `BeforeHandle` and `AfterHandle` around `Handle`
- Preserve response semantics for successful handler

**Step 4: Run tests to verify pass**
Run: `go test ./sdk -run "TestServiceHandle" -v`
Expected: PASS

### Task 3: Remove Hard-Coded Handle Inbox Bridge and Keep Emit Behavior Stable

**Files:**
- Modify: `sdk/client.go`
- Test: `sdk/service_test.go`
- Test: `sdk/client_test.go`

**Step 1: Write/adjust failing tests**
- Handle no longer auto-runs inbox bridge by default
- Existing emit tests remain green

**Step 2: Run tests to verify fail**
Run: `go test ./sdk -run "TestServiceHandleRequestAutoBridgeAfterSuccess|TestClientPublish" -v`
Expected: mixed fail/pass until migration complete

**Step 3: Implement minimal code**
- Remove hard-coded `bridgeInboxFromRequest` call from `Handle`
- Keep `Emit` behavior unchanged for now

**Step 4: Run tests to verify pass**
Run: `go test ./sdk -run "TestServiceHandle|TestClientPublish" -v`
Expected: PASS

### Task 4: Full Regression

**Files:**
- Test only

**Step 1: Run full sdk tests**
Run: `go test ./sdk/...`
Expected: PASS
