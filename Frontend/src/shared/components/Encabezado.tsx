import { Ionicons } from '@expo/vector-icons';
import { router } from 'expo-router';
import { ReactNode } from 'react';
import { Pressable, Text, View } from 'react-native';
import { colors } from '../theme/colors';
import { fonts } from '../theme/fonts';
import { useEscala } from '../utils/escala';

type Props = { titulo: string; subtitulo?: string; derecha?: ReactNode; volver?: boolean };

// Encabezado de las pantallas secundarias: flecha para volver, título y una acción opcional a la derecha
export default function Encabezado({ titulo, subtitulo, derecha, volver = true }: Props) {
  const { e } = useEscala();
  return (
    <View style={{ flexDirection: 'row', alignItems: 'center', gap: 12 }}>
      {volver && (
        <Pressable hitSlop={10} onPress={() => router.back()}>
          <Ionicons name="chevron-back" size={e(30)} color={colors.lavanda} />
        </Pressable>
      )}
      <View style={{ flex: 1 }}>
        <Text numberOfLines={1} adjustsFontSizeToFit style={{ fontFamily: fonts.negrita, fontSize: e(28), color: colors.lavanda }}>{titulo}</Text>
        {subtitulo ? (
          <Text numberOfLines={1} style={{ fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda, opacity: 0.8 }}>{subtitulo}</Text>
        ) : null}
      </View>
      {derecha}
    </View>
  );
}
