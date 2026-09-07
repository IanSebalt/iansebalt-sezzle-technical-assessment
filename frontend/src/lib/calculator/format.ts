/**
 * float64 carries roughly 15-17 significant decimal digits, so past 15 the extra keystrokes buy
 * noise rather than precision and the keypad stops accepting them.
 */
export const MAX_ENTRY_DIGITS = 15

export const MAX_ENTRY_MESSAGE = `Maximum ${MAX_ENTRY_DIGITS} digits.`

const SIGNIFICANT_DIGITS = 12

/**
 * Renders a float64 the way a calculator would: no trailing zeros and no `0.30000000000000004`.
 * toPrecision already falls back to exponential notation once plain digits stop being readable,
 * which is exactly where we want the switch, so the thresholds are left to it.
 */
export function formatResult(value: number): string {
  if (value === 0) return '0'

  return trimTrailingZeros(value.toPrecision(SIGNIFICANT_DIGITS))
}

export function countDigits(entry: string): number {
  let digits = 0
  for (const character of entry) {
    if (character >= '0' && character <= '9') digits += 1
  }
  return digits
}

function trimTrailingZeros(text: string): string {
  const [mantissa = text, exponent] = text.split('e')

  const trimmed = mantissa.includes('.') ? mantissa.replace(/\.?0+$/, '') : mantissa

  return exponent === undefined ? trimmed : `${trimmed}e${exponent}`
}
