import { createContext, ReactNode, useCallback, useContext, useEffect, useRef, useState } from 'react';
import { Animated, Text } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors } from '../theme/colors';
import { fonts } from '../theme/fonts';
import { useEscala } from '../utils/escala';
import Icono from './Icono';

// Avisos flotantes de Tally (estilo del mockup, con los colores de la app).
// Cualquier pantalla dentro de las pestañas llama: const avisar = useAviso(); avisar('¡Ingreso registrado!')
type Aviso = { id: number; mensaje: string; icono: string };
const ContextoAviso = createContext<(mensaje: string, icono?: string) => void>(() => {});
export const useAviso = () => useContext(ContextoAviso);

export function AvisoProvider({ children }: { children: ReactNode }) {
  const [cola, setCola] = useState<Aviso[]>([]);
  const contador = useRef(0);
  const avisar = useCallback((mensaje: string, icono = 'check-circle') => {
    contador.current += 1;
    setCola((c) => [...c, { id: contador.current, mensaje, icono }]);
  }, []);
  const terminar = useCallback(() => setCola((c) => c.slice(1)), []);

  return (
    <ContextoAviso.Provider value={avisar}>
      {children}
      {cola[0] && <Burbuja key={cola[0].id} aviso={cola[0]} onTerminar={terminar} />}
    </ContextoAviso.Provider>
  );
}

// Se muestra 2 segundos y se va; si hay más avisos en cola, aparece el siguiente
function Burbuja({ aviso, onTerminar }: { aviso: Aviso; onTerminar: () => void }) {
  const { t } = useEscala();
  const insets = useSafeAreaInsets();
  const opacidad = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    Animated.timing(opacidad, { toValue: 1, duration: 180, useNativeDriver: true }).start();
    const id = setTimeout(() => {
      Animated.timing(opacidad, { toValue: 0, duration: 180, useNativeDriver: true }).start(() => onTerminar());
    }, 2000);
    return () => clearTimeout(id);
  }, [opacidad, onTerminar]);

  return (
    <Animated.View
      pointerEvents="none"
      style={{
        position: 'absolute', top: insets.top + t(12), alignSelf: 'center', zIndex: 50, opacity: opacidad,
        transform: [{ translateY: opacidad.interpolate({ inputRange: [0, 1], outputRange: [-8, 0] }) }],
        flexDirection: 'row', alignItems: 'center', gap: t(8), backgroundColor: colors.lavandaSuave,
        borderWidth: 2, borderColor: colors.lavanda,
        paddingHorizontal: t(16), paddingVertical: t(9), borderRadius: 999, maxWidth: '90%',
        boxShadow: '0px 8px 20px rgba(60, 56, 82, 0.35)',
      }}
    >
      <Icono nombre={aviso.icono} tamano={t(16)} color={colors.lavanda} />
      <Text style={{ color: colors.lavanda, fontFamily: fonts.extra, fontSize: t(12) }}>{aviso.mensaje}</Text>
    </Animated.View>
  );
}
