import { Ionicons } from '@expo/vector-icons';
import { Pressable, Text, View } from 'react-native';
import Chip from '../../features/movimientos/components/Chip';
import { colors } from '../theme/colors';
import { fonts } from '../theme/fonts';
import { aDMA, aLargo, hoy, sumarDias } from '../utils/fechas';
import { useEscala } from '../utils/escala';

type Props = { etiqueta: string; fecha: Date; onCambiar: (f: Date) => void };

const atajos = [
  { texto: 'Mañana', dias: 1 },
  { texto: '1 semana', dias: 7 },
  { texto: '15 días', dias: 15 },
  { texto: '30 días', dias: 30 },
];

// Selector de fecha simple, sin librerías: botones − / + por día y atajos rápidos
export default function SelectorFecha({ etiqueta, fecha, onCambiar }: Props) {
  const { e } = useEscala();
  const minimo = hoy();
  const puedeRetroceder = fecha.getTime() > minimo.getTime();

  return (
    <View style={{ gap: 8 }}>
      <Text style={{ fontFamily: fonts.medio, fontSize: e(15), color: colors.lavanda }}>{etiqueta}</Text>
      <View style={{ flexDirection: 'row', alignItems: 'center', backgroundColor: colors.superficie, borderRadius: 16, borderWidth: 1.5, borderColor: colors.lavandaSuave, padding: 8 }}>
        <Pressable hitSlop={8} disabled={!puedeRetroceder} onPress={() => onCambiar(sumarDias(fecha, -1))} style={{ opacity: puedeRetroceder ? 1 : 0.3 }}>
          <Ionicons name="remove-circle" size={e(34)} color={colors.lavanda} />
        </Pressable>
        <View style={{ flex: 1, alignItems: 'center' }}>
          <Text style={{ fontFamily: fonts.negrita, fontSize: e(18), color: colors.lavanda }}>{aDMA(fecha)}</Text>
          <Text style={{ fontFamily: fonts.normal, fontSize: e(13), color: colors.lavanda, opacity: 0.8 }}>{aLargo(fecha)}</Text>
        </View>
        <Pressable hitSlop={8} onPress={() => onCambiar(sumarDias(fecha, 1))}>
          <Ionicons name="add-circle" size={e(34)} color={colors.lavanda} />
        </Pressable>
      </View>
      <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }}>
        {atajos.map((a) => {
          const destino = sumarDias(minimo, a.dias);
          return <Chip key={a.texto} texto={a.texto} activo={destino.getTime() === fecha.getTime()} onPress={() => onCambiar(destino)} />;
        })}
      </View>
    </View>
  );
}
