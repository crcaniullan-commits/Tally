import { Categoria } from '../categorias/categorias';

export type Periodo = 'semana' | 'mes' | 'anio';

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

// ---------- Dashboard del mockup: filtros Ingresos / Egresos / Ing.-Egr. ----------
export type FiltroDashboard = 'ingresos' | 'egresos' | 'ambos';

export type ItemDesglose = { nombre: string; icono: string; color: string; colorSuave: string; total: number; porcentaje: number };

// Inicio del período anterior (para comparar "vs mes ant.")
export function inicioPeriodoAnterior(periodo: Periodo, hoy = new Date()): Date {
  const actual = inicioPeriodo(periodo, hoy);
  if (periodo === 'semana') return new Date(actual.getFullYear(), actual.getMonth(), actual.getDate() - 7);
  if (periodo === 'mes') return new Date(actual.getFullYear(), actual.getMonth() - 1, 1);
  return new Date(actual.getFullYear() - 1, 0, 1);
}

// Separa los movimientos que son del período actual y los del anterior
export function separarPorPeriodo<T extends { fecha: string }>(lista: T[], periodo: Periodo) {
  const inicio = inicioPeriodo(periodo).getTime();
  return {
    actual: lista.filter((x) => desdeTexto(x.fecha).getTime() >= inicio),
    anterior: lista.filter((x) => desdeTexto(x.fecha).getTime() < inicio),
  };
}

// Agrupa por categoría y calcula el porcentaje de cada una (de mayor a menor)
export function desglosarPorCategoria(lista: { monto: number; category_id: string | null }[], categorias: Categoria[]): ItemDesglose[] {
  const total = lista.reduce((s, x) => s + x.monto, 0);
  const grupos = new Map<string, number>();
  for (const x of lista) grupos.set(x.category_id ?? '', (grupos.get(x.category_id ?? '') ?? 0) + x.monto);
  return [...grupos.entries()]
    .map(([id, suma]) => {
      const c = categorias.find((x) => x.id === id);
      return {
        nombre: c?.nombre ?? 'Sin categoría',
        icono: c?.icono ?? 'tag',
        color: c?.color ?? '#8778cd',
        colorSuave: c?.colorSuave ?? '#eaedfa',
        total: suma,
        porcentaje: total > 0 ? Math.round((suma / total) * 100) : 0,
      };
    })
    .sort((a, b) => b.total - a.total);
}

// ---------- Dashboard "por fecha": compara tramos de tiempo ----------
// Semana → las semanas del mes en curso; Mes → los meses del año en curso; Año → los últimos 4 años
export type Tramo = { nombre: string; inicio: Date; fin: Date };
const MESES_TRAMO = ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun', 'Jul', 'Ago', 'Sep', 'Oct', 'Nov', 'Dic'];
const MESES_LARGOS = ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'];

export function tramosPorFecha(periodo: Periodo, hoy = new Date()): { descripcion: string; tramos: Tramo[] } {
  const anio = hoy.getFullYear();
  const mes = hoy.getMonth();
  if (periodo === 'semana') {
    const ultimoDia = new Date(anio, mes + 1, 0).getDate();
    const tramos: Tramo[] = [];
    for (let d = 1, n = 1; d <= ultimoDia; d += 7, n++) {
      tramos.push({ nombre: `Sem ${n}`, inicio: new Date(anio, mes, d), fin: new Date(anio, mes, Math.min(d + 7, ultimoDia + 1)) });
    }
    return { descripcion: `por semana de ${MESES_LARGOS[mes]}`, tramos };
  }
  if (periodo === 'mes') {
    return {
      descripcion: `por mes de ${anio}`,
      tramos: Array.from({ length: mes + 1 }, (_, m) => ({ nombre: MESES_TRAMO[m], inicio: new Date(anio, m, 1), fin: new Date(anio, m + 1, 1) })),
    };
  }
  return {
    descripcion: 'de los últimos 4 años',
    tramos: [3, 2, 1, 0].map((atras) => ({ nombre: String(anio - atras), inicio: new Date(anio - atras, 0, 1), fin: new Date(anio - atras + 1, 0, 1) })),
  };
}

export function sumarPorTramo(lista: { monto: number; fecha: string }[], tramos: Tramo[]): number[] {
  return tramos.map((tr) => lista
    .filter((x) => { const f = desdeTexto(x.fecha); return f >= tr.inicio && f < tr.fin; })
    .reduce((s, x) => s + x.monto, 0));
}
