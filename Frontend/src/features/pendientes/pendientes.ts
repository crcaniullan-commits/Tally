import { peticion } from '../../shared/api/cliente';
import { API } from '../../shared/api/rutas';
import { guardarLocal, leerLocal } from '../../shared/utils/almacen';
import { desdeISO } from '../../shared/utils/fechas';
import { formatearMonto } from '../../shared/utils/formato';
import { cancelarAviso, programarAvisoPendiente } from '../notificaciones/notificaciones';

// Pagos pendientes / fiado (Historia 008): ventas que el cliente pagará después.
// Se guardan en el backend (módulo debtors).

// Formato del backend
type DeudorApi = {
  id: string;
  nombre: string;
  monto: number;
  descripcion: string | null;
  pagado: boolean;
  fecha_limite: string | null; // "2026-11-20T12:00:00Z"
  created_at: string;
};

// Formato que usa la app
export type Pendiente = {
  id: string;
  cliente: string;
  monto: number;
  vence: string;     // "AAAA-MM-DD"
  nota: string;
  categoria: string; // Categoría con que se registra el ingreso al pagarse
};

export const MAX_CLIENTE = 30; // Límite del backend para el nombre

// Ids de las notificaciones programadas (viven solo en el celular)
type Avisos = Record<string, string>;
const claveAvisos = (empresaId: string) => `avisos_${empresaId}`;

// Configuración de los recordatorios de pagos pendientes (se elige en Configuración)
export type ConfigAvisosPendientes = { activo: boolean; diasAntes: number };
const CLAVE_CONFIG = 'avisos_pendientes';
export const leerConfigAvisos = () => leerLocal<ConfigAvisosPendientes>(CLAVE_CONFIG, { activo: true, diasAntes: 0 });
export const guardarConfigAvisos = (c: ConfigAvisosPendientes) => guardarLocal(CLAVE_CONFIG, c);

// Programa (o reprograma) el recordatorio de un pago según la configuración
async function programarAviso(empresaId: string, id: string, cliente: string, monto: number, vence: string) {
  const avisos = await leerLocal<Avisos>(claveAvisos(empresaId), {});
  await cancelarAviso(avisos[id]);
  delete avisos[id];
  const config = await leerConfigAvisos();
  if (config.activo) {
    const avisoId = await programarAvisoPendiente(desdeISO(vence), cliente, formatearMonto(monto), config.diasAntes).catch(() => undefined);
    if (avisoId) avisos[id] = avisoId;
  }
  await guardarLocal(claveAvisos(empresaId), avisos);
}

const aPendiente = (d: DeudorApi): Pendiente => ({
  id: d.id,
  cliente: d.nombre,
  monto: d.monto,
  vence: (d.fecha_limite ?? d.created_at).slice(0, 10),
  nota: d.descripcion ?? '',
  categoria: 'Ventas',
});

// GET → solo los que aún no se pagan, el que vence primero arriba
export async function listarPendientes(empresaId: string, token: string): Promise<Pendiente[]> {
  const lista = (await peticion<DeudorApi[]>(API.deudores(empresaId), { token })) ?? [];
  return lista.filter((d) => !d.pagado).map(aPendiente).sort((a, b) => a.vence.localeCompare(b.vence));
}

// POST → crea el pago pendiente y programa el aviso del día del vencimiento
export async function agregarPendiente(
  empresaId: string,
  token: string,
  datos: { cliente: string; monto: number; vence: string; nota: string },
): Promise<Pendiente> {
  const creado = await peticion<DeudorApi>(API.deudores(empresaId), {
    metodo: 'POST',
    cuerpo: {
      nombre: datos.cliente.trim().slice(0, MAX_CLIENTE),
      monto: datos.monto,
      descripcion: (datos.nota.trim() || `Venta a ${datos.cliente.trim()}`).slice(0, 100), // El backend la exige
      fecha_limite: `${datos.vence}T12:00:00Z`, // Mediodía: evita que la zona horaria cambie el día
    },
    token,
  });
  await programarAviso(empresaId, creado.id, datos.cliente, datos.monto, datos.vence);
  return aPendiente(creado);
}

// PATCH → edita cliente, monto y fecha (el backend no permite cambiar la nota) y reprograma el aviso
export async function actualizarPendiente(
  empresaId: string,
  token: string,
  id: string,
  datos: { cliente: string; monto: number; vence: string },
): Promise<void> {
  await peticion(API.deudor(empresaId, id), {
    metodo: 'PATCH',
    cuerpo: { nombre: datos.cliente.trim().slice(0, MAX_CLIENTE), monto: datos.monto, fecha_limite: `${datos.vence}T12:00:00Z` },
    token,
  });
  await programarAviso(empresaId, id, datos.cliente, datos.monto, datos.vence);
}

// Al cambiar la configuración en Configuración se reprograman todos los avisos
export async function reprogramarAvisos(empresaId: string, token: string): Promise<void> {
  const lista = await listarPendientes(empresaId, token);
  for (const p of lista) await programarAviso(empresaId, p.id, p.cliente, p.monto, p.vence);
}

// PATCH /paid → lo marca como pagado (sale de la lista) y cancela su aviso.
// El backend aún no tiene DELETE para deudas: "Quitar" también usa esta ruta.
export async function marcarPagado(empresaId: string, token: string, pendiente: Pendiente): Promise<void> {
  await peticion(API.pagarDeudor(empresaId, pendiente.id), { metodo: 'PATCH', token });
  const avisos = await leerLocal<Avisos>(claveAvisos(empresaId), {});
  await cancelarAviso(avisos[pendiente.id]);
  delete avisos[pendiente.id];
  await guardarLocal(claveAvisos(empresaId), avisos);
}
