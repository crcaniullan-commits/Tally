import { Text, View } from 'react-native';
import Svg, { Circle, G } from 'react-native-svg';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { pesos } from '../../../shared/utils/formato';
import { Comparado } from '../calcularDashboard';
import { ItemDesglose } from '../resumen';

// Gráficos del Dashboard (Historia 006), dibujados con react-native-svg y vistas de React Native

// Dona con el total al centro y leyenda al lado (a qué corresponde cada color)
export function Torta({ items, total, etiqueta }: { items: ItemDesglose[]; total: number; etiqueta: string }) {
  const { t } = useEscala();
  const radio = 54;
  const circunferencia = 2 * Math.PI * radio;
  let acumulado = 0;
  const lado = t(132);
  return (
    <View style={{ width: '100%', flexDirection: 'row', alignItems: 'center', gap: t(12) }}>
      <View style={{ width: lado, height: lado, alignItems: 'center', justifyContent: 'center' }}>
        <Svg width={lado} height={lado} viewBox="0 0 150 150">
          <G rotation={-90} origin="75, 75">
            <Circle cx={75} cy={75} r={radio} fill="none" stroke={colors.lavandaSuave} strokeWidth={18} />
            {items.map((item) => {
              const largo = (item.porcentaje / 100) * circunferencia;
              const desfase = -(acumulado / 100) * circunferencia;
              acumulado += item.porcentaje;
              return (
                <Circle key={item.nombre} cx={75} cy={75} r={radio} fill="none" stroke={item.color} strokeWidth={18}
                  strokeDasharray={`${largo} ${circunferencia}`} strokeDashoffset={desfase} />
              );
            })}
          </G>
        </Svg>
        {/* Monto completo; si es muy largo, la letra se achica sola para caber dentro del círculo */}
        <View style={{ position: 'absolute', alignItems: 'center', width: lado * 0.56 }}>
          <Text numberOfLines={1} adjustsFontSizeToFit minimumFontScale={0.45}
            style={{ fontFamily: fonts.extra, fontSize: t(13), color: colors.texto, textAlign: 'center', width: '100%' }}>
            {pesos(total)}
          </Text>
          <Text style={{ fontFamily: fonts.medio, fontSize: t(10), color: colors.textoSecundario }}>{etiqueta}</Text>
        </View>
      </View>
      <View style={{ flex: 1, gap: t(6) }}>
        {items.filter((x) => x.total > 0).map((item) => (
          <View key={item.nombre} style={{ flexDirection: 'row', alignItems: 'center', gap: t(6) }}>
            <View style={{ width: t(10), height: t(10), borderRadius: t(3), backgroundColor: item.color }} />
            <Text numberOfLines={1} style={{ flex: 1, fontFamily: fonts.medio, fontSize: t(11), color: colors.texto }}>{item.nombre}</Text>
            <Text style={{ fontFamily: fonts.extra, fontSize: t(11), color: colors.texto }}>{item.porcentaje}%</Text>
          </View>
        ))}
      </View>
    </View>
  );
}

// Barras verticales con su porcentaje y nombre
export function BarrasVerticales({ items }: { items: ItemDesglose[] }) {
  const { t } = useEscala();
  const maximo = Math.max(...items.map((i) => i.porcentaje), 1);
  return (
    <View style={{ width: '100%', flexDirection: 'row', alignItems: 'flex-end', justifyContent: 'space-around', gap: t(8), paddingTop: t(4) }}>
      {items.map((item) => (
        <View key={item.nombre} style={{ flex: 1, alignItems: 'center', gap: t(6) }}>
          <Text style={{ fontFamily: fonts.extra, fontSize: t(10), color: colors.texto }}>{item.porcentaje}%</Text>
          <View style={{ width: '100%', maxWidth: t(38), height: t(112), borderRadius: t(12), padding: t(4), justifyContent: 'flex-end', backgroundColor: item.colorSuave }}>
            <View style={{ width: '100%', height: `${Math.max(18, Math.round((item.porcentaje / maximo) * 100))}%`, borderRadius: t(8), backgroundColor: item.color }} />
          </View>
          <Text numberOfLines={1} style={{ fontFamily: fonts.negrita, fontSize: t(10), color: colors.textoSecundario, maxWidth: t(64) }}>{item.nombre}</Text>
        </View>
      ))}
    </View>
  );
}

// Barras horizontales: una fila por elemento, crece hacia abajo
export function BarrasHorizontales({ items }: { items: ItemDesglose[] }) {
  const { t } = useEscala();
  const maximo = Math.max(...items.map((i) => i.porcentaje), 1);
  return (
    <View style={{ width: '100%', gap: t(8) }}>
      {items.map((item) => (
        <View key={item.nombre} style={{ gap: t(3) }}>
          <View style={{ flexDirection: 'row', justifyContent: 'space-between' }}>
            <Text numberOfLines={1} style={{ flex: 1, fontFamily: fonts.negrita, fontSize: t(11), color: colors.texto }}>{item.nombre}</Text>
            <Text style={{ fontFamily: fonts.extra, fontSize: t(11), color: colors.texto }}>{item.porcentaje}%</Text>
          </View>
          <View style={{ height: t(12), borderRadius: 999, backgroundColor: item.colorSuave, overflow: 'hidden' }}>
            <View style={{ width: `${Math.max(3, Math.round((item.porcentaje / maximo) * 100))}%`, height: '100%', borderRadius: 999, backgroundColor: item.color }} />
          </View>
        </View>
      ))}
    </View>
  );
}

function LeyendaComparada() {
  const { t } = useEscala();
  return (
    <View style={{ flexDirection: 'row', justifyContent: 'center', gap: t(16) }}>
      {[['Ingresos', colors.verde], ['Egresos', colors.rosado]].map(([n, c]) => (
        <View key={n} style={{ flexDirection: 'row', alignItems: 'center', gap: t(4) }}>
          <View style={{ width: t(10), height: t(10), borderRadius: t(3), backgroundColor: c }} />
          <Text style={{ fontFamily: fonts.medio, fontSize: t(11), color: colors.texto }}>{n}</Text>
        </View>
      ))}
    </View>
  );
}

// Ingresos vs egresos por fecha: dos barras por período
export function ComparadoVertical({ comparado }: { comparado: Comparado[] }) {
  const { t } = useEscala();
  const maximo = Math.max(1, ...comparado.flatMap((c) => [c.ing, c.egr]));
  return (
    <View style={{ gap: t(8) }}>
      <View style={{ flexDirection: 'row', alignItems: 'flex-end', height: t(130), gap: t(4) }}>
        {comparado.map((c) => (
          <View key={c.nombre} style={{ flex: 1, alignItems: 'center', gap: t(4) }}>
            <View style={{ flexDirection: 'row', alignItems: 'flex-end', gap: t(2), height: t(110) }}>
              <View style={{ width: t(9), height: `${Math.max(2, (c.ing / maximo) * 100)}%`, borderRadius: t(4), backgroundColor: colors.verde }} />
              <View style={{ width: t(9), height: `${Math.max(2, (c.egr / maximo) * 100)}%`, borderRadius: t(4), backgroundColor: colors.rosado }} />
            </View>
            <Text numberOfLines={1} style={{ fontFamily: fonts.negrita, fontSize: t(9), color: colors.textoSecundario }}>{c.nombre}</Text>
          </View>
        ))}
      </View>
      <LeyendaComparada />
    </View>
  );
}

// Ingresos vs egresos por fecha, en filas (monto completo; la letra se achica si no cabe)
export function ComparadoHorizontal({ comparado }: { comparado: Comparado[] }) {
  const { t } = useEscala();
  const maximo = Math.max(1, ...comparado.flatMap((c) => [c.ing, c.egr]));
  return (
    <View style={{ gap: t(10) }}>
      {comparado.map((c) => (
        <View key={c.nombre} style={{ gap: t(3) }}>
          <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: colors.texto }}>{c.nombre}</Text>
          {([[c.ing, colors.verde, colors.verdeSuave], [c.egr, colors.rosado, colors.rosadoSuave]] as const).map(([v, col, suave], i) => (
            <View key={i} style={{ flexDirection: 'row', alignItems: 'center', gap: t(6) }}>
              <View style={{ flex: 1, height: t(9), borderRadius: 999, backgroundColor: suave, overflow: 'hidden' }}>
                <View style={{ width: `${Math.max(2, (v / maximo) * 100)}%`, height: '100%', borderRadius: 999, backgroundColor: col }} />
              </View>
              <Text numberOfLines={1} adjustsFontSizeToFit minimumFontScale={0.6}
                style={{ fontFamily: fonts.extra, fontSize: t(10), color: col, width: t(78), textAlign: 'right' }}>{pesos(v)}</Text>
            </View>
          ))}
        </View>
      ))}
      <LeyendaComparada />
    </View>
  );
}
