import dayjs from 'dayjs'

export function formatDate(date: string | Date, format = 'YYYY-MM-DD'): string {
  return dayjs(date).format(format)
}

export function formatDateTime(date: string | Date): string {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

export function hoursToDays(hours: number): number {
  return Math.round((hours / 8) * 100) / 100
}

export function formatMoney(amount: number): string {
  return amount.toFixed(2)
}
