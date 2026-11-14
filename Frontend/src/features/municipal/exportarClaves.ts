import XLSX from 'xlsx-js-style';
import { bordes, C, celda, Estilo, estiloEncabezado, estiloSubtitulo, estiloTitulo, fechaExcel, guardarYCompartir } from '../../shared/utils/excel';
import { ClaveEmitida } from './municipal.service';

// Estado de una clave para el municipio
export function estadoClave(c: ClaveEmitida): { texto: string; color: string } {
  if (c.revocada) return { texto: 'Anulada', color: C.rosado };
  if (c.redeemed_at) return { texto: 'Usada', color: C.azul };
  if (c.expires_at && new Date(c.expires_at) < new Date()) return { texto: 'Vencida', color: C.gris };
  return { texto: 'Disponible', color: C.verde };
}

// Excel con las claves emitidas para que el municipio las distribuya (Definición de Terminado, Historia 009)
export async function exportarClaves(claves: ClaveEmitida[]): Promise<void> {
  const columnas = ['Correo', 'Clave', 'Emitida', 'Vence', 'Estado'];
  const vacia = (s: Estilo) => Array.from({ length: columnas.length - 1 }, () => celda('', s));
  const filas: unknown[][] = [
    [celda('Tally · Claves de activación', estiloTitulo), ...vacia(estiloTitulo)],
    [celda(`Descargado el ${fechaExcel(new Date())}  ·  ${claves.length} claves`, estiloSubtitulo), ...vacia(estiloSubtitulo)],
    [],
    columnas.map((c) => celda(c, estiloEncabezado)),
  ];
  claves.forEach((c, i) => {
    const base = { font: { color: { rgb: C.texto } }, fill: { fgColor: { rgb: i % 2 === 0 ? C.blanco : C.fondo } }, border: bordes };
    const centro = { ...base, alignment: { horizontal: 'center' } };
    const estado = estadoClave(c);
    filas.push([
      celda(c.email, base),
      celda(c.code, { ...centro, font: { bold: true, color: { rgb: C.lavanda } } }),
      celda(fechaExcel(new Date(c.created_at)), centro),
      celda(c.expires_at ? fechaExcel(new Date(c.expires_at)) : '—', centro),
      celda(estado.texto, { ...centro, font: { bold: true, color: { rgb: estado.color } } }),
    ]);
  });

  const hoja = XLSX.utils.aoa_to_sheet(filas);
  hoja['!merges'] = [{ s: { r: 0, c: 0 }, e: { r: 0, c: 4 } }, { s: { r: 1, c: 0 }, e: { r: 1, c: 4 } }];
  hoja['!cols'] = [{ wch: 32 }, { wch: 12 }, { wch: 12 }, { wch: 12 }, { wch: 12 }];
  hoja['!rows'] = [{ hpt: 30 }, { hpt: 18 }, {}, { hpt: 22 }];
  const libro = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(libro, hoja, 'Claves');
  await guardarYCompartir(libro, `Tally_claves_${fechaExcel(new Date())}`);
}
