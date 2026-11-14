import { Pressable, Text, View } from 'react-native';
import Icono from '../../../shared/components/Icono';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { Categoria } from '../categorias';

type Props = { categoria: Categoria; activa: boolean; onPress: () => void };

// Tag de categoría del mockup: ícono en cuadrito de color + nombre; borde del color de la categoría si está elegida
export default function ChipCategoria({ categoria: c, activa, onPress }: Props) {
  const { t } = useEscala();
  return (
    <Pressable
      onPress={onPress}
      style={{
        flexDirection: 'row', alignItems: 'center', gap: t(6), paddingHorizontal: t(10), paddingVertical: t(6),
        borderRadius: t(12), borderWidth: activa ? 2 : 1, borderColor: activa ? c.color : colors.lavandaOnda,
        backgroundColor: activa ? colors.fondoApp : colors.superficie,
        margin: activa ? 0 : 1, // Compensa el borde más grueso para que no "salte"
      }}
    >
      <View style={{ width: t(20), height: t(20), borderRadius: t(6), backgroundColor: c.colorSuave, alignItems: 'center', justifyContent: 'center' }}>
        <Icono nombre={c.icono} tamano={t(12)} color={c.color} />
      </View>
      <Text style={{ fontFamily: activa ? fonts.negrita : fonts.medio, fontSize: t(11), color: colors.texto }}>{c.nombre}</Text>
    </Pressable>
  );
}
