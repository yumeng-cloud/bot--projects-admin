/** Same-origin path only; reject open redirects. Default home. */
export function safeReturnPath(from: unknown): string {
  if (typeof from !== 'string' || !from.startsWith('/') || from.startsWith('//')) return '/'
  if (from === '/login') return '/'
  return from
}
