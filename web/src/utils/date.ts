import dayjs from 'dayjs'

export function formatDate(date: string | Date): string {
  return dayjs(date).format('YYYY-MM-DD')
}

export function formatMonth(date: string | Date): string {
  return dayjs(date).format('YYYY-MM')
}

export function formatDateTime(date: string | Date): string {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}
