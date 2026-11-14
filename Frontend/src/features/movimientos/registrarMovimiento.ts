import { Categoria, TipoMovimiento } from '../categorias/categorias';
import { obtenerIdReal } from '../categorias/categorias.service';
import { obtenerMetaMensual } from '../metas/metas';
import { agregarPendiente } from '../pendientes/pendientes';
import { MetodoPago } from './metodosPago';
import { crearGasto, crearIngreso } from './movimientos.service';

export type DatosRegistro = {
  tipo: TipoMovimiento;
  monto: number;
  categoria: Categoria;
  metodo: string;
  nota: string;
  pendiente: { cliente: string; vence: string } | null; // Solo si es un pago pendiente (Historia 008)
};

export type ResultadoRegistro = {
  mensaje: string;
  icono: string;
  categoriaCreada: boolean; // Se creó en el backend una categoría predeterminada (hay que recargar la lista)
};

// Registro rápido (Historia 004): guarda un ingreso, un egreso o un pago pendiente
export async function registrarMovimiento(d: DatosRegistro, empresaId: string, token: string): Promise<ResultadoRegistro> {
  if (d.pendiente) {
    await agregarPendiente(empresaId, token, { cliente: d.pendiente.cliente.trim(), monto: d.monto, vence: d.pendiente.vence, nota: d.nota.trim() });
    return { mensaje: '¡Pago pendiente guardado!', icono: 'clock', categoriaCreada: false };
  }
  // Las predeterminadas se crean en el backend la primera vez que se usan
  const idReal = await obtenerIdReal(d.categoria, empresaId, token);
  const categoriaCreada = idReal !== d.categoria.id;
  if (d.tipo === 'ingreso') {
    await crearIngreso({ monto: d.monto, metodo: d.metodo as MetodoPago, categoriaId: idReal, nota: d.nota }, empresaId, token);
    return { mensaje: '¡Ingreso registrado!', icono: 'check-circle', categoriaCreada };
  }
  await crearGasto({ monto: d.monto, categoriaId: idReal, nota: d.nota }, empresaId, token);
  return { mensaje: '¡Egreso registrado!', icono: 'check-circle', categoriaCreada };
}

// ¿Con esta venta se cumplió la meta del mes? (Historia 007). Solo es true justo cuando se cruza la meta.
export async function seCumplioLaMeta(montoNuevo: number, empresaId: string, token: string): Promise<boolean> {
  const datos = await obtenerMetaMensual(empresaId, token).catch(() => null);
  if (!datos) return false;
  const actual = datos.estado.monto_actual;
  return actual >= datos.meta.monto_meta && actual - montoNuevo < datos.meta.monto_meta;
}
