import { guardarLocal, leerLocal } from '../../shared/utils/almacen';
import { formatearMonto } from '../../shared/utils/formato';
import { desdeISO } from '../../shared/utils/fechas';
import { cancelarAviso, programarAvisoPendiente } from '../notificaciones/notificaciones';

// Pagos pendientes / fiado (Historia 008): ventas que el cliente pagará después.
// Pendiente backend: el módulo debtors está vacío; mientras tanto se guardan en el celular.
export type Pendiente = {
  id: string;
  cliente: string;
  monto: number;
  vence: string;            // "AAAA-MM-DD"
  nota: string;
  categoria: string;        // Nombre de la categoría (se usa al marcarlo como pagado)
  notificacionId?: string;  // Aviso programado para el día del vencimiento
};

const clave = (empresaId: string) => `pendientes_${empresaId}`;

export async function listarPendientes(empresaId: string): Promise<Pendiente[]> {
  const lista = await leerLocal<Pendiente[]>(clave(empresaId), []);
  return lista.sort((a, b) => a.vence.localeCompare(b.vence)); // El que vence primero, arriba
}

export async function agregarPendiente(empresaId: string, datos: Omit<Pendiente, 'id' | 'notificacionId'>): Promise<Pendiente> {
  const notificacionId = await programarAvisoPendiente(desdeISO(datos.vence), datos.cliente, formatearMonto(datos.monto)).catch(() => undefined);
  const nuevo: Pendiente = { ...datos, id: `${Date.now()}`, notificacionId };
  const lista = await leerLocal<Pendiente[]>(clave(empresaId), []);
  await guardarLocal(clave(empresaId), [...lista, nuevo]);
  return nuevo;
}

export async function quitarPendiente(empresaId: string, pendiente: Pendiente): Promise<void> {
  await cancelarAviso(pendiente.notificacionId);
  const lista = await leerLocal<Pendiente[]>(clave(empresaId), []);
  await guardarLocal(clave(empresaId), lista.filter((p) => p.id !== pendiente.id));
}
