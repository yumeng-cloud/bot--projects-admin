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

export { dayjs, TZ }
