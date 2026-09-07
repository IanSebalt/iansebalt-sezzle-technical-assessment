import type { KeyDefinition, KeyId } from './keypadTypes'

/** Visual order: the keypad renders this list straight into a four-column grid. */
export const KEY_DEFINITIONS: readonly KeyDefinition[] = [
  { id: 'clear', label: 'C', ariaLabel: 'Clear', variant: 'action', keyboard: ['Escape', 'c', 'C'] },
  { id: 'backspace', label: '⌫', ariaLabel: 'Delete last digit', variant: 'action', keyboard: ['Backspace'] },
  { id: 'sqrt', label: '√', ariaLabel: 'Square root', variant: 'operator', keyboard: ['r', 'R'] },
  { id: 'power', label: '^', ariaLabel: 'To the power of', variant: 'operator', keyboard: ['^'] },

  { id: '7', label: '7', ariaLabel: 'Seven', variant: 'digit', keyboard: ['7'] },
  { id: '8', label: '8', ariaLabel: 'Eight', variant: 'digit', keyboard: ['8'] },
  { id: '9', label: '9', ariaLabel: 'Nine', variant: 'digit', keyboard: ['9'] },
  { id: 'divide', label: '÷', ariaLabel: 'Divide', variant: 'operator', keyboard: ['/'] },

  { id: '4', label: '4', ariaLabel: 'Four', variant: 'digit', keyboard: ['4'] },
  { id: '5', label: '5', ariaLabel: 'Five', variant: 'digit', keyboard: ['5'] },
  { id: '6', label: '6', ariaLabel: 'Six', variant: 'digit', keyboard: ['6'] },
  { id: 'multiply', label: '×', ariaLabel: 'Multiply', variant: 'operator', keyboard: ['*', 'x'] },

  { id: '1', label: '1', ariaLabel: 'One', variant: 'digit', keyboard: ['1'] },
  { id: '2', label: '2', ariaLabel: 'Two', variant: 'digit', keyboard: ['2'] },
  { id: '3', label: '3', ariaLabel: 'Three', variant: 'digit', keyboard: ['3'] },
  { id: 'subtract', label: '−', ariaLabel: 'Subtract', variant: 'operator', keyboard: ['-'] },

  { id: '0', label: '0', ariaLabel: 'Zero', variant: 'digit', keyboard: ['0'] },
  { id: 'decimal', label: '.', ariaLabel: 'Decimal point', variant: 'digit', keyboard: ['.'] },
  { id: 'percentage', label: '%', ariaLabel: 'Percent of', variant: 'operator', keyboard: ['%'] },
  { id: 'add', label: '+', ariaLabel: 'Add', variant: 'operator', keyboard: ['+'] },

  // Equals spans the final row; Keypad.module.css keys off its id.
  { id: 'equals', label: '=', ariaLabel: 'Equals', variant: 'accent', keyboard: ['Enter', '='] },
]

// Every KeyId appears exactly once in KEY_DEFINITIONS, which keys.test.ts enforces.
const labels = Object.fromEntries(
  KEY_DEFINITIONS.map((key) => [key.id, key.label]),
) as Record<KeyId, string>

const byKeyboardKey = new Map<string, KeyId>(
  KEY_DEFINITIONS.flatMap((key) => key.keyboard.map((stroke) => [stroke, key.id] as const)),
)

export function keyLabel(id: KeyId): string {
  return labels[id]
}

export function keyIdForKeyboardKey(stroke: string): KeyId | undefined {
  return byKeyboardKey.get(stroke)
}
