import { StrictMode, useState, useCallback } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'
import AuthPage from './components/AuthPage.jsx'
import { logout } from './api/auth'

function Root() {
  const [token, setToken] = useState(() => localStorage.getItem('token'))

  const onAuthSuccess = useCallback(() => {
    setToken(localStorage.getItem('token'))
  }, [])

  const onLogout = useCallback(() => {
    logout()
    setToken(null)
  }, [])

  if (!token) {
    return <AuthPage onSuccess={onAuthSuccess} />
  }
  return <App onLogout={onLogout} />
}

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <Root />
  </StrictMode>,
)
