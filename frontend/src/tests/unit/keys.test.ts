import { describe, expect, it } from 'vitest'

import { KEY_DEFINITIONS, keyIdForKeyboardKey, keyLabel } from '../../lib/calculator/keys'
import { OPERATIONS } from '../../lib/api/types'

describe('key definitions', () => {
  it('gives every key a unique id', () => {
    const ids = KEY_DEFINITIONS.map((key) => key.id)

    expect(new Set(ids).size).toBe(ids.length)
  })

  it('gives every key a label and an accessible name', () => {
    for (const key of KEY_DEFINITIONS) {
      expect(key.label, `label for ${key.id}`).not.toBe('')
      expect(key.ariaLabel, `aria-label for ${key.id}`).not.toBe('')
    }
  })

  it('offers a key for every operation the API supports', () => {
    const ids = new Set(KEY_DEFINITIONS.map((key) => key.id))

    for (const operation of OPERATIONS) {
      expect(ids.has(operation), `missing a key for ${operation}`).toBe(true)
    }
  })

  it('lays the keys out in rows of four, with equals spanning the last row', () => {
    const last = KEY_DEFINITIONS.at(-1)

    expect(last?.id).toBe('equals')
    expect((KEY_DEFINITIONS.length - 1) % 4).toBe(0)
  })

  it('binds every keyboard stroke to exactly one key', () => {
    const strokes = KEY_DEFINITIONS.flatMap((key) => key.keyboard)

    expect(new Set(strokes).size).toBe(strokes.length)
  })

  it('resolves keyboard strokes back to their key', () => {
    expect(keyIdForKeyboardKey('7')).toBe('7')
    expect(keyIdForKeyboardKey('/')).toBe('divide')
    expect(keyIdForKeyboardKey('Enter')).toBe('equals')
    expect(keyIdForKeyboardKey('Escape')).toBe('clear')
    expect(keyIdForKeyboardKey('Backspace')).toBe('backspace')
    expect(keyIdForKeyboardKey('F5')).toBeUndefined()
  })

  it('exposes labels for rendering an expression', () => {
    expect(keyLabel('divide')).toBe('÷')
    expect(keyLabel('multiply')).toBe('×')
  })
})
