import { ApiError, isAbortError } from './errors'
import { OPERATIONS } from './types'
import type { ApiErrorBody, CalculateRequest, CalculateResponse, Operation } from './types'

const CALCULATE_ENDPOINT = '/api/v1/calculate'

// A proxy sits in front of the API in both dev (Vite) and production (nginx). When the backend is
// down the browser gets one of these from the proxy, not a failed fetch, so they mean the same
// thing to a user: the calculator service could not be reached.
const GATEWAY_FAILURES = new Set([502, 503, 504])

// A proxy in front of a dead backend can take the better part of a minute to give up. The keypad is
// disabled while a request is in flight, so the client stops waiting long before that.
const REQUEST_TIMEOUT_MS = 10_000

/**
 * Every failure leaves here as an ApiError carrying a message fit to show a user, so callers never
 * have to interpret a status code. Aborts are the one exception: they are the caller's own doing.
 */
export async function calculate(
  request: CalculateRequest,
  signal?: AbortSignal,
): Promise<CalculateResponse> {
  let response: Response

  const deadline = AbortSignal.timeout(REQUEST_TIMEOUT_MS)

  try {
    response = await fetch(CALCULATE_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal: signal ? AbortSignal.any([signal, deadline]) : deadline,
    })
  } catch (cause) {
    // A timeout aborts with a TimeoutError, so only the caller's own abort passes through here.
    if (isAbortError(cause)) throw cause
    throw ApiError.network()
  }

  const body: unknown = await response.json().catch(() => null)

  if (!response.ok) {
    if (isApiErrorBody(body)) {
      throw new ApiError(body.error.code, body.error.message, response.status)
    }
    throw GATEWAY_FAILURES.has(response.status)
      ? ApiError.network(response.status)
      : ApiError.unreadable(response.status)
  }

  if (!isCalculateResponse(body)) {
    throw ApiError.unreadable(response.status)
  }

  return body
}

function isApiErrorBody(body: unknown): body is ApiErrorBody {
  if (!isRecord(body) || !isRecord(body.error)) return false

  return typeof body.error.code === 'string' && typeof body.error.message === 'string'
}

function isCalculateResponse(body: unknown): body is CalculateResponse {
  if (!isRecord(body)) return false

  return (
    isOperation(body.operation) &&
    isFiniteNumber(body.a) &&
    // b is absent for unary operations, so it is optional rather than merely nullable.
    (body.b === undefined || isFiniteNumber(body.b)) &&
    isFiniteNumber(body.result)
  )
}

function isOperation(value: unknown): value is Operation {
  return typeof value === 'string' && (OPERATIONS as readonly string[]).includes(value)
}

function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
