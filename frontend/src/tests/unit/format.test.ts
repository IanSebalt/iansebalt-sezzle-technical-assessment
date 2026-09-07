import { describe, expect, it } from 'vitest'

import { countDigits, formatResult, MAX_ENTRY_DIGITS } from '../../lib/calculator/format'

describe('formatResult', () => {
  it('renders integers without a trailing decimal', () => {
    expect(formatResult(5)).toBe('5')
    expect(formatResult(1024)).toBe('1024')
    expect(formatResult(-12)).toBe('-12')
  })

  it('renders zero as a bare zero, including negative zero', () => {
    expect(formatResult(0)).toBe('0')
    expect(formatResult(-0)).toBe('0')
  })

  it('keeps meaningful decimals', () => {
    expect(formatResult(2.5)).toBe('2.5')
    expect(formatResult(0.125)).toBe('0.125')
  })

  it('trims floating point noise', () => {
    expect(formatResult(0.1 + 0.2)).toBe('0.3')
    expect(formatResult(1.005 * 3)).toBe('3.015')
  })

  it('keeps a repeating decimal readable rather than exact', () => {
    expect(formatResult(1 / 3)).toBe('0.333333333333')
  })

  it('switches to exponential notation for very large magnitudes', () => {
    expect(formatResult(1e20)).toBe('1e+20')
    expect(formatResult(-2.5e15)).toBe('-2.5e+15')
  })

  it('switches to exponential notation for very small magnitudes', () => {
    expect(formatResult(1e-12)).toBe('1e-12')
    expect(formatResult(1e-9)).toBe('1e-9')
  })

  it('stays in plain notation while the digits are still readable', () => {
    expect(formatResult(999999999999)).toBe('999999999999')
    expect(formatResult(0.000001)).toBe('0.000001')
  })
})

describe('countDigits', () => {
  it('counts only digits, ignoring the sign and decimal point', () => {
    expect(countDigits('123')).toBe(3)
    expect(countDigits('-12.34')).toBe(4)
    expect(countDigits('0.')).toBe(1)
  })

  it('agrees with the advertised entry limit', () => {
    expect(countDigits('9'.repeat(MAX_ENTRY_DIGITS))).toBe(MAX_ENTRY_DIGITS)
  })
})
