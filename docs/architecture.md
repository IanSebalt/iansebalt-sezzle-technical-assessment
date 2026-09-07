# Architecture Notes

Two services, each with one job: a Go microservice that owns all arithmetic, and a React SPA that
owns interaction. The interesting decisions are about where responsibilities *stop*.

## Backend layering

```
cmd/server  ──▶  httpapi  ──▶  calculator   (domain)
                    │      ──▶  apierror     (wire error contract)
                    │      ──▶  config       (environment)
```

Dependencies point one way only:

- **`calculator`** imports `math` and `errors`. Nothing else. It does not know that HTTP exists, has
  never heard of a status code, and contains no user-facing copy. Its failures are four sentinel
  errors: `ErrUnsupportedOperation`, `ErrDivisionByZero`, `ErrNegativeSquareRoot`,
  `ErrNonFiniteResult`.
- **`apierror`** owns the wire contract: a code, a status, and the sentence a person reads. It
  imports `net/http` for status constants and nothing from the domain — `UnsupportedOperation`
  receives the list of allowed names as an argument rather than importing `calculator`, which would
  have reversed the arrow.
- **`httpapi`** is the only package that knows both. It decodes, validates the envelope, calls the
  domain, and renders.
- **`cmd/server`** reads config, builds the router, and manages the process lifecycle.

### How a domain error becomes an HTTP response

1. `calculator.Evaluate` returns a sentinel. No status, no code, no copy.
2. `httpapi/errmap.go` holds the *only* translation table in the codebase:
   `errors.Is(err, calculator.ErrDivisionByZero)` → `apierror.DivisionByZero()`. Anything
   unrecognised becomes `apierror.Internal()` and is logged — an unexpected error never reaches a
   user with its details attached.
3. `httpapi/render.go` writes the carried status and the `{"error":{…}}` envelope.

The payoff is that adding an operation touches `calculator` alone, and adding an error touches
`apierror/catalog.go` plus one line of `errmap.go`. `errmap_test.go` includes a test asserting that
an unrecognised error's text does not leak into the response.

### Why the domain guards against non-finite results

`float64` overflows to `±Inf` and produces `NaN` silently — `1e308 * 10`, `power(10, 400)`,
`power(-8, 1/3)`. Neither value can be encoded in JSON, so `Evaluate` checks the result and returns
`ErrNonFiniteResult` rather than letting the encoder emit an invalid body. The check lives in the
domain because it is a fact about arithmetic, not about transport.

### Pointers in the request DTO

`calculateRequest` uses `*string` and `*float64` so that "field absent" and "field sent as zero" are
distinguishable. `{"operation":"add","a":0}` is a missing operand; `{"operation":"add","a":0,"b":0}`
is a valid sum. A plain `float64` could not tell the two apart.

## Frontend layering

```
components/  ──▶  hooks/  ──▶  lib/api        (fetch + typed contract)
                          ──▶  lib/calculator (pure state machine)
```

- **`lib/calculator`** is pure TypeScript: no React, no `fetch`, and — deliberately — **no
  arithmetic**. `keypadReducer` is a state machine over one binary operation; the only numbers it
  ever writes to the display are digits the user typed and results the server returned. A unit test
  (`never computes a result on its own`) locks that property in place.
- **`lib/api`** owns the network. Every failure leaves it as an `ApiError` carrying a message fit to
  show a user, so no component ever interprets a status code. Aborts are the single exception: they
  are the caller's own doing and propagate untouched.
  Two failures the server never gets to describe are worded by the client itself: a rejected `fetch`
  and a gateway status (502/503/504) both become *"Cannot reach the calculator service."* The
  gateway case matters because a proxy sits in front of the API in both environments — with the
  backend down, the browser receives a 502 from Vite or nginx rather than a failed request, so
  treating it as an unreadable response would have told the user the wrong thing. Anything else
  unparseable becomes *"Unexpected response from the server."* The client also gives up after 10
  seconds, and nginx after 5 — left at their defaults, nginx spends about 40 seconds retrying a
  backend that is simply not there, and the keypad stays disabled for all of it.
- **`hooks/useCalculator`** joins the two: it runs the reducer, and when the reducer produces a
  `pending` request it performs the call and dispatches the answer back.
- **`components/`** render state and forward presses. `Calculator` is the only component that holds
  policy (which keys are disabled and why); `Keypad` receives a predicate, and `Key` receives a
  boolean.

### The keypad state machine

The UI computes nothing, so the reducer never evaluates — it *builds a request*:

```
digits      ──▶ entry, entryValue := null      (the typed digits are the value)
operator    ──▶ operandA := currentOperand(), operation := operator, await the second operand
digits      ──▶ entry, entryValue := null
'='         ──▶ pending = { operation, a: operandA, b: currentOperand() }
                                    │
                          useCalculator performs the call
                                    │
resolved    ──▶ entry := formatted result, entryValue := the exact result
```

`currentOperand()` reads `entryValue` when the number on screen came from the server, and falls
back to parsing `entry` while the user is typing. That split matters: `entry` is rounded to 12
significant digits for display, so re-parsing it would feed that rounding into the next request and
make `10 ÷ 3 = × 3 =` return `9.99999999999`.

Three rules make it predictable:

- **The operator can be changed at any time before `=`.** `12 + − 3 =` is `12 − 3`, and so is
  `12 + 3` followed by `−`. One uniform rule, nothing silently discarded.
- **`√` applies to the number on screen and leaves a half-built expression intact.** `9 + 16 √ =`
  computes `√16` on the server, shows `4`, and then `9 + 4`. Still one binary operation per `=`.
- **Presses are ignored while a request is in flight, except `Clear`,** which also aborts it.

Each pending request carries an incrementing id. The reducer ignores any answer whose id is not the
one it is waiting for, so a slow reply that lands after the user moved on cannot overwrite the
display. The effect additionally aborts the outgoing request on cleanup — belt and braces, because
only the id check is a guarantee.

### Why keys are disabled instead of validated

The invalid states a keypad can reach are knowable in advance, so the UI removes them rather than
letting a user reach them and then complaining: `=` is disabled until the expression is complete,
`.` is disabled once the number already has one, and every key but `Clear` is disabled mid-request.
The one constraint that cannot be expressed as a disabled key — the 15-digit entry cap — is stated
in words when it is reached. Errors are reserved for things only the server can decide.

## No CORS anywhere

In development, Vite proxies `/api` to `localhost:9080`. In the container, nginx proxies `/api/` to
`backend:9080`. The browser only ever sees a same-origin request, so there is no CORS middleware to
write, configure, or get wrong. The API client uses the relative path `/api/v1/calculate` and needs
no base-URL configuration in either environment.

## Deliberate non-goals

Not built, because the brief does not describe them and speculative structure is harder to remove
than to add: authentication, rate limiting, request-history persistence, a health endpoint,
end-to-end browser tests, arbitrary-precision arithmetic, and expression parsing with operator
precedence.
