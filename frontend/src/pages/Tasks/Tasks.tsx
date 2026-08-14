import { useCallback, useEffect, useMemo, useState } from 'react'
import './Tasks.css'
import { api } from '../../api/client'
import type { Task, TaskType } from '../../api/types'
import { useGameEvent } from '../../realtime/RealtimeContext'

const taskContent: Record<TaskType, { title: string; description: string; action: string }> = {
  view: { title: 'Посмотреть 5 объявлений', description: 'Каждое нажатие засчитывает один просмотр.', action: 'Посмотреть' },
  favorite: { title: 'Добавить в избранное', description: 'Добавьте любое объявление в избранное.', action: 'Добавить' },
  create_listing: { title: 'Разместить объявление', description: 'Создайте новое объявление на mock-платформе.', action: 'Разместить' },
  create_listing_in_category: { title: 'Разместить в категории', description: 'Создайте объявление в предложенной категории.', action: 'Разместить' },
}

export default function Tasks() {
  const event = useGameEvent()
  const [tasks, setTasks] = useState<Task[]>([])
  const [pending, setPending] = useState<TaskType | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    try {
      setTasks(await api.getTasks())
      setError(null)
    } catch {
      setError('Не удалось загрузить задания.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const timer = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(timer)
  }, [load])
  useEffect(() => {
    if (event?.name !== 'tasks_updated') {return}
    const timer = window.setTimeout(() => { void load() }, 0)
    return () => window.clearTimeout(timer)
  }, [event, load])

  const completed = useMemo(() => tasks.filter((task) => task.status === 'completed').length, [tasks])
  const complete = async (task: Task) => {
    setPending(task.type)
    try {
      await api.applyTaskAction(task.type)
      await load()
    } catch {
      setError('Не удалось выполнить действие. Попробуйте ещё раз.')
    } finally {
      setPending(null)
    }
  }

  if (loading) {return <div className="tasks-container"><div className="page-state">Загружаем задания...</div></div>}

  return (
    <div className="tasks-container">
      <div className="tasks">
        <header className="tasks__content"><h1 className="tasks__header">Ежедневные задания</h1><p className="tasks__description">Выполняй задания, чтобы развивать питомца</p></header>
        {error && <p className="tasks__error" role="alert">{error}</p>}
        <div className="tasks__content-container">
            <h2 className="tasks__stats-title">Задания на сегодня</h2>
            <div className="tasks__progress">
              <span className="tasks__progress-label">Выполнено: {completed} из {tasks.length}</span>
              <div className="tasks__progress-bar"><div className="tasks__progress-fill" style={{ width: `${tasks.length ? completed / tasks.length * 100 : 0}%` }} /></div>
            </div>
          <div className="tasks__list">
            {tasks.length === 0 && !error && <div className="page-state">Сегодня заданий нет.</div>}
            {tasks.map((task) => <TaskItem key={task.type} task={task} pending={pending === task.type} onAction={() => { void complete(task) }} />)}
          </div>
        </div>
      </div>
    </div>
  )
}

function TaskItem({ task, pending, onAction }: { task: Task; pending: boolean; onAction: () => void }) {
  const content = taskContent[task.type]
  const completed = task.status === 'completed'
  return (
    <article className={`tasks__card-item ${completed ? 'tasks__card-item--completed' : ''}`}>
      <div className="tasks__card-icon">+{task.xpReward}</div>
      <div className="tasks__card-info">
        <h2 className="tasks__card-title">{content.title}</h2>
        <p className="tasks__card-description">{content.description}</p>
        <span className="tasks__card-progress">Прогресс: {task.progress} / {task.requiredCount}</span>
      </div>
      <button className="tasks__complete-button" type="button" disabled={completed || pending} onClick={onAction}>
        {completed ? 'Выполнено' : pending ? 'Выполняем...' : content.action}
      </button>
    </article>
  )
}
