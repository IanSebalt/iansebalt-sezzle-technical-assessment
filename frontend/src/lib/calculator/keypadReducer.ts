import type { CalculateRequest } from '../api/types'
import { countDigits, formatResult, MAX_ENTRY_DIGITS, MAX_ENTRY_MESSAGE } from './format'
import type {
  BinaryOperation,
  DigitKeyId,
  KeyId,
  KeypadAction,
  KeypadState,
} from './keypadTypes'

export const initialState: KeypadState = {
  entry: '0',
  entryValue: null,
  entryStarted: false,
  operandA: null,
  operation: null,
  showsResult: false,
  pending: null,
  error: null,
  constraint: null,
  nextRequestId: 1,
}

/**
 * A pure state machine over one binary operation: operand, operator, operand, equals. It never
 * performs arithmetic — the only numbers it writes to the display are typed digits and results the
 * server sent back.
 */
export function keypadReducer(state: KeypadState, action: KeypadAction): KeypadState {
  switch (action.type) {
    case 'press':
      return press(state, action.key)

    case 'resolved': {
      if (state.pending?.id !== action.id) return state
      const { clearsExpression } = state.pending

      return {
        ...state,
        entry: formatResult(action.result),
        entryValue: action.result,
        entryStarted: !clearsExpression,
        operandA: clearsExpression ? null : state.operandA,
        operation: clearsExpression ? null : state.operation,
        showsResult: true,
        pending: null,
        error: null,
        constraint: null,
      }
    }

    case 'failed': {
      if (state.pending?.id !== action.id) return state

      return { ...state, pending: null, error: action.message, constraint: null }
    }
  }
}

function press(state: KeypadState, key: KeyId): KeypadState {
  if (key === 'clear') return { ...initialState, nextRequestId: state.nextRequestId }

  // A calculation is in flight; only clearing may interrupt it.
  if (state.pending !== null) return state

  if (isDigit(key)) return appendDigit(state, key)
  if (key === 'backspace') return deleteLastDigit(state)
  if (key === 'decimal') return appendDecimal(state)
  if (key === 'equals') return submitExpression(state)
  if (key === 'sqrt') return submitSquareRoot(state)

  return setOperation(state, key)
}

function appendDigit(state: KeypadState, digit: DigitKeyId): KeypadState {
  const base = state.showsResult || state.entry === '0' ? '' : state.entry
  const entry = base + digit

  if (countDigits(entry) > MAX_ENTRY_DIGITS) {
    return { ...state, constraint: MAX_ENTRY_MESSAGE }
  }

  return {
    ...state,
    entry,
    entryValue: null,
    entryStarted: true,
    showsResult: false,
    error: null,
    constraint: null,
  }
}

/** A result cannot be edited digit by digit, so backspacing one clears the entry instead. */
function deleteLastDigit(state: KeypadState): KeypadState {
  if (state.showsResult) {
    return {
      ...state,
      entry: '0',
      entryValue: null,
      entryStarted: false,
      showsResult: false,
      error: null,
      constraint: null,
    }
  }

  const trimmed = state.entry.slice(0, -1)
  const entry = trimmed === '' ? '0' : trimmed

  return {
    ...state,
    entry,
    entryValue: null,
    entryStarted: entry !== '0',
    error: null,
    constraint: null,
  }
}

function appendDecimal(state: KeypadState): KeypadState {
  if (state.showsResult) {
    return {
      ...state,
      entry: '0.',
      entryValue: null,
      entryStarted: true,
      showsResult: false,
      error: null,
      constraint: null,
    }
  }
  if (state.entry.includes('.')) return state

  return {
    ...state,
    entry: `${state.entry}.`,
    entryValue: null,
    entryStarted: true,
    error: null,
    constraint: null,
  }
}

/** The operator may be changed at any point before `=`; only the first press captures operand A. */
function setOperation(state: KeypadState, operation: BinaryOperation): KeypadState {
  if (state.operandA !== null) {
    return { ...state, operation, error: null, constraint: null }
  }

  return {
    ...state,
    operandA: currentOperand(state),
    operation,
    entry: '0',
    entryValue: null,
    entryStarted: false,
    showsResult: false,
    error: null,
    constraint: null,
  }
}

function submitExpression(state: KeypadState): KeypadState {
  if (state.operandA === null || state.operation === null || !state.entryStarted) return state

  return withPendingRequest(state, {
    operation: state.operation,
    a: state.operandA,
    b: currentOperand(state),
  }, true)
}

function submitSquareRoot(state: KeypadState): KeypadState {
  return withPendingRequest(state, { operation: 'sqrt', a: currentOperand(state) }, false)
}

function withPendingRequest(
  state: KeypadState,
  request: CalculateRequest,
  clearsExpression: boolean,
): KeypadState {
  return {
    ...state,
    error: null,
    constraint: null,
    pending: { id: state.nextRequestId, request, clearsExpression },
    nextRequestId: state.nextRequestId + 1,
  }
}

/** Prefers the server's exact value over the rounded string the display shows. */
function currentOperand(state: KeypadState): number {
  return state.entryValue ?? Number(state.entry)
}

function isDigit(key: KeyId): key is DigitKeyId {
  return key.length === 1 && key >= '0' && key <= '9'
}
