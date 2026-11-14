import { Pressable, Text, View } from 'react-native';
import Icono from '../../../shared/components/Icono';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { ID_PENDIENTE, metodosPago } from '../metodosPago';

type Props = { metodo: string; onCambiar: (metodo: string) => void };

// Método de pago (grilla de 4) y botón "Pago pendiente" (Historia 008). Solo se muestra en ingresos.
export default function SelectorMetodo({ metodo, onCambiar }: Props) {
  const { t } = useEscala();
  const esPendiente = metodo === ID_PENDIENTE;
  return (
    <View style={{ gap: t(10) }}>
      <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: colors.lavanda }}>Método de pago</Text>
      <View style={{ flexDirection: 'row', gap: t(6) }}>
        {metodosPago.map((m) => {
          const activo = m.id === metodo;
          return (
            <Pressable key={m.id} onPress={() => onCambiar(m.id)} style={{
              flex: 1, flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: t(4),
              paddingVertical: t(6), paddingHorizontal: t(2), borderRadius: t(12),
              borderWidth: activo ? 2 : 1, borderColor: activo ? colors.lavanda : colors.lavandaOnda,
              backgroundColor: activo ? colors.lavandaSuave : colors.superficie, margin: activo ? 0 : 1,
            }}>
              <Icono nombre={m.icono} tamano={t(14)} color={activo ? colors.lavanda : colors.textoSecundario} />
              <Text numberOfLines={1} style={{ fontFamily: activo ? fonts.negrita : fonts.medio, fontSize: t(10), color: activo ? colors.lavanda : colors.textoSecundario }}>
                {m.corto}
              </Text>
            </Pressable>
          );
        })}
      </View>

      {/* Pago pendiente: siempre destacado con borde naranjo */}
      <Pressable
        onPress={() => onCambiar(esPendiente ? 'efectivo' : ID_PENDIENTE)}
        style={({ pressed }) => ({
          flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: t(6), paddingVertical: t(7),
          borderRadius: t(12), borderWidth: 2, borderColor: colors.naranjo,
          backgroundColor: esPendiente ? colors.naranjo : pressed ? colors.naranjoSuave : colors.superficie,
        })}
      >
        <Icono nombre={esPendiente ? 'check-circle' : 'clock'} tamano={t(14)} color={esPendiente ? colors.superficie : colors.naranjo} />
        <Text style={{ fontFamily: fonts.extra, fontSize: t(11), color: esPendiente ? colors.superficie : colors.naranjo }}>Pago pendiente</Text>
      </Pressable>
    </View>
  );
}
