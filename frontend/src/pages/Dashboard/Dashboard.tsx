import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import './Dashboard.css'
import { ApiError, api } from '../../api/client'
import type { DeltaLeaderboard, Leaderboard, Pet, Reward, Streak, Summary } from '../../api/types'
import GameOverModal from '../../components/GameOverModal/GameOverModal'
import PetVisual, { stageLabel } from '../../components/PetVisual/PetVisual'
import { useGameEvent } from '../../realtime/RealtimeContext'

type PetAction = 'charge' | 'pet'

export default function Dashboard() {
  const event = useGameEvent()
  const [pet, setPet] = useState<Pet | null>(null)
  const [summary, setSummary] = useState<Summary | null>(null)
  const [leaderboard, setLeaderboard] = useState<Leaderboard | null>(null)
  const [weeklyLeaderboard, setWeeklyLeaderboard] = useState<DeltaLeaderboard | null>(null)
  const [streak, setStreak] = useState<Streak | null>(null)
  const [rewards, setRewards] = useState<Reward[]>([])
  const [loading, setLoading] = useState(true)
  const [pendingAction, setPendingAction] = useState<PetAction | null>(null)
  const [showGameOver, setShowGameOver] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const previousDead = useRef<boolean | undefined>(undefined)

  const acceptPet = useCallback((value: Pet) => {
    if (value.isDead && previousDead.current !== true) {
      setShowGameOver(true)
    }
    previousDead.current = value.isDead
    setPet(value)
  }, [])

  const load = useCallback(async (showLoading = false) => {
    if (showLoading) {setLoading(true)}
    setError(null)
    try {
      const [petValue, summaryValue, leaderboardValue, weeklyValue, streakValue, rewardValues] = await Promise.all([
        api.getPet(),
        api.getSummary(),
        api.getLeaderboard(),
        api.getWeeklyLeaderboard(),
        api.getStreak(),
        api.getRewards(),
      ])
      acceptPet(petValue)
      setSummary(summaryValue)
      setLeaderboard(leaderboardValue)
      setWeeklyLeaderboard(weeklyValue)
      setStreak(streakValue)
      setRewards(rewardValues)
    } catch {
      setError('Не удалось загрузить состояние питомца.')
    } finally {
      setLoading(false)
    }
  }, [acceptPet])

  useEffect(() => {
    const timer = window.setTimeout(() => { void load(true) }, 0)
    return () => window.clearTimeout(timer)
  }, [load])

  useEffect(() => {
    if (!event) {return}
    const timer = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(timer)
  }, [event, load])

  useEffect(() => {
    const timer = window.setInterval(() => {
      void api.getPet().then(acceptPet).catch(() => {
        // A later refresh or WebSocket event will retry; keep the last backend state meanwhile.
      })
    }, 30000)
    return () => window.clearInterval(timer)
  }, [acceptPet])

  const availableRewards = useMemo(() => rewards.filter((reward) => reward.status === 'available').length, [rewards])
  const levelProgress = pet ? pet.xp % 100 : 0

  const refreshDeadPet = async () => {
    try {
      const value = await api.getPet()
      acceptPet(value)
      if (value.isDead) {setShowGameOver(true)}
    } catch {
      setError('Не удалось обновить состояние питомца.')
    }
  }

  const runAction = async (action: PetAction) => {
    if (pendingAction !== null) {return}
    setPendingAction(action)
    setError(null)
    try {
      const result = action === 'charge' ? await api.chargePet() : await api.petPet()
      acceptPet(result.pet)
      await load()
    } catch (caught) {
      if (caught instanceof ApiError && caught.status === 409 && caught.code === 'pet_dead') {
        await refreshDeadPet()
      } else {
        setError(action === 'charge' ? 'Не удалось зарядить питомца.' : 'Не удалось погладить питомца.')
      }
    } finally {
      setPendingAction(null)
    }
  }

  if (loading) {return <div className="dashboard-page"><div className="page-state">Загружаем питомца...</div></div>}
  if (!pet || !summary) {return <div className="dashboard-page"><div className="page-state page-state--error">{error ?? 'Данные питомца недоступны.'}</div></div>}

  return (
    <div className="dashboard-page">
      <div className="dashboard">
        {error && <p className="dashboard__error" role="alert">{error}</p>}
        <section className="dashboard__pet-card">
          <div className="dashboard__pet-image">
            <PetVisual stage={pet.stage} />
          </div>
          <div className="dashboard__info">
            <div className="dashboard__pet-name">K1-T4</div>
            <div className="dashboard__pet-phase">{stageLabel(pet.stage)}, уровень {pet.level}</div>
            <div className="dashboard__progress-bars">
              <Progress label="Опыт" value={`${levelProgress} / 100 XP`} percent={levelProgress} kind="xp" />
              <Progress label="Зарядка" value={`${pet.energy}%`} percent={pet.energy} kind="battery" />
            </div>
            <div className="dashboard__actions">
              <button className="dashboard__action dashboard__action--pet" type="button" disabled={pendingAction !== null} onClick={() => { void runAction('pet') }}>
                {pendingAction === 'pet' ? 'Гладим...' : 'Погладить'}
              </button>
              <button className="dashboard__action dashboard__action--charge" type="button" disabled={pendingAction !== null} onClick={() => { void runAction('charge') }}>
                {pendingAction === 'charge' ? 'Заряжаем...' : 'Зарядить'}
              </button>
            </div>
          </div>
        </section>

        <section className="dashboard__stats" aria-label="Игровая статистика">
          <Stat title="Задач за день" value={`${summary.completedTasks} / 3`} accent="blue" />
          <Stat title="Опыта за день" value={`${summary.xpEarned} XP`} accent="purple" />
          <Stat title="Доступно наград" value={String(availableRewards)} accent="green" />
          <Stat title="Место в лидерборде" value={leaderboard?.currentUser ? `#${String(leaderboard.currentUser.rank)}` : 'Нет данных'} accent="pink" />
          <Stat title="Серия зарядок" value={streak ? `${streak.currentStreak}, рекорд ${streak.longestStreak}` : 'Нет данных'} accent="orange" />
          <Stat title="Опыта за 7 дней" value={`${weeklyLeaderboard?.currentUser?.xpDelta ?? 0} XP`} accent="cyan" />
        </section>
      </div>
      {showGameOver && <GameOverModal onClose={() => setShowGameOver(false)} />}
    </div>
  )
}

function Progress({ label, value, percent, kind }: { label: string; value: string; percent: number; kind: 'xp' | 'battery' }) {
  return (
    <div className="dashboard__progress">
      <div className="dashboard__progress-info"><span>{label}</span><span>{value}</span></div>
      <div className="dashboard__progress-bar"><div className={`dashboard__progress-fill dashboard__progress-${kind}`} style={{ width: `${percent}%` }} /></div>
    </div>
  )
}

function Stat({ title, value, accent }: { title: string; value: string; accent: string }) {
  return <article className={`dashboard__stat-card dashboard__stat-card--${accent}`}><h2>{title}</h2><p>{value}</p></article>
}
