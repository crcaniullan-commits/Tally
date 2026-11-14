import { useFocusEffect } from 'expo-router';
import { useCallback, useEffect, useState } from 'react';
import { Text, View } from 'react-native';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { formatearMonto } from '../../../shared/utils/formato';
import { useSesion } from '../../auth/SesionContext';
import { gastosDesde, ingresosDesde } from '../../movimientos/movimientos.service';

const DESDE_SIEMPRE = '2000-01-01';

// Recuento general (todo el historial): no cambia con los filtros de Registros.
// version: se incrementa cuando se edita algo, para volver a calcular.
export default function ResumenGeneral({ version = 0 }: { version?: number }) {
  const { token, empresa } = useSesion();
  const { t } = useEscala();
  const [totales, setTotales] = useState<{ ingresos: number; egresos: number } | null>(null);

  const cargar = useCallback(async () => {
    if (!token || !empresa) return;
    try {
      const [ing, egr] = await Promise.all([ingresosDesde(DESDE_SIEMPRE, empresa.id, token), gastosDesde(DESDE_SIEMPRE, empresa.id, token)]);
      setTotales({ ingresos: ing.reduce((s, x) => s + x.monto, 0), egresos: egr.reduce((s, x) => s + x.monto, 0) });
    } catch {
      // Si falla se quedan los últimos valores
    }
  }, [token, empresa]);

  useFocusEffect(useCallback(() => { cargar(); }, [cargar]));
  useEffect(() => { if (version) cargar(); }, [version, cargar]);

  const balance = (totales?.ingresos ?? 0) - (totales?.egresos ?? 0);
  return (
    <View style={{ flexDirection: 'row', gap: t(8) }}>
      {[
        { titulo: 'Ingresos', valor: totales?.ingresos ?? 0, color: colors.ingreso },
        { titulo: 'Egresos', valor: totales?.egresos ?? 0, color: colors.egreso },
        { titulo: 'Balance', valor: balance, color: colors.azul },
      ].map((x) => (
        <View key={x.titulo} style={{
          flex: 1, backgroundColor: colors.superficie, padding: t(10), borderRadius: t(12), borderWidth: 1, borderColor: colors.lavandaOnda,
          boxShadow: '0px 1px 2px rgba(0, 0, 0, 0.05)',
        }}>
          <Text style={{ fontFamily: fonts.negrita, fontSize: t(10), color: colors.textoSecundario }}>{x.titulo}</Text>
          <Text numberOfLines={1} adjustsFontSizeToFit style={{ fontFamily: fonts.extra, fontSize: t(14), color: x.color }}>
            {totales ? `${x.valor < 0 ? '-' : ''}$${formatearMonto(Math.abs(x.valor))}` : '…'}
          </Text>
        </View>
      ))}
    </View>
  );
}
