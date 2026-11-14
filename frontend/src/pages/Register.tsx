
import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  Activity,
  User,
  Mail,
  Lock,
  ArrowRight,
  CheckCircle2,
} from 'lucide-react'
import { useAuth } from '../context/AuthContext'

export default function Register() {
  const { register } = useAuth()
  const navigate = useNavigate()

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [registered, setRegistered] = useState(false)

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      await register(name.trim(), email.trim(), password)
      setRegistered(true)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Registration failed',
      )
    } finally {
      setLoading(false)
    }
  }

  if (registered) {
    return (
      <main className="auth-page">
        <div className="auth-card">
          <Link to="/" className="brand">
            <span className="brand-icon">
              <Activity size={22} />
            </span>
            MyPoll
          </Link>

          <div className="auth-heading">
            <span className="eyebrow">REGISTRATION COMPLETE</span>
            <h1>Account created successfully!</h1>
            <p>
              Your MyPoll account is ready. Please sign in with your
              registered email and password to continue.
            </p>
          </div>

          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '12px',
              padding: '15px',
              marginBottom: '18px',
              borderRadius: '10px',
              background: '#effaf4',
              color: '#21875c',
              fontSize: '13px',
            }}
          >
            <CheckCircle2 size={22} />
            <span>Registration successful. Please sign in.</span>
          </div>

          <button
            type="button"
            className="primary-button"
            onClick={() => navigate('/login', { replace: true })}
          >
            Continue to Sign In
            <ArrowRight size={18} />
          </button>
        </div>
      </main>
    )
  }

  return (
    <main className="auth-page">
      <div className="auth-card">
        <Link to="/" className="brand">
          <span className="brand-icon">
            <Activity size={22} />
          </span>
          MyPoll
        </Link>

        <div className="auth-heading">
          <span className="eyebrow">GET STARTED</span>
          <h1>Create your MyPoll account</h1>
          <p>
            Create your account to start building polls and collecting
            real-time opinions.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="auth-form">
          <label htmlFor="register-name">Full name</label>
          <div className="field">
            <User size={18} />
            <input
              id="register-name"
              type="text"
              placeholder="Your name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoComplete="name"
              required
            />
          </div>

          <label htmlFor="register-email">Email address</label>
          <div className="field">
            <Mail size={18} />
            <input
              id="register-email"
              type="email"
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              required
            />
          </div>

          <label htmlFor="register-password">Password</label>
          <div className="field">
            <Lock size={18} />
            <input
              id="register-password"
              type="password"
              placeholder="Create a password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              minLength={8}
              autoComplete="new-password"
              required
            />
          </div>

          {error && <p className="form-error">{error}</p>}

          <button
            type="submit"
            className="primary-button"
            disabled={loading}
          >
            {loading ? 'Creating account...' : 'Create account'}
            {!loading && <ArrowRight size={18} />}
          </button>
        </form>

        <p className="auth-footer">
          Already registered? <Link to="/login">Sign in</Link>
        </p>
      </div>
    </main>
  )
}
