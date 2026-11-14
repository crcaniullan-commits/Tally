import { ReactNode } from 'react';
import { Text, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { colors } from '../theme/colors';
import { fonts } from '../theme/fonts';
import { useEscala } from '../utils/escala';

type Props = {
  titulo: string;
  subtitulo?: string;   // Normalmente el nombre del emprendimiento
  derecha?: ReactNode;  // Botón opcional a la derecha (ej: Descargar Excel)
  children?: ReactNode; // Contenido extra dentro de la barra (Inicio)
};

// Barra superior de cada sección (va dentro del ScrollView: se desplaza junto con el contenido): lavanda, esquinas inferiores redondeadas y burbujas suaves de la marca
export default function BarraTitulo({ titulo, subtitulo, derecha, children }: Props) {
  const { t } = useEscala();
  const insets = useSafeAreaInsets();
  return (
    <View style={{
      backgroundColor: colors.lavanda, paddingVertical: t(12), paddingHorizontal: t(16),
      marginTop: insets.top + t(10), marginHorizontal: t(12),
      borderRadius: t(18), overflow: 'hidden',
      boxShadow: '0px 6px 14px rgba(135, 120, 205, 0.35)',
    }}>
      {/* Burbujas decorativas */}
      <View style={{ position: 'absolute', width: t(140), height: t(140), borderRadius: 999, backgroundColor: 'rgba(255,255,255,0.08)', top: -t(50), right: -t(30) }} />
      <View style={{ position: 'absolute', width: t(80), height: t(80), borderRadius: 999, backgroundColor: 'rgba(255,255,255,0.07)', bottom: -t(36), right: t(90) }} />

      <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: t(10) }}>
        <View style={{ flex: 1 }}>
          {subtitulo ? (
            <Text numberOfLines={1} style={{ fontFamily: fonts.medio, fontSize: t(11), color: colors.lavandaSuave }}>{subtitulo}</Text>
          ) : null}
          <Text numberOfLines={1} style={{ fontFamily: fonts.extra, fontSize: t(17), color: colors.superficie }}>{titulo}</Text>
        </View>
        {derecha}
      </View>
      {children}
    </View>
  );
}
