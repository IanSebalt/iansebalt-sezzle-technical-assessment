# Sezzle Calculator

A full-stack calculator: a **React + TypeScript** keypad frontend consuming a **Go** backend
microservice over REST.

Every arithmetic result is computed by the Go service. The frontend holds keypad state — which digits
you typed, which operator is pending — but never does the maths itself, and a unit test enforces
that. Built against the brief in [`docs/main-task/main-task.txt`](docs/main-task/main-task.txt).

```
┌──────────────────────┐        POST /api/v1/calculate        ┌─────────────────────┐
│  React SPA           │  ──────────────────────────────────▶ │  Go microservice    │
│  keypad + validation │  ◀────────────────────────────────── │  arithmetic + rules │
└──────────────────────┘        result, or a stated error     └─────────────────────┘
```

## Requirements

| Tool | Version used | Needed for |
|---|---|---|
| Go | 1.22 or newer (1.27.1 used here) | backend |
| Node.js | 24.14.0 (npm 11.9.0) | frontend |
| Docker + Compose | 29.7.2 / v5.5.0 | optional containerised run |

Go ≥ 1.22 is required — the router uses `net/http`'s method-and-path patterns.

## Quick start

### With Docker (both services together)

```bash
docker compose up --build
```

Then open **http://localhost:8081**. The API is also published directly on
**http://localhost:9080** if you want to `curl` it.

### Locally, without Docker

Two terminals:

```bash
# Terminal 1 — backend on :9080
make run-backend

# Terminal 2 — frontend on :5173
cd frontend && npm install     # first time only
make run-frontend
```

Then open **http://localhost:5173**. Vite proxies `/api` to the backend, so the browser only ever
makes same-origin requests and there is no CORS setup to do.

Port 9080 already taken? `make run-backend PORT=9090`, and adjust `backendUrl` in
`frontend/vite.config.ts` (or set `BACKEND_URL=http://localhost:9090`).

## Everything you can run

The frontend targets need dependencies installed first (`cd frontend && npm install`).

```bash
make help          # list every target
make test          # Go + Vitest suites
make coverage      # coverage for both layers
make lint          # go vet, TypeScript typecheck, oxlint
make fmt           # gofmt
make up / make down    # Docker Compose stack
```

## Using the calculator

| Operation | Operands | Definition |
|---|---|---|
| `add` | a, b | a + b |
| `subtract` | a, b | a − b |
| `multiply` | a, b | a × b |
| `divide` | a, b | a ÷ b — rejects b = 0 |
| `power` | a, b | a ^ b |
| `sqrt` | a | √a — rejects a < 0; `b` is ignored if sent |
| `percentage` | a, b | (a × b) ÷ 100, i.e. **"a% of b"** — `15 % 200 = 30` |

The keypad evaluates **one binary operation per `=`**: type a number, choose an operator, type the
second number, press `=`. The operator can be changed at any point before `=`. `√` applies to the
number currently shown and leaves a half-built expression intact, so `9 + 16 √ =` gives 13.

The whole app is usable by keyboard alone:

| Key | Does |
|---|---|
| `0`–`9`, `.` | Enter a number |
| `+`, `-`, `*` (or `x`), `/` | Add, subtract, multiply, divide |
| `^` | Exponentiation |
| `%` | Percentage |
| `r` | Square root |
| `Enter` or `=` | Calculate |
| `Backspace` | Delete the last digit |
| `Esc` or `c` | Clear |

Letter keys are case-insensitive.

## API

Full contract, every error code, and copy-pasteable examples: **[`docs/api.md`](docs/api.md)**.

```bash
# 10 ÷ 4
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"divide","a":10,"b":4}'
# {"operation":"divide","a":10,"b":4,"result":2.5}

# 15% of 200
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"percentage","a":15,"b":200}'
# {"operation":"percentage","a":15,"b":200,"result":30}

# √81 — unary, so no "b" is sent or returned
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","a":81}'
# {"operation":"sqrt","a":81,"result":9}

# Division by zero
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"divide","a":10,"b":0}'
# {"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero."}}
```

Every failure uses the same envelope — `{"error":{"code","message"}}` — where `code` is stable for
branching and `message` is a finished sentence the UI shows verbatim.

## Tests and coverage

```bash
make test        # 50 Go test functions, 89 frontend tests
make coverage
```

Backend **97.2%** of statements (100% in every package except the `main()` entrypoint);
frontend **100%** statements / functions / lines, 99.13% branches. Figures, the exact gaps, and why
they are gaps: **[`docs/coverage.md`](docs/coverage.md)**.

## Design decisions

Architectural reasoning in full: **[`docs/architecture.md`](docs/architecture.md)**. The decisions
that most shape the code:

**One endpoint, not seven.** `POST /api/v1/calculate` takes the operation as a field, so there is one
decode path, one validation path and one client call site. Adding an operation touches the domain
package and nothing else.

**The domain knows nothing about HTTP.** `internal/calculator` imports only `math` and `errors`; it
returns sentinel errors carrying no status, code, or user-facing text. A single file,
`internal/httpapi/errmap.go`, translates those into the wire contract. Unrecognised errors become a
generic 500 and are logged rather than leaked — there is a test for that.

**All copy lives in one place.** `internal/apierror/catalog.go` holds every code, status and message.
A test asserts the codes are unique and every message is a properly punctuated sentence, so the API
cannot drift into showing users a bare error code.

**The frontend never computes.** The keypad reducer is a pure state machine that builds a *request*;
the only numbers it writes to the display are typed digits and server results. This is what the
assessment is actually testing, so it is enforced by a test rather than by convention.

**Invalid states are unreachable, not merely rejected.** `=` is disabled until the expression is
complete, `.` is disabled once the number has one, and every key but `Clear` is disabled while a
request is in flight. Errors are reserved for the things only the server can decide.

**No CORS, ever.** Vite proxies `/api` in development and nginx proxies it in the container, so the
browser only makes same-origin requests. The client uses a relative path and needs no configuration.

**Zero third-party Go dependencies.** `go.mod` has no `require` block. Go 1.22's `ServeMux` covers
the routing, `net/http/httptest` covers the tests, and `log/slog` covers logging.

**`float64` throughout, with the caveat stated.** It matches JSON's only numeric type, and `sqrt` and
non-integer `power` are inherently floating-point. Results are rendered to 12 significant digits, so
`0.1 + 0.2` displays as `0.3`. Exact decimal arithmetic was rejected because it could not have
covered `sqrt` or `power` and the guarantee would have been misleading. Non-finite outcomes
(overflow to `±Inf`, `NaN`) are caught and reported as errors rather than encoded as invalid JSON.

## Assumptions

These were not specified in the brief and were decided deliberately:

- **`percentage` means "a% of b"** — `(a/100)*b`. The term is genuinely ambiguous, so the definition
  is pinned here, in the API docs, and in a hint under the keypad.
- **The keypad does one binary operation per `=`.** No chaining and no operator precedence, so
  `2 + 3 × 4` is not an expression the UI can build. This keeps every result server-computed and the
  reducer small; an expression parser would have meant either a new endpoint or arithmetic in the
  browser.
- **Operands are entered as non-negative numbers.** There is no `+/−` key. Negative values arise as
  results (`4 − 9`) and can then be fed into the next operation, which is how `√` of a negative is
  reachable.
- **Entry is capped at 15 digits**, and the cap is stated on screen when reached rather than
  silently swallowing the keypress.
- **`sqrt` ignores a `b` it was sent** instead of rejecting the request, and omits `b` from its
  response.
- **The request body is capped at 4 KiB** and unknown JSON fields are rejected, so `"opperation"` is
  reported rather than read as a missing field.

## Repository layout

```
backend/
  cmd/server/          entrypoint: config, router, graceful shutdown
  internal/calculator/ pure arithmetic domain — imports only math and errors
  internal/apierror/   the wire error contract: code, status, message
  internal/httpapi/    HTTP transport: decode, validate, map errors, render
  internal/config/     environment parsing
frontend/
  src/components/      one folder per component, styles beside it
  src/hooks/           useCalculator (state + network), useKeyboard
  src/lib/api/         typed client and error types
  src/lib/calculator/  pure keypad state machine, key layout, formatting
  src/tests/           all tests, mirrored by layer
docs/                  roadmap, API reference, architecture notes, coverage
Makefile               single entry point for tests, coverage, runs and Docker
docker-compose.yml     backend + nginx-served frontend
```

## Documentation

- [`docs/api.md`](docs/api.md) — endpoint contract, every error code, curl examples
- [`docs/architecture.md`](docs/architecture.md) — layering, error mapping, the keypad state machine
- [`docs/coverage.md`](docs/coverage.md) — coverage figures and how to reproduce them
- [`docs/roadmap.md`](docs/roadmap.md) — the plan this was built to, decisions taken and deferred
