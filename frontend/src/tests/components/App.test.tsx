import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import App from '../../App'
import { stubJsonResponse } from '../fetchStub'

describe('App', () => {
  it('presents the calculator with its heading and keyboard guidance', () => {
    stubJsonResponse({ operation: 'add', a: 1, b: 1, result: 2 })
    render(<App />)

    expect(screen.getByRole('heading', { level: 1, name: 'Calculator' })).toBeInTheDocument()
    expect(screen.getByRole('group', { name: 'Calculator keypad' })).toBeInTheDocument()
    expect(screen.getByText(/Enter/)).toBeInTheDocument()
  })
})
