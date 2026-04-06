# Go Event SDK Phase 1 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go Event SDK Phase 1 with producer/consumer split, internal inbox mapping, NATS+JetStream transport, simplified retry, and manual integration tests.

**Architecture:** Keep public surface small (`Client`, `Service`, business-oriented examples), push subjects/envelope/retry/dispatch into internal packages, and isolate transport behind an interface so unit tests run without NATS. Implement behavior via strict TDD: each component starts with failing tests, then minimal implementation.

**Tech Stack:** Go 1.22+, `github.com/nats-io/nats.go`, standard `testing` package.

---

### Task 1: Bootstrap Module and Public Contracts

**Files:**
- Create: `go.mod`
- Create: `sdk/types.go`
- Create: `sdk/interfaces.go`
- Create: `sdk/config.go`

**Step 1: Write the failing test**
- Create `sdk/public_contract_test.go` asserting core types (`Command`, `Event`, `Handler`, `RunConfig`) and interfaces compile and have expected fields/methods.

**Step 2: Run test to verify it fails**
- Run: `go test ./sdk -run TestPublicContracts -v`
- Expected: FAIL due to missing files/types.

**Step 3: Write minimal implementation**
- Add module and public types/interfaces/config with only fields/methods needed by tests.

**Step 4: Run test to verify it passes**
- Run: `go test ./sdk -run TestPublicContracts -v`
- Expected: PASS.

**Step 5: Commit**
- `git add go.mod sdk/types.go sdk/interfaces.go sdk/config.go sdk/public_contract_test.go`
- `git commit -m "feat: add sdk public contracts"`

### Task 2: Internal Subject Resolver

**Files:**
- Create: `sdk/internal/subject/resolver.go`
- Test: `sdk/internal/subject/resolver_test.go`

**Step 1: Write the failing test**
- Test inbox command mapping and event subject derivation from namespace.

**Step 2: Run test to verify it fails**
- Run: `go test ./sdk/internal/subject -v`
- Expected: FAIL due to missing resolver.

**Step 3: Write minimal implementation**
- Implement resolver with fixed inbox command map and event subject helpers.

**Step 4: Run test to verify it passes**
- Run: `go test ./sdk/internal/subject -v`
- Expected: PASS.

**Step 5: Commit**
- `git add sdk/internal/subject/resolver.go sdk/internal/subject/resolver_test.go`
- `git commit -m "feat: add internal subject resolver"`

### Task 3: Registry

**Files:**
- Create: `sdk/internal/registry/registry.go`
- Test: `sdk/internal/registry/registry_test.go`

**Step 1: Write the failing test**
- Verify register/get behavior and unknown type handling.

**Step 2: Run test to verify it fails**
- Run: `go test ./sdk/internal/registry -v`
- Expected: FAIL due to missing registry.

**Step 3: Write minimal implementation**
- Implement thread-safe handler registry.

**Step 4: Run test to verify it passes**
- Run: `go test ./sdk/internal/registry -v`
- Expected: PASS.

**Step 5: Commit**
- `git add sdk/internal/registry/registry.go sdk/internal/registry/registry_test.go`
- `git commit -m "feat: add handler registry"`

### Task 4: Envelope and Retry Policy

**Files:**
- Create: `sdk/internal/envelope/envelope.go`
- Create: `sdk/internal/retry/policy.go`
- Test: `sdk/internal/envelope/envelope_test.go`
- Test: `sdk/internal/retry/policy_test.go`

**Step 1: Write the failing test**
- Envelope auto-populates IDs/timestamps; retry policy enforces max attempt=3.

**Step 2: Run test to verify it fails**
- Run: `go test ./sdk/internal/envelope ./sdk/internal/retry -v`
- Expected: FAIL due to missing implementation.

**Step 3: Write minimal implementation**
- Implement envelope helpers and retry policy (`CanRetry`, `NextAttempt`).

**Step 4: Run test to verify it passes**
- Run: `go test ./sdk/internal/envelope ./sdk/internal/retry -v`
- Expected: PASS.

**Step 5: Commit**
- `git add sdk/internal/envelope/* sdk/internal/retry/*`
- `git commit -m "feat: add envelope and retry policy"`

### Task 5: Dispatcher

**Files:**
- Create: `sdk/internal/dispatcher/dispatcher.go`
- Test: `sdk/internal/dispatcher/dispatcher_test.go`

**Step 1: Write the failing test**
- Cover decode error, missing handler, handler success, handler error + retry trigger.

**Step 2: Run test to verify it fails**
- Run: `go test ./sdk/internal/dispatcher -v`
- Expected: FAIL due to missing dispatcher.

**Step 3: Write minimal implementation**
- Implement dispatch logic with injected retry publisher and logger callback.

**Step 4: Run test to verify it passes**
- Run: `go test ./sdk/internal/dispatcher -v`
- Expected: PASS.

**Step 5: Commit**
- `git add sdk/internal/dispatcher/*`
- `git commit -m "feat: add event dispatcher"`

### Task 6: NATS Transport

**Files:**
- Create: `sdk/internal/transport/nats/transport.go`
- Test: `sdk/internal/transport/nats/transport_test.go`

**Step 1: Write the failing test**
- Validate constructor argument checks and nil guards around request/publish/subscribe calls using function seams.

**Step 2: Run test to verify it fails**
- Run: `go test ./sdk/internal/transport/nats -v`
- Expected: FAIL.

**Step 3: Write minimal implementation**
- Implement transport struct wrapping `nats.Conn` and `nats.JetStreamContext`.

**Step 4: Run test to verify it passes**
- Run: `go test ./sdk/internal/transport/nats -v`
- Expected: PASS.

**Step 5: Commit**
- `git add sdk/internal/transport/nats/*`
- `git commit -m "feat: add nats transport"`

### Task 7: Client (Producer)

**Files:**
- Create: `sdk/client.go`
- Test: `sdk/client_test.go`

**Step 1: Write the failing test**
- Verify `SendCommand` uses resolver+request path and `PublishEvent` wraps payload with metadata and publishes to event subject.

**Step 2: Run test to verify it fails**
- Run: `go test ./sdk -run 'TestClient' -v`
- Expected: FAIL.

**Step 3: Write minimal implementation**
- Implement client with transport/resolver dependencies.

**Step 4: Run test to verify it passes**
- Run: `go test ./sdk -run 'TestClient' -v`
- Expected: PASS.

**Step 5: Commit**
- `git add sdk/client.go sdk/client_test.go`
- `git commit -m "feat: implement producer client"`

### Task 8: Service (Consumer)

**Files:**
- Create: `sdk/service.go`
- Test: `sdk/service_test.go`

**Step 1: Write the failing test**
- Verify handler registration and run loop wiring into transport subscription + dispatcher handling.

**Step 2: Run test to verify it fails**
- Run: `go test ./sdk -run 'TestService' -v`
- Expected: FAIL.

**Step 3: Write minimal implementation**
- Implement service with registry + dispatcher and transport consumer.

**Step 4: Run test to verify it passes**
- Run: `go test ./sdk -run 'TestService' -v`
- Expected: PASS.

**Step 5: Commit**
- `git add sdk/service.go sdk/service_test.go`
- `git commit -m "feat: implement consumer service"`

### Task 9: Business-Semantic Examples

**Files:**
- Create: `examples/producer_example_test.go`
- Create: `examples/consumer_example_test.go`

**Step 1: Write the failing test**
- Add compile/examples that demonstrate business methods only.

**Step 2: Run test to verify it fails**
- Run: `go test ./examples -v`
- Expected: FAIL due to missing symbols.

**Step 3: Write minimal implementation**
- Provide adapter/publisher/service example code with no inbox/subject/envelope leakage.

**Step 4: Run test to verify it passes**
- Run: `go test ./examples -v`
- Expected: PASS.

**Step 5: Commit**
- `git add examples/*`
- `git commit -m "docs: add business semantic sdk examples"`

### Task 10: Manual Integration Tests and Docs

**Files:**
- Create: `integration/integration_test.go`
- Create: `integration/README.md`
- Create: `Makefile`

**Step 1: Write the failing test**
- Add integration tests (build tag `integration`) that require running inbox docker-compose externally.

**Step 2: Run test to verify it fails**
- Run: `go test -tags integration ./integration -v`
- Expected: FAIL/SKIP until environment exists.

**Step 3: Write minimal implementation**
- Implement command/event integration cases and setup from env vars.

**Step 4: Run test to verify it passes (manual environment)**
- Run: `go test -tags integration ./integration -v`
- Expected: PASS when inbox stack is up.

**Step 5: Commit**
- `git add integration/* Makefile`
- `git commit -m "test: add manual integration tests and run targets"`

### Task 11: Final Verification

**Files:**
- Modify: `README.md` (if needed)

**Step 1: Run full unit test suite**
- Run: `go test ./...`
- Expected: all unit tests PASS; integration tests SKIP unless `-tags integration` used.

**Step 2: Run formatting/lint baseline**
- Run: `gofmt -w $(find . -name '*.go' -not -path './.git/*')`
- Run: `go test ./...`
- Expected: PASS after formatting.

**Step 3: Optional manual integration verification**
- Run: `make test-integration`
- Expected: PASS when inbox stack is running.

**Step 4: Commit final polish**
- `git add .`
- `git commit -m "chore: finalize sdk phase1 with tests and docs"`
