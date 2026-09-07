import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Calculator } from '../../components/Calculator/Calculator'
import { MAX_ENTRY_DIGITS, MAX_ENTRY_MESSAGE } from '../../lib/calculator/format'
import { stubBackend, stubRejection } from '../fetchStub'

/** Stands in for the Go service, including the error envelopes it returns. */
function stubCalculatorApi() {
  return stubBackend((request) => {
    const { operation, a, b } = request as { operation: string; a: number; b?: number }

    const rejected = (code: string, message: string) => ({
      status: 422,
      body: { error: { code, message } },
    })

    switch (operation) {
      case 'add':
        return { body: { operation, a, b, result: a + (b ?? 0) } }
      case 'subtract':
        return { body: { operation, a, b, result: a - (b ?? 0) } }
      case 'multiply':
        return { body: { operation, a, b, result: a * (b ?? 0) } }
      case 'percentage':
        return { body: { operation, a, b, result: (a / 100) * (b ?? 0) } }
      case 'divide':
        return b === 0
          ? rejected('DIVISION_BY_ZERO', 'Cannot divide by zero.')
          : { body: { operation, a, b, result: a / (b as number) } }
      case 'sqrt':
        return a < 0
          ? rejected('NEGATIVE_SQUARE_ROOT', 'Cannot take the square root of a negative number.')
          : { body: { operation, a, result: Math.sqrt(a) } }
      default:
        return {
          status: 400,
          body: { error: { code: 'UNSUPPORTED_OPERATION', message: 'Unsupported operation.' } },
        }
    }
  })
}

async function pressKeys(names: string[]) {
  for (const name of names) {
    await userEvent.click(screen.getByRole('button', { name }))
  }
}

function display() {
  return screen.getByRole('status')
}

describe('Calculator', () => {
  it('shows the result the service computed', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await pressKeys(['One', 'Zero', 'Divide', 'Four', 'Equals'])

    await waitFor(() => expect(display()).toHaveTextContent('2.5'))
  })

  it('shows the pending expression while the second operand is typed', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await pressKeys(['One', 'Zero', 'Divide', 'Four'])

    expect(screen.getByText('10 ÷')).toBeInTheDocument()
    expect(display()).toHaveTextContent('4')
  })

  it('explains a division by zero instead of showing a code', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await pressKeys(['One', 'Zero', 'Divide', 'Zero', 'Equals'])

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent('Cannot divide by zero.')
  })

  it('explains the square root of a negative number', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await pressKeys(['Four', 'Subtract', 'Nine', 'Equals'])
    await waitFor(() => expect(display()).toHaveTextContent('-5'))

    await pressKeys(['Square root'])

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent('Cannot take the square root of a negative number.')
  })

  it('explains that the service is unreachable', async () => {
    stubRejection(new TypeError('Failed to fetch'))
    render(<Calculator />)

    await pressKeys(['One', 'Add', 'One', 'Equals'])

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent('Cannot reach the calculator service.')
  })

  it('clears the error once the next calculation succeeds', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await pressKeys(['One', 'Divide', 'Zero', 'Equals'])
    await screen.findByRole('alert')

    await pressKeys(['Clear', 'Two', 'Add', 'Two', 'Equals'])

    await waitFor(() => expect(display()).toHaveTextContent('4'))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('computes a percentage as "a% of b"', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await pressKeys(['One', 'Five', 'Percent of', 'Two', 'Zero', 'Zero', 'Equals'])

    await waitFor(() => expect(display()).toHaveTextContent('30'))
  })

  it('keeps a half-built expression when a square root is taken mid-way', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await pressKeys(['Nine', 'Add', 'One', 'Six', 'Square root'])
    await waitFor(() => expect(display()).toHaveTextContent('4'))
    expect(screen.getByText('9 +')).toBeInTheDocument()

    await pressKeys(['Equals'])
    await waitFor(() => expect(display()).toHaveTextContent('13'))
  })
})

describe('Calculator input constraints', () => {
  it('disables equals until the expression is complete', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    const equals = screen.getByRole('button', { name: 'Equals' })
    expect(equals).toBeDisabled()

    await pressKeys(['One', 'Add'])
    expect(equals).toBeDisabled()

    await pressKeys(['Two'])
    expect(equals).toBeEnabled()
  })

  it('disables the decimal point once the number already has one', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    const decimal = screen.getByRole('button', { name: 'Decimal point' })
    expect(decimal).toBeEnabled()

    await pressKeys(['One', 'Decimal point'])
    expect(decimal).toBeDisabled()
  })

  it('explains the digit limit rather than ignoring the press', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await pressKeys(Array.from({ length: MAX_ENTRY_DIGITS + 1 }, () => 'Nine'))

    expect(screen.getByText(MAX_ENTRY_MESSAGE)).toBeInTheDocument()
    expect(display()).toHaveTextContent('9'.repeat(MAX_ENTRY_DIGITS))
  })
})

describe('Calculator keyboard support', () => {
  it('accepts an expression typed on the keyboard', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await userEvent.keyboard('12+3{Enter}')

    await waitFor(() => expect(display()).toHaveTextContent('15'))
  })

  it('clears on Escape', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await userEvent.keyboard('123')
    expect(display()).toHaveTextContent('123')

    await userEvent.keyboard('{Escape}')
    expect(display()).toHaveTextContent('0')
  })

  it('uses the same key definitions as the on-screen keypad', async () => {
    stubCalculatorApi()
    render(<Calculator />)

    await userEvent.keyboard('81r')

    await waitFor(() => expect(display()).toHaveTextContent('9'))
  })
})
