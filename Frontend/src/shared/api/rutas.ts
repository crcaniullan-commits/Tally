// TODAS las rutas de la API en un solo lugar.
// Si el backend cambia sus rutas (por ejemplo, a español en la rama exel_and_feed),
// solo se edita este archivo. Entre paréntesis, el nombre equivalente en español.

const app = '/app';                    // (/aplicacion)
const empresa = (id: string) => `${app}/empresas/${id}`;

export const API = {
  // Autenticación (/autenticacion)
  registro: '/auth',
  login: '/auth/login',

  // Usuario (/usuarios)
  usuario: `${app}/users/`,
  usuarioPorRut: (rut: string) => `${app}/users/municipal/${encodeURIComponent(rut)}`,

  // Claves municipales (/claves, /IDAcceso, /revocar)
  claves: `${app}/key/`,
  reenviarClave: (id: string) => `${app}/key/AccessID/${id}`,
  revocarClave: (code: string) => `${app}/key/revoke/${code}`,

  // Empresas
  empresas: `${app}/empresas/`,
  renombrarEmpresa: (id: string, nombre: string) => `${empresa(id)}/${encodeURIComponent(nombre)}`,

  // Movimientos y módulos de la empresa
  categorias: (id: string) => `${empresa(id)}/categories/`,   // (/categorias)
  categoria: (id: string, catId: string) => `${empresa(id)}/categories/${catId}`,
  ingresos: (id: string) => `${empresa(id)}/incomes/`,        // (/ingresos)
  ingreso: (id: string, movId: string) => `${empresa(id)}/incomes/${movId}`,
  gastos: (id: string) => `${empresa(id)}/expenses/`,         // (/gastos)
  gasto: (id: string, movId: string) => `${empresa(id)}/expenses/${movId}`,
  metas: (id: string) => `${empresa(id)}/goal/`,              // (/metas)
  meta: (id: string, metaId: string) => `${empresa(id)}/goal/${metaId}/`,
  estadoMeta: (id: string, metaId: string) => `${empresa(id)}/goal/${metaId}/state`,
  deudores: (id: string) => `${empresa(id)}/debtors/`,        // (/deudas)
  deudor: (id: string, deudaId: string) => `${empresa(id)}/debtors/${deudaId}`,
  pagarDeudor: (id: string, deudaId: string) => `${empresa(id)}/debtors/${deudaId}/paid`,
};

// Nombres de los filtros de los listados (en español: limite, desplazamiento, desde, hasta)
export const FILTROS = { limite: 'limit', desplazamiento: 'offset', desde: 'since', hasta: 'until', categoria: 'category_id' };
