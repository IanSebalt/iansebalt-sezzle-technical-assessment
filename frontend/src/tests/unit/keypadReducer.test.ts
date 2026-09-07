import { describe, expect, it } from 'vitest'

import { MAX_ENTRY_DIGITS, MAX_ENTRY_MESSAGE } from '../../lib/calculator/format'
import { initialState, keypadReducer } from '../../lib/calculator/keypadReducer'
import type { KeyId, KeypadState } from '../../lib/calculator/keypadTypes'

function pressAll(keys: KeyId[], from: KeypadState = initialState): KeypadState {
  return keys.reduce((state, key) => keypadReducer(state, { type: 'press', key }), from)
}

function resolve(state: KeypadState, result: number): KeypadState {
  const id = state.pending?.id ?? -1
  return keypadReducer(state, { type: 'resolved', id, result })
}

describe('entering a number', () => {
  it('appends digits and collapses the leading zero', () => {
    expect(pressAll(['1', '2', '3']).entry).toBe('123')
    expect(pressAll(['0', '5']).entry).toBe('5')
    expect(pressAll(['0', '0']).entry).toBe('0')
  })

  it('allows only one decimal point', () => {
    expect(pressAll(['1', 'decimal', '5']).entry).toBe('1.5')
    expect(pressAll(['1', 'decimal', 'decimal', '5']).entry).toBe('1.5')
  })

  it('keeps the zero when a decimal point opens the number', () => {
    expect(pressAll(['decimal', '5']).entry).toBe('0.5')
  })

  it('caps the entry and explains the constraint instead of silently ignoring the press', () => {
    const digits = Array.from({ length: MAX_ENTRY_DIGITS }, () => '9' as KeyId)
    const atLimit = pressAll(digits)

    expect(atLimit.entry).toBe('9'.repeat(MAX_ENTRY_DIGITS))
    expect(atLimit.constraint).toBeNull()

    const overLimit = keypadReducer(atLimit, { type: 'press', key: '9' })

    expect(overLimit.entry).toBe(atLimit.entry)
    expect(overLimit.constraint).toBe(MAX_ENTRY_MESSAGE)
  })
})

describe('correcting a mistyped number', () => {
  it('drops the last digit', () => {
    expect(pressAll(['1', '2', '3', 'backspace']).entry).toBe('12')
  })

  it('drops a decimal point like any other character', () => {
    expect(pressAll(['1', 'decimal', 'backspace']).entry).toBe('1')
  })

  it('falls back to zero once the last character goes', () => {
    const emptied = pressAll(['7', 'backspace'])

    expect(emptied.entry).toBe('0')
    expect(emptied.entryStarted).toBe(false)
  })

  it('leaves an already empty entry alone', () => {
    expect(pressAll(['backspace']).entry).toBe('0')
  })

  it('re-opens the expression so equals waits for a fresh operand', () => {
    const state = pressAll(['1', '2', 'add', '3', 'backspace'])

    expect(state.operandA).toBe(12)
    expect(state.operation).toBe('add')
    expect(pressAll(['equals'], state).pending).toBeNull()
  })

  it('clears a result rather than editing its digits', () => {
    const result = resolve(pressAll(['1', '0', 'divide', '4', 'equals']), 2.5)
    const backspaced = pressAll(['backspace'], result)

    expect(backspaced.entry).toBe('0')
    expect(backspaced.entryValue).toBeNull()
    expect(backspaced.showsResult).toBe(false)
  })

  it('clears the digit limit warning', () => {
    const digits = Array.from({ length: MAX_ENTRY_DIGITS + 1 }, () => '9' as KeyId)
    const overLimit = pressAll(digits)

    expect(overLimit.constraint).toBe(MAX_ENTRY_MESSAGE)
    expect(pressAll(['backspace'], overLimit).constraint).toBeNull()
  })
})

describe('building an expression', () => {
  it('stores the operand and the pending operator', () => {
    const state = pressAll(['1', '2', 'add'])

    expect(state.operandA).toBe(12)
    expect(state.operation).toBe('add')
    expect(state.entry).toBe('0')
    expect(state.entryStarted).toBe(false)
  })

  it('replaces the pending operator when a second one is pressed before equals', () => {
    const state = pressAll(['1', '2', 'add', 'subtract', '3', 'equals'])

    expect(state.operandA).toBe(12)
    expect(state.pending?.request).toEqual({ operation: 'subtract', a: 12, b: 3 })
  })

  it('lets the operator be changed after the second operand is typed', () => {
    const state = pressAll(['1', '2', 'add', '3', 'multiply', 'equals'])

    expect(state.pending?.request).toEqual({ operation: 'multiply', a: 12, b: 3 })
  })

  it('builds a pending request on equals', () => {
    const state = pressAll(['1', '0', 'divide', '4', 'equals'])

    expect(state.pending).toEqual({
      id: 1,
      request: { operation: 'divide', a: 10, b: 4 },
      clearsExpression: true,
    })
  })

  it('ignores equals while the expression is incomplete', () => {
    expect(pressAll(['1', '2', 'equals']).pending).toBeNull()
    expect(pressAll(['1', '2', 'add', 'equals']).pending).toBeNull()
    expect(pressAll(['equals']).pending).toBeNull()
  })

  it('never computes a result on its own', () => {
    const state = pressAll(['1', '0', 'divide', '4', 'equals'])

    expect(state.entry).toBe('4')
    expect(state.showsResult).toBe(false)
  })
})

describe('working with a result', () => {
  it('shows the formatted result the server returned', () => {
    const state = resolve(pressAll(['1', '0', 'divide', '4', 'equals']), 2.5)

    expect(state.entry).toBe('2.5')
    expect(state.showsResult).toBe(true)
    expect(state.pending).toBeNull()
    expect(state.operandA).toBeNull()
    expect(state.operation).toBeNull()
  })

  it('reuses the result as the next operand', () => {
    const result = resolve(pressAll(['1', '0', 'divide', '4', 'equals']), 2.5)
    const next = pressAll(['add', '5', 'equals'], result)

    expect(next.pending?.request).toEqual({ operation: 'add', a: 2.5, b: 5 })
  })

  it('chains the exact result rather than the rounded value on screen', () => {
    const divided = resolve(pressAll(['1', '0', 'divide', '3', 'equals']), 10 / 3)

    expect(divided.entry).toBe('3.33333333333')
    expect(divided.entryValue).toBe(10 / 3)

    const chained = pressAll(['multiply', '3', 'equals'], divided)

    expect(chained.pending?.request.a).toBe(10 / 3)
    expect(chained.pending?.request.a).not.toBe(Number(divided.entry))
  })

  it('takes the square root of the exact result, not of the rounded display', () => {
    const divided = resolve(pressAll(['1', '0', 'divide', '3', 'equals']), 10 / 3)

    expect(pressAll(['sqrt'], divided).pending?.request).toEqual({
      operation: 'sqrt',
      a: 10 / 3,
    })
  })

  it('forgets the exact value as soon as the user types over the result', () => {
    const divided = resolve(pressAll(['1', '0', 'divide', '3', 'equals']), 10 / 3)
    const typed = pressAll(['7'], divided)

    expect(typed.entryValue).toBeNull()
    expect(pressAll(['add', '1', 'equals'], typed).pending?.request.a).toBe(7)
  })

  it('starts a fresh decimal number when a decimal point follows a result', () => {
    const result = resolve(pressAll(['1', '0', 'divide', '4', 'equals']), 2.5)
    const next = pressAll(['decimal', '5'], result)

    expect(next.entry).toBe('0.5')
    expect(next.showsResult).toBe(false)
  })

  it('starts a fresh entry when a digit follows a result', () => {
    const result = resolve(pressAll(['1', '0', 'divide', '4', 'equals']), 2.5)

    expect(pressAll(['7'], result).entry).toBe('7')
  })

  it('clears the error once a new number is entered', () => {
    const failed = keypadReducer(pressAll(['1', 'divide', '0', 'equals']), {
      type: 'failed',
      id: 1,
      message: 'Cannot divide by zero.',
    })

    expect(failed.error).toBe('Cannot divide by zero.')
    expect(pressAll(['5'], failed).error).toBeNull()
  })
})

describe('square root', () => {
  it('submits immediately for the number on screen', () => {
    const state = pressAll(['8', '1', 'sqrt'])

    expect(state.pending).toEqual({
      id: 1,
      request: { operation: 'sqrt', a: 81 },
      clearsExpression: false,
    })
  })

  it('leaves a half-built expression intact so it can be finished', () => {
    const rooted = resolve(pressAll(['9', 'add', '1', '6', 'sqrt']), 4)

    expect(rooted.entry).toBe('4')
    expect(rooted.operandA).toBe(9)
    expect(rooted.operation).toBe('add')

    expect(pressAll(['equals'], rooted).pending?.request).toEqual({
      operation: 'add',
      a: 9,
      b: 4,
    })
  })
})

describe('while a calculation is in flight', () => {
  it('ignores further presses', () => {
    const inFlight = pressAll(['1', '0', 'divide', '4', 'equals'])

    expect(pressAll(['7'], inFlight)).toBe(inFlight)
    expect(pressAll(['add'], inFlight)).toBe(inFlight)
    expect(pressAll(['equals'], inFlight)).toBe(inFlight)
  })

  it('still allows clearing', () => {
    const inFlight = pressAll(['1', '0', 'divide', '4', 'equals'])
    const cleared = pressAll(['clear'], inFlight)

    expect(cleared.pending).toBeNull()
    expect(cleared.entry).toBe('0')
  })

  it('ignores an answer that belongs to an older request', () => {
    const inFlight = pressAll(['1', '0', 'divide', '4', 'equals'])

    expect(keypadReducer(inFlight, { type: 'resolved', id: 99, result: 1234 })).toBe(inFlight)
    expect(keypadReducer(inFlight, { type: 'failed', id: 99, message: 'stale' })).toBe(inFlight)
  })
})

describe('clearing', () => {
  it('resets everything the user can see', () => {
    const busy = resolve(pressAll(['1', '0', 'divide', '4', 'equals']), 2.5)
    const cleared = pressAll(['clear'], busy)

    expect(cleared).toEqual({ ...initialState, nextRequestId: cleared.nextRequestId })
    expect(cleared.entry).toBe('0')
    expect(cleared.error).toBeNull()
  })

  it('keeps issuing fresh request ids so a late answer cannot be mistaken for a new one', () => {
    const first = pressAll(['1', 'add', '1', 'equals'])
    const second = pressAll(['clear', '2', 'add', '2', 'equals'], first)

    expect(second.pending?.id).not.toBe(first.pending?.id)
  })
})
