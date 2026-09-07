import styles from './Display.module.css'

interface DisplayProps {
  expression: string
  value: string
  busy: boolean
}

/**
 * Long results (exponential notation, in particular) are stepped down so they stay fully visible.
 * The thresholds are the longest string that still fits the panel at each size, measured at the
 * 22rem card width where the panel is narrowest.
 */
function sizeFor(value: string): 'normal' | 'medium' | 'small' {
  if (value.length > 14) return 'small'
  if (value.length > 10) return 'medium'
  return 'normal'
}

export function Display({ expression, value, busy }: DisplayProps) {
  return (
    <div className={styles.display}>
      <p className={styles.expression}>{expression}</p>
      <output
        className={styles.value}
        data-size={sizeFor(value)}
        aria-live="polite"
        aria-busy={busy}
      >
        {value}
      </output>
    </div>
  )
}
