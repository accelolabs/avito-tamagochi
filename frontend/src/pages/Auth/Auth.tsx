import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import './Auth.css'
import { ApiError } from '../../api/client'
import { useAuth } from '../../auth/AuthContext'

const displayNamePattern = /^[A-Za-z_-]+$/

function errorMessage(error: unknown) {
  if (error instanceof ApiError) {
    if (error.code === 'email_already_exists') {return 'Этот email уже зарегистрирован.'}
    if (error.code === 'invalid_credentials') {return 'Неверный email или пароль.'}
    if (error.code === 'validation_error') {return 'Проверьте введённые данные.'}
  }
  return 'Не удалось выполнить запрос. Попробуйте ещё раз.'
}

export default function Auth() {
  const [isRegistration, setIsRegistration] = useState(true)
  const { initializationError } = useAuth()

  return (
    <div className="auth__page">
      <div className="auth__container">
        <header className="auth__header">
          <div className="auth__logo" aria-hidden="true">🐶</div>
          <h1>Авито Тамагочи</h1>
          <p className="auth__catch-phrase">Твой виртуальный питомец в мире Авито</p>
        </header>

        <div className="auth__toggle">
          <button type="button" className={`auth__toggle-button ${isRegistration ? 'auth__toggle-button--active' : ''}`} onClick={() => setIsRegistration(true)}>
            Регистрация
          </button>
          <button type="button" className={`auth__toggle-button ${!isRegistration ? 'auth__toggle-button--active' : ''}`} onClick={() => setIsRegistration(false)}>
            Вход
          </button>
          <div className={`auth__toggle-slider ${!isRegistration ? 'auth__toggle-slider--right' : ''}`} />
        </div>

        {initializationError && <p className="auth__error" role="alert">{initializationError}</p>}
        {isRegistration
          ? <RegistrationForm switchToLogin={() => setIsRegistration(false)} />
          : <LoginForm switchToRegistration={() => setIsRegistration(true)} />}

        <footer className="auth__footer">2026 Avito Tamagotchi</footer>
      </div>
    </div>
  )
}

function RegistrationForm({ switchToLogin }: { switchToLogin: () => void }) {
  const { register } = useAuth()
  const navigate = useNavigate()
  const [displayName, setDisplayName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError(null)
    if (!displayNamePattern.test(displayName) || displayName.length < 2 || displayName.length > 32) {
      setError('Имя: от 2 до 32 латинских букв, дефис или подчёркивание.')
      return
    }
    setSubmitting(true)
    try {
      await register({ displayName, email, password })
      void navigate('/', { replace: true })
    } catch (requestError) {
      setError(errorMessage(requestError))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="auth__form-container">
      <form className="auth__form" onSubmit={(event) => { void submit(event) }}>
        <AuthField label="Имя пользователя" type="text" value={displayName} onChange={setDisplayName} placeholder="Avito_User" autoComplete="username" />
        <AuthField label="Электронная почта" type="email" value={email} onChange={setEmail} placeholder="mail@example.ru" autoComplete="email" />
        <AuthField label="Пароль" type="password" value={password} onChange={setPassword} placeholder="Не менее 8 символов" autoComplete="new-password" minLength={8} maxLength={128} />
        {error && <p className="auth__error" role="alert">{error}</p>}
        <button className="auth__submit" disabled={submitting}>{submitting ? 'Регистрируем...' : 'Зарегистрироваться'}</button>
      </form>
      <div className="auth__form-footer">
        Уже есть аккаунт? <button type="button" className="auth__footer-link" onClick={switchToLogin}>Войти</button>
      </div>
    </div>
  )
}

function LoginForm({ switchToRegistration }: { switchToRegistration: () => void }) {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await login({ email, password })
      void navigate('/', { replace: true })
    } catch (requestError) {
      setError(errorMessage(requestError))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="auth__form-container">
      <form className="auth__form" onSubmit={(event) => { void submit(event) }}>
        <AuthField label="Электронная почта" type="email" value={email} onChange={setEmail} placeholder="mail@example.ru" autoComplete="email" />
        <AuthField label="Пароль" type="password" value={password} onChange={setPassword} placeholder="Ваш пароль" autoComplete="current-password" minLength={8} maxLength={128} />
        {error && <p className="auth__error" role="alert">{error}</p>}
        <button className="auth__submit" disabled={submitting}>{submitting ? 'Входим...' : 'Войти'}</button>
      </form>
      <div className="auth__form-footer">
        Нет аккаунта? <button type="button" className="auth__footer-link" onClick={switchToRegistration}>Зарегистрироваться</button>
      </div>
    </div>
  )
}

interface AuthFieldProps {
  label: string
  type: 'text' | 'email' | 'password'
  value: string
  onChange: (value: string) => void
  placeholder: string
  autoComplete: string
  minLength?: number
  maxLength?: number
}

function AuthField({ label, type, value, onChange, placeholder, autoComplete, minLength, maxLength }: AuthFieldProps) {
  const id = `auth-${label.toLowerCase().replaceAll(' ', '-')}`
  return (
    <div className="auth__field">
      <label htmlFor={id}>{label}</label>
      <input id={id} type={type} value={value} onChange={(event) => onChange(event.target.value)} placeholder={placeholder} autoComplete={autoComplete} minLength={minLength} maxLength={maxLength} required />
    </div>
  )
}
