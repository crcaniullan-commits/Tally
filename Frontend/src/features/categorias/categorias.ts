import { colors } from '../../shared/theme/colors';
import type { EstilosPropios } from './estilosPropios';

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
  icono: string;      // Nombre del ícono (Lucide, ver shared/components/Icono)
  color: string;      // Color intenso: ícono y borde al seleccionar
  colorSuave: string; // Color suave: fondo del ícono
  tipo: TipoMovimiento | 'ambos'; // 'ambos' = categoría propia antigua (sin tipo guardado)
  clave: string;      // Dónde se guarda su estilo: id predeterminado (pred-*) o id del backend
  propia: boolean;    // true = creada por el usuario (se puede eliminar)
};

export const MAX_NOMBRE_CATEGORIA = 30; // Límite del backend
export const PREFIJO_LOCAL = 'pred-';

// Categorías predeterminadas (Historia 003): siempre visibles, 3 por tipo
export const categoriasPredeterminadas: Categoria[] = [
  { id: 'pred-ventas', nombre: 'Ventas', icono: 'shopping-bag', color: colors.verde, colorSuave: colors.verdeSuave, tipo: 'ingreso', clave: 'pred-ventas', propia: false },
  { id: 'pred-servicios', nombre: 'Servicios', icono: 'handshake', color: colors.azul, colorSuave: colors.azulSuave, tipo: 'ingreso', clave: 'pred-servicios', propia: false },
  { id: 'pred-pedidos', nombre: 'Pedidos', icono: 'clipboard-list', color: colors.naranjo, colorSuave: colors.naranjoSuave, tipo: 'ingreso', clave: 'pred-pedidos', propia: false },
  { id: 'pred-insumos', nombre: 'Insumos', icono: 'package', color: colors.naranjo, colorSuave: colors.naranjoSuave, tipo: 'egreso', clave: 'pred-insumos', propia: false },
  { id: 'pred-transporte', nombre: 'Transporte', icono: 'truck', color: colors.azul, colorSuave: colors.azulSuave, tipo: 'egreso', clave: 'pred-transporte', propia: false },
  { id: 'pred-gastos-fijos', nombre: 'Gastos fijos', icono: 'building-2', color: colors.rosado, colorSuave: colors.rosadoSuave, tipo: 'egreso', clave: 'pred-gastos-fijos', propia: false },
];

const estiloPropio = { icono: 'tag', color: colors.lavanda, colorSuave: colors.lavandaSuave };
export const normalizar = (texto: string) => texto.trim().toLowerCase();

// Une las predeterminadas con las del backend:
// - si una predeterminada ya existe en el backend, toma su id real (se busca por su nombre original)
// - las demás del backend son categorías propias del usuario, con el tipo guardado en el celular
// - los estilos guardados (color, ícono y nombre visible) se aplican encima
export function combinarCategorias(lista: CategoriaApi[], estilos: EstilosPropios = {}): Categoria[] {
  const predeterminadas = categoriasPredeterminadas.map((p) => {
    const enBackend = lista.find((c) => normalizar(c.nombre) === normalizar(p.nombre));
    const e = estilos[p.clave];
    return { ...p, id: enBackend?.id ?? p.id, ...(e ? { icono: e.icono, color: e.color, colorSuave: e.colorSuave, nombre: e.nombre || p.nombre } : {}) };
  });
  const nombresPred = categoriasPredeterminadas.map((p) => normalizar(p.nombre));
  const propias: Categoria[] = lista
    .filter((c) => !nombresPred.includes(normalizar(c.nombre)))
    .map((c) => {
      const e = estilos[c.id];
      return {
        id: c.id, nombre: c.nombre, clave: c.id, propia: true, tipo: e?.tipo ?? 'ambos',
        icono: e?.icono ?? estiloPropio.icono, color: e?.color ?? estiloPropio.color, colorSuave: e?.colorSuave ?? estiloPropio.colorSuave,
      };
    });
  return [...predeterminadas, ...propias];
}
