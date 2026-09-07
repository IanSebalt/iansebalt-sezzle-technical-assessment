# API Reference

The calculator exposes a single endpoint. Every arithmetic result the UI shows comes from it.

```
POST /api/v1/calculate
Content-Type: application/json
```

Base URL in local development: `http://localhost:9080` (override with `PORT`).
Through the Docker stack: `http://localhost:8081/api/v1/calculate`, proxied same-origin by nginx.

## Request

| Field | Type | Required | Notes |
|---|---|---|---|
| `operation` | string | yes | One of the operations below |
| `a` | number | yes | First operand |
| `b` | number | for binary operations | Ignored by `sqrt` |

Unknown fields are rejected rather than silently dropped, so a typo like `"opperation"` is reported
instead of being read as a missing field. The body is capped at 4 KiB.

## Operations

| `operation` | Operands | Definition | Rejects |
|---|---|---|---|
| `add` | a, b | a + b | — |
| `subtract` | a, b | a − b | — |
| `multiply` | a, b | a × b | — |
| `divide` | a, b | a ÷ b | b = 0 |
| `power` | a, b | a ^ b | — |
| `sqrt` | a | √a | a < 0 |
| `percentage` | a, b | (a ÷ 100) × b — **"a% of b"** | — |

Every operation additionally rejects a result that is not a finite number (overflow to ±∞, or an
undefined result such as `power(-8, 1/3)`), because neither is expressible in JSON.

`percentage` is defined explicitly because the word has no single agreed meaning:
`15 % 200 = 30`, read as *15% of 200*.

## Response

```json
{ "operation": "divide", "a": 10, "b": 4, "result": 2.5 }
```

`b` is omitted for unary operations:

```json
{ "operation": "sqrt", "a": 81, "result": 9 }
```

## Errors

Every failure — without exception — uses one envelope:

```json
{ "error": { "code": "DIVISION_BY_ZERO", "message": "Cannot divide by zero." } }
```

`code` is stable and meant for branching. `message` is a complete sentence written for a person and
is what the UI displays verbatim. Clients should branch on `code`, never on the status.

| Code | HTTP | Message |
|---|---|---|
| `METHOD_NOT_ALLOWED` | 405 | Only POST is supported for this endpoint. |
| `NOT_FOUND` | 404 | Endpoint not found. |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | Content-Type must be application/json. |
| `PAYLOAD_TOO_LARGE` | 413 | Request body is too large. |
| `MALFORMED_JSON` | 400 | Request body must be a valid JSON object. |
| `UNKNOWN_FIELD` | 400 | Request contains an unrecognised field. Expected: "operation", "a" and "b". |
| `MISSING_OPERATION` | 400 | Field "operation" is required. |
| `INVALID_OPERATION` | 400 | Field "operation" must be text. |
| `INVALID_OPERAND` | 400 | Field "a" must be a number the calculator can represent. |
| `MISSING_OPERAND` | 400 | Field "b" is required for this operation. |
| `UNSUPPORTED_OPERATION` | 400 | Unsupported operation. Use one of: add, subtract, multiply, divide, power, sqrt, percentage. |
| `DIVISION_BY_ZERO` | 422 | Cannot divide by zero. |
| `NEGATIVE_SQUARE_ROOT` | 422 | Cannot take the square root of a negative number. |
| `NON_FINITE_RESULT` | 422 | The result is too large or undefined to represent. |
| `INTERNAL_ERROR` | 500 | Something went wrong. Please try again. |

`INVALID_OPERAND` and `MISSING_OPERAND` name whichever field was at fault.

**Why 400 and 422 both appear:** 400 means the request itself is wrong — bad shape, wrong type, an
operation that does not exist. 422 means the request was understood perfectly and the arithmetic is
simply impossible. A client can retry a 422 with different numbers; a 400 needs a different request.

A 405 also carries `Allow: POST`.

## Examples

Every command below was run against the service and the responses are copied verbatim.

```bash
# Addition
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"add","a":2,"b":3}'
# {"operation":"add","a":2,"b":3,"result":5}

# Division producing a fraction
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"divide","a":10,"b":4}'
# {"operation":"divide","a":10,"b":4,"result":2.5}

# Exponentiation
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"power","a":2,"b":10}'
# {"operation":"power","a":2,"b":10,"result":1024}

# Square root — unary, so no "b" is sent or returned
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","a":81}'
# {"operation":"sqrt","a":81,"result":9}

# Percentage — 15% of 200
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"percentage","a":15,"b":200}'
# {"operation":"percentage","a":15,"b":200,"result":30}
```

Failure cases:

```bash
# Division by zero -> 422
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"divide","a":10,"b":0}'
# {"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero."}}

# Square root of a negative -> 422
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"sqrt","a":-9}'
# {"error":{"code":"NEGATIVE_SQUARE_ROOT","message":"Cannot take the square root of a negative number."}}

# Overflow -> 422
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"multiply","a":1e308,"b":10}'
# {"error":{"code":"NON_FINITE_RESULT","message":"The result is too large or undefined to represent."}}

# Unknown operation -> 400, and the reply lists what is allowed
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"modulo","a":1,"b":2}'
# {"error":{"code":"UNSUPPORTED_OPERATION","message":"Unsupported operation. Use one of: add, subtract, multiply, divide, power, sqrt, percentage."}}

# Non-numeric operand -> 400, naming the field
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"add","a":"ten","b":2}'
# {"error":{"code":"INVALID_OPERAND","message":"Field \"a\" must be a number the calculator can represent."}}

# Missing operand -> 400
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"operation":"add","a":1}'
# {"error":{"code":"MISSING_OPERAND","message":"Field \"b\" is required for this operation."}}

# Wrong content type -> 415
curl -s -X POST http://localhost:9080/api/v1/calculate \
  -H 'Content-Type: text/plain' \
  -d '{"operation":"add","a":1,"b":2}'
# {"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json."}}

# Wrong method -> 405 with an Allow header
curl -s -X GET http://localhost:9080/api/v1/calculate
# {"error":{"code":"METHOD_NOT_ALLOWED","message":"Only POST is supported for this endpoint."}}
```

## Numbers

Operands and results are IEEE-754 doubles (`float64` in Go, the only numeric type in JSON), so the
usual binary floating-point caveats apply: `0.1 + 0.2` is `0.30000000000000004`, not `0.3`. The
frontend renders results to 12 significant digits, which hides that artefact without hiding real
precision. Exact decimal arithmetic was not adopted because `sqrt` and non-integer `power` are
inherently floating-point, so the guarantee would have been partial and misleading.
