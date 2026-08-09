/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'

export type GameEventName = 'pet_updated' | 'tasks_updated' | 'rewards_updated'

interface GameEvent {
  name: GameEventName
  sequence: number
}

const RealtimeContext = createContext<GameEvent | null>(null)
const knownEvents = new Set<GameEventName>(['pet_updated', 'tasks_updated', 'rewards_updated'])

export function RealtimeProvider({ children }: { children: ReactNode }) {
  const [event, setEvent] = useState<GameEvent | null>(null)
  const sequence = useRef(0)

  useEffect(() => {
    let socket: WebSocket | null = null
    let reconnectTimer: number | null = null
    let attempts = 0
    let active = true

    const connect = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      socket = new WebSocket(`${protocol}//${window.location.host}/ws`)
      socket.onopen = () => {
        attempts = 0
      }
      socket.onmessage = (message) => {
        try {
          const payload = JSON.parse(String(message.data)) as { event?: string }
          if (payload.event && knownEvents.has(payload.event as GameEventName)) {
            sequence.current += 1
            setEvent({ name: payload.event as GameEventName, sequence: sequence.current })
          }
        } catch {
          // Ignore malformed real-time messages; REST remains the source of truth.
        }
      }
      socket.onclose = () => {
        if (!active) return
        attempts += 1
        const delay = Math.min(1000 * 2 ** (attempts - 1), 30000)
        reconnectTimer = window.setTimeout(connect, delay)
      }
    }

    connect()
    return () => {
      active = false
      if (reconnectTimer !== null) window.clearTimeout(reconnectTimer)
      socket?.close()
    }
  }, [])

  const value = useMemo(() => event, [event])
  return <RealtimeContext.Provider value={value}>{children}</RealtimeContext.Provider>
}

export function useGameEvent() {
  return useContext(RealtimeContext)
}
