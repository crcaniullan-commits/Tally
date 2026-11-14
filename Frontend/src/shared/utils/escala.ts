import { useWindowDimensions } from 'react-native';

// Ajusta los tamaños al ancho de la pantalla.
// e(): tamaños de las pantallas de bienvenida/login (diseño original, un poco reducido).
// t(): tamaños del mockup de la app (pensado para 364 px de ancho útil): t(12) = texto "xs" del mockup.
export function useEscala() {
  const { width, height } = useWindowDimensions();
  const COMPACTO = 0.86;
  const factor = Math.min(Math.max(width / 390, 0.82), 1.15) * COMPACTO;
  const factorMockup = Math.min(Math.max(width / 364, 0.95), 1.2);
  const e = (tamano: number) => Math.round(tamano * factor);
  const t = (tamano: number) => Math.round(tamano * factorMockup);
  return { e, t, ancho: width, alto: height };
}
