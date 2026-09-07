/** Failures the browser detects on its own, where the server never got to state a message. */
export const CLIENT_ERROR_CODES = {
  network: 'NETWORK_UNAVAILABLE',
  unreadable: 'UNREADABLE_RESPONSE',
} as const

export class ApiError extends Error {
  readonly code: string
  readonly status: number

  constructor(code: string, message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }

  static network(status = 0): ApiError {
    return new ApiError(
      CLIENT_ERROR_CODES.network,
      'Cannot reach the calculator service. Check your connection and try again.',
      status,
    )
  }

  static unreadable(status: number): ApiError {
    return new ApiError(
      CLIENT_ERROR_CODES.unreadable,
      'Unexpected response from the server.',
      status,
    )
  }
}

/** Structural check: DOMException does not reliably subclass Error across runtimes. */
export function isAbortError(cause: unknown): boolean {
  return (
    typeof cause === 'object' &&
    cause !== null &&
    (cause as { name?: unknown }).name === 'AbortError'
  )
}
