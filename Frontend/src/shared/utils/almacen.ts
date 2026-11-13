import { File, Paths } from 'expo-file-system';

// Guarda datos de la app en el celular, en archivos JSON dentro de la carpeta privada de la app.
// Se usa para lo que el backend todavía no guarda (metas, pagos pendientes, preferencias).
const archivo = (clave: string) => new File(Paths.document, `tally_${clave}.json`);

export async function leerLocal<T>(clave: string, porDefecto: T): Promise<T> {
  try {
    const a = archivo(clave);
    if (!a.exists) return porDefecto;
    return JSON.parse(await a.text()) as T;
  } catch {
    return porDefecto; // Si el archivo está dañado, se parte de cero
  }
}

export async function guardarLocal<T>(clave: string, valor: T): Promise<void> {
  const a = archivo(clave);
  if (!a.exists) a.create();
  a.write(JSON.stringify(valor));
}
