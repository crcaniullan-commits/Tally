import { useFocusEffect } from 'expo-router';
import { useCallback, useState } from 'react';
import { ActivityIndicator, Pressable, RefreshControl, ScrollView, Text, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { traducirError } from '../../features/auth/errores';
import { useSesion } from '../../features/auth/SesionContext';
import { categoriasPredeterminadas, combinarCategorias } from '../../features/categorias/categorias';
import { listarCategorias } from '../../features/categorias/categorias.service';
import GraficoBarras from '../../features/dashboard/components/GraficoBarras';
import { aTexto, calcularResumen, inicioPeriodo, Periodo, Resumen } from '../../features/dashboard/resumen';
import TarjetaMeta from '../../features/metas/components/TarjetaMeta';
import { gastosDesde, ingresosDesde } from '../../features/movimientos/movimientos.service';
import { colors } from '../../shared/theme/colors';
import { fonts } from '../../shared/theme/fonts';
import { sombras } from '../../shared/theme/sombras';
import { useEscala } from '../../shared/utils/escala';
import { formatearMonto } from '../../shared/utils/formato';

const periodos: { id: Periodo; nombre: string }[] = [
  { id: 'semana', nombre: 'Semana' },
  { id: 'mes', nombre: 'Mes' },
  { id: 'anio', nombre: 'Año' },
];

const tarjeta = { backgroundColor: colors.superficie, borderRadius: 20, padding: 16, ...sombras.suave };

export default function Dashboard() {
  const { token, empresa, manejarError } = useSesion();
  const { e } = useEscala();
  const insets = useSafeAreaInsets();
  const [periodo, setPeriodo] = useState<Periodo>('mes');
  const [resumen, setResumen] = useState<Resumen | null>(null);
  const [ventasMes, setVentasMes] = useState(0); // Para la meta mensual
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const titulo = { fontFamily: fonts.negrita, fontSize: e(22), color: colors.lavanda };

  const cargar = useCallback(async () => {
    if (!token || !empresa) return;
    setCargando(true);
    setError(null);
    try {
      const desde = aTexto(inicioPeriodo(periodo));
      const desdeMes = aTexto(inicioPeriodo('mes'));
      const [ingresos, gastos, cats, ingresosMes] = await Promise.all([
        ingresosDesde(desde, empresa.id, token),
        gastosDesde(desde, empresa.id, token),
        listarCategorias(empresa.id, token).catch(() => []),
        periodo === 'mes' ? Promise.resolve(null) : ingresosDesde(desdeMes, empresa.id, token),
      ]);
      setVentasMes((ingresosMes ?? ingresos).reduce((s, i) => s + i.monto, 0));
      const categorias = cats.length ? combinarCategorias(cats) : categoriasPredeterminadas;
      setResumen(calcularResumen(periodo, ingresos, gastos, categorias));
    } catch (err) {
      manejarError(err);
      setError(traducirError(err));
    } finally {
      setCargando(false);
    }
  }, [token, empresa, periodo, manejarError]);

  // Se recarga cada vez que se entra a la pestaña (por si se registró algo en Inicio)
  useFocusEffect(useCallback(() => { cargar(); }, [cargar]));

  return (
    <ScrollView
      style={{ flex: 1, backgroundColor: colors.fondoApp }}
      contentContainerStyle={{ paddingHorizontal: 20, paddingTop: insets.top + 16, paddingBottom: 32, gap: 20 }}
      refreshControl={<RefreshControl refreshing={false} onRefresh={cargar} />}
    >
      <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
        <View style={{ flex: 1 }}>
          <Text style={{ fontFamily: fonts.negrita, fontSize: e(30), color: colors.lavanda }}>Salud financiera</Text>
          <Text numberOfLines={1} style={{ fontFamily: fonts.medio, fontSize: e(17), color: colors.lavanda, opacity: 0.8 }}>
            {empresa?.nombre_empresa ?? ''}
          </Text>
        </View>
      </View>

      {/* Selector de período */}
      <View style={{ flexDirection: 'row', borderRadius: 26, backgroundColor: colors.lavandaSuave }}>
        {periodos.map((p) => {
          const activo = p.id === periodo;
          return (
            <Pressable key={p.id} onPress={() => setPeriodo(p.id)}
              style={{ flex: 1, paddingVertical: e(12), borderRadius: 26, alignItems: 'center', backgroundColor: activo ? colors.lavanda : 'transparent' }}>
              <Text style={{ fontFamily: fonts.negrita, fontSize: e(17), color: activo ? colors.superficie : colors.lavanda }}>{p.nombre}</Text>
            </Pressable>
          );
        })}
      </View>

      {cargando && !resumen ? (
        <ActivityIndicator color={colors.lavanda} style={{ paddingVertical: 40 }} />
      ) : error ? (
        <Text style={{ fontFamily: fonts.medio, fontSize: e(16), color: colors.error, textAlign: 'center' }}>{error}</Text>
      ) : resumen && (
        <>
          {/* Totales del período */}
          <View style={{ flexDirection: 'row', gap: 12 }}>
            {[
              { t: 'Ingresos', v: resumen.ingresos, c: colors.ingreso },
              { t: 'Gastos', v: resumen.gastos, c: colors.egreso },
            ].map((x) => (
              <View key={x.t} style={[tarjeta, { flex: 1, gap: 4 }]}>
                <Text style={{ fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda }}>{x.t}</Text>
                <Text numberOfLines={1} adjustsFontSizeToFit style={{ fontFamily: fonts.negrita, fontSize: e(24), color: x.c }}>
                  ${formatearMonto(x.v)}
                </Text>
              </View>
            ))}
          </View>

          <View style={[tarjeta, { gap: 4, alignItems: 'center' }]}>
            <Text style={{ fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda }}>Balance (ganancia)</Text>
            <Text numberOfLines={1} adjustsFontSizeToFit
              style={{ fontFamily: fonts.titulo, fontSize: e(48), color: resumen.balance >= 0 ? colors.ingreso : colors.egreso }}>
              {resumen.balance < 0 ? '−' : ''}${formatearMonto(Math.abs(resumen.balance))}
            </Text>
          </View>

          {/* Meta de venta del mes (Historia 007) */}
          {empresa && <TarjetaMeta empresaId={empresa.id} ventasDelMes={ventasMes} />}

          {/* Ingresos vs gastos */}
          <View style={[tarjeta, { gap: 14 }]}>
            <Text style={titulo}>Ingresos vs gastos</Text>
            <GraficoBarras barras={resumen.barras} />
          </View>

          {/* Gastos por categoría */}
          <View style={[tarjeta, { gap: 14 }]}>
            <Text style={titulo}>Gastos por categoría</Text>
            {resumen.gastosPorCategoria.length === 0 ? (
              <Text style={{ fontFamily: fonts.normal, fontSize: e(16), color: colors.lavanda }}>Sin gastos en este período.</Text>
            ) : resumen.gastosPorCategoria.map((g) => (
              <View key={g.nombre} style={{ gap: 6 }}>
                <View style={{ flexDirection: 'row', justifyContent: 'space-between', gap: 8 }}>
                  <Text numberOfLines={1} style={{ flex: 1, fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda }}>{g.nombre}</Text>
                  <Text style={{ fontFamily: fonts.negrita, fontSize: e(16), color: colors.lavanda }}>
                    ${formatearMonto(g.total)} · {g.porcentaje}%
                  </Text>
                </View>
                <View style={{ height: 12, borderRadius: 6, backgroundColor: colors.lavandaSuave, overflow: 'hidden' }}>
                  <View style={{ width: `${g.porcentaje}%`, height: '100%', backgroundColor: g.color, borderRadius: 6 }} />
                </View>
              </View>
            ))}
          </View>
        </>
      )}
    </ScrollView>
  );
}
