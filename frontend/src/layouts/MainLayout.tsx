import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import './Sidebar.css'
import { useAuth } from '../auth/AuthContext'
import { RealtimeProvider } from '../realtime/RealtimeContext'

const navigation = [
  { to: '/', label: 'Питомец', icon: '🐾', end: true },
  { to: '/tasks', label: 'Задания', icon: '✓' },
  { to: '/rewards', label: 'Награды', icon: '★' },
  { to: '/leaderboard', label: 'Лидерборд', icon: '▥' },
]

function Sidebar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [loggingOut, setLoggingOut] = useState(false)

  const handleLogout = async () => {
    setLoggingOut(true)
    try {
      await logout()
      navigate('/auth', { replace: true })
    } finally {
      setLoggingOut(false)
    }
  }

  return (
    <nav className="sidebar">
      <header className="sidebar__header">
        <span className="sidebar__logo" aria-hidden="true">🐶</span>
        <div>
          <h1 className="sidebar__title">Авито Тамагочи</h1>
          <p className="sidebar__user">{user?.displayName}</p>
        </div>
      </header>

      <div className="sidebar__links">
        {navigation.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.end}
            className={({ isActive }) => `sidebar__link ${isActive ? 'sidebar__link--active' : ''}`}
          >
            <span className="sidebar__link-icon" aria-hidden="true">{item.icon}</span>
            {item.label}
          </NavLink>
        ))}
      </div>

      <footer className="sidebar__footer">
        <button className="sidebar__exit" type="button" disabled={loggingOut} onClick={handleLogout}>
          <span aria-hidden="true">↪</span>
          {loggingOut ? 'Выходим…' : 'Выйти'}
        </button>
      </footer>
    </nav>
  )
}

export default function MainLayout() {
  return (
    <RealtimeProvider>
      <Sidebar />
      <main>
        <Outlet />
      </main>
    </RealtimeProvider>
  )
}
