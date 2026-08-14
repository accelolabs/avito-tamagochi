import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import './Sidebar.css'
import { useAuth } from '../auth/AuthContext'
import { RealtimeProvider } from '../realtime/RealtimeContext'
import RainbowText from '../components/RainbowText/RainbowText'

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
        <div>
          <h1 className="sidebar__title">
            <svg width="152" height="50" viewBox="0 0 152 50" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path fill-rule="evenodd" clip-rule="evenodd" d="M81.1211 0.981525C78.7935 2.55382 78.0468 3.88167 78.0665 6.41559C78.117 12.8855 86.7086 15.2183 90.1078 9.68493C91.2235 7.86948 91.2235 4.91501 90.1078 3.09956C88.2438 0.0653091 83.9691 -0.942145 81.1211 0.981525ZM18.2813 1.69065C18.032 2.15363 13.8993 12.7103 9.09828 25.1503C4.29724 37.5902 0.235155 48.1087 0.0709441 48.525C-0.206162 49.2286 0.125681 49.2732 4.83221 49.1615L9.89154 49.0414L11.6641 44.3734L13.4366 39.7053L23.4792 39.5916L33.5221 39.4783L35.3656 44.3657L37.2096 49.2536H42.0624C44.7312 49.2536 46.9147 49.1093 46.9147 48.9328C46.9147 48.7567 42.809 37.9191 37.7912 24.8498L28.6672 1.08762L23.7011 0.968369C19.1199 0.858033 18.7004 0.91405 18.2813 1.69065ZM97.8031 10.8481V15.304H95.2373H92.6715V19.3355V23.3671H95.1894H97.7077L97.9177 32.3849C98.0361 37.4651 98.347 42.1026 98.6301 43.0056C100.233 48.1197 106.702 50.8722 112.908 49.0818L115.336 48.3811V44.4353V40.4891L112.308 40.5214C109.52 40.5515 109.199 40.462 108.272 39.3951C106.982 37.9089 106.707 35.8974 106.867 29.096L106.997 23.5792L111.182 23.4574L115.368 23.3361L115.245 19.4259L115.122 15.5162L110.953 15.3948L106.783 15.2735V10.8329V6.39224H102.293H97.8031V10.8481ZM20.2805 21.9878L17.0591 30.3692L23.2799 30.4871C26.7014 30.5516 29.6029 30.5033 29.7282 30.3793C29.9184 30.1905 23.8482 13.6065 23.5886 13.6065C23.5412 13.6065 22.0526 17.3783 20.2805 21.9878ZM129.829 15.4674C123.507 17.2854 118.485 22.4305 117.063 28.5482C114.816 38.2145 121.394 47.8978 131.471 49.7574C136.706 50.7237 142.84 48.7482 146.792 44.8236C155.562 36.1156 152.893 21.7799 141.511 16.4642C138.31 14.9692 133.107 14.5253 129.829 15.4674ZM41.9486 16.2589C42.0906 16.7838 44.9673 24.3749 48.3413 33.1276L54.4761 49.0414L59.355 49.1492L64.2339 49.2575L70.4345 32.6879C73.8449 23.575 76.7015 15.9291 76.7828 15.697C76.8429 15.5246 68.3968 15.489 67.312 15.5162L63.4209 25.7771C61.2811 31.4208 59.417 36.1478 59.2793 36.2815C59.1412 36.4156 57.2339 31.7501 55.0402 25.9145L51.0521 15.304H46.3712C41.7258 15.304 41.6925 15.3112 41.9486 16.2589ZM79.8425 32.2788V49.2536H84.3326H88.8228V32.2788V15.304H84.3326H79.8425V32.2788ZM131.372 24.4292C130.431 24.7594 128.953 25.78 128.088 26.6967C123.531 31.523 125.856 39.1986 132.343 40.7429C139.365 42.4145 145.266 34.9638 141.947 28.6191C139.83 24.5718 135.666 22.9219 131.372 24.4292Z" fill="black"/>
            </svg>
            <RainbowText text="Тамагочи" /></h1>
          <p className="sidebar__user">@{user?.displayName}</p>
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
