import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'
import 'dayjs/locale/zh-cn'

dayjs.extend(utc)
dayjs.extend(timezone)
dayjs.locale('zh-cn')

const TZ = 'Asia/Shanghai'

export function formatShanghai(iso?: string | null, pattern = 'YYYY-MM-DD HH:mm') {
  if (!iso) return '—'
  return dayjs.utc(iso).tz(TZ).format(pattern)
}

/** Wiki meta line (W7): Asia/Shanghai readable, no bare RFC3339Z. */
export function formatWikiTime(iso?: string | null): string {
  if (!iso) return '—'
  const d = dayjs.utc(iso).tz(TZ)
  if (!d.isValid()) return '—'
  const now = dayjs().tz(TZ)
  const time = d.format('HH:mm')
  if (d.isSame(now, 'day')) return `今天 ${time}`
  if (d.isSame(now.subtract(1, 'day'), 'day')) return `昨天 ${time}`
  if (d.isSame(now, 'year')) return d.format('M月D日 HH:mm')
  return d.format('YYYY-MM-DD HH:mm')
}

export { dayjs, TZ }
