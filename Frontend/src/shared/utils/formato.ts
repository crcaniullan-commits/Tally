// Formatea un monto en pesos chilenos: 35000 → "35.000"
export function formatearMonto(monto: number): string {
    return String(monto).replace(/\B(?=(\d{3})+(?!\d))/g, '.');
}
// Monto con signo y símbolo: -35000 → "−$35.000"
export const pesos = (n: number) => `${n < 0 ? '−' : ''}$${formatearMonto(Math.abs(n))}`;
