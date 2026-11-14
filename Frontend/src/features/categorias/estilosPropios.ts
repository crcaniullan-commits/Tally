import { guardarLocal, leerLocal } from '../../shared/utils/almacen';

// Apariencia que el usuario elige para sus categorías. El backend solo guarda el nombre (y no distingue
// ingreso/egreso), así que el color, el ícono, el tipo y el nombre visible de las predeterminadas se guardan en el celular.
export type EstiloCategoria = {
  icono: string;
  color: string;
  colorSuave: string;
  nombre?: string;                 // Nombre visible de una predeterminada renombrada
  tipo?: 'ingreso' | 'egreso';     // Tipo de una categoría propia
};
export type EstilosPropios = Record<string, EstiloCategoria>; // clave de la categoría → estilo

// Colores (intenso + suave) que combinan con la marca. Los 5 primeros son los de Tally.
export const OPCIONES_COLOR = [
  { color: '#3ecd98', colorSuave: '#d3fdee' }, // verde
  { color: '#8778cd', colorSuave: '#eaedfa' }, // lavanda
  { color: '#f29f5e', colorSuave: '#ffe8d5' }, // naranjo
  { color: '#f82f8f', colorSuave: '#ffdfef' }, // rosado
  { color: '#1a8dc0', colorSuave: '#dcf9fd' }, // azul
  { color: '#14b8a6', colorSuave: '#d5f7f2' }, // turquesa
  { color: '#e0a800', colorSuave: '#fff4cc' }, // amarillo
  { color: '#f0645a', colorSuave: '#ffe2df' }, // coral
  { color: '#a259d9', colorSuave: '#f3e6fd' }, // morado
  { color: '#4fb3e8', colorSuave: '#e0f3fc' }, // celeste
  { color: '#b07a4f', colorSuave: '#f5e8dc' }, // café
  { color: '#7a7f99', colorSuave: '#eceef4' }, // gris
];

// Íconos para emprendimientos (Lucide)
export const OPCIONES_ICONO = [
  'tag', 'shopping-bag', 'shopping-cart', 'store', 'handshake', 'clipboard-list', 'package', 'truck',
  'cake', 'cookie', 'coffee', 'utensils', 'pizza', 'croissant', 'ice-cream', 'sandwich',
  'apple', 'wine', 'beer', 'leaf', 'shirt', 'gem', 'glasses', 'scissors',
  'sparkles', 'brush', 'palette', 'flower', 'gift', 'heart', 'baby', 'dog',
  'paw-print', 'car', 'bike', 'fuel', 'wrench', 'hammer', 'smartphone', 'laptop',
  'camera', 'printer', 'book-open', 'graduation-cap', 'dumbbell', 'stethoscope', 'home', 'building-2',
  'zap', 'droplet', 'wifi', 'megaphone', 'receipt', 'banknote', 'piggy-bank', 'wallet',
  'users', 'briefcase', 'plane', 'hand-coins',
];

const clave = (empresaId: string) => `estilos_categorias_${empresaId}`;

export const leerEstilos = (empresaId: string) => leerLocal<EstilosPropios>(clave(empresaId), {});

export async function guardarEstilo(empresaId: string, claveCategoria: string, estilo: EstiloCategoria): Promise<void> {
  const estilos = await leerEstilos(empresaId);
  await guardarLocal(clave(empresaId), { ...estilos, [claveCategoria]: estilo });
}
