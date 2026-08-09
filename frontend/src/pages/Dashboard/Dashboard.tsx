import { useCallback, useEffect, useMemo, useState } from 'react'
import './Dashboard.css'
import { api } from '../../api/client'
import type { Leaderboard, Pet, Reward, Summary } from '../../api/types'
import PetVisual, { stageLabel } from '../../components/PetVisual/PetVisual'
import { useGameEvent } from '../../realtime/RealtimeContext'

const energyLifetimeMs = 48 * 60 * 60 * 1000

function currentEnergy(lastChargedAt: string) {
  const elapsed = Date.now() - new Date(lastChargedAt).getTime()
  return Math.max(0, Math.min(100, 100 - Math.trunc(elapsed / energyLifetimeMs * 100)))
}

export default function Dashboard() {
  const event = useGameEvent()
  const [pet, setPet] = useState<Pet | null>(null)
  const [summary, setSummary] = useState<Summary | null>(null)
  const [leaderboard, setLeaderboard] = useState<Leaderboard | null>(null)
  const [rewards, setRewards] = useState<Reward[]>([])
  const [energy, setEnergy] = useState(0)
  const [loading, setLoading] = useState(true)
  const [charging, setCharging] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async (showLoading = false) => {
    if (showLoading) {setLoading(true)}
    setError(null)
    try {
      const [petValue, summaryValue, leaderboardValue, rewardValues] = await Promise.all([
        api.getPet(), api.getSummary(), api.getLeaderboard(), api.getRewards(),
      ])
      setPet(petValue)
      setSummary(summaryValue)
      setLeaderboard(leaderboardValue)
      setRewards(rewardValues)
      setEnergy(currentEnergy(petValue.lastChargedAt))
    } catch {
      setError('Не удалось загрузить состояние питомца.')
    } finally {
      setLoading(false)
    }
  }, [])

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
    if (!pet) {return}
    const update = () => setEnergy(currentEnergy(pet.lastChargedAt))
    const timer = window.setInterval(update, 30000)
    return () => window.clearInterval(timer)
  }, [pet])

  const availableRewards = useMemo(() => rewards.filter((reward) => reward.status === 'available').length, [rewards])
  const levelProgress = pet ? pet.xp % 100 : 0

  const charge = async () => {
    setCharging(true)
    setError(null)
    try {
      const updated = await api.chargePet()
      setPet(updated)
      setEnergy(updated.energy)
      await load()
    } catch {
      setError('Не удалось зарядить питомца.')
    } finally {
      setCharging(false)
    }
  }

  if (loading) {return <div className="dashboard-page"><div className="page-state">Загружаем питомца...</div></div>}
  if (!pet || !summary) {return <div className="dashboard-page"><div className="page-state page-state--error">{error ?? 'Данные питомца недоступны.'}</div></div>}

  return (
    <div className="dashboard-page">
      <div className="dashboard">
        {error && <p className="dashboard__error" role="alert">{error}</p>}
        <section className="dashboard__pet-card">
          <div className="dashboard__pet-image"><PetVisual stage={pet.stage} /></div>
          <div className="dashboard__pet-name">K1-T4</div>
          <div className="dashboard__pet-phase">{stageLabel(pet.stage)}, уровень {pet.level}</div>
          <div className="dashboard__progress-bars">
            <Progress label="XP до следующего уровня" value={`${levelProgress} / 100`} percent={levelProgress} kind="xp" />
            <Progress label="Батарея" value={`${energy}%`} percent={energy} kind="battery" />
          </div>
          <button className="dashboard__charge" type="button" disabled={charging} onClick={() => { void charge() }}>
            {charging ? 'Заряжаем...' : 'Зарядить'}
          </button>
        </section>

        <section className="dashboard__stats" aria-label="Дневная статистика">
          <Stat title="Выполнено заданий" value={`${summary.completedTasks} / 3`} accent="purple" />
          <Stat title="Заработано сегодня" value={`${summary.xpEarned} XP`} accent="blue" />
          <Stat title="Доступно наград" value={String(availableRewards)} accent="green" />
          <Stat title="Место в лидерборде" value={leaderboard?.currentUser ? String(leaderboard.currentUser.rank) : 'Нет данных'} accent="pink" />
        </section>
      </div>
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
