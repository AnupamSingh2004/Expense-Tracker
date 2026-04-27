/** Convert rupee string entered by user → paise integer for API */
export function rupeesToPaise(rupees: string): number {
  const parsed = parseFloat(rupees);
  if (isNaN(parsed) || parsed <= 0) throw new Error('Invalid amount');
  return Math.round(parsed * 100);
}

/** Format paise integer → display string e.g. 5050 → "₹50.50" */
export function formatPaise(paise: number): string {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    minimumFractionDigits: 2,
  }).format(paise / 100);
}
