import { api } from './client'
import type { Team, TeamSeat } from './boardTypes'

export function listTeams(workspaceId: string) {
  return api<{ teams: Team[] }>(`/workspaces/${workspaceId}/teams`)
}

export function getTeam(workspaceId: string, teamId: string) {
  return api<{ team: Team }>(`/workspaces/${workspaceId}/teams/${teamId}`)
}

export function getRoster(workspaceId: string, teamId: string) {
  return api<{ roster: TeamSeat[] }>(`/workspaces/${workspaceId}/teams/${teamId}/roster`)
}
