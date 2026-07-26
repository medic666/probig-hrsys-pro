import dayjs from 'dayjs'

export function formatDate(date: string | Date | null | undefined, fmt = 'YYYY-MM-DD'): string {
  if (!date) return ''
  return dayjs(date).format(fmt)
}

export function formatDateTime(date: string | Date | null | undefined): string {
  return formatDate(date, 'YYYY-MM-DD HH:mm:ss')
}

export function hoursToDays(hours: number): number {
  const perDay = 8
  return parseFloat((hours / perDay).toFixed(2))
}

export function daysToHours(days: number): number {
  const perDay = 8
  return parseFloat((days * perDay).toFixed(1))
}

export function formatMoney(value: number | null | undefined): string {
  if (value === null || value === undefined) return '0.00'
  return Number(value).toFixed(2)
}
