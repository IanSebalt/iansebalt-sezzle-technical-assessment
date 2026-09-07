import { useCallback, useEffect, useReducer } from 'react'

import { calculate } from '../lib/api/client'
import { ApiError, isAbortError } from '../lib/api/errors'
import { initialState, keypadReducer } from '../lib/calculator/keypadReducer'
import type { KeyId } from '../lib/calculator/keypadTypes'

const UNEXPECTED_FAILURE = 'Something went wrong. Please try again.'

export function useCalculator() {
  const [state, dispatch] = useReducer(keypadReducer, initialState)
  const { pending } = state

  useEffect(() => {
    if (pending === null) return

    const controller = new AbortController()

    calculate(pending.request, controller.signal)
      .then((response) => dispatch({ type: 'resolved', id: pending.id, result: response.result }))
      .catch((cause: unknown) => {
        // An abort means a newer request replaced this one; its answer is no longer wanted.
        if (isAbortError(cause)) return
        dispatch({ type: 'failed', id: pending.id, message: messageFor(cause) })
      })

    return () => controller.abort()
  }, [pending])

  const press = useCallback((key: KeyId) => dispatch({ type: 'press', key }), [])

  return { state, press, isCalculating: pending !== null }
}

function messageFor(cause: unknown): string {
  return cause instanceof ApiError ? cause.message : UNEXPECTED_FAILURE
}
