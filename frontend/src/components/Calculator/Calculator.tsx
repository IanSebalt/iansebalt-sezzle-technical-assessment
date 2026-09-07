import { useCallback } from 'react'

import { useCalculator } from '../../hooks/useCalculator'
import { useKeyboard } from '../../hooks/useKeyboard'
import { formatResult } from '../../lib/calculator/format'
import { keyLabel } from '../../lib/calculator/keys'
import type { KeyId, KeypadState } from '../../lib/calculator/keypadTypes'
import { Display } from '../Display/Display'
import { ErrorBanner } from '../ErrorBanner/ErrorBanner'
import { Keypad } from '../Keypad/Keypad'
import styles from './Calculator.module.css'

export function Calculator() {
  const { state, press, isCalculating } = useCalculator()
  useKeyboard(press)

  const isDisabled = useCallback(
    (id: KeyId) => {
      if (isCalculating) return id !== 'clear'
      if (id === 'decimal') return !state.showsResult && state.entry.includes('.')
      if (id === 'equals') return !isExpressionComplete(state)
      return false
    },
    [isCalculating, state],
  )

  return (
    <section className={styles.calculator} aria-label="Calculator">
      <Display expression={describeExpression(state)} value={state.entry} busy={isCalculating} />
      <ErrorBanner message={state.error} />
      <Keypad isDisabled={isDisabled} onPress={press} />

      <p className={styles.constraint} aria-live="polite">
        {state.constraint}
      </p>

      <ul className={styles.legend}>
        <li>Square root (√) applies to the number shown.</li>
        <li>a % b gives a% of b.</li>
      </ul>
    </section>
  )
}

function describeExpression(state: KeypadState): string {
  if (state.operandA === null || state.operation === null) return ''

  return `${formatResult(state.operandA)} ${keyLabel(state.operation)}`
}

function isExpressionComplete(state: KeypadState): boolean {
  return state.operandA !== null && state.operation !== null && state.entryStarted
}
