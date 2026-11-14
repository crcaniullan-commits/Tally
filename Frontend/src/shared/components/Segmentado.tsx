import { Pressable, Text, View } from 'react-native';
import { colors } from '../theme/colors';
import { fonts } from '../theme/fonts';
import { useEscala } from '../utils/escala';

export type OpcionSegmentada<T extends string> = { id: T; nombre: string; color?: string };

type Props<T extends string> = {
  opciones: OpcionSegmentada<T>[];
  valor: T;
  onCambiar: (id: T) => void;
  colorInactivo?: string;  // Color del texto de las opciones no elegidas
  borde?: boolean;         // Borde lavanda-onda alrededor (dashboard)
  tamano?: number;         // Tamaño del texto en px del mockup
  alto?: number;           // Relleno vertical de cada opción
};

// Selector segmentado del mockup: fondo lavanda suave y la opción elegida rellena con su color
export default function Segmentado<T extends string>({
  opciones, valor, onCambiar, colorInactivo = colors.lavanda, borde = false, tamano = 11, alto = 6,
}: Props<T>) {
  const { t } = useEscala();
  return (
    <View style={{
      flexDirection: 'row', gap: t(4), backgroundColor: colors.lavandaSuave, padding: t(4), borderRadius: t(12),
      borderWidth: borde ? 1 : 0, borderColor: colors.lavandaOnda,
    }}>
      {opciones.map((o) => {
        const activo = o.id === valor;
        return (
          <Pressable
            key={o.id}
            onPress={() => onCambiar(o.id)}
            style={{
              flex: 1, paddingVertical: t(alto), borderRadius: t(8), alignItems: 'center',
              backgroundColor: activo ? (o.color ?? colors.lavanda) : 'transparent',
              boxShadow: activo ? '0px 1px 2px rgba(0, 0, 0, 0.08)' : undefined,
            }}
          >
            <Text numberOfLines={1} style={{
              fontFamily: activo ? fonts.extra : fonts.negrita, fontSize: t(tamano),
              color: activo ? colors.superficie : colorInactivo,
            }}>
              {o.nombre}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}
