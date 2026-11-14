import { Text, View } from 'react-native';
import Icono from '../../../shared/components/Icono';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { pesos } from '../../../shared/utils/formato';
import { ItemDesglose } from '../resumen';

// Lista bajo el gráfico: ícono, nombre, porcentaje y monto. Comparando fechas muestra el balance de cada período.
export default function ListaDesglose({ items, comparando }: { items: ItemDesglose[]; comparando: boolean }) {
  const { t } = useEscala();
  return (
    <View style={{ gap: t(10), paddingTop: t(4), borderTopWidth: 1, borderTopColor: colors.lavandaSuave }}>
      {comparando && <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: colors.textoSecundario }}>Balance de cada período</Text>}
      {items.map((item) => (
        <View key={item.nombre} style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', paddingTop: t(4) }}>
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: t(10), flex: 1 }}>
            <View style={{ width: t(32), height: t(32), borderRadius: t(12), backgroundColor: item.colorSuave, alignItems: 'center', justifyContent: 'center' }}>
              <Icono nombre={item.icono} tamano={t(16)} color={item.color} />
            </View>
            <Text numberOfLines={1} style={{ fontFamily: fonts.negrita, fontSize: t(12), color: colors.texto, flexShrink: 1 }}>{item.nombre}</Text>
          </View>
          <View style={{ flexDirection: 'row', alignItems: 'center', gap: t(12) }}>
            {!comparando && <Text style={{ fontFamily: fonts.medio, fontSize: t(12), color: colors.textoSecundario }}>{item.porcentaje}%</Text>}
            <Text style={{ fontFamily: fonts.extra, fontSize: t(12), minWidth: t(64), textAlign: 'right', color: comparando ? (item.total >= 0 ? colors.verde : colors.rosado) : colors.texto }}>
              {comparando && item.total >= 0 ? '+' : ''}{pesos(item.total)}
            </Text>
          </View>
        </View>
      ))}
    </View>
  );
}
