import { ApiError, isAbortError } from './errors'
import type { ApiErrorBody, CalculateRequest, CalculateResponse } from './types'

const CALCULATE_ENDPOINT = '/api/v1/calculate'

/**
 * Every failure leaves here as an ApiError carrying a message fit to show a user, so callers never
 * have to interpret a status code. Aborts are the one exception: they are the caller's own doing.
 */
export async function calculate(
  request: CalculateRequest,
  signal?: AbortSignal,
): Promise<CalculateResponse> {
  let response: Response

  try {
    response = await fetch(CALCULATE_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal,
    })
  } catch (cause) {
    if (isAbortError(cause)) throw cause
    throw ApiError.network()
  }

  const body: unknown = await response.json().catch(() => null)

  if (!response.ok) {
    throw isApiErrorBody(body)
      ? new ApiError(body.error.code, body.error.message, response.status)
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
  return isRecord(body) && typeof body.result === 'number' && Number.isFinite(body.result)
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
