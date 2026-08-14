import { useCallback, useEffect, useState } from 'react'
import './Rewards.css'
import { api } from '../../api/client'
import type { Reward } from '../../api/types'
import { useGameEvent } from '../../realtime/RealtimeContext'

const rewardContent = {
  promotion: { title: 'Продвижение объявления', description: 'Поднимите своё объявление и получите больше просмотров.' },
  free_delivery: { title: 'Бесплатная доставка', description: 'Воспользуйтесь бесплатной доставкой для одной покупки.' },
} as const

export default function Rewards() {
  const event = useGameEvent()
  const [rewards, setRewards] = useState<Reward[]>([])
  const [petLevel, setPetLevel] = useState(1)
  const [pending, setPending] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      const [rewardValues, pet] = await Promise.all([api.getRewards(), api.getPet()])
      setRewards(rewardValues)
      setPetLevel(pet.level)
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
    if (event?.name !== 'rewards_updated') {return}
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

  const nextRewardLevel = Math.max(2, petLevel + 1)

  if (loading) {return <div className="rewards-page"><div className="page-state">Загружаем награды...</div></div>}

  return (
    <div className="rewards-page">
      <div className="rewards">
        <header className="rewards__header"><h1 className="rewards__title">Награды</h1><p className="rewards__description">Персональные бонусы, открытые за новые уровни</p></header>
        {error && <p className="rewards__error" role="alert">{error}</p>}
        <div className="rewards__list">
          {rewards.map((reward) => <RewardCard key={reward.id} reward={reward} pending={pending === reward.id} onUse={() => { void redeemReward(reward) }} />)}
          <div className='rewards__card rewards__card--next'>
            <div className='rewards__card-info'>
              <svg width="32" height="42" viewBox="0 0 32 42" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M4 42C2.9 42 1.95833 41.6083 1.175 40.825C0.391667 40.0417 0 39.1 0 38V18C0 16.9 0.391667 15.9583 1.175 15.175C1.95833 14.3917 2.9 14 4 14H6V10C6 7.23334 6.975 4.875 8.925 2.925C10.875 0.975 13.2333 0 16 0C18.7667 0 21.125 0.975 23.075 2.925C25.025 4.875 26 7.23334 26 10V14H28C29.1 14 30.0417 14.3917 30.825 15.175C31.6083 15.9583 32 16.9 32 18V38C32 39.1 31.6083 40.0417 30.825 40.825C30.0417 41.6083 29.1 42 28 42H4ZM4 38H28V18H4V38ZM16 32C17.1 32 18.0417 31.6083 18.825 30.825C19.6083 30.0417 20 29.1 20 28C20 26.9 19.6083 25.9583 18.825 25.175C18.0417 24.3917 17.1 24 16 24C14.9 24 13.9583 24.3917 13.175 25.175C12.3917 25.9583 12 26.9 12 28C12 29.1 12.3917 30.0417 13.175 30.825C13.9583 31.6083 14.9 32 16 32ZM10 14H22V10C22 8.33334 21.4167 6.91667 20.25 5.75C19.0833 4.58334 17.6667 4 16 4C14.3333 4 12.9167 4.58334 11.75 5.75C10.5833 6.91667 10 8.33334 10 10V14ZM4 38V18V38Z" fill="#6E7883"/>
              </svg>
              <h2 className="rewards__card-title">Следующая награда</h2>
              <p className='rewards__card-description'>Откроется на уровне {nextRewardLevel}</p>
            </div>
          </div>
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
        <div className="rewards__icon" aria-hidden="true">
          <span
            className="rewards__icon-image"
            style={{ maskImage: 'url(/reward.svg)', WebkitMaskImage: 'url(/reward.svg)' }}
          />
        </div>
      </div>
      <div className="rewards__card-info">
        <span className="rewards__level">Награда за уровень {reward.level}</span>
        <h2 className="rewards__card-title">{content.title}</h2>
        <p className="rewards__card-description">{content.description}</p>
      </div>
      <button className="rewards__use" type="button" disabled={used || pending} onClick={onUse}>
        {used ? 'Уже использовано' : pending ? 'Используем...' : 'Использовать'}
      </button>
    </article>
  )
}
