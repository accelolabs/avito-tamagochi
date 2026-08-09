export interface User {
  id: string
  email: string
  displayName: string
  createdAt: string
}

export type PetStage = 'egg' | 'child' | 'teen' | 'adult'

export interface Pet {
  xp: number
  level: number
  stage: PetStage
  energy: number
  lastChargedAt: string
}

export type TaskType = 'view' | 'favorite' | 'create_listing' | 'create_listing_in_category'

export interface Task {
  type: TaskType
  progress: number
  requiredCount: number
  status: 'in_progress' | 'completed'
  xpReward: number
}

export interface Reward {
  id: string
  level: number
  type: 'promotion' | 'free_delivery'
  status: 'available' | 'used'
  unlockedAt: string
  usedAt: string | null
}

export interface Summary {
  xpEarned: number
  completedTasks: number
  charges: number
  level: number
  currentXP: number
  energy: number
  unlockedRewards: string[]
}

export interface LeaderboardEntry {
  rank: number
  userId: string
  displayName: string
  xp: number
  level: number
}

export interface Leaderboard {
  entries: LeaderboardEntry[]
  currentUser: LeaderboardEntry | null
}

export interface ApiErrorPayload {
  code?: string
  message?: string
}
