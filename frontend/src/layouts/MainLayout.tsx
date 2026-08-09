import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import './Sidebar.css'
import { useAuth } from '../auth/AuthContext'
import { RealtimeProvider } from '../realtime/RealtimeContext'

const navigation = [
  { to: '/', label: 'Питомец', icon: '/pet.svg', end: true },
  { to: '/tasks', label: 'Задания', icon: '/task.svg' },
  { to: '/rewards', label: 'Награды', icon: '/reward.svg' },
  { to: '/leaderboard', label: 'Лидерборд', icon: '/leaderboard.svg' },
]

function Sidebar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [loggingOut, setLoggingOut] = useState(false)

  const handleLogout = async () => {
    setLoggingOut(true)
    try {
      await logout()
      void navigate('/auth', { replace: true })
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
            <span
              className="sidebar__link-icon"
              aria-hidden="true"
              style={{ maskImage: `url(${item.icon})`, WebkitMaskImage: `url(${item.icon})` }}
            />
            {item.label}
          </NavLink>
        ))}
      </div>

      <footer className="sidebar__footer">
        <button className="sidebar__exit" type="button" disabled={loggingOut} onClick={() => { void handleLogout() }}>
          <span
            className="sidebar__exit-icon"
            aria-hidden="true"
            style={{ maskImage: 'url(/logout.svg)', WebkitMaskImage: 'url(/logout.svg)' }}
          />
          {loggingOut ? 'Выходим...' : 'Выйти'}
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
