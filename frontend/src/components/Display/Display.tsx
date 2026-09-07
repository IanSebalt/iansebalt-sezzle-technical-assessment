import styles from './Display.module.css'

interface DisplayProps {
  expression: string
  value: string
  busy: boolean
}

export function Display({ expression, value, busy }: DisplayProps) {
  return (
    <div className={styles.display}>
      <p className={styles.expression}>{expression || ' '}</p>
      <output className={styles.value} aria-live="polite" aria-busy={busy}>
        {value}
      </output>
    </div>
  )
}
