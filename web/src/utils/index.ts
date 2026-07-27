import dayjs from 'dayjs'

export function formatDate(date: string | Date, format = 'YYYY-MM-DD'): string {
  return dayjs(date).format(format)
}

export function formatDateTime(date: string | Date): string {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

export function formatMonth(date: string | Date): string {
  return dayjs(date).format('YYYY-MM')
}

export function hoursToDays(hours: number): number {
  return Math.round((hours / 8) * 100) / 100
}

export function daysToHours(days: number): number {
  return Math.round((days * 8) * 10) / 10
}

export function formatMoney(amount: number): string {
  return amount.toFixed(2)
}

export function parseMoney(value: string | number): number {
  const num = typeof value === 'string' ? parseFloat(value) : value
  return Math.round(num * 100) / 100
}
