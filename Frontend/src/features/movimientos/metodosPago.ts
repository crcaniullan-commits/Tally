// Métodos de pago aceptados por el backend (campo payment_method)
export type MetodoPago = 'efectivo' | 'debito' | 'credito' | 'transferencia';

// corto: texto de la grilla de 4 columnas del mockup
export const metodosPago: { id: MetodoPago; nombre: string; corto: string; icono: string }[] = [
    { id: 'efectivo', nombre: 'Efectivo', corto: 'Efectivo', icono: 'banknote' },
    { id: 'debito', nombre: 'Débito', corto: 'Débito', icono: 'credit-card' },
    { id: 'credito', nombre: 'Crédito', corto: 'Crédito', icono: 'wallet' },
    { id: 'transferencia', nombre: 'Transferencia', corto: 'Transf.', icono: 'arrow-left-right' },
];
// Venta que el cliente paga después (Historia 008). No es un método del backend:
// se guarda como pago pendiente hasta que se marque como pagado.
export const ID_PENDIENTE = 'pendiente';
export const metodoPendiente = { id: ID_PENDIENTE, nombre: 'Pendiente', corto: 'Pendiente', icono: 'clock' };
