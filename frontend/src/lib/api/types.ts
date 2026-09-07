export const OPERATIONS = [
  'add',
  'subtract',
  'multiply',
  'divide',
  'power',
  'sqrt',
  'percentage',
] as const

export type Operation = (typeof OPERATIONS)[number]

export interface CalculateRequest {
  operation: Operation
  a: number
  b?: number
}

/** Mirrors the backend response DTO; `b` is absent for unary operations. */
export interface CalculateResponse {
  operation: Operation
  a: number
  b?: number
  result: number
}

export interface ApiErrorBody {
  error: {
    code: string
    message: string
  }
}
