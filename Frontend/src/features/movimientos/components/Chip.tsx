import { Pressable, Text } from 'react-native';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';

type ChipProps = { texto: string; activo: boolean; onPress: () => void; color?: string };

// Botón pequeño tipo "píldora" para filtros y selecciones rápidas
export default function Chip({ texto, activo, onPress, color = colors.lavanda }: ChipProps) {
  const { e } = useEscala();
  return (
    <Pressable
      onPress={onPress}
      style={{
        paddingVertical: e(8), paddingHorizontal: e(14), borderRadius: 999,
        backgroundColor: activo ? color : colors.superficie,
        borderWidth: 1.5, borderColor: activo ? color : colors.lavandaSuave,
      }}
    >
      <Text style={{ fontFamily: activo ? fonts.negrita : fonts.medio, fontSize: e(15), color: activo ? colors.superficie : colors.lavanda }}>
        {texto}
      </Text>
    </Pressable>
  );
}
