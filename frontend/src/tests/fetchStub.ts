import { vi } from 'vitest'

function stub(response: () => Promise<Response>) {
  const fetchMock = vi.fn(response)
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

export function stubJsonResponse(body: unknown, status = 200) {
  return stub(async () =>
    new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    }),
  )
}

export function stubTextResponse(body: string, status = 200) {
  return stub(async () =>
    new Response(body, { status, headers: { 'Content-Type': 'text/html' } }),
  )
}

export function stubRejection(cause: unknown) {
  return stub(async () => {
    throw cause
  })
}
