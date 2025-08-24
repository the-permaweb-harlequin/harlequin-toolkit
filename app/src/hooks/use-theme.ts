import { useEffect, useState } from 'react'
import useLocalStorageState from './use-localstorage-state'

type Theme = 'light' | 'dark'

export function useTheme() {
  const [theme, setTheme] = useLocalStorageState<Theme>('theme', 'light')
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    const root = document.documentElement
    root.classList.remove('light', 'dark')
    root.classList.add(theme)
  }, [theme])

  const toggleTheme = () => {
    setTheme(theme === 'light' ? 'dark' : 'light')
  }

  // Return light as default during SSR to avoid hydration mismatch
  return {
    theme: mounted ? theme : 'light',
    toggleTheme,
    setTheme,
  }
}
