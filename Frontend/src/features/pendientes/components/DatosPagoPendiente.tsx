import { useState } from 'react';
import { Pressable, Text, TextInput, View } from 'react-native';
import Calendario from '../../../shared/components/Calendario';
import Icono from '../../../shared/components/Icono';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { aDMA, hoy, sumarDias } from '../../../shared/utils/fechas';
import { MAX_CLIENTE } from '../pendientes';

const PLAZOS = [
  { texto: 'Mañana', dias: 1 },
  { texto: '7 días', dias: 7 },
  { texto: '15 días', dias: 15 },
];

type Props = {
  cliente: string;
  onCliente: (v: string) => void;
  vence: Date;
  onVence: (f: Date) => void;
};

// Datos de un pago pendiente en Inicio: quién debe y cuándo paga (plazos rápidos u otra fecha con calendario)
export default function DatosPagoPendiente({ cliente, onCliente, vence, onVence }: Props) {
  const { t } = useEscala();
  const [plazo, setPlazo] = useState<number | null>(7); // null = fecha elegida en el calendario
  const [calendarioAbierto, setCalendarioAbierto] = useState(false);

  return (
    <View style={{ gap: t(6), backgroundColor: colors.naranjoSuave, borderRadius: t(12), padding: t(8) }}>
      <TextInput value={cliente} onChangeText={onCliente} maxLength={MAX_CLIENTE}
        placeholder="¿Quién te debe? (nombre del cliente)" placeholderTextColor={colors.textoSecundario}
        style={{
          paddingVertical: t(7), paddingHorizontal: t(12), fontSize: t(12), fontFamily: fonts.normal, color: colors.texto,
          backgroundColor: colors.superficie, borderWidth: 1, borderColor: colors.lavandaOnda, borderRadius: t(12),
        }} />
      <Text style={{ fontFamily: fonts.negrita, fontSize: t(10), color: colors.naranjo }}>Te paga el {aDMA(vence)}</Text>
      <View style={{ flexDirection: 'row', gap: t(4) }}>
        {[...PLAZOS, { texto: 'Otra fecha', dias: null }].map((p) => {
          const activo = p.dias === plazo;
          return (
            <Pressable key={p.texto} onPress={() => {
              if (p.dias === null) { setPlazo(null); setCalendarioAbierto(!calendarioAbierto); return; }
              setPlazo(p.dias); onVence(sumarDias(hoy(), p.dias)); setCalendarioAbierto(false);
            }} style={{
              flex: 1, flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: t(3), paddingVertical: t(5), borderRadius: t(8),
              backgroundColor: activo ? colors.naranjo : colors.superficie,
            }}>
              {p.dias === null && <Icono nombre="calendar" tamano={t(11)} color={activo ? colors.superficie : colors.naranjo} />}
              <Text style={{ fontFamily: fonts.negrita, fontSize: t(10), color: activo ? colors.superficie : colors.naranjo }}>{p.texto}</Text>
            </Pressable>
          );
        })}
      </View>
      {calendarioAbierto && <Calendario fecha={vence} onElegir={(f) => { onVence(f); setCalendarioAbierto(false); }} />}
    </View>
  );
}
