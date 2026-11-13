import { Text, View } from 'react-native';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { Barra } from '../resumen';

const ALTO = 160; // Alto máximo de las barras

// Gráfico de barras dobles (ingresos en verde, gastos en rosado) hecho con Views:
// cada barra mide su valor dividido por el mayor valor del período
export default function GraficoBarras({ barras }: { barras: Barra[] }) {
  const maximo = Math.max(1, ...barras.flatMap((b) => [b.ingresos, b.gastos]));
  const alto = (valor: number) => (valor > 0 ? Math.max(4, (valor / maximo) * ALTO) : 0);

  return (
    <View style={{ gap: 10 }}>
      <View style={{ flexDirection: 'row', alignItems: 'flex-end', height: ALTO, gap: 4 }}>
        {barras.map((b) => (
          <View key={b.etiqueta} style={{ flex: 1, flexDirection: 'row', alignItems: 'flex-end', justifyContent: 'center', gap: 2 }}>
            <View style={{ width: '40%', height: alto(b.ingresos), backgroundColor: colors.ingreso, borderRadius: 4 }} />
            <View style={{ width: '40%', height: alto(b.gastos), backgroundColor: colors.egreso, borderRadius: 4 }} />
          </View>
        ))}
      </View>

      <View style={{ flexDirection: 'row', gap: 4 }}>
        {barras.map((b) => (
          <Text key={b.etiqueta} numberOfLines={1} adjustsFontSizeToFit
            style={{ flex: 1, textAlign: 'center', fontFamily: fonts.medio, fontSize: 13, color: colors.lavanda }}>
            {b.etiqueta}
          </Text>
        ))}
      </View>

      <View style={{ flexDirection: 'row', justifyContent: 'center', gap: 20 }}>
        {[{ t: 'Ingresos', c: colors.ingreso }, { t: 'Gastos', c: colors.egreso }].map((l) => (
          <View key={l.t} style={{ flexDirection: 'row', alignItems: 'center', gap: 6 }}>
            <View style={{ width: 12, height: 12, borderRadius: 3, backgroundColor: l.c }} />
            <Text style={{ fontFamily: fonts.medio, fontSize: 14, color: colors.lavanda }}>{l.t}</Text>
          </View>
        ))}
      </View>
    </View>
  );
}
