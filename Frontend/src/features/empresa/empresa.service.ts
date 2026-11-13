import { ApiError, peticion } from '../../shared/api/cliente';

// Empresa (emprendimiento) del usuario: ingresos, gastos y categorías pertenecen a ella
export type Empresa = {
  id: string;
  nombre_empresa: string;
  user_id: string;
};

export const MAX_NOMBRE_EMPRESA = 30; // Límite del backend (validate max=30)
const NOMBRE_POR_DEFECTO = 'Mi emprendimiento';

// POST /v1/app/empresas/ → crea la empresa del usuario
export function crearEmpresa(nombre: string, token: string): Promise<Empresa> {
  return peticion<Empresa>('/app/empresas/', {
    metodo: 'POST',
    cuerpo: { nombre_empresa: nombre.trim().slice(0, MAX_NOMBRE_EMPRESA) || NOMBRE_POR_DEFECTO },
    token,
  });
}

// GET /v1/app/empresas/ → lista las empresas del usuario (404 si no tiene)
export function listarEmpresas(token: string): Promise<Empresa[]> {
  return peticion<Empresa[]>('/app/empresas/', { token });
}

// PATCH /v1/app/empresas/{id}/{nombre} → cambia el nombre del emprendimiento
export function cambiarNombreEmpresa(empresaId: string, nombre: string, token: string): Promise<string> {
  const limpio = nombre.trim().slice(0, MAX_NOMBRE_EMPRESA);
  return peticion<string>(`/app/empresas/${empresaId}/${encodeURIComponent(limpio)}`, { metodo: 'PATCH', token });
}

// Devuelve la primera empresa del usuario; si no tiene ninguna, crea una
export async function obtenerOCrearEmpresa(token: string): Promise<Empresa> {
  try {
    const empresas = await listarEmpresas(token);
    if (empresas && empresas.length > 0) return empresas[0];
  } catch (error) {
    const sinEmpresas = error instanceof ApiError && error.status === 404;
    if (!sinEmpresas) throw error;
  }
  return crearEmpresa(NOMBRE_POR_DEFECTO, token);
}
