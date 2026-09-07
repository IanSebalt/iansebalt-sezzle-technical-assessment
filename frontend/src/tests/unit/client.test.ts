import { describe, expect, it } from 'vitest'

import { calculate } from '../../lib/api/client'
import { ApiError, CLIENT_ERROR_CODES } from '../../lib/api/errors'
import { stubJsonResponse, stubRejection, stubTextResponse } from '../fetchStub'

describe('calculate', () => {
  it('sends POST with a JSON content type and the request body', async () => {
    const fetchMock = stubJsonResponse({ operation: 'add', a: 2, b: 3, result: 5 })

    await calculate({ operation: 'add', a: 2, b: 3 })

    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit]
    expect(url).toBe('/api/v1/calculate')
    expect(init.method).toBe('POST')
    expect(init.headers).toMatchObject({ 'Content-Type': 'application/json' })
    expect(JSON.parse(String(init.body))).toEqual({ operation: 'add', a: 2, b: 3 })
  })

  it('omits b for a unary operation', async () => {
    const fetchMock = stubJsonResponse({ operation: 'sqrt', a: 9, result: 3 })

    await calculate({ operation: 'sqrt', a: 9 })

    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit]
    expect(JSON.parse(String(init.body))).toEqual({ operation: 'sqrt', a: 9 })
  })

  it('returns the parsed success payload', async () => {
    stubJsonResponse({ operation: 'divide', a: 10, b: 4, result: 2.5 })

    await expect(calculate({ operation: 'divide', a: 10, b: 4 })).resolves.toEqual({
      operation: 'divide',
      a: 10,
      b: 4,
      result: 2.5,
    })
  })

  it('throws an ApiError carrying the code and message the server chose', async () => {
    stubJsonResponse(
      { error: { code: 'DIVISION_BY_ZERO', message: 'Cannot divide by zero.' } },
      422,
    )

    const error = await calculate({ operation: 'divide', a: 1, b: 0 }).catch((cause) => cause)

    expect(error).toBeInstanceOf(ApiError)
    expect(error.code).toBe('DIVISION_BY_ZERO')
    expect(error.message).toBe('Cannot divide by zero.')
    expect(error.status).toBe(422)
  })

  it('falls back to a readable message when an error body is not the documented envelope', async () => {
    stubTextResponse('<html>502 Bad Gateway</html>', 502)

    const error = await calculate({ operation: 'add', a: 1, b: 2 }).catch((cause) => cause)

    expect(error).toBeInstanceOf(ApiError)
    expect(error.code).toBe(CLIENT_ERROR_CODES.unreadable)
    expect(error.message).toBe('Unexpected response from the server.')
  })

  it('falls back to a readable message when an error envelope is malformed', async () => {
    stubJsonResponse({ error: { code: 42 } }, 400)

    const error = await calculate({ operation: 'add', a: 1, b: 2 }).catch((cause) => cause)

    expect(error.code).toBe(CLIENT_ERROR_CODES.unreadable)
  })

  it('rejects a successful response whose result is not a usable number', async () => {
    stubJsonResponse({ operation: 'add', a: 1, b: 2, result: 'three' })

    const error = await calculate({ operation: 'add', a: 1, b: 2 }).catch((cause) => cause)

    expect(error.code).toBe(CLIENT_ERROR_CODES.unreadable)
  })

  it('reports a network failure when fetch rejects', async () => {
    stubRejection(new TypeError('Failed to fetch'))

    const error = await calculate({ operation: 'add', a: 1, b: 2 }).catch((cause) => cause)

    expect(error).toBeInstanceOf(ApiError)
    expect(error.code).toBe(CLIENT_ERROR_CODES.network)
    expect(error.message).toBe(
      'Cannot reach the calculator service. Check your connection and try again.',
    )
  })

  it('propagates an abort untouched so the caller can ignore it', async () => {
    const abort = new DOMException('The operation was aborted.', 'AbortError')
    stubRejection(abort)

    const error = await calculate({ operation: 'add', a: 1, b: 2 }).catch((cause) => cause)

    expect(error).toBe(abort)
    expect(error).not.toBeInstanceOf(ApiError)
  })
})
