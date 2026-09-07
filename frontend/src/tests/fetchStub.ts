import { vi } from 'vitest'

function stub(response: (...args: unknown[]) => Promise<Response>) {
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

export interface StubbedReply {
  status?: number
  body: unknown
}

/** Answers like the real backend would: the reply is chosen from the request that was sent. */
export function stubBackend(reply: (request: Record<string, unknown>) => StubbedReply) {
  return stub(async (...args: unknown[]) => {
    const init = args[1] as RequestInit
    const { status = 200, body } = reply(JSON.parse(String(init.body)) as Record<string, unknown>)

    return new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    })
  })
}
