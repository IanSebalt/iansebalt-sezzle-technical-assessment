import type { KeyDefinition, KeyId } from '../../lib/calculator/keypadTypes'
import styles from './Key.module.css'

interface KeyProps {
  definition: KeyDefinition
  disabled: boolean
  onPress: (id: KeyId) => void
}

export function Key({ definition, disabled, onPress }: KeyProps) {
  return (
    <button
      type="button"
      className={[styles.key, styles[definition.variant]].filter(Boolean).join(' ')}
      data-key={definition.id}
      aria-label={definition.ariaLabel}
      disabled={disabled}
      onClick={() => onPress(definition.id)}
    >
      <span aria-hidden="true">{definition.label}</span>
    </button>
  )
}
