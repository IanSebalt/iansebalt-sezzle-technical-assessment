import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Keypad } from '../../components/Keypad/Keypad'
import { KEY_DEFINITIONS } from '../../lib/calculator/keys'

describe('Keypad', () => {
  it('renders every key with its accessible name', () => {
    render(<Keypad isDisabled={() => false} onPress={vi.fn()} />)

    for (const definition of KEY_DEFINITIONS) {
      expect(screen.getByRole('button', { name: definition.ariaLabel })).toBeInTheDocument()
    }
  })

  it('groups the keys under a single accessible label', () => {
    render(<Keypad isDisabled={() => false} onPress={vi.fn()} />)

    expect(screen.getByRole('group', { name: 'Calculator keypad' })).toBeInTheDocument()
  })

  it('disables exactly the keys the predicate rejects', () => {
    render(<Keypad isDisabled={(id) => id === 'decimal'} onPress={vi.fn()} />)

    expect(screen.getByRole('button', { name: 'Decimal point' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'Seven' })).toBeEnabled()
  })

  it('forwards the pressed key upwards', async () => {
    const onPress = vi.fn()
    render(<Keypad isDisabled={() => false} onPress={onPress} />)

    await userEvent.click(screen.getByRole('button', { name: 'Seven' }))

    expect(onPress).toHaveBeenCalledExactlyOnceWith('7')
  })
})
