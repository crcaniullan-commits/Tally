import { File, Paths } from 'expo-file-system';
import * as Sharing from 'expo-sharing';
import XLSX from 'xlsx-js-style';

// Colores y estilos comunes de los Excel de Tally (librería xlsx-js-style)
export const C = {
  lavanda: '8778CD', lavandaSuave: 'EAEDFA', fondo: 'FAF9FD', blanco: 'FFFFFF', texto: '545454', gris: '8A8A8A',
  verde: '2FAE80', verdeSuave: 'D3FDEE', rosado: 'F82F8F', rosadoSuave: 'FFDFEF', azul: '1A8DC0', azulSuave: 'DCF9FD',
  naranjo: 'E58A3E', borde: 'DFE3F9',
};
const borde = { style: 'thin', color: { rgb: C.borde } };
export const bordes = { top: borde, bottom: borde, left: borde, right: borde };
export type Estilo = Record<string, unknown>;
export const celda = (v: string | number, s: Estilo, z?: string) => ({ v, t: typeof v === 'number' ? 'n' : 's', s, ...(z ? { z } : {}) });
export const estiloTitulo = { font: { bold: true, sz: 16, color: { rgb: C.blanco } }, fill: { fgColor: { rgb: C.lavanda } }, alignment: { vertical: 'center' } };
export const estiloSubtitulo = { font: { sz: 10, color: { rgb: C.lavanda } }, fill: { fgColor: { rgb: C.lavandaSuave } } };
export const estiloEncabezado = {
  font: { bold: true, color: { rgb: C.blanco } }, fill: { fgColor: { rgb: C.lavanda } },
  alignment: { horizontal: 'center', vertical: 'center' }, border: bordes,
};
const dos = (n: number) => String(n).padStart(2, '0');
export const fechaExcel = (f: Date) => `${dos(f.getDate())}-${dos(f.getMonth() + 1)}-${f.getFullYear()}`;
export const horaExcel = (f: Date) => `${dos(f.getHours())}:${dos(f.getMinutes())}`;

// Guarda el libro como .xlsx (lo abren Excel, Google Sheets y LibreOffice) y abre el menú para compartirlo
export async function guardarYCompartir(libro: XLSX.WorkBook, nombreArchivo: string): Promise<void> {
  const contenido = XLSX.write(libro, { type: 'array', bookType: 'xlsx', compression: true }) as ArrayBuffer;
  const archivo = new File(Paths.cache, `${nombreArchivo}.xlsx`);
  if (archivo.exists) archivo.delete();
  archivo.create();
  archivo.write(new Uint8Array(contenido));
  await Sharing.shareAsync(archivo.uri, {
    mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    dialogTitle: 'Guardar o compartir tu Excel',
    UTI: 'org.openxmlformats.spreadsheetml.sheet',
  });
}
