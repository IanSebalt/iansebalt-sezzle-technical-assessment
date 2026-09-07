import type { CalculateRequest, Operation } from '../api/types'

export type DigitKeyId = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9'

/** Operations that take a second operand; `sqrt` is the only unary one. */
export type BinaryOperation = Exclude<Operation, 'sqrt'>

export type KeyId = DigitKeyId | 'decimal' | 'clear' | 'equals' | Operation

export type KeyVariant = 'digit' | 'operator' | 'action' | 'accent'

export interface KeyDefinition {
  id: KeyId
  label: string
  ariaLabel: string
  variant: KeyVariant
  /** KeyboardEvent.key values that press this key. */
  keyboard: string[]
}

export interface PendingRequest {
  id: number
  request: CalculateRequest
  /** `=` consumes the whole expression; a square root only replaces the number on screen. */
  clearsExpression: boolean
}

export interface KeypadState {
  /** The number the display shows. Only ever built from typed digits or a server result. */
  entry: string
  /** False between pressing an operator and typing the operand that follows it. */
  entryStarted: boolean
  operandA: number | null
  operation: BinaryOperation | null
  /** True when `entry` holds a server result, so the next digit starts a fresh number. */
  showsResult: boolean
  pending: PendingRequest | null
  error: string | null
  constraint: string | null
  nextRequestId: number
}

export type KeypadAction =
  | { type: 'press'; key: KeyId }
  | { type: 'resolved'; id: number; result: number }
  | { type: 'failed'; id: number; message: string }
