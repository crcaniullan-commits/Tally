import { guardarLocal, leerLocal } from '../../shared/utils/almacen';
import { claveMes } from '../../shared/utils/fechas';

// Meta de venta mensual (Historia 007).
// Pendiente backend: el módulo goals está vacío; mientras tanto se guarda en el celular.
type Metas = Record<string, number>; // "2026-11" → monto de la meta

const clave = (empresaId: string) => `metas_${empresaId}`;

export async function obtenerMetaDelMes(empresaId: string): Promise<number | null> {
  const metas = await leerLocal<Metas>(clave(empresaId), {});
  return metas[claveMes()] ?? null;
}

export async function guardarMetaDelMes(empresaId: string, monto: number): Promise<void> {
  const metas = await leerLocal<Metas>(clave(empresaId), {});
  metas[claveMes()] = monto;
  await guardarLocal(clave(empresaId), metas);
}
