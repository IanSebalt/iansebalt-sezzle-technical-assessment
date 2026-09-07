# Roadmap — Sezzle Full-Stack Calculator

> **Status: complete.** All nine phases landed, each as its own commit. Figures and outcomes are in
> [`coverage.md`](coverage.md); the contract is in [`api.md`](api.md) and the reasoning in
> [`architecture.md`](architecture.md). The record below is kept as built, with the deferred
> decisions resolved at the bottom.

## Context

`docs/main-task/main-task.txt` asks for a full-stack calculator: a React (TypeScript) frontend
consuming a Go backend microservice over REST. The graded qualities are, in the spec's own words,
*"correctness, clarity, and maintainability over extra features"* — clean idiomatic code, unit
tests on both layers with a coverage report, and a README covering setup, API examples and design
rationale. Suggested effort ~2–4 hours.

The repository is empty: `docs/main-task/main-task.txt` and `.gitignore` only, **no commits yet**,
remote already set to `github.com/IanSebalt/iansebalt-sezzle-technical-assessment`.

Target outcome: a two-service repo a reviewer can clone, run in one command, and read end-to-end
without confusion — where every arithmetic result is computed by the Go service, every failure mode
surfaces as a sentence a human understands, and the folder layout makes the separation between
transport, domain and tests self-evident.

**Verified environment:** Go 1.27.1 · Node 24.14.0 / npm 11.9.0 · Docker 29.7.2 + Compose v5.5.0.
Go and Docker are installed but *not on `PATH`* — every invocation must be prefixed with
`export PATH="/usr/local/go/bin:$HOME/.docker/bin:$PATH"`.

---

## Decisions taken (locked before any code)

Each was an open question in the spec, decided by you rather than assumed.

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | Operation set | **All 7** — add, subtract, multiply, divide, power, sqrt, percentage | The optional three carry the interesting edge cases (√ of a negative, overflow to ±Inf) — what the error-handling requirement is really testing |
| 2 | API shape | **Single `POST /api/v1/calculate`** | One validation path, one handler, one client call site; adding an operation touches only the domain layer |
| 3 | UI model | **Classic keypad calculator** | Most intuitive to a reviewer opening the app |
| 4 | Go HTTP stack | **stdlib `net/http`** (1.22+ `ServeMux` method+path patterns) | Zero third-party deps; middleware is just `func(http.Handler) http.Handler` |
| 5 | Keypad semantics | **One binary operation per `=`** — operand → operator → operand; a second operator *replaces* the pending one; √ is unary and fires immediately | Keeps the frontend a thin client so every result is server-computed, and the reducer stays trivially testable |
| 6 | Frontend stack | **Vite + React + TS + CSS Modules** | Vite is the current standard (CRA deprecated); CSS Modules need no dependency and scope per component |
| 7 | Frontend tests | **Vitest + React Testing Library** | Shares Vite's transform pipeline — no second build config; v8 coverage built in |
| 8 | Go tests | **stdlib `testing` + `httptest`**, table-driven | Idiomatic; `go test -cover` gives the report with no extra tooling |
| 9 | Percentage | **Binary, `(a/100)*b`** — "a% of b" | Ambiguous term, pinned in the README so the evaluator is never guessing |
| 10 | Docker | **Two multi-stage images + Compose** — Go static binary; Vite build served by nginx proxying `/api` | Genuinely separate services, matching "backend microservice" |
| 11 | Numerics | **`float64` throughout**, non-finite results caught and reported as errors | Matches JSON/JS number semantics; √ and `^` are inherently floating-point. `0.1 + 0.2` imprecision disclosed in the README |
| 12 | Coverage | **Root `Makefile`** (`make test`, `make coverage`) + committed `docs/coverage.md` with real figures | The deliverable becomes an artifact, not a promise |

**Backend port:** default **9080**, env-configurable. Port 8080 on this machine has previously been
held by a foreign Docker container, which makes a local server look broken when it never bound.

**Commits:** format `type(module): message`, English, **no `Co-Authored-By` trailer** (your explicit
instruction overrides the session default).

---

## Folder tree

Dependency direction is one-way and enforced by review:
`cmd/server → httpapi → {calculator, apierror, config}`. `apierror` imports `net/http` for status
constants only. `calculator` imports `math`/`errors` only — the domain never knows about HTTP,
error codes, or user-facing copy.

```
/ (repo root)
├── .gitignore                     # + node_modules, dist, coverage, *.out, bin/
├── .editorconfig                  # shared indentation/EOL for Go + TS
├── Makefile                       # single entry point: test / coverage / run / docker
├── README.md                      # setup, run, API examples, design decisions, keyboard map
├── docker-compose.yml             # backend + nginx-served frontend
│
├── docs/
│   ├── main-task/main-task.txt    # (existing) original brief
│   ├── roadmap.md                 # this roadmap, committed, checkboxes ticked as phases land
│   ├── api.md                     # endpoint contract, curl per operation, full error catalogue
│   ├── architecture.md            # layering, error mapping, keypad state machine rationale
│   └── coverage.md                # real coverage numbers + how to reproduce
│
├── backend/
│   ├── go.mod                     # module github.com/iansebalt/sezzle-calculator, go 1.27, zero requires
│   ├── Dockerfile                 # multi-stage: golang builder → distroless/static:nonroot
│   ├── .dockerignore
│   ├── cmd/server/main.go         # config → router → http.Server → graceful shutdown
│   └── internal/
│       ├── config/
│       │   ├── config.go          # env parsing: PORT (default 9080), read/write/idle timeouts
│       │   └── config_test.go
│       ├── calculator/            # pure domain, zero transport knowledge
│       │   ├── operation.go       # Operation type, 7 constants, ParseOperation, Arity
│       │   ├── calculator.go      # Evaluate(op, a, b) dispatch + finite-result guard
│       │   ├── errors.go          # sentinels: ErrUnsupportedOperation, ErrDivisionByZero,
│       │   │                      #   ErrNegativeSquareRoot, ErrNonFiniteResult
│       │   ├── operation_test.go
│       │   └── calculator_test.go
│       ├── apierror/              # the wire error contract — single source of truth
│       │   ├── apierror.go        # type Error{Status int; Code, Message string}
│       │   ├── catalog.go         # one constructor per catalogue entry
│       │   └── apierror_test.go   # invariants: unique codes, non-empty copy, sane statuses
│       └── httpapi/               # HTTP transport only
│           ├── router.go          # "POST /api/v1/calculate" + 405/404 fallbacks + middleware
│           ├── calculate.go       # handler: decode → parse → evaluate → render
│           ├── dto.go             # request/response structs + decodeRequest
│           ├── errmap.go          # domain error → *apierror.Error   ← the ONLY mapping point
│           ├── render.go          # writeJSON / writeError (always the {"error":{…}} envelope)
│           ├── middleware.go      # recoverPanic → 500, logs via slog
│           ├── router_test.go     # package httpapi_test (black-box)
│           ├── calculate_test.go  # package httpapi_test
│           ├── dto_test.go        # package httpapi_test
│           └── errmap_test.go     # package httpapi (white-box, unexported mapError)
│
└── frontend/
    ├── package.json               # dev, build, preview, test, test:coverage, lint, typecheck
    ├── tsconfig.json              # strict: true, noUncheckedIndexedAccess
    ├── tsconfig.node.json
    ├── vite.config.ts             # React plugin, dev proxy /api → localhost:9080, vitest block
    ├── eslint.config.js
    ├── index.html
    ├── nginx.conf                 # SPA fallback + proxy_pass /api/ → backend:9080
    ├── Dockerfile                 # multi-stage: node builder → nginx:alpine
    ├── .dockerignore
    ├── public/favicon.svg
    └── src/
        ├── main.tsx
        ├── App.tsx                # page shell: header, <Calculator/>, hints footer
        ├── App.module.css
        ├── styles/
        │   ├── tokens.css         # CSS custom properties: color, spacing, radii, key sizes
        │   └── global.css         # reset, base type, focus-visible, prefers-color-scheme
        ├── components/            # one folder per component, styles beside it
        │   ├── Calculator/        # composes the hooks + Display/Keypad/ErrorBanner
        │   ├── Display/           # expression line + current value, aria-live="polite"
        │   ├── Keypad/            # renders the grid from keys.ts, forwards presses
        │   ├── Key/               # single <button>: label, aria-label, disabled/pressed
        │   └── ErrorBanner/       # role="alert" message area
        ├── hooks/
        │   ├── useCalculator.ts   # useReducer + effect on state.pending → API, abort/stale guard
        │   └── useKeyboard.ts     # window keydown → KeyId via keys.ts bindings
        ├── lib/
        │   ├── api/
        │   │   ├── types.ts       # CalculateRequest/Response, ApiErrorBody (mirrors the Go DTOs)
        │   │   ├── errors.ts      # ApiError class + network/parse fallbacks
        │   │   └── client.ts      # calculate(req, signal) → Response | throws ApiError
        │   └── calculator/
        │       ├── keypadTypes.ts # KeyId, Operator, State, Action, PendingRequest
        │       ├── keys.ts        # key layout + labels + keyboard bindings (single source)
        │       ├── keypadReducer.ts # PURE state machine — no React, no fetch, no arithmetic
        │       └── format.ts      # display formatting + digit-entry constraints
        └── tests/                 # all frontend tests, mirrored by layer
            ├── setup.ts           # RTL cleanup, jest-dom matchers, fetch stub helper
            ├── unit/              # keypadReducer, format, keys, client
            ├── hooks/             # useCalculator
            └── components/        # Display, Key, Keypad, Calculator
```

The API client targets the relative path `/api/v1/calculate` — same-origin in both dev (Vite proxy)
and prod (nginx), so there is no base-URL env var to configure.

### How a domain error becomes an HTTP response

1. `calculator.Evaluate` returns a sentinel (`ErrDivisionByZero`) — no status, no code, no copy.
2. `httpapi/errmap.go` holds the only translation table: `errors.Is(err, calculator.ErrDivisionByZero)`
   → `apierror.DivisionByZero()`. Unmatched errors → `apierror.Internal()` **and** get logged.
3. `render.go` serialises `{"error":{"code","message"}}` and writes the carried status.

Adding an operation touches `calculator` only; adding an error touches `catalog.go` plus one line
of `errmap.go`.

---

## Error catalogue

| Code | HTTP | User-facing message | Detected by |
|---|---|---|---|
| `METHOD_NOT_ALLOWED` | 405 | Only POST is supported for this endpoint. | `router.go` catch-all on the path; sets `Allow: POST` |
| `NOT_FOUND` | 404 | Endpoint not found. | `router.go` `/` fallback |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | Content-Type must be application/json. | `dto.go` |
| `PAYLOAD_TOO_LARGE` | 413 | Request body is too large. | `dto.go` — `http.MaxBytesReader` (4 KiB) |
| `MALFORMED_JSON` | 400 | Request body must be a valid JSON object. | `dto.go` — syntax error, EOF, trailing data |
| `UNKNOWN_FIELD` | 400 | Request contains an unrecognised field. Expected: operation, a, b. | `dto.go` — `DisallowUnknownFields` |
| `MISSING_OPERATION` | 400 | Field "operation" is required. | `dto.go` — nil `*string` |
| `INVALID_OPERATION` | 400 | Field "operation" must be text. | `dto.go` — added in Phase 2: a non-string `operation` is neither malformed JSON nor an unsupported operation |
| `INVALID_OPERAND` | 400 | Operands "a" and "b" must be finite numbers. | `dto.go` — `*json.UnmarshalTypeError` (covers `"ten"`, `true`, `1e400`) |
| `MISSING_OPERAND` | 400 | This operation requires two numbers (a and b). | `calculate.go` — nil `*float64` while arity is 2 |
| `UNSUPPORTED_OPERATION` | 400 | Unsupported operation. Use one of: add, subtract, multiply, divide, power, sqrt, percentage. | domain `ParseOperation`, mapped in `errmap.go` |
| `DIVISION_BY_ZERO` | 422 | Cannot divide by zero. | domain `Evaluate` |
| `NEGATIVE_SQUARE_ROOT` | 422 | Cannot take the square root of a negative number. | domain `Evaluate` |
| `NON_FINITE_RESULT` | 422 | The result is too large or undefined to represent. | domain finite guard (overflow, NaN) |
| `INTERNAL_ERROR` | 500 | Something went wrong. Please try again. | `middleware.go` recover + `errmap.go` default |

Status rationale (documented in `docs/api.md`): **400** = the request itself is wrong (shape, type,
unknown value); **422** = the request is well-formed but the arithmetic is impossible. The frontend
branches on `code`, never on status.

**Frontend-only messages** (no round-trip): *Cannot reach the calculator service. Check your
connection and try again.* (fetch rejects) · *Unexpected response from the server.* (non-JSON body,
e.g. an nginx 502).

**Input constraints shown as affordances, not errors:** `.` disabled once the entry has a decimal
point · `=` disabled until the expression is complete · all keys disabled while a request is in
flight · hint line *Maximum 15 digits.* when the entry cap is reached.

---

## Phased checklist

Each phase ends in a committable, verifiable state.

**Phase 0 — Workspace scaffold**
- [x] Extend `.gitignore`; add `.editorconfig`
- [x] `Makefile` with `help`, `test`, `coverage`, `run-backend`, `run-frontend`, `fmt`, `lint`
- [x] `README.md` skeleton; `docs/roadmap.md` (this plan) + `api.md` / `architecture.md` / `coverage.md` placeholders
- [x] Verify `make help`
- [x] `chore(repo): scaffold workspace, makefile and docs skeleton`

**Phase 1 — Backend domain**
- [x] `backend/go.mod` (no requires)
- [x] `internal/calculator/{operation,errors,calculator}.go`
- [x] Table-driven tests: all 7 operations + every domain error path
- [x] Verify `go test ./internal/calculator/... -cover` (target 100%)
- [x] `feat(calculator): add domain operations with table-driven tests`

**Phase 2 — Error contract + HTTP transport**
- [x] `internal/apierror/{apierror,catalog}.go` implementing the full catalogue
- [x] `internal/httpapi/{dto,errmap,render,calculate,middleware,router}.go`
- [x] Black-box `httptest` tests for every catalogue row + all 7 success shapes
- [x] Verify `go test ./... -cover`; every error code asserted at least once
- [x] `feat(api): add http transport and error catalogue for calculate endpoint`

**Phase 3 — Server wiring**
- [x] `internal/config` (+ test): `PORT` default 9080, timeouts
- [x] `cmd/server/main.go`: slog, server timeouts, `signal.NotifyContext` graceful shutdown
- [x] Verify `make run-backend` + curl every documented example (success, 422, 405, 415, 413)
- [x] `feat(server): wire configurable http server with graceful shutdown`

**Phase 4 — Frontend scaffold + API client**
- [x] Vite React-TS app, strict tsconfig, CSS Modules, `styles/{tokens,global}.css`
- [x] Vitest + RTL config, `src/tests/setup.ts`, dev proxy `/api → localhost:9080`
- [x] `lib/api/{types,errors,client}.ts` + client tests
- [x] Verify `npm run typecheck && npm test`
- [x] `feat(web): scaffold vite app and typed calculate api client`

**Phase 5 — Keypad state machine + hooks**
- [x] `lib/calculator/{keypadTypes,keys,keypadReducer,format}.ts` — pure, no React, no fetch
- [x] `hooks/useCalculator.ts` (reducer + effect on `pending`, AbortController + stale-response guard)
- [x] `hooks/useKeyboard.ts` driven by the same `keys.ts` bindings
- [x] Verify reducer/format/hook tests green, reducer coverage ≥ 95%
- [x] `feat(web): add keypad state machine and calculator hook`

**Phase 6 — UI components**
- [x] `Key`, `Keypad`, `Display`, `ErrorBanner`, `Calculator`, `App` + CSS Modules
- [x] Responsive grid, ≥44px touch targets, `aria-live` / `role="alert"`, visible focus
- [x] Hints row: *√ applies to the number shown* · *a % b = a% of b*
- [x] Component + integration tests with stubbed `fetch`
- [x] Verify `npm test`, manual mobile-width check, full keyboard-only run
- [x] `feat(web): add keypad calculator ui with accessible error messaging`

**Phase 7 — Docker**
- [x] `backend/Dockerfile` (distroless static, nonroot, `CGO_ENABLED=0 -trimpath -ldflags="-s -w"`)
- [x] `frontend/{Dockerfile,nginx.conf}` (SPA fallback + `/api/` → `backend:9080`)
- [x] `docker-compose.yml`, `.dockerignore` per context
- [x] Verify `docker compose up --build`, exercise the UI and a direct curl
- [x] `chore(docker): add multi-stage images and compose stack`

**Phase 8 — Coverage + documentation**
- [x] `make coverage` → `go tool cover -func` totals + Vitest v8 summary
- [x] `docs/coverage.md` with real per-package numbers, date, reproduction commands
- [x] `docs/api.md` (contract + curl per operation + catalogue); `docs/architecture.md` (layering, error mapping, state machine)
- [x] `README.md`: quickstart (local + docker), API examples, keyboard map, design decisions, explicit non-goals
- [x] Tick every box in `docs/roadmap.md`
- [x] `docs(repo): add api reference, architecture notes and coverage report`

---

## Tests to write

**Domain (`internal/calculator`, table-driven)** — `TestParseOperation` (7 valid, unknown, empty,
wrong case) · `TestOperationArity` · `TestEvaluate_Arithmetic` (all 7, integers/floats/negatives/
identities) · `TestEvaluate_Percentage` · `TestEvaluate_PowerEdgeCases` (`0^0=1`, negative and
fractional exponents) · `TestEvaluate_DivisionByZero` (`b=0`, `b=-0.0`, `0/0`) ·
`TestEvaluate_NegativeSquareRoot` (plus `sqrt(0)` succeeds) · `TestEvaluate_NonFiniteResult`
(`1e308*10`, `power(10,400)`, `power(-8,1/3)` NaN) · `TestEvaluate_UnaryIgnoresB` ·
`TestEvaluate_UnsupportedOperation`

**Transport (`httptest`)** — `TestCalculate_Success` (7 ops, exact JSON, `b` omitted for sqrt) ·
`_MalformedJSON` · `_TrailingDataAfterJSON` · `_UnknownField` · `_WrongContentType` ·
`_MissingOperation` · `_UnsupportedOperation` · `_MissingOperandB` · `_NonNumericOperand` ·
`_OutOfRangeOperandLiteral` (`1e400`) · `_DivisionByZero` · `_NegativeSquareRoot` ·
`_OverflowResult` · `_BodyTooLarge` · `TestErrorEnvelopeShape` (every failure returns the envelope
+ JSON content type) · `TestRouter_MethodNotAllowed` (GET/PUT/DELETE → 405 + `Allow`) ·
`TestRouter_NotFound` · `TestRecoverPanic_Returns500` · `TestMapError` (white-box sentinel table) ·
`TestConfig_LoadDefaultsAndOverrides`

**Frontend hook** — posts once on `=` · exposes loading state · surfaces the server message for
`DIVISION_BY_ZERO` · network fallback message when fetch rejects · clears a previous error on the
next success · ignores a stale response after a newer request resolves first · fires immediately
on `√`

**Frontend units** — *keypadReducer*: appends digits and collapses leading zeros · only one decimal
point · caps entry length and flags the constraint · stores operand + pending operator · **replaces
the pending operator when a second is pressed before `=`** · builds a pending request on `=` ·
ignores `=` on an incomplete expression · reuses the result as the next operand · starts a fresh
entry when a digit follows a result · `C` resets · **never produces a numeric result on its own**
(guards the no-local-math rule). *format*: integers without trailing decimal · trims floating-point
noise · exponential for extreme magnitudes · renders `-0` as `0`. *keys*: unique id/label/accessible
name · every keyboard binding maps to a real key. *client*: sends POST with JSON content type ·
returns parsed payload · throws `ApiError` carrying server code+message · parse fallback for a
non-JSON error body · propagates abort silently.

**Frontend components** — *Display*: current entry · pending expression · announces via `aria-live`.
*Key*: calls `onPress` with its id · disabled while busy. *Keypad*: renders every key with its
accessible name · disables `.` when the entry already has one. *Calculator* (integration, stubbed
fetch): `10 ÷ 4 =` shows 2.5 from the API · division-by-zero message in an alert · typing digits and
pressing Enter submits · Escape clears · √ of a negative shows the domain error.

---

## Deferred decisions — resolved

Each was left open deliberately and settled at the point the roadmap named.

| Decision | Resolved as | Where |
|---|---|---|
| `GET /api/v1/health` endpoint | **Omitted.** No compose healthcheck was added, so nothing needed it; distroless carries no shell to probe with | Phase 7 |
| Per-request access logging | **5xx and panics only.** Local debugging never needed more | Phase 3 |
| Coverage thresholds gating `make coverage` | **Report only, no gate.** The figures are committed instead | Phase 8 |
| Display precision policy | **12 significant digits, trailing zeros trimmed.** The explicit exponential thresholds were *deleted* — `toPrecision(12)` already switches notation at exactly the right magnitudes, so they were duplicated logic | Phase 5 |
| MSW vs. stubbed `fetch` | **Stubbed `fetch`**, plus a `stubBackend` helper that replies the way the Go service does. No extra dependency, and component tests stayed readable | Phase 6 |
| `UNKNOWN_FIELD` as its own code | **Kept**, and `INVALID_OPERATION` was added alongside it. Folding them into `MALFORMED_JSON` would have told a user with a typo'd field name nothing useful | Phase 2 |
| Dark mode depth | **`prefers-color-scheme` tokens only.** Verified in both themes in a real browser | Phase 6 |
| Published compose host ports | **web `8081:80`, backend `9080:9080`** — both confirmed free | Phase 7 |
| `b: null` vs. omitted for `sqrt` | **Omitted** via `omitempty`, with a test asserting the key is absent rather than null | Phase 2 |

### Decided during implementation, beyond the original plan

| Decision | Outcome |
|---|---|
| Linter | The Vite template ships **oxlint**, not ESLint. Kept it — one less dependency and it is the template default, so `eslint.config.js` from the plan does not exist |
| A whole-body JSON type mismatch | A body like `[1,2]` produced `Field "" must be a number…`. Now reported as `MALFORMED_JSON`, which is what it actually is |
| Long results overflowing the display | `99^99` clipped its own exponent. The display now steps its type size down by value length (`data-size` attribute, asserted in a test) |
| `cmd/server` test coverage | The plan left the entrypoint untested. It now has a test that boots the real server on an ephemeral port, calls it, sends `SIGTERM` and asserts a clean shutdown |
| A downed backend behind a proxy | Found by killing the backend with the page open: Vite (and nginx) answer **502**, so the browser gets a response rather than a failed fetch, and the user was told "Unexpected response from the server." 502/503/504 now map to "Cannot reach the calculator service." |
| How long a dead backend takes to report | With default timeouts nginx took **39s** to return its 502, leaving the keypad disabled throughout. `proxy_connect_timeout 5s` plus a 10s client deadline bring it to ~5s, and the keypad recovers |

**Explicit non-goals** (documented in the README, not deferred): CORS middleware (dev uses the Vite
proxy, prod is same-origin via nginx), auth, rate limiting, persistence or history, E2E tests,
arbitrary-precision arithmetic.

## Verification

Per phase, as listed above. End-to-end before the final commit:

1. `make test` — Go and Vitest suites green.
2. `make coverage` — real figures, transcribed into `docs/coverage.md`.
3. `make run-backend` (port 9080) + `make run-frontend`; in the browser: compute `10 ÷ 4`, then
   `10 ÷ 0`, `√ -9`, `1e308 × 10` and confirm each renders its catalogue message, not a raw code.
4. `curl` every example in the README verbatim and diff the responses against the documented ones.
5. Stop the backend and press `=` — confirm the network fallback message appears.
6. Keyboard-only run (digits, operators, Enter, Escape) and a 375px-wide viewport check.
7. `docker compose up --build` — repeat steps 3 and 5 against the containerised stack.
