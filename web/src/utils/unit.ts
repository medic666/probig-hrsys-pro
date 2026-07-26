const HOURS_PER_DAY = 8

export function hoursToDays(hours: number): number {
  return Math.round((hours / HOURS_PER_DAY) * 100) / 100
}

export function daysToHours(days: number): number {
  return Math.round(days * HOURS_PER_DAY * 10) / 10
}
