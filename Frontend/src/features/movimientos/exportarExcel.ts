import XLSX from 'xlsx-js-style';
import {
  bordes, C, celda, Estilo, estiloEncabezado, estiloSubtitulo, estiloTitulo, fechaExcel as fecha, guardarYCompartir, horaExcel as hora,
} from '../../shared/utils/excel';
import { Movimiento } from './components/HistorialMovimientos';

// Excel con colores de Tally (Historia 005)
const MONEDA = '"$"#,##0;-"$"#,##0';
const COLUMNAS = ['Fecha', 'Hora', 'Tipo', 'Categoría', 'Método de pago', 'Comentario', 'Monto'];

// Genera el archivo y abre el menú para compartirlo (WhatsApp, correo, Drive...). Devuelve true si se completó.
export async function exportarExcel(movimientos: Movimiento[], nombreArchivo: string, empresa: string, periodo: string): Promise<boolean> {
  const totalIngresos = movimientos.filter((m) => m.tipo === 'ingreso').reduce((s, m) => s + m.monto, 0);
  const totalEgresos = movimientos.filter((m) => m.tipo === 'egreso').reduce((s, m) => s + m.monto, 0);
  const n = COLUMNAS.length;
  const vacia = (s: Estilo) => Array.from({ length: n - 1 }, () => celda('', s));

  // 1) Título y datos del reporte
  const titulo = estiloTitulo;
  const sub = estiloSubtitulo;
  const filas: unknown[][] = [
    [celda('Tally · Reporte de movimientos', titulo), ...vacia(titulo)],
    [celda(`${empresa}  ·  Período: ${periodo}  ·  Descargado el ${fecha(new Date())}`, sub), ...vacia(sub)],
    [],
  ];

  // 2) Encabezados
  const encabezado = estiloEncabezado;
  filas.push(COLUMNAS.map((c) => celda(c, encabezado)));

  // 3) Movimientos con filas alternadas y tipo/monto en color
  movimientos.forEach((m, i) => {
    const fondo = { fgColor: { rgb: i % 2 === 0 ? C.blanco : C.fondo } };
    const base = { font: { color: { rgb: C.texto } }, fill: fondo, border: bordes, alignment: { vertical: 'center' } };
    const esIngreso = m.tipo === 'ingreso';
    const color = esIngreso ? C.verde : C.rosado;
    filas.push([
      celda(fecha(m.fecha), { ...base, alignment: { horizontal: 'center' } }),
      celda(hora(m.fecha), { ...base, alignment: { horizontal: 'center' } }),
      celda(esIngreso ? 'Ingreso' : 'Egreso', { ...base, font: { bold: true, color: { rgb: color } }, alignment: { horizontal: 'center' } }),
      celda(m.categoria, base),
      celda(m.metodo || '—', base),
      celda(m.nota || '—', { ...base, alignment: { wrapText: true, vertical: 'center' } }),
      celda(esIngreso ? m.monto : -m.monto, { ...base, font: { bold: true, color: { rgb: color } } }, MONEDA),
    ]);
  });
  if (movimientos.length === 0) {
    const s = { font: { italic: true, color: { rgb: C.gris } }, border: bordes };
    filas.push([celda('Sin movimientos en este período', s), ...vacia(s)]);
  }

  // 4) Resumen
  const inicioResumen = filas.length + 1;
  filas.push([]);
  const resumen = (texto: string, valor: number, color: string, fondo: string) => {
    const s = { font: { bold: true, sz: 12, color: { rgb: color } }, fill: { fgColor: { rgb: fondo } }, border: bordes };
    return [...Array.from({ length: n - 2 }, () => celda('', {})), celda(texto, s), celda(valor, s, MONEDA)];
  };
  const tituloResumen = { font: { bold: true, color: { rgb: C.blanco } }, fill: { fgColor: { rgb: C.lavanda } }, border: bordes };
  filas.push([...Array.from({ length: n - 2 }, () => celda('', {})), celda('Resumen', tituloResumen), celda('', tituloResumen)]);
  filas.push(resumen('Total ingresos', totalIngresos, C.verde, C.verdeSuave));
  filas.push(resumen('Total egresos', -totalEgresos, C.rosado, C.rosadoSuave));
  filas.push(resumen('Balance', totalIngresos - totalEgresos, C.azul, C.azulSuave));

  const hoja = XLSX.utils.aoa_to_sheet(filas);
  hoja['!merges'] = [
    { s: { r: 0, c: 0 }, e: { r: 0, c: n - 1 } },
    { s: { r: 1, c: 0 }, e: { r: 1, c: n - 1 } },
    { s: { r: inicioResumen, c: n - 2 }, e: { r: inicioResumen, c: n - 1 } },
    ...(movimientos.length === 0 ? [{ s: { r: 4, c: 0 }, e: { r: 4, c: n - 1 } }] : []),
  ];
  hoja['!cols'] = [{ wch: 12 }, { wch: 8 }, { wch: 10 }, { wch: 18 }, { wch: 16 }, { wch: 34 }, { wch: 14 }];
  hoja['!rows'] = [{ hpt: 30 }, { hpt: 18 }, {}, { hpt: 22 }];

  const libro = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(libro, hoja, 'Movimientos');
  await guardarYCompartir(libro, nombreArchivo);
  return true;
}
