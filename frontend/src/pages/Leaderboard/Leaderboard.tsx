import { useCallback, useEffect, useState, type ReactNode } from 'react'
import './Leaderboard.css'
import { api } from '../../api/client'
import type {
  DeltaLeaderboard,
  DeltaLeaderboardEntry,
  Leaderboard as LeaderboardData,
  LeaderboardEntry,
  StreakLeaderboard,
  StreakLeaderboardEntry,
} from '../../api/types'
import { useAuth } from '../../auth/AuthContext'
import { useGameEvent } from '../../realtime/RealtimeContext'

export default function Leaderboard() {
  const { user } = useAuth()
  const event = useGameEvent()
  const [total, setTotal] = useState<LeaderboardData | null>(null)
  const [week, setWeek] = useState<DeltaLeaderboard | null>(null)
  const [month, setMonth] = useState<DeltaLeaderboard | null>(null)
  const [streak, setStreak] = useState<StreakLeaderboard | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      const [totalValue, weekValue, monthValue, streakValue] = await Promise.all([
        api.getLeaderboard(),
        api.getWeeklyLeaderboard(),
        api.getMonthlyLeaderboard(),
        api.getStreakLeaderboard(),
      ])
      setTotal(totalValue)
      setWeek(weekValue)
      setMonth(monthValue)
      setStreak(streakValue)
      setError(null)
    } catch {
      setError('Не удалось загрузить лидерборды.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const timer = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(timer)
  }, [load])

  useEffect(() => {
    if (!event) {return}
    const timer = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(timer)
  }, [event, load])

  if (loading) {return <div className="leaderboard__page"><div className="page-state">Загружаем лидерборды...</div></div>}

  return (
    <div className="leaderboard__page">
      <div className="leaderboard__container">
        <header className="leaderboard__info"><h1 className="leaderboard__title">Лидерборд</h1><p className="leaderboard__description">Результаты игроков и активные серии</p></header>
        {error && <p className="leaderboard__error" role="alert">{error}</p>}
        {total && <TotalTable title="Общий рейтинг" description="Топ-10 по общему опыту" data={total} userId={user?.id} />}
        {week && <DeltaTable title="Последние 7 × 24 часов" description="Опыт, полученный за последние семь суток" data={week} userId={user?.id} />}
        {month && <DeltaTable title="Последние 30 × 24 часов" description="Опыт, полученный за последние тридцать суток" data={month} userId={user?.id} />}
        {streak && <StreakTable title="Активные серии" description="Топ-10 активных серий зарядок" data={streak} userId={user?.id} />}
      </div>
    </div>
  )
}

function BoardSection({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return <section className="leaderboard__section"><div><h2>{title}</h2><p>{description}</p></div>{children}</section>
}

function TotalTable({ title, description, data, userId }: { title: string; description: string; data: LeaderboardData; userId?: string }) {
  return (
    <BoardSection title={title} description={description}>
      <div className="leaderboard-wrapper"><table className="leaderboard">
        <thead><tr><th>Место</th><th>Игрок</th><th>Уровень</th><th>Общий XP</th></tr></thead>
        <tbody>
          {data.entries.length === 0 && <EmptyRow columns={4} />}
          {data.entries.map((player) => <TotalRow key={player.userId} player={player} current={player.userId === userId} />)}
        </tbody>
      </table></div>
      <CurrentUser>{data.currentUser ? <>Ваш результат: #{data.currentUser.rank}, {data.currentUser.displayName}, уровень {data.currentUser.level}, {formatNumber(data.currentUser.xp)} XP</> : <>Ваш результат в этом рейтинге пока отсутствует.</>}</CurrentUser>
    </BoardSection>
  )
}

function TotalRow({ player, current }: { player: LeaderboardEntry; current: boolean }) {
  return <tr className={current ? 'leaderboard__row--your-result' : ''}><td>{player.rank}</td><td>{player.displayName}</td><td>{player.level}</td><td>{formatNumber(player.xp)}</td></tr>
}

function DeltaTable({ title, description, data, userId }: { title: string; description: string; data: DeltaLeaderboard; userId?: string }) {
  return (
    <BoardSection title={title} description={description}>
      <div className="leaderboard-wrapper"><table className="leaderboard leaderboard--delta">
        <thead><tr><th>Место</th><th>Игрок</th><th>Уровень</th><th>Общий XP</th><th>XP за период</th></tr></thead>
        <tbody>
          {data.entries.length === 0 && <EmptyRow columns={5} />}
          {data.entries.map((player) => <DeltaRow key={player.userId} player={player} current={player.userId === userId} />)}
        </tbody>
      </table></div>
      <CurrentUser>{data.currentUser ? <>Ваш результат: #{data.currentUser.rank}, {data.currentUser.displayName}, уровень {data.currentUser.level}, {formatNumber(data.currentUser.xp)} XP всего, +{formatNumber(data.currentUser.xpDelta)} XP</> : <>Ваш результат в этом рейтинге пока отсутствует.</>}</CurrentUser>
    </BoardSection>
  )
}

function DeltaRow({ player, current }: { player: DeltaLeaderboardEntry; current: boolean }) {
  return <tr className={current ? 'leaderboard__row--your-result' : ''}><td>{player.rank}</td><td>{player.displayName}</td><td>{player.level}</td><td>{formatNumber(player.xp)}</td><td>+{formatNumber(player.xpDelta)}</td></tr>
}

function StreakTable({ title, description, data, userId }: { title: string; description: string; data: StreakLeaderboard; userId?: string }) {
  return (
    <BoardSection title={title} description={description}>
      <div className="leaderboard-wrapper"><table className="leaderboard leaderboard--streak">
        <thead><tr><th>Место</th><th>Игрок</th><th>Текущая серия</th><th>Рекорд</th><th>Начало серии</th><th>Последняя зарядка</th></tr></thead>
        <tbody>
          {data.entries.length === 0 && <EmptyRow columns={6} />}
          {data.entries.map((player) => <StreakRow key={player.userId} player={player} current={player.userId === userId} />)}
        </tbody>
      </table></div>
      <CurrentUser>{data.currentUser ? <>Ваш результат: #{data.currentUser.rank}, {data.currentUser.displayName}, серия {data.currentUser.currentStreak}, рекорд {data.currentUser.longestStreak}, с {formatDate(data.currentUser.streakStartedAt)}, последняя зарядка {formatDate(data.currentUser.lastChargeDate)}</> : <>У вас пока нет активной серии.</>}</CurrentUser>
    </BoardSection>
  )
}

function StreakRow({ player, current }: { player: StreakLeaderboardEntry; current: boolean }) {
  return <tr className={current ? 'leaderboard__row--your-result' : ''}><td>{player.rank}</td><td>{player.displayName}</td><td>{player.currentStreak}</td><td>{player.longestStreak}</td><td>{formatDate(player.streakStartedAt)}</td><td>{formatDate(player.lastChargeDate)}</td></tr>
}

function EmptyRow({ columns }: { columns: number }) {
  return <tr><td className="leaderboard__empty" colSpan={columns}>В рейтинге пока нет участников.</td></tr>
}

function CurrentUser({ children }: { children: ReactNode }) {
  return <p className="leaderboard__current-user">{children}</p>
}

function formatNumber(value: number) {
  return value.toLocaleString('ru-RU')
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('ru-RU').format(new Date(`${value}T00:00:00`))
}
