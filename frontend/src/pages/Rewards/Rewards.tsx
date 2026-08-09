import { useCallback, useEffect, useState } from 'react'
import './Rewards.css'
import { api } from '../../api/client'
import type { Reward } from '../../api/types'
import { useGameEvent } from '../../realtime/RealtimeContext'

const rewardContent = {
  promotion: { title: 'Продвижение объявления', description: 'Поднимите своё объявление и получите больше просмотров.', icon: '🚀' },
  free_delivery: { title: 'Бесплатная доставка', description: 'Воспользуйтесь бесплатной доставкой для одной покупки.', icon: '📦' },
} as const

export default function Rewards() {
  const event = useGameEvent()
  const [rewards, setRewards] = useState<Reward[]>([])
  const [pending, setPending] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setRewards(await api.getRewards())
      setError(null)
    } catch {
      setError('Не удалось загрузить награды.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const timer = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(timer)
  }, [load])
  useEffect(() => {
    if (event?.name !== 'rewards_updated') return
    const timer = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(timer)
  }, [event, load])

  const redeemReward = async (reward: Reward) => {
    setPending(reward.id)
    try {
      await api.useReward(reward.id)
      await load()
    } catch {
      setError('Не удалось использовать награду.')
    } finally {
      setPending(null)
    }
  }

  if (loading) return <div className="rewards-page"><div className="page-state">Загружаем награды…</div></div>

  return (
    <div className="rewards-page">
      <div className="rewards">
        <header className="rewards__header"><h1 className="rewards__title">Награды</h1><p className="rewards__description">Персональные бонусы, открытые за новые уровни</p></header>
        {error && <p className="rewards__error" role="alert">{error}</p>}
        <div className="rewards__list">
          {rewards.length === 0 && !error && <div className="rewards__empty">Первая награда откроется на втором уровне.</div>}
          {rewards.map((reward) => <RewardCard key={reward.id} reward={reward} pending={pending === reward.id} onUse={() => redeemReward(reward)} />)}
        </div>
      </div>
    </div>
  )
}

function RewardCard({ reward, pending, onUse }: { reward: Reward; pending: boolean; onUse: () => void }) {
  const content = rewardContent[reward.type]
  const used = reward.status === 'used'
  return (
    <article className={`rewards__card ${used ? 'rewards__card--used' : ''}`}>
      <div className="rewards__card-header">
        <div className="rewards__icon" aria-hidden="true">{content.icon}</div>
        <div className="rewards__status">{used ? 'Использовано' : 'Доступно'}</div>
      </div>
      <div className="rewards__card-info">
        <h2 className="rewards__card-title">{content.title}</h2>
        <p className="rewards__card-description">{content.description}</p>
        <span className="rewards__level">Награда за уровень {reward.level}</span>
      </div>
      <button className="rewards__use" type="button" disabled={used || pending} onClick={onUse}>
        {used ? 'Уже использовано' : pending ? 'Используем…' : 'Использовать'}
      </button>
    </article>
  )
}
