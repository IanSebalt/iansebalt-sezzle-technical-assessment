import { renderHook } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { useKeyboard } from '../../hooks/useKeyboard'

describe('useKeyboard', () => {
  it('reports a mapped key press', async () => {
    const onPress = vi.fn()
    renderHook(() => useKeyboard(onPress))

    await userEvent.keyboard('7')

    expect(onPress).toHaveBeenCalledExactlyOnceWith('7')
  })

  it('ignores keys the keypad does not define', async () => {
    const onPress = vi.fn()
    renderHook(() => useKeyboard(onPress))

    await userEvent.keyboard('{F5}{Tab}z')

    expect(onPress).not.toHaveBeenCalled()
  })

  it('leaves browser and system shortcuts alone', async () => {
    const onPress = vi.fn()
    renderHook(() => useKeyboard(onPress))

    await userEvent.keyboard('{Control>}1{/Control}')
    await userEvent.keyboard('{Meta>}4{/Meta}')
    await userEvent.keyboard('{Alt>}7{/Alt}')

    expect(onPress).not.toHaveBeenCalled()
  })

  it('stops listening once unmounted', async () => {
    const onPress = vi.fn()
    const { unmount } = renderHook(() => useKeyboard(onPress))

    unmount()
    await userEvent.keyboard('7')

    expect(onPress).not.toHaveBeenCalled()
  })
})
