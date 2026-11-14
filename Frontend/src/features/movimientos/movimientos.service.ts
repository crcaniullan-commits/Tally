import { API, FILTROS } from '../../shared/api/rutas';
import { peticion } from '../../shared/api/cliente';
import { MetodoPago } from './metodosPago';

// Formato que entrega el backend
export type IngresoApi = {
  id: string;
  monto: number;
  payment_method: MetodoPago;
  descripcion: string | null;
  category_id: string | null;
  fecha: string;
  created_at: string;
};

export type GastoApi = {
  id: string;
  monto: number;
  category_id: string;
  descripcion: string | null;
  fecha: string;
  created_at: string;
};

const nota = (texto: string) => (texto.trim() ? texto.trim() : undefined); // Sin nota → no se envía

// POST /incomes → registra un ingreso
export function crearIngreso(
  datos: { monto: number; metodo: MetodoPago; categoriaId: string; nota: string },
  empresaId: string,
  token: string,
): Promise<IngresoApi> {
  return peticion<IngresoApi>(API.ingresos(empresaId), {
    metodo: 'POST',
    cuerpo: { monto: datos.monto, payment_method: datos.metodo, category_id: datos.categoriaId, descripcion: nota(datos.nota) },
    token,
  });
}

// POST /expenses → registra un gasto (el backend aún no guarda método de pago en los gastos)
export function crearGasto(
  datos: { monto: number; categoriaId: string; nota: string },
  empresaId: string,
  token: string,
): Promise<GastoApi> {
  return peticion<GastoApi>(API.gastos(empresaId), {
    metodo: 'POST',
    cuerpo: { monto: datos.monto, category_id: datos.categoriaId, descripcion: nota(datos.nota) },
    token,
  });
}

// GET → últimos 30 ingresos y gastos (máximo que permite el backend por página)
export async function listarIngresos(empresaId: string, token: string): Promise<IngresoApi[]> {
  return (await peticion<IngresoApi[]>(`${API.ingresos(empresaId)}?${FILTROS.limite}=30`, { token })) ?? [];
}

export async function listarGastos(empresaId: string, token: string): Promise<GastoApi[]> {
  return (await peticion<GastoApi[]>(`${API.gastos(empresaId)}?${FILTROS.limite}=30`, { token })) ?? [];
}

// Trae TODOS los movimientos desde una fecha, página por página (el backend entrega máximo 30 por vez)
async function traerTodo<T>(ruta: string, desde: string, token: string): Promise<T[]> {
  const POR_PAGINA = 30;
  const MAX_PAGINAS = 30; // Tope de seguridad: hasta 900 movimientos por período
  const todos: T[] = [];
  for (let pagina = 0; pagina < MAX_PAGINAS; pagina++) {
    const lote = (await peticion<T[]>(`${ruta}?${FILTROS.limite}=${POR_PAGINA}&${FILTROS.desplazamiento}=${pagina * POR_PAGINA}&${FILTROS.desde}=${desde}`, { token })) ?? [];
    todos.push(...lote);
    if (lote.length < POR_PAGINA) break;
  }
  return todos;
}

// desde: fecha en formato AAAA-MM-DD
export function ingresosDesde(desde: string, empresaId: string, token: string) {
  return traerTodo<IngresoApi>(API.ingresos(empresaId), desde, token);
}

export function gastosDesde(desde: string, empresaId: string, token: string) {
  return traerTodo<GastoApi>(API.gastos(empresaId), desde, token);
}

// DELETE → elimina un ingreso o un gasto (Historia 012)
export function eliminarMovimiento(tipo: 'ingreso' | 'egreso', id: string, empresaId: string, token: string): Promise<unknown> {
  const ruta = tipo === 'ingreso' ? API.ingreso(empresaId, id) : API.gasto(empresaId, id);
  return peticion(ruta, { metodo: 'DELETE', token });
}

// Editar (Historia 012): el backend aún no tiene PATCH, así que se crea el movimiento corregido
// y después se elimina el original. Se crea primero para que nunca se pierda el dato.
export async function reemplazarMovimiento(
  original: { tipo: 'ingreso' | 'egreso'; id: string },
  nuevo: { monto: number; metodo: MetodoPago | null; categoriaId: string; nota: string },
  empresaId: string,
  token: string,
): Promise<void> {
  if (original.tipo === 'ingreso') {
    await crearIngreso({ monto: nuevo.monto, metodo: nuevo.metodo ?? 'efectivo', categoriaId: nuevo.categoriaId, nota: nuevo.nota }, empresaId, token);
  } else {
    await crearGasto({ monto: nuevo.monto, categoriaId: nuevo.categoriaId, nota: nuevo.nota }, empresaId, token);
  }
  await eliminarMovimiento(original.tipo, original.id, empresaId, token);
}

// ¿La categoría tiene movimientos? (Historia 003: no se puede eliminar si los tiene)
export async function categoriaTieneMovimientos(categoriaId: string, empresaId: string, token: string): Promise<boolean> {
  const filtro = `?${FILTROS.limite}=1&${FILTROS.categoria}=${categoriaId}`;
  const [ingresos, gastos] = await Promise.all([
    peticion<IngresoApi[]>(`${API.ingresos(empresaId)}${filtro}`, { token }),
    peticion<GastoApi[]>(`${API.gastos(empresaId)}${filtro}`, { token }),
  ]);
  return (ingresos?.length ?? 0) > 0 || (gastos?.length ?? 0) > 0;
}
