import { peticion } from '../../shared/api/cliente';
import { Categoria, CategoriaApi, PREFIJO_LOCAL } from './categorias';

const ruta = (empresaId: string) => `/app/empresas/${empresaId}/categories/`;

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
  const creada = await crearCategoria(categoria.nombre, empresaId, token);
  return creada.id;
}

// PATCH → cambia el nombre de una categoría propia
export function renombrarCategoria(id: string, nombre: string, empresaId: string, token: string): Promise<unknown> {
  return peticion(`${ruta(empresaId)}${id}`, { metodo: 'PATCH', cuerpo: { nombre: nombre.trim() }, token });
}

// DELETE → elimina una categoría propia
export function eliminarCategoria(id: string, empresaId: string, token: string): Promise<unknown> {
  return peticion(`${ruta(empresaId)}${id}`, { metodo: 'DELETE', token });
}
