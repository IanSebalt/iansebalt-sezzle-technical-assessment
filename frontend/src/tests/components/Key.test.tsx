import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Key } from '../../components/Key/Key'
import type { KeyDefinition } from '../../lib/calculator/keypadTypes'

const divide: KeyDefinition = {
  id: 'divide',
  label: '÷',
  ariaLabel: 'Divide',
  variant: 'operator',
  keyboard: ['/'],
}

describe('Key', () => {
  it('is named for assistive technology and labelled for the eye', () => {
    render(<Key definition={divide} disabled={false} onPress={vi.fn()} />)

    expect(screen.getByRole('button', { name: 'Divide' })).toHaveTextContent('÷')
  })

  it('reports its own id when pressed', async () => {
    const onPress = vi.fn()
    render(<Key definition={divide} disabled={false} onPress={onPress} />)

    await userEvent.click(screen.getByRole('button', { name: 'Divide' }))

    expect(onPress).toHaveBeenCalledExactlyOnceWith('divide')
  })

  it('cannot be pressed while disabled', async () => {
    const onPress = vi.fn()
    render(<Key definition={divide} disabled onPress={onPress} />)

    const button = screen.getByRole('button', { name: 'Divide' })
    expect(button).toBeDisabled()

    await userEvent.click(button)
    expect(onPress).not.toHaveBeenCalled()
  })
})
