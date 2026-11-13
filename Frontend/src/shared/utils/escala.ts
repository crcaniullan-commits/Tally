import { useWindowDimensions } from 'react-native';

// Ajusta los tamaños al ancho de la pantalla. El diseño está pensado para 390 px de ancho:
// en celulares más chicos todo se achica un poco y en los más grandes crece un poco.
export function useEscala() {
  const { width, height } = useWindowDimensions();
  const factor = Math.min(Math.max(width / 390, 0.82), 1.15);
  const e = (tamano: number) => Math.round(tamano * factor);
  return { e, ancho: width, alto: height };
}
