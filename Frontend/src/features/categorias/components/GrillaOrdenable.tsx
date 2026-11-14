import { ReactNode, useRef, useState } from 'react';
import { View } from 'react-native';
import { Gesture, GestureDetector } from 'react-native-gesture-handler';
import Animated, { useAnimatedStyle, useSharedValue } from 'react-native-reanimated';

type Caja = { x: number; y: number; w: number; h: number };

type Props<T> = {
  items: T[];
  clave: (item: T) => string;
  renderItem: (item: T) => ReactNode;
  onReordenar: (claves: string[]) => void;      // Se llama al soltar, con el nuevo orden
  onArrastrando?: (activo: boolean) => void;    // Para bloquear el scroll de la pantalla mientras se arrastra
};

// Grilla centrada de categorías que se reordena como los íconos de Android:
// mantener presionado → el tag "flota" → arrastrarlo → soltarlo donde se quiere dejar.
export default function GrillaOrdenable<T>({ items, clave, renderItem, onReordenar, onArrastrando }: Props<T>) {
  const cajas = useRef<Record<string, Caja>>({});
  const [orden, setOrden] = useState<string[] | null>(null); // Orden temporal mientras se arrastra
  const [arrastrada, setArrastrada] = useState<string | null>(null);
  const ordenRef = useRef<string[] | null>(null);       // Copias en ref: los gestos siempre leen el valor actual
  const arrastradaRef = useRef<string | null>(null);
  const lugares = useRef<(Caja | undefined)[]>([]);    // Posiciones fijas tomadas al empezar a arrastrar
  const x = useSharedValue(0);
  const y = useSharedValue(0);

  const porClave = new Map(items.map((i) => [clave(i), i]));
  const visibles = (orden ?? items.map(clave)).map((c) => porClave.get(c)).filter(Boolean) as T[];

  const cambiarOrden = (o: string[] | null) => { ordenRef.current = o; setOrden(o); };

  function empezar(c: string) {
    arrastradaRef.current = c;
    setArrastrada(c);
    cambiarOrden(items.map(clave));
    lugares.current = items.map((i) => cajas.current[clave(i)]);
    onArrastrando?.(true);
  }

  // Busca el lugar más cercano al dedo usando las posiciones del inicio del arrastre
  // (si se usaran las posiciones que cambian mientras se mueve, los tags saltarían de un lado a otro)
  function mover(cx: number, cy: number) {
    const actual = ordenRef.current;
    const movida = arrastradaRef.current;
    if (!actual || !movida) return;
    let destino = -1;
    let menor = Infinity;
    lugares.current.forEach((b, i) => {
      if (!b) return;
      const d = (b.x + b.w / 2 - cx) ** 2 + (b.y + b.h / 2 - cy) ** 2;
      if (d < menor) { menor = d; destino = i; }
    });
    if (destino < 0 || destino === actual.indexOf(movida)) return;
    const nuevo = actual.filter((c) => c !== movida);
    nuevo.splice(destino, 0, movida);
    cambiarOrden(nuevo);
  }

  function terminar() {
    if (!arrastradaRef.current) return;
    if (ordenRef.current) onReordenar(ordenRef.current);
    arrastradaRef.current = null;
    setArrastrada(null);
    cambiarOrden(null);
    onArrastrando?.(false);
  }

  const estiloFlotante = useAnimatedStyle(() => ({ transform: [{ translateX: x.value }, { translateY: y.value }, { scale: 1.1 }] }));
  const cajaArrastrada = arrastrada ? cajas.current[arrastrada] : undefined;
  const itemArrastrado = arrastrada ? porClave.get(arrastrada) : undefined;

  return (
    <View style={{ flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'center', gap: 6 }}>
      {visibles.map((item) => {
        const c = clave(item);
        const gesto = Gesture.Pan()
          .runOnJS(true) // Los cálculos se hacen en JavaScript (son pocos tags)
          .activateAfterLongPress(300)
          .onStart(() => {
            const b = cajas.current[c];
            x.value = b?.x ?? 0;
            y.value = b?.y ?? 0;
            empezar(c);
          })
          .onUpdate((e) => {
            const b = cajas.current[c];
            if (!b) return;
            x.value = b.x + e.translationX;
            y.value = b.y + e.translationY;
            mover(b.x + b.w / 2 + e.translationX, b.y + b.h / 2 + e.translationY);
          })
          .onFinalize(() => terminar());
        return (
          <GestureDetector key={c} gesture={gesto}>
            <View
              collapsable={false}
              onLayout={(e) => {
                const { x: lx, y: ly, width, height } = e.nativeEvent.layout;
                if (!arrastradaRef.current) cajas.current[c] = { x: lx, y: ly, w: width, h: height };
              }}
              style={{ opacity: c === arrastrada ? 0.25 : 1 }} // Hueco donde quedará el tag
            >
              {renderItem(item)}
            </View>
          </GestureDetector>
        );
      })}

      {/* Copia del tag que sigue al dedo, "flotando" sobre los demás */}
      {itemArrastrado && cajaArrastrada && (
        <Animated.View
          pointerEvents="none"
          style={[{
            position: 'absolute', left: 0, top: 0, zIndex: 10, elevation: 8,
            boxShadow: '0px 8px 16px rgba(60, 56, 82, 0.3)', borderRadius: 12,
          }, estiloFlotante]}
        >
          {renderItem(itemArrastrado)}
        </Animated.View>
      )}
    </View>
  );
}
