import { colors } from '../../shared/theme/colors';
import { Categoria } from '../categorias/categorias';
import { OPCIONES_COLOR } from '../categorias/estilosPropios';
import { GastoApi, IngresoApi } from '../movimientos/movimientos.service';
import {
  aTexto, desglosarPorCategoria, FiltroDashboard, inicioPeriodoAnterior, ItemDesglose, Periodo, separarPorPeriodo, sumarPorTramo, tramosPorFecha,
} from './resumen';

export type TipoGrafico = 'torta' | 'barras' | 'horizontal';
export type Agrupacion = 'categoria' | 'fecha';
export type DatosDashboard = { ingresos: IngresoApi[]; gastos: GastoApi[]; categorias: Categoria[] };
export type Comparado = { nombre: string; ing: number; egr: number };

export const PERIODOS: { id: Periodo; nombre: string; texto: string; corto: string }[] = [
  { id: 'semana', nombre: 'Semana', texto: 'Esta semana', corto: 'sem.' },
  { id: 'mes', nombre: 'Mes', texto: 'Este mes', corto: 'mes' },
  { id: 'anio', nombre: 'Año', texto: 'Este año', corto: 'año' },
];

const suma = (lista: { monto: number }[]) => lista.reduce((s, x) => s + x.monto, 0);
const MAX_VISIBLES = 5; // Torta y barras verticales: 5 categorías + "Otras"

// Desde qué fecha hay que pedir datos: el período anterior (para "vs mes ant.") o el primer tramo "por fecha"
export function fechaDesde(periodo: Periodo): string {
  const primerTramo = tramosPorFecha(periodo).tramos[0].inicio;
  const anterior = inicioPeriodoAnterior(periodo);
  return aTexto(primerTramo < anterior ? primerTramo : anterior);
}

// Todos los cálculos del Dashboard (Historia 006). La pantalla solo dibuja el resultado.
export function calcularDashboard(datos: DatosDashboard, filtro: FiltroDashboard, periodo: Periodo, agrupacion: Agrupacion) {
  const info = PERIODOS.find((p) => p.id === periodo)!;
  const ing = separarPorPeriodo(datos.ingresos, periodo);
  const gas = separarPorPeriodo(datos.gastos, periodo);
  const totalIng = suma(ing.actual);
  const totalGas = suma(gas.actual);

  let total: number;
  let subtitulo: string;
  let etiquetaCentro: string;
  let insignia: { texto: string; color: string; fondo: string };
  let items: ItemDesglose[];

  if (filtro === 'ingresos') {
    total = totalIng;
    subtitulo = `Total Ingresos • ${info.texto}`;
    etiquetaCentro = 'Ingresos';
    const anterior = suma(ing.anterior);
    const cambio = anterior > 0 ? Math.round(((totalIng - anterior) / anterior) * 100) : null;
    insignia = cambio === null
      ? { texto: `Sin datos ${info.corto} ant.`, color: colors.verde, fondo: colors.verdeSuave }
      : { texto: `${cambio >= 0 ? '+' : ''}${cambio}% vs ${info.corto} ant.`, color: cambio >= 0 ? colors.verde : colors.rosado, fondo: cambio >= 0 ? colors.verdeSuave : colors.rosadoSuave };
    items = desglosarPorCategoria(ing.actual, datos.categorias);
  } else if (filtro === 'egresos') {
    total = totalGas;
    subtitulo = `Total Egresos • ${info.texto}`;
    etiquetaCentro = 'Egresos';
    insignia = { texto: totalIng > 0 ? `${Math.round((totalGas / totalIng) * 100)}% de ingresos` : 'Sin ingresos', color: colors.rosado, fondo: colors.rosadoSuave };
    items = desglosarPorCategoria(gas.actual, datos.categorias);
  } else {
    total = totalIng - totalGas;
    subtitulo = `Balance Neto • ${info.texto}`;
    etiquetaCentro = 'Balance';
    insignia = { texto: totalIng > 0 ? `${Math.round((total / totalIng) * 100)}% margen` : 'Sin ingresos', color: colors.lavanda, fondo: colors.lavandaSuave };
    const ambos = totalIng + totalGas;
    items = [
      { nombre: 'Ingresos Totales', icono: 'arrow-up-right', color: colors.verde, colorSuave: colors.verdeSuave, total: totalIng, porcentaje: ambos ? Math.round((totalIng / ambos) * 100) : 0 },
      { nombre: 'Egresos Totales', icono: 'arrow-down-right', color: colors.rosado, colorSuave: colors.rosadoSuave, total: totalGas, porcentaje: ambos ? Math.round((totalGas / ambos) * 100) : 0 },
    ].filter((x) => x.total > 0);
  }

  // Vista "por fecha": semanas del mes, meses del año o últimos 4 años
  const porFecha = agrupacion === 'fecha';
  const { descripcion, tramos } = tramosPorFecha(periodo);
  let comparado: Comparado[] = [];
  if (porFecha) {
    const ingT = sumarPorTramo(datos.ingresos, tramos);
    const egrT = sumarPorTramo(datos.gastos, tramos);
    if (filtro === 'ambos') {
      comparado = tramos.map((tr, i) => ({ nombre: tr.nombre, ing: ingT[i], egr: egrT[i] }));
      items = comparado.some((c) => c.ing || c.egr) ? comparado.map((c) => ({
        nombre: c.nombre, icono: 'calendar', color: colors.lavanda, colorSuave: colors.lavandaSuave, total: c.ing - c.egr, porcentaje: 0,
      })) : [];
    } else {
      const valores = filtro === 'ingresos' ? ingT : egrT;
      const totalTramos = valores.reduce((a, b) => a + b, 0);
      items = totalTramos === 0 ? [] : tramos.map((tr, i) => ({
        nombre: tr.nombre, icono: 'calendar',
        color: OPCIONES_COLOR[i % OPCIONES_COLOR.length].color, colorSuave: OPCIONES_COLOR[i % OPCIONES_COLOR.length].colorSuave,
        total: valores[i], porcentaje: Math.round((valores[i] / totalTramos) * 100),
      }));
    }
  }
  const comparando = porFecha && filtro === 'ambos'; // Ingresos y egresos lado a lado: solo barras

  // Torta y barras verticales: las 5 más grandes + "Otras" (para que se vean ordenadas con muchas categorías)
  const agrupados: ItemDesglose[] = porFecha || items.length <= MAX_VISIBLES + 1 ? items : [
    ...items.slice(0, MAX_VISIBLES),
    (() => {
      const resto = items.slice(MAX_VISIBLES);
      return {
        nombre: `Otras (${resto.length})`, icono: 'tag', color: '#b9bccd', colorSuave: '#eceef4',
        total: resto.reduce((a, x) => a + x.total, 0), porcentaje: resto.reduce((a, x) => a + x.porcentaje, 0),
      };
    })(),
  ];

  const nombreFiltro = filtro === 'ingresos' ? 'Ingresos' : filtro === 'egresos' ? 'Egresos' : 'Ingresos vs Egresos';
  const tituloGrafico = porFecha ? `${nombreFiltro} ${descripcion}` : filtro === 'ambos' ? 'Ingresos vs Egresos' : `${nombreFiltro} por categoría`;

  return { total, subtitulo, etiquetaCentro, insignia, items, agrupados, comparado, comparando, tituloGrafico };
}
