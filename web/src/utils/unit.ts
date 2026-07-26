export function hoursToDays(hours: number): number {
  return parseFloat((hours / 8).toFixed(2))
}

export function daysToHours(days: number): number {
  return parseFloat((days * 8).toFixed(1))
}
