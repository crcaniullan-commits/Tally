import { colors } from '../../shared/theme/colors';

export type TipoMovimiento = 'ingreso' | 'egreso';

// Categoría tal como la entrega el backend (GET /categories)
export type CategoriaApi = {
  id: string;
  nombre: string;
  empresa_id: string | null; // null = categoría global
};

// Categoría lista para mostrar en la app
export type Categoria = {
  id: string;         // id del backend; las predeterminadas que aún no existen allá usan "pred-<nombre>"
  nombre: string;
  icono: string;      // Nombre del ícono (MaterialCommunityIcons)
  color: string;      // Color intenso: ícono y borde al seleccionar
  colorSuave: string; // Color suave: fondo del ícono
  tipo: TipoMovimiento | 'ambos'; // 'ambos' = categoría creada por el usuario
};

export const MAX_NOMBRE_CATEGORIA = 30; // Límite del backend
export const PREFIJO_LOCAL = 'pred-';

// Categorías predeterminadas (Historia 003): siempre visibles, 3 por tipo
export const categoriasPredeterminadas: Categoria[] = [
  { id: 'pred-ventas', nombre: 'Ventas', icono: 'shopping-outline', color: colors.verde, colorSuave: colors.verdeSuave, tipo: 'ingreso' },
  { id: 'pred-servicios', nombre: 'Servicios', icono: 'handshake-outline', color: colors.azul, colorSuave: colors.azulSuave, tipo: 'ingreso' },
  { id: 'pred-pedidos', nombre: 'Pedidos', icono: 'clipboard-text-outline', color: colors.naranjo, colorSuave: colors.naranjoSuave, tipo: 'ingreso' },
  { id: 'pred-insumos', nombre: 'Insumos', icono: 'package-variant-closed', color: colors.naranjo, colorSuave: colors.naranjoSuave, tipo: 'egreso' },
  { id: 'pred-transporte', nombre: 'Transporte', icono: 'truck-delivery', color: colors.azul, colorSuave: colors.azulSuave, tipo: 'egreso' },
  { id: 'pred-gastos-fijos', nombre: 'Gastos fijos', icono: 'home-city-outline', color: colors.rosado, colorSuave: colors.rosadoSuave, tipo: 'egreso' },
];

const estiloPropio = { icono: 'tag-outline', color: colors.lavanda, colorSuave: colors.lavandaSuave };
export const normalizar = (texto: string) => texto.trim().toLowerCase();

// Une las predeterminadas con las del backend:
// - si una predeterminada ya existe en el backend, toma su id real
// - las demás del backend son categorías propias del usuario
export function combinarCategorias(lista: CategoriaApi[]): Categoria[] {
  const predeterminadas = categoriasPredeterminadas.map((p) => {
    const enBackend = lista.find((c) => normalizar(c.nombre) === normalizar(p.nombre));
    return enBackend ? { ...p, id: enBackend.id } : p;
  });
  const nombresPred = categoriasPredeterminadas.map((p) => normalizar(p.nombre));
  const propias: Categoria[] = lista
    .filter((c) => !nombresPred.includes(normalizar(c.nombre)))
    .map((c) => ({ id: c.id, nombre: c.nombre, tipo: 'ambos', ...estiloPropio }));
  return [...predeterminadas, ...propias];
}
