import { useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import { colors } from '../theme/colors';
import { fonts } from '../theme/fonts';
import { useEscala } from '../utils/escala';
import { hoy } from '../utils/fechas';
import Icono from './Icono';

const MESES = ['Enero', 'Febrero', 'Marzo', 'Abril', 'Mayo', 'Junio', 'Julio', 'Agosto', 'Septiembre', 'Octubre', 'Noviembre', 'Diciembre'];
const DIAS = ['L', 'M', 'M', 'J', 'V', 'S', 'D'];

type Props = { fecha: Date; onElegir: (f: Date) => void; color?: string };

// Mini calendario sin librerías: se elige un día (desde hoy en adelante) y se avisa con onElegir
export default function Calendario({ fecha, onElegir, color = colors.naranjo }: Props) {
  const { t } = useEscala();
  const [mes, setMes] = useState(new Date(fecha.getFullYear(), fecha.getMonth(), 1));
  const minimo = hoy();
  const primerDia = (mes.getDay() + 6) % 7; // Lunes = 0
  const diasDelMes = new Date(mes.getFullYear(), mes.getMonth() + 1, 0).getDate();
  const celdas: (number | null)[] = [...Array(primerDia).fill(null), ...Array.from({ length: diasDelMes }, (_, i) => i + 1)];
  while (celdas.length % 7) celdas.push(null);
  const puedeRetroceder = mes > new Date(minimo.getFullYear(), minimo.getMonth(), 1);

  return (
    <View style={{ backgroundColor: colors.superficie, borderRadius: t(12), padding: t(8), gap: t(4), borderWidth: 1, borderColor: colors.lavandaOnda }}>
      <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
        <Pressable hitSlop={8} disabled={!puedeRetroceder} onPress={() => setMes(new Date(mes.getFullYear(), mes.getMonth() - 1, 1))}
          style={{ padding: t(4), opacity: puedeRetroceder ? 1 : 0.3, transform: [{ rotate: '180deg' }] }}>
          <Icono nombre="chevron-right" tamano={t(16)} color={color} />
        </Pressable>
        <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.texto }}>{MESES[mes.getMonth()]} {mes.getFullYear()}</Text>
        <Pressable hitSlop={8} onPress={() => setMes(new Date(mes.getFullYear(), mes.getMonth() + 1, 1))} style={{ padding: t(4) }}>
          <Icono nombre="chevron-right" tamano={t(16)} color={color} />
        </Pressable>
      </View>
      <View style={{ flexDirection: 'row' }}>
        {DIAS.map((d, i) => (
          <Text key={i} style={{ flex: 1, textAlign: 'center', fontFamily: fonts.negrita, fontSize: t(10), color: colors.textoSecundario }}>{d}</Text>
        ))}
      </View>
      {Array.from({ length: celdas.length / 7 }, (_, fila) => (
        <View key={fila} style={{ flexDirection: 'row' }}>
          {celdas.slice(fila * 7, fila * 7 + 7).map((dia, i) => {
            if (!dia) return <View key={i} style={{ flex: 1 }} />;
            const f = new Date(mes.getFullYear(), mes.getMonth(), dia);
            const pasado = f < minimo;
            const elegido = f.toDateString() === fecha.toDateString();
            return (
              <Pressable key={i} disabled={pasado} onPress={() => onElegir(f)} style={{ flex: 1, alignItems: 'center', paddingVertical: t(2) }}>
                <View style={{ width: t(28), height: t(28), borderRadius: 999, alignItems: 'center', justifyContent: 'center', backgroundColor: elegido ? color : 'transparent' }}>
                  <Text style={{ fontFamily: elegido ? fonts.extra : fonts.medio, fontSize: t(11), color: elegido ? colors.superficie : pasado ? colors.lavandaOnda : colors.texto }}>{dia}</Text>
                </View>
              </Pressable>
            );
          })}
        </View>
      ))}
    </View>
  );
}
