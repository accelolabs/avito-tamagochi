import { useCallback, useEffect, useState } from 'react'
import './Leaderboard.css'
import { api } from '../../api/client'
import type { Leaderboard as LeaderboardData } from '../../api/types'
import { useAuth } from '../../auth/AuthContext'
import { useGameEvent } from '../../realtime/RealtimeContext'

export default function Leaderboard() {
  const { user } = useAuth()
  const event = useGameEvent()
  const [data, setData] = useState<LeaderboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setData(await api.getLeaderboard())
      setError(null)
    } catch {
      setError('Не удалось загрузить лидерборд.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const timer = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(timer)
  }, [load])
  useEffect(() => {
    if (event?.name !== 'pet_updated') {return}
    const timer = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(timer)
  }, [event, load])

  if (loading) {return <div className="leaderboard__page"><div className="page-state">Загружаем лидерборд…</div></div>}

  return (
    <div className="leaderboard__page">
      <div className="leaderboard__container">
        <header className="leaderboard__info"><h1 className="leaderboard__title">Лидерборд</h1><p className="leaderboard__description">Топ-10 игроков по общему опыту</p></header>
        {error && <p className="leaderboard__error" role="alert">{error}</p>}
        {data && (
          <div className="leaderboard-wrapper">
            <table className="leaderboard">
              <thead><tr><th>Место</th><th>Игрок</th><th>Уровень</th><th>Опыт (XP)</th></tr></thead>
              <tbody>
                {data.entries.map((player) => (
                  <tr key={player.userId} className={player.userId === user?.id ? 'leaderboard__row--your-result' : ''}>
                    <td>{player.rank}</td><td>{player.displayName}</td><td>{player.level}</td><td>{player.xp.toLocaleString('ru-RU')}</td>
                  </tr>
                ))}
              </tbody>
              <tfoot><tr><td colSpan={4}>Ваше место: {data.currentUser?.rank ?? 'ещё не определено'} · {data.currentUser?.xp ?? 0} XP</td></tr></tfoot>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
