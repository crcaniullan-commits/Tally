import { API } from '../../shared/api/rutas';
import { peticion } from '../../shared/api/cliente';
import { Categoria, CategoriaApi, categoriasPredeterminadas, PREFIJO_LOCAL } from './categorias';

const ruta = (empresaId: string) => API.categorias(empresaId);

// GET → categorías globales + las propias de la empresa
export async function listarCategorias(empresaId: string, token: string): Promise<CategoriaApi[]> {
  return (await peticion<CategoriaApi[]>(ruta(empresaId), { token })) ?? [];
}

// POST → crea una categoría propia de la empresa
export function crearCategoria(nombre: string, empresaId: string, token: string): Promise<CategoriaApi> {
  return peticion<CategoriaApi>(ruta(empresaId), { metodo: 'POST', cuerpo: { nombre: nombre.trim() }, token });
}

// Devuelve el id real de una categoría. Si es una predeterminada que aún no existe
// en el backend, la crea en ese momento (solo pasa la primera vez que se usa).
export async function obtenerIdReal(categoria: Categoria, empresaId: string, token: string): Promise<string> {
  if (!categoria.id.startsWith(PREFIJO_LOCAL)) return categoria.id;
  // Se crea con su nombre original (el nombre visible puede estar cambiado solo en el celular)
  const original = categoriasPredeterminadas.find((p) => p.clave === categoria.clave)?.nombre ?? categoria.nombre;
  const creada = await crearCategoria(original, empresaId, token);
  return creada.id;
}

// PATCH → cambia el nombre de una categoría propia
export function renombrarCategoria(id: string, nombre: string, empresaId: string, token: string): Promise<unknown> {
  return peticion(API.categoria(empresaId, id), { metodo: 'PATCH', cuerpo: { nombre: nombre.trim() }, token });
}

// DELETE → elimina una categoría propia
export function eliminarCategoria(id: string, empresaId: string, token: string): Promise<unknown> {
  return peticion(API.categoria(empresaId, id), { metodo: 'DELETE', token });
}
