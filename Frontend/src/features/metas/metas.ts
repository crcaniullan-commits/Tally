import { ApiError, peticion } from '../../shared/api/cliente';
import { API } from '../../shared/api/rutas';

// Meta de venta mensual (Historia 007). Se guarda en el backend (módulo goals):
// el backend suma los ingresos del período y calcula el avance.

export type Meta = {
  id: string;
  nombre: string;
  periodo: 'dia' | 'semana' | 'mes';
  monto_meta: number;
  created_at: string;
};

export type EstadoMeta = {
  monto_actual: number;
  porcentaje_actual: number;
  monto_pendiente: number;
  porcentaje_faltante: number;
};

async function listarMetas(empresaId: string, token: string): Promise<Meta[]> {
  try {
    return (await peticion<Meta[]>(API.metas(empresaId), { token })) ?? [];
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) return []; // Aún no tiene metas
    throw error;
  }
}

// La meta mensual vigente (la más reciente) y su avance, o null si no hay
export async function obtenerMetaMensual(empresaId: string, token: string): Promise<{ meta: Meta; estado: EstadoMeta } | null> {
  const mensuales = (await listarMetas(empresaId, token))
    .filter((m) => m.periodo === 'mes')
    .sort((a, b) => b.created_at.localeCompare(a.created_at));
  if (mensuales.length === 0) return null;
  const meta = mensuales[0];
  const estado = await peticion<EstadoMeta>(API.estadoMeta(empresaId, meta.id), { token });
  return { meta, estado };
}

// Define o cambia la meta del mes. El backend no permite editarla: se borran las
// metas mensuales anteriores y se crea la nueva.
export async function guardarMetaMensual(empresaId: string, monto: number, token: string): Promise<void> {
  const anteriores = (await listarMetas(empresaId, token)).filter((m) => m.periodo === 'mes');
  await peticion<Meta>(API.metas(empresaId), {
    metodo: 'POST',
    cuerpo: { nombre: 'Meta mensual', periodo: 'mes', monto_meta: monto },
    token,
  });
  await Promise.all(anteriores.map((m) => peticion(API.meta(empresaId, m.id), { metodo: 'DELETE', token }).catch(() => {})));
}
