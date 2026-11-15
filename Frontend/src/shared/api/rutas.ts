// TODAS las rutas de la API en un solo lugar (backend en español desde el PR #12 de Cristopher).
// Si el backend cambia sus rutas, solo se edita este archivo.

const app = '/aplicacion';
const empresa = (id: string) => `${app}/empresas/${id}`;

export const API = {
  // Autenticación
  registro: '/autenticacion',
  login: '/autenticacion/inicio-sesion',
  recuperarSolicitar: '/autenticacion/recuperar-password',          // Envía el código al correo
  recuperarVerificar: '/autenticacion/recuperar-password/verificar', // Revisa el código
  recuperarConfirmar: '/autenticacion/recuperar-password/confirmar', // Guarda la nueva contraseña

  // Usuario
  usuario: `${app}/usuarios/`,
  usuarioPorRut: (rut: string) => `${app}/usuarios/municipal/${encodeURIComponent(rut)}`,

  // Claves municipales
  claves: `${app}/claves/`,
  reenviarClave: (id: string) => `${app}/claves/IDAcceso/${id}`,
  revocarClave: (codigo: string) => `${app}/claves/revocar/${codigo}`,

  // Empresas
  empresas: `${app}/empresas/`,
  renombrarEmpresa: (id: string, nombre: string) => `${empresa(id)}/${encodeURIComponent(nombre)}`,

  // Movimientos y módulos de la empresa
  categorias: (id: string) => `${empresa(id)}/categorias/`,
  categoria: (id: string, catId: string) => `${empresa(id)}/categorias/${catId}`,
  ingresos: (id: string) => `${empresa(id)}/ingresos/`,
  ingreso: (id: string, movId: string) => `${empresa(id)}/ingresos/${movId}`,
  gastos: (id: string) => `${empresa(id)}/gastos/`,
  gasto: (id: string, movId: string) => `${empresa(id)}/gastos/${movId}`,
  metas: (id: string) => `${empresa(id)}/metas/`,
  meta: (id: string, metaId: string) => `${empresa(id)}/metas/${metaId}/`,
  estadoMeta: (id: string, metaId: string) => `${empresa(id)}/metas/${metaId}/estado`,
  deudores: (id: string) => `${empresa(id)}/deudas/`,
  deudor: (id: string, deudaId: string) => `${empresa(id)}/deudas/${deudaId}`,
  pagarDeudor: (id: string, deudaId: string) => `${empresa(id)}/deudas/${deudaId}/pagada`,
};

// Nombres de los filtros de los listados
export const FILTROS = { limite: 'limite', desplazamiento: 'desplazamiento', desde: 'desde', hasta: 'hasta', categoria: 'categoria_id' };
