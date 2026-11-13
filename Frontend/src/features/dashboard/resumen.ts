import { Categoria } from '../categorias/categorias';
import { GastoApi, IngresoApi } from '../movimientos/movimientos.service';

export type Periodo = 'semana' | 'mes' | 'anio';

export type Barra = { etiqueta: string; ingresos: number; gastos: number };
export type GastoCategoria = { nombre: string; color: string; total: number; porcentaje: number };

export type Resumen = {
  ingresos: number;
  gastos: number;
  balance: number;
  barras: Barra[];
  gastosPorCategoria: GastoCategoria[];
};

const DIAS = ['Lun', 'Mar', 'Mié', 'Jue', 'Vie', 'Sáb', 'Dom'];
const MESES = ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun', 'Jul', 'Ago', 'Sep', 'Oct', 'Nov', 'Dic'];

// Fecha AAAA-MM-DD en hora local (la que espera el backend en ?since=)
export const aTexto = (f: Date) =>
  `${f.getFullYear()}-${String(f.getMonth() + 1).padStart(2, '0')}-${String(f.getDate()).padStart(2, '0')}`;

// "2026-11-02" → Date local (sin desfase por zona horaria)
const desdeTexto = (texto: string) => {
  const [a, m, d] = texto.slice(0, 10).split('-').map(Number);
  return new Date(a, m - 1, d);
};

// Primer día del período actual: lunes de esta semana, día 1 del mes o 1 de enero
export function inicioPeriodo(periodo: Periodo, hoy = new Date()): Date {
  if (periodo === 'semana') {
    const diasDesdeLunes = (hoy.getDay() + 6) % 7;
    return new Date(hoy.getFullYear(), hoy.getMonth(), hoy.getDate() - diasDesdeLunes);
  }
  if (periodo === 'mes') return new Date(hoy.getFullYear(), hoy.getMonth(), 1);
  return new Date(hoy.getFullYear(), 0, 1);
}

// Divide el período en columnas: 7 días, 5 semanas del mes o 12 meses
function columnas(periodo: Periodo): { etiquetas: string[]; indice: (f: Date) => number } {
  if (periodo === 'semana') return { etiquetas: DIAS, indice: (f) => (f.getDay() + 6) % 7 };
  if (periodo === 'mes') {
    return {
      etiquetas: ['Sem 1', 'Sem 2', 'Sem 3', 'Sem 4', 'Sem 5'],
      indice: (f) => Math.min(Math.floor((f.getDate() - 1) / 7), 4),
    };
  }
  return { etiquetas: MESES, indice: (f) => f.getMonth() };
}

export function calcularResumen(
  periodo: Periodo,
  ingresos: IngresoApi[],
  gastos: GastoApi[],
  categorias: Categoria[],
): Resumen {
  const { etiquetas, indice } = columnas(periodo);
  const barras: Barra[] = etiquetas.map((etiqueta) => ({ etiqueta, ingresos: 0, gastos: 0 }));

  let totalIngresos = 0;
  for (const i of ingresos) {
    totalIngresos += i.monto;
    barras[indice(desdeTexto(i.fecha))].ingresos += i.monto;
  }

  let totalGastos = 0;
  const porCategoria = new Map<string, number>();
  for (const g of gastos) {
    totalGastos += g.monto;
    barras[indice(desdeTexto(g.fecha))].gastos += g.monto;
    porCategoria.set(g.category_id, (porCategoria.get(g.category_id) ?? 0) + g.monto);
  }

  const gastosPorCategoria: GastoCategoria[] = [...porCategoria.entries()]
    .map(([id, total]) => {
      const c = categorias.find((x) => x.id === id);
      return {
        nombre: c?.nombre ?? 'Sin categoría',
        color: c?.color ?? '#8778cd',
        total,
        porcentaje: totalGastos > 0 ? Math.round((total / totalGastos) * 100) : 0,
      };
    })
    .sort((a, b) => b.total - a.total);

  return { ingresos: totalIngresos, gastos: totalGastos, balance: totalIngresos - totalGastos, barras, gastosPorCategoria };
}
