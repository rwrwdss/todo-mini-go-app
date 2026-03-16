import { useState } from 'react'
import { login, register, saveAuth } from '../api/auth'

const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export default function AuthPage({ onSuccess }) {
  const [tab, setTab] = useState('login')
  const [loginEmail, setLoginEmail] = useState('')
  const [loginPassword, setLoginPassword] = useState('')
  const [regName, setRegName] = useState('')
  const [regEmail, setRegEmail] = useState('')
  const [regPassword, setRegPassword] = useState('')
  const [regPassword2, setRegPassword2] = useState('')
  const [regTerms, setRegTerms] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState({})

  function clearErrors() {
    setError('')
    setFieldErrors({})
  }

  function handleLogin(e) {
    e?.preventDefault()
    clearErrors()
    const err = {}
    if (!emailRe.test(loginEmail.trim())) err.loginEmail = 'Enter a valid email'
    if (!loginPassword) err.loginPw = 'Password required'
    if (Object.keys(err).length) {
      setFieldErrors(err)
      return
    }
    setLoading(true)
    login(loginEmail.trim().toLowerCase(), loginPassword)
      .then((data) => {
        saveAuth(data.token, data.user)
        onSuccess?.()
      })
      .catch((e) => {
        setError(e.message || 'Login failed')
      })
      .finally(() => setLoading(false))
  }

  function handleRegister(e) {
    e?.preventDefault()
    clearErrors()
    const err = {}
    if (regName.trim().length < 2) err.regName = 'Name required (2+ chars)'
    if (!emailRe.test(regEmail.trim())) err.regEmail = 'Enter a valid email'
    if (regPassword.length < 8) err.regPw = 'Min. 8 characters'
    if (regPassword !== regPassword2) err.regPw2 = 'Passwords do not match'
    if (!regTerms) err.regTerms = 'Accept the terms'
    if (Object.keys(err).length) {
      setFieldErrors(err)
      return
    }
    setLoading(true)
    register(regEmail.trim().toLowerCase(), regPassword, regName.trim())
      .then((data) => {
        saveAuth(data.token, data.user)
        onSuccess?.()
      })
      .catch((e) => {
        setError(e.message || 'Registration failed')
      })
      .finally(() => setLoading(false))
  }

  return (
    <div className="auth-page">
      <div className="auth-card-wrap">
        <div className="auth-card">
          <div className="auth-tabs">
            <div
              className={`auth-tab-slider ${tab === 'register' ? 'right' : ''}`}
              style={{ width: '50%' }}
            />
            <button
              type="button"
              className={`auth-tab ${tab === 'login' ? 'active' : ''}`}
              onClick={() => { setTab('login'); clearErrors(); }}
            >
              Sign in
            </button>
            <button
              type="button"
              className={`auth-tab ${tab === 'register' ? 'active' : ''}`}
              onClick={() => { setTab('register'); clearErrors(); }}
            >
              Create account
            </button>
          </div>

          {error && <div className="auth-msg">{error}</div>}

          {tab === 'login' ? (
            <form onSubmit={handleLogin}>
              <div className="auth-form-group">
                <label className="auth-label">Email</label>
                <input
                  type="email"
                  className={`auth-fi ${fieldErrors.loginEmail ? 'error' : ''}`}
                  placeholder="you@example.com"
                  value={loginEmail}
                  onChange={(e) => setLoginEmail(e.target.value)}
                  autoComplete="email"
                />
                {fieldErrors.loginEmail && (
                  <div className="auth-field-err">{fieldErrors.loginEmail}</div>
                )}
              </div>
              <div className="auth-form-group">
                <label className="auth-label">Password</label>
                <input
                  type="password"
                  className={`auth-fi ${fieldErrors.loginPw ? 'error' : ''}`}
                  placeholder="••••••••"
                  value={loginPassword}
                  onChange={(e) => setLoginPassword(e.target.value)}
                  autoComplete="current-password"
                />
                {fieldErrors.loginPw && (
                  <div className="auth-field-err">{fieldErrors.loginPw}</div>
                )}
              </div>
              <button
                type="submit"
                className="auth-btn-submit"
                disabled={loading}
              >
                {loading ? 'Signing in…' : 'Sign in'}
              </button>
              <div className="auth-footer">
                No account?{' '}
                <button type="button" className="auth-footer a" onClick={() => { setTab('register'); clearErrors(); }} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0 }}>
                  Create one
                </button>
              </div>
            </form>
          ) : (
            <form onSubmit={handleRegister}>
              <div className="auth-form-group">
                <label className="auth-label">Name</label>
                <input
                  type="text"
                  className={`auth-fi ${fieldErrors.regName ? 'error' : ''}`}
                  placeholder="Your name"
                  value={regName}
                  onChange={(e) => setRegName(e.target.value)}
                  autoComplete="name"
                />
                {fieldErrors.regName && (
                  <div className="auth-field-err">{fieldErrors.regName}</div>
                )}
              </div>
              <div className="auth-form-group">
                <label className="auth-label">Email</label>
                <input
                  type="email"
                  className={`auth-fi ${fieldErrors.regEmail ? 'error' : ''}`}
                  placeholder="you@example.com"
                  value={regEmail}
                  onChange={(e) => setRegEmail(e.target.value)}
                  autoComplete="email"
                />
                {fieldErrors.regEmail && (
                  <div className="auth-field-err">{fieldErrors.regEmail}</div>
                )}
              </div>
              <div className="auth-form-group">
                <label className="auth-label">Password</label>
                <input
                  type="password"
                  className={`auth-fi ${fieldErrors.regPw ? 'error' : ''}`}
                  placeholder="Min. 8 characters"
                  value={regPassword}
                  onChange={(e) => setRegPassword(e.target.value)}
                  autoComplete="new-password"
                />
                {fieldErrors.regPw && (
                  <div className="auth-field-err">{fieldErrors.regPw}</div>
                )}
              </div>
              <div className="auth-form-group">
                <label className="auth-label">Confirm password</label>
                <input
                  type="password"
                  className={`auth-fi ${fieldErrors.regPw2 ? 'error' : ''}`}
                  placeholder="Repeat password"
                  value={regPassword2}
                  onChange={(e) => setRegPassword2(e.target.value)}
                  autoComplete="new-password"
                />
                {fieldErrors.regPw2 && (
                  <div className="auth-field-err">{fieldErrors.regPw2}</div>
                )}
              </div>
              <div className="auth-form-group">
                <label style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={regTerms}
                    onChange={(e) => setRegTerms(e.target.checked)}
                  />
                  <span className="auth-footer">I agree to the terms</span>
                </label>
                {fieldErrors.regTerms && (
                  <div className="auth-field-err">{fieldErrors.regTerms}</div>
                )}
              </div>
              <button
                type="submit"
                className="auth-btn-submit"
                disabled={loading}
              >
                {loading ? 'Creating account…' : 'Create account'}
              </button>
              <div className="auth-footer">
                Already have an account?{' '}
                <button type="button" className="auth-footer a" onClick={() => { setTab('login'); clearErrors(); }} style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0 }}>
                  Sign in
                </button>
              </div>
            </form>
          )}
        </div>
      </div>
    </div>
  )
}
