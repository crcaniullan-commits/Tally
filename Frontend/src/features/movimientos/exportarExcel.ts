import { File, Paths } from 'expo-file-system';
import * as Sharing from 'expo-sharing';
import * as XLSX from 'xlsx';
import { Movimiento } from './components/HistorialMovimientos';

const dosDigitos = (n: number) => String(n).padStart(2, '0');
const fecha = (f: Date) => `${dosDigitos(f.getDate())}-${dosDigitos(f.getMonth() + 1)}-${f.getFullYear()}`;

// Genera un archivo Excel con los movimientos y abre el menú para compartirlo
// (WhatsApp, correo, Drive, etc.) — Historia 005
export async function exportarExcel(movimientos: Movimiento[], nombreArchivo: string) {
  const filas = movimientos.map((m) => ({
    Fecha: fecha(m.fecha),
    Tipo: m.tipo === 'ingreso' ? 'Ingreso' : 'Gasto',
    Categoría: m.categoria,
    'Método de pago': m.metodo || '-',
    Nota: m.nota || '-',
    Monto: m.tipo === 'ingreso' ? m.monto : -m.monto, // Gastos en negativo para poder sumar
  }));

  const totalIngresos = movimientos.filter((m) => m.tipo === 'ingreso').reduce((s, m) => s + m.monto, 0);
  const totalGastos = movimientos.filter((m) => m.tipo === 'egreso').reduce((s, m) => s + m.monto, 0);

  const hoja = XLSX.utils.json_to_sheet(filas);
  XLSX.utils.sheet_add_aoa(hoja, [
    [],
    ['Total ingresos', '', '', '', '', totalIngresos],
    ['Total gastos', '', '', '', '', -totalGastos],
    ['Balance', '', '', '', '', totalIngresos - totalGastos],
  ], { origin: -1 });
  hoja['!cols'] = [{ wch: 12 }, { wch: 9 }, { wch: 16 }, { wch: 16 }, { wch: 30 }, { wch: 12 }];

  const libro = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(libro, hoja, 'Movimientos');
  const contenido = XLSX.write(libro, { type: 'array', bookType: 'xlsx' }) as ArrayBuffer;

  const archivo = new File(Paths.cache, `${nombreArchivo}.xlsx`);
  if (archivo.exists) archivo.delete();
  archivo.create();
  archivo.write(new Uint8Array(contenido));

  await Sharing.shareAsync(archivo.uri, {
    mimeType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    dialogTitle: 'Compartir movimientos',
    UTI: 'org.openxmlformats.spreadsheetml.sheet',
  });
}
