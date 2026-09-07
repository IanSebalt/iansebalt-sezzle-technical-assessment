# Sezzle Calculator

A full-stack calculator: a **React + TypeScript** keypad frontend consuming a **Go** backend
microservice over REST. Every arithmetic result is computed server-side — the frontend holds
keypad state but never does the maths itself.

Built against the brief in [`docs/main-task/main-task.txt`](docs/main-task/main-task.txt).

> **Status:** in progress. Setup instructions, API examples and the coverage report land as the
> corresponding phases complete — see [`docs/roadmap.md`](docs/roadmap.md) for the checklist.

## Operations

| Operation | Operands | Definition |
|---|---|---|
| `add` | a, b | a + b |
| `subtract` | a, b | a − b |
| `multiply` | a, b | a × b |
| `divide` | a, b | a ÷ b — rejects b = 0 |
| `power` | a, b | a ^ b |
| `sqrt` | a | √a — rejects a < 0; `b` is not accepted |
| `percentage` | a, b | (a ÷ 100) × b, i.e. **"a% of b"** |

`percentage` has no universally agreed meaning, so it is pinned explicitly: `15 % 200 = 30`.

## Repository layout

```
backend/    Go microservice — domain, HTTP transport, config (see docs/architecture.md)
frontend/   Vite + React + TypeScript SPA
docs/       Roadmap, API reference, architecture notes, coverage report
Makefile    Single entry point for tests, coverage, local runs and Docker
```

## Documentation

- [`docs/roadmap.md`](docs/roadmap.md) — phased plan, decisions taken, decisions deferred
- [`docs/api.md`](docs/api.md) — endpoint contract, curl examples, full error catalogue
- [`docs/architecture.md`](docs/architecture.md) — layering and error-mapping rationale
- [`docs/coverage.md`](docs/coverage.md) — coverage figures and how to reproduce them
