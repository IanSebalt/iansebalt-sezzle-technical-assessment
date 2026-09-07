import { useEffect } from 'react'

import { keyIdForKeyboardKey } from '../lib/calculator/keys'
import type { KeyId } from '../lib/calculator/keypadTypes'

/** Maps physical key presses onto the same key definitions the on-screen keypad renders. */
export function useKeyboard(onPress: (key: KeyId) => void): void {
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.metaKey || event.ctrlKey || event.altKey) return

      const keyId = keyIdForKeyboardKey(event.key)
      if (keyId === undefined) return

      // "/" opens quick-find in some browsers, and Enter would resubmit a focused button.
      event.preventDefault()
      onPress(keyId)
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onPress])
}
