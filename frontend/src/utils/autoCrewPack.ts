/**
 * AUTO_CREW · 全班入场包：对每个空席签发带 role 的 ymt_，打成一份 Markdown。
 * 循环现有 POST …/teams/{id}/invites（role_hint/seat_slug）；禁嵌 ymk_。
 */
import { ApiError } from '../api/client'
import {
  createTeamInvite,
  seatDisplayName,
  seatOccupancy,
  type TeamInvite,
  type TeamSeat,
} from '../api/teams'
import { buildBotJoinPrompt, buildJoinUrl } from './oneTapJoin'

export type CrewPackSeatOk = {
  slug: string
  displayName: string
  code: string
  joinUrl: string
  invite: TeamInvite
}

export type CrewPackSeatFail = {
  slug: string
  displayName: string
  reason: string
}

export type CrewPackResult = {
  teamName: string
  markdown: string
  ok: CrewPackSeatOk[]
  failed: CrewPackSeatFail[]
}

/** 空席 / 可补位（非 active）才进包 */
export function emptySeatsForPack(seats: TeamSeat[]): TeamSeat[] {
  return (seats || []).filter((s) => seatOccupancy(s) !== 'active')
}

export function buildCrewPackMarkdown(params: {
  teamName: string
  seats: Array<{ displayName: string; slug: string; code: string }>
}): string {
  const n = params.seats.length
  const lines: string[] = [
    `# 班底入场包 · ${params.teamName}`,
    `（含 ${n} 个入场码，勿发公开群）`,
    '',
  ]
  for (const s of params.seats) {
    const url = buildJoinUrl({ code: s.code, role: s.slug })
    const prompt = buildBotJoinPrompt({
      code: s.code,
      role: s.slug,
      roleDisplayName: s.displayName,
    })
    lines.push(`## ${s.displayName}（${s.slug}）`)
    lines.push(`入场链接：${url}`)
    lines.push('')
    lines.push(prompt)
    lines.push('')
  }
  return `${lines.join('\n').trimEnd()}\n`
}

/**
 * 对每个空席循环签发；部分失败仍组装已成功席，失败列表供人话提示。
 */
export async function issueCrewPack(params: {
  workspaceId: string
  teamId: string
  teamName: string
  seats: TeamSeat[]
  /** 仅签发勾选席；缺省=全部空席（调用方应先确认） */
  selectedSlugs?: string[]
}): Promise<CrewPackResult> {
  let targets = emptySeatsForPack(params.seats)
  if (params.selectedSlugs) {
    const want = new Set(params.selectedSlugs.map((s) => s.trim()).filter(Boolean))
    targets = targets.filter((s) => want.has((s.role_slug || '').trim()))
  }
  const ok: CrewPackSeatOk[] = []
  const failed: CrewPackSeatFail[] = []

  for (const seat of targets) {
    const slug = (seat.role_slug || '').trim()
    const displayName = seatDisplayName(seat)
    if (!slug) {
      failed.push({ slug: seat.id, displayName, reason: '缺少席位 slug' })
      continue
    }
    try {
      const { invite } = await createTeamInvite(
        params.workspaceId,
        params.teamId,
        {
          allow_expand_seat: true,
          role_hint: slug,
          seat_slug: slug,
        },
        crypto.randomUUID(),
      )
      const code = (invite.code || '').trim()
      if (!code) {
        failed.push({ slug, displayName, reason: '签发未返回明文入场码' })
        continue
      }
      ok.push({
        slug,
        displayName,
        code,
        joinUrl: buildJoinUrl({ code, role: slug }),
        invite,
      })
    } catch (e: unknown) {
      const reason =
        e instanceof ApiError
          ? e.message
          : e instanceof Error
            ? e.message
            : '签发失败'
      failed.push({ slug, displayName, reason })
    }
  }

  const markdown =
    ok.length > 0
      ? buildCrewPackMarkdown({
          teamName: params.teamName,
          seats: ok.map((s) => ({
            displayName: s.displayName,
            slug: s.slug,
            code: s.code,
          })),
        })
      : ''

  return {
    teamName: params.teamName,
    markdown,
    ok,
    failed,
  }
}
