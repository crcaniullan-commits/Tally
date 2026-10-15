// Métodos de pago aceptados por el backend (campo payment_method)
export type MetodoPago = 'efectivo' | 'debito' | 'credito' | 'transferencia';

export const metodosPago: { id: MetodoPago; nombre: string; icono: string }[] = [
    { id: 'efectivo', nombre: 'Efectivo', icono: 'cash' },
    { id: 'debito', nombre: 'Débito', icono: 'credit-card-outline' },
    { id: 'credito', nombre: 'Crédito', icono: 'credit-card-plus-outline' },
    { id: 'transferencia', nombre: 'Transferencia', icono: 'cellphone-arrow-down' },
];