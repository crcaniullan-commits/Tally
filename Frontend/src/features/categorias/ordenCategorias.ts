import { guardarLocal, leerLocal } from '../../shared/utils/almacen';
import { Categoria } from './categorias';

// Orden de las categorías que el usuario elige en Inicio (se guarda en el celular, por tipo)
export type OrdenCategorias = { ingreso: string[]; egreso: string[] };
const clave = (empresaId: string) => `orden_categorias_${empresaId}`;

export const leerOrden = (empresaId: string) => leerLocal<OrdenCategorias>(clave(empresaId), { ingreso: [], egreso: [] });
export const guardarOrden = (empresaId: string, orden: OrdenCategorias) => guardarLocal(clave(empresaId), orden);

// Ordena según lo guardado; las categorías nuevas quedan al final
export function ordenar(lista: Categoria[], orden: string[]): Categoria[] {
  const posicion = (c: Categoria) => {
    const i = orden.indexOf(c.clave);
    return i === -1 ? Number.MAX_SAFE_INTEGER : i;
  };
  return [...lista].sort((a, b) => posicion(a) - posicion(b));
}

// Preferencia de Configuración: categorías en carrusel (una fila) o en grilla (todas visibles)
export type VistaCategorias = 'carrusel' | 'grilla';
const CLAVE_VISTA = 'vista_categorias';
export const leerVistaCategorias = () => leerLocal<VistaCategorias>(CLAVE_VISTA, 'carrusel');
export const guardarVistaCategorias = (v: VistaCategorias) => guardarLocal(CLAVE_VISTA, v);
