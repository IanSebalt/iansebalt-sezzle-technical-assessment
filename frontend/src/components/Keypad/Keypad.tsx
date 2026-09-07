import { KEY_DEFINITIONS } from '../../lib/calculator/keys'
import type { KeyId } from '../../lib/calculator/keypadTypes'
import { Key } from '../Key/Key'
import styles from './Keypad.module.css'

interface KeypadProps {
  isDisabled: (id: KeyId) => boolean
  onPress: (id: KeyId) => void
}

export function Keypad({ isDisabled, onPress }: KeypadProps) {
  return (
    <div className={styles.keypad} role="group" aria-label="Calculator keypad">
      {KEY_DEFINITIONS.map((definition) => (
        <Key
          key={definition.id}
          definition={definition}
          disabled={isDisabled(definition.id)}
          onPress={onPress}
        />
      ))}
    </div>
  )
}
