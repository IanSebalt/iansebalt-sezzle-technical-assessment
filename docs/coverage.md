# Coverage Report

Figures below are from an actual run on **2026-09-07**, after the post-implementation review (Go 1.27.1, Node 24.14.0, Vitest 5).

Reproduce everything with:

```bash
make coverage
```

## Backend — `go test ./... -coverprofile=coverage.out`

| Package | Statement coverage |
|---|---|
| `internal/calculator` | **100.0%** |
| `internal/apierror` | **100.0%** |
| `internal/httpapi` | **100.0%** |
| `internal/config` | **100.0%** |
| `cmd/server` | 84.2% |
| **Total** | **97.4%** |

55 test functions, most of them table-driven, so the number of assertions is considerably higher.

The only uncovered function is `main()` in `cmd/server/main.go`: it constructs the logger and calls
`os.Exit`, which cannot run inside a test process. The logic it delegates to — `run()` — is covered,
including a test that starts the real server on an ephemeral port, performs an HTTP request against
it, sends `SIGTERM`, and asserts a clean graceful shutdown.

Every error code in the catalogue is asserted by at least one test.

```bash
make coverage-backend            # per-function breakdown
cd backend && go tool cover -html=coverage.out    # annotated source
```

## Frontend — `vitest run --coverage` (v8)

| Metric | Coverage |
|---|---|
| Statements | **100%** (164/164) |
| Functions | **100%** (53/53) |
| Lines | **100%** (133/133) |
| Branches | **99.22%** (128/129) |

105 tests across 11 files: pure units for the reducer, formatting and key definitions; the API
client; both hooks; and component tests up to a full integration test that drives the real component
tree against a stub shaped like the Go service.

The single uncovered branch is `useCalculator.ts:37` — the fallback message for a thrown value that
is not an `ApiError`. The API client converts every failure into an `ApiError` before it reaches the
hook, so the branch is unreachable in practice. It is kept as a guard rather than deleted, on the
grounds that an error path should not itself be able to throw.

```bash
cd frontend && npm run test:coverage
```

## What is not covered by tests

CSS, the Vite and nginx configuration, and the Dockerfiles carry no unit tests. These were verified
by running the stack: `docker compose up --build`, then exercising the UI in a browser at desktop
and 375px widths, in light and dark themes, by mouse and by keyboard alone, including the
division-by-zero, negative-square-root and overflow paths, and with the backend container stopped
mid-session to confirm the unreachable-service message and that the keypad recovers.
