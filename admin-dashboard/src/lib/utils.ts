export function formatPrice(price: number): string {
  // Format with Tunisian representation (199 DT or 99.90 DT)
  // Check if the price has decimal places
  const hasDecimal = price % 1 !== 0;

  return new Intl.NumberFormat('fr-TN', {
    style: 'decimal',
    minimumFractionDigits: 0,
    maximumFractionDigits: hasDecimal ? 2 : 0, // Show up to 2 decimal places only if needed
  }).format(price) + ' DT';
}

export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}