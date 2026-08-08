import { createContext, useContext, useEffect, useState } from 'react'

const ThemeContext = createContext(null)

export function ThemeProvider({ children }) {
  const [dark, setDark] = useState(() => {
    const stored = localStorage.getItem('theme')
    // Only honour an explicit 'light' choice if the user toggled it themselves
    // (i.e. 'light' was written by this app, not inherited from an old session)
    if (stored === 'light' && localStorage.getItem('theme-user-set') === 'true') return false
    return true // dark by default
  })

  useEffect(() => {
    const root = document.documentElement
    if (dark) {
      root.classList.add('dark')
      localStorage.setItem('theme', 'dark')
      localStorage.setItem('theme-user-set', 'true')
    } else {
      root.classList.remove('dark')
      localStorage.setItem('theme', 'light')
      localStorage.setItem('theme-user-set', 'true')
    }
  }, [dark])

  return (
    <ThemeContext.Provider value={{ dark, toggle: () => setDark((d) => !d) }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used inside <ThemeProvider>')
  return ctx
}
