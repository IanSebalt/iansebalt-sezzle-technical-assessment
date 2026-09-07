import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Display } from '../../components/Display/Display'

describe('Display', () => {
  it('shows the number currently on screen', () => {
    render(<Display expression="" value="123" busy={false} />)

    expect(screen.getByRole('status')).toHaveTextContent('123')
  })

  it('shows the pending expression above the number', () => {
    render(<Display expression="10 ÷" value="4" busy={false} />)

    expect(screen.getByText('10 ÷')).toBeInTheDocument()
    expect(screen.getByRole('status')).toHaveTextContent('4')
  })

  it('announces changes politely', () => {
    render(<Display expression="" value="7" busy={false} />)

    expect(screen.getByRole('status')).toHaveAttribute('aria-live', 'polite')
  })

  it('steps the type down so a long result stays fully visible', () => {
    const { rerender } = render(<Display expression="" value="123" busy={false} />)
    expect(screen.getByRole('status')).toHaveAttribute('data-size', 'normal')

    rerender(<Display expression="" value="12345678901" busy={false} />)
    expect(screen.getByRole('status')).toHaveAttribute('data-size', 'medium')

    rerender(<Display expression="" value="3.6972963765e+197" busy={false} />)
    expect(screen.getByRole('status')).toHaveAttribute('data-size', 'small')
  })

  it('marks itself busy while a calculation is in flight', () => {
    const { rerender } = render(<Display expression="" value="7" busy={false} />)
    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'false')

    rerender(<Display expression="" value="7" busy />)
    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true')
  })
})
