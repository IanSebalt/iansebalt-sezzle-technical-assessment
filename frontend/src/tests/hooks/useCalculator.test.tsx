import { act, renderHook, waitFor } from '@testing-library/react'
import type { RenderHookResult } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { useCalculator } from '../../hooks/useCalculator'
import type { KeyId } from '../../lib/calculator/keypadTypes'
import { stubJsonResponse, stubRejection } from '../fetchStub'

type Calculator = ReturnType<typeof useCalculator>

function pressKeys(view: RenderHookResult<Calculator, unknown>, keys: KeyId[]) {
  act(() => {
    for (const key of keys) view.result.current.press(key)
  })
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function deferred<T>() {
  let settle!: (value: T) => void
  const promise = new Promise<T>((resolve) => {
    settle = resolve
  })
  return { promise, settle }
}

describe('useCalculator', () => {
  it('posts the built request once when equals is pressed', async () => {
    const fetchMock = stubJsonResponse({ operation: 'divide', a: 10, b: 4, result: 2.5 })
    const view = renderHook(() => useCalculator())

    pressKeys(view, ['1', '0', 'divide', '4', 'equals'])

    await waitFor(() => expect(view.result.current.state.entry).toBe('2.5'))
    expect(fetchMock).toHaveBeenCalledTimes(1)

    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit]
    expect(JSON.parse(String(init.body))).toEqual({ operation: 'divide', a: 10, b: 4 })
  })

  it('reports that a calculation is in flight, then that it is done', async () => {
    const pending = deferred<Response>()
    vi.stubGlobal('fetch', vi.fn(() => pending.promise))

    const view = renderHook(() => useCalculator())
    pressKeys(view, ['2', 'add', '2', 'equals'])

    expect(view.result.current.isCalculating).toBe(true)

    await act(async () => {
      pending.settle(jsonResponse({ operation: 'add', a: 2, b: 2, result: 4 }))
    })

    await waitFor(() => expect(view.result.current.isCalculating).toBe(false))
    expect(view.result.current.state.entry).toBe('4')
  })

  it('surfaces the message the server chose for a rejected calculation', async () => {
    stubJsonResponse({ error: { code: 'DIVISION_BY_ZERO', message: 'Cannot divide by zero.' } }, 422)
    const view = renderHook(() => useCalculator())

    pressKeys(view, ['1', '0', 'divide', '0', 'equals'])

    await waitFor(() => expect(view.result.current.state.error).toBe('Cannot divide by zero.'))
    expect(view.result.current.state.entry).toBe('0')
  })

  it('falls back to a network message when the service cannot be reached', async () => {
    stubRejection(new TypeError('Failed to fetch'))
    const view = renderHook(() => useCalculator())

    pressKeys(view, ['1', 'add', '1', 'equals'])

    await waitFor(() =>
      expect(view.result.current.state.error).toBe(
        'Cannot reach the calculator service. Check your connection and try again.',
      ),
    )
  })

  it('clears a previous error on the next successful calculation', async () => {
    stubJsonResponse({ error: { code: 'DIVISION_BY_ZERO', message: 'Cannot divide by zero.' } }, 422)
    const view = renderHook(() => useCalculator())

    pressKeys(view, ['1', 'divide', '0', 'equals'])
    await waitFor(() => expect(view.result.current.state.error).not.toBeNull())

    stubJsonResponse({ operation: 'add', a: 1, b: 1, result: 2 })
    pressKeys(view, ['clear', '1', 'add', '1', 'equals'])

    await waitFor(() => expect(view.result.current.state.entry).toBe('2'))
    expect(view.result.current.state.error).toBeNull()
  })

  it('sends a request as soon as square root is pressed', async () => {
    const fetchMock = stubJsonResponse({ operation: 'sqrt', a: 81, result: 9 })
    const view = renderHook(() => useCalculator())

    pressKeys(view, ['8', '1', 'sqrt'])

    await waitFor(() => expect(view.result.current.state.entry).toBe('9'))

    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit]
    expect(JSON.parse(String(init.body))).toEqual({ operation: 'sqrt', a: 81 })
  })

  it('ignores an answer that arrives after its request was abandoned', async () => {
    const abandoned = deferred<Response>()
    const current = deferred<Response>()
    vi.stubGlobal(
      'fetch',
      vi.fn().mockReturnValueOnce(abandoned.promise).mockReturnValueOnce(current.promise),
    )

    const view = renderHook(() => useCalculator())

    pressKeys(view, ['1', 'add', '1', 'equals'])
    pressKeys(view, ['clear', '2', 'add', '2', 'equals'])

    await act(async () => {
      current.settle(jsonResponse({ operation: 'add', a: 2, b: 2, result: 4 }))
    })
    await waitFor(() => expect(view.result.current.state.entry).toBe('4'))

    await act(async () => {
      abandoned.settle(jsonResponse({ operation: 'add', a: 1, b: 1, result: 2 }))
    })

    expect(view.result.current.state.entry).toBe('4')
  })
})
