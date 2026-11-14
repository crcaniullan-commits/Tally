import { useState } from 'react';
import { ActivityIndicator, Pressable, RefreshControl, ScrollView, Text, View } from 'react-native';
import { useSesion } from '../../features/auth/SesionContext';
import { Agrupacion, calcularDashboard, PERIODOS, TipoGrafico } from '../../features/dashboard/calcularDashboard';
import { BarrasHorizontales, BarrasVerticales, ComparadoHorizontal, ComparadoVertical, Torta } from '../../features/dashboard/components/Graficos';
import ListaDesglose from '../../features/dashboard/components/ListaDesglose';
import { FiltroDashboard, Periodo } from '../../features/dashboard/resumen';
import { useDatosDashboard } from '../../features/dashboard/useDatosDashboard';
import TarjetaMeta from '../../features/metas/components/TarjetaMeta';
import BarraTitulo from '../../shared/components/BarraTitulo';
import EstadoEmpresa from '../../shared/components/EstadoEmpresa';
import Icono from '../../shared/components/Icono';
import Segmentado from '../../shared/components/Segmentado';
import { colors } from '../../shared/theme/colors';
import { fonts } from '../../shared/theme/fonts';
import { useEscala } from '../../shared/utils/escala';
import { pesos } from '../../shared/utils/formato';

const BOTONES_GRAFICO = [['torta', 'pie-chart'], ['barras', 'bar-chart-3'], ['horizontal', 'bar-chart-h']] as const;

// Dashboard (Historia 006): meta, filtros, gráfico y desglose.
// Los datos y cálculos viven en features/dashboard (useDatosDashboard y calcularDashboard).
export default function Dashboard() {
  const { token, empresa } = useSesion();
  const { t } = useEscala();
  const [filtro, setFiltro] = useState<FiltroDashboard>('ingresos');
  const [grafico, setGrafico] = useState<TipoGrafico>('torta');
  const [agrupacion, setAgrupacion] = useState<Agrupacion>('categoria');
  const [periodo, setPeriodo] = useState<Periodo>('mes');
  const { datos, cargando, error, cargar, recargas } = useDatosDashboard(periodo);

  const r = datos ? calcularDashboard(datos, filtro, periodo, agrupacion) : null;
  // Comparando ingresos y egresos por fecha no hay torta (no sirve para comparar)
  const graficoVisible: TipoGrafico = r?.comparando && grafico === 'torta' ? 'barras' : grafico;

  return (
    <ScrollView
      style={{ flex: 1, backgroundColor: colors.fondoApp }}
      contentContainerStyle={{ paddingBottom: t(20) }}
      refreshControl={<RefreshControl refreshing={false} onRefresh={cargar} />}
      showsVerticalScrollIndicator={false}
    >
      <BarraTitulo titulo="Dashboard" />
      <View style={{ paddingHorizontal: t(16), paddingTop: t(12), gap: t(12) }}>
        <EstadoEmpresa />

        {/* Meta de venta del mes (Historia 007) */}
        {empresa && token && <TarjetaMeta empresaId={empresa.id} token={token} recargar={recargas} />}

        {/* Qué se quiere ver y en qué período */}
        <Segmentado
          opciones={[
            { id: 'ingresos', nombre: 'Ingresos', color: colors.verde },
            { id: 'egresos', nombre: 'Egresos', color: colors.rosado },
            { id: 'ambos', nombre: 'Ing. / Egr.', color: colors.lavanda },
          ]}
          valor={filtro}
          onCambiar={setFiltro}
          colorInactivo={colors.texto}
          borde
        />
        <Segmentado opciones={PERIODOS.map((p) => ({ id: p.id, nombre: p.nombre }))} valor={periodo} onCambiar={setPeriodo} borde alto={4} />

        {/* Tarjeta principal: total, indicador, gráfico y desglose */}
        <View style={{ backgroundColor: colors.superficie, borderRadius: t(16), padding: t(16), gap: t(16), borderWidth: 1, borderColor: colors.lavandaOnda, boxShadow: '0px 1px 3px rgba(0, 0, 0, 0.06)' }}>
          {cargando && !r ? (
            <ActivityIndicator color={colors.lavanda} style={{ paddingVertical: t(60) }} />
          ) : error || !r ? (
            <Text style={{ fontFamily: fonts.medio, fontSize: t(12), color: colors.error, textAlign: 'center', paddingVertical: t(40) }}>{error}</Text>
          ) : (
            <>
              {/* Total a la izquierda; a la derecha, el tipo de gráfico sobre el indicador */}
              <View style={{ flexDirection: 'row', justifyContent: 'space-between', gap: t(8) }}>
                <View style={{ flex: 1, justifyContent: 'space-between' }}>
                  <Text numberOfLines={1} adjustsFontSizeToFit style={{ fontFamily: fonts.extra, fontSize: t(24), color: colors.texto, letterSpacing: -0.5 }}>{pesos(r.total)}</Text>
                  <Text style={{ fontFamily: fonts.medio, fontSize: t(12), color: colors.textoSecundario }}>{r.subtitulo}</Text>
                </View>
                <View style={{ alignItems: 'flex-end', gap: t(6) }}>
                  <View style={{ flexDirection: 'row', backgroundColor: colors.lavandaSuave, padding: t(3), borderRadius: t(10), borderWidth: 1, borderColor: colors.lavandaOnda }}>
                    {BOTONES_GRAFICO.filter(([id]) => !(r.comparando && id === 'torta')).map(([id, icono]) => {
                      const activo = graficoVisible === id;
                      return (
                        <Pressable key={id} onPress={() => setGrafico(id)} style={{ padding: t(5), borderRadius: t(7), backgroundColor: activo ? colors.lavanda : 'transparent' }}>
                          <Icono nombre={icono} tamano={t(14)} color={activo ? colors.superficie : colors.lavanda} />
                        </Pressable>
                      );
                    })}
                  </View>
                  <Text style={{ fontFamily: fonts.extra, fontSize: t(11), color: r.insignia.color, backgroundColor: r.insignia.fondo, paddingHorizontal: t(10), paddingVertical: t(4), borderRadius: t(8), overflow: 'hidden' }}>
                    {r.insignia.texto}
                  </Text>
                </View>
              </View>

              {/* Ver el gráfico por categoría o comparando fechas */}
              <Segmentado
                opciones={[{ id: 'categoria', nombre: 'Por categoría' }, { id: 'fecha', nombre: 'Por fecha' }]}
                valor={agrupacion}
                onCambiar={setAgrupacion}
                alto={4}
              />

              <View style={{ paddingVertical: t(4), alignItems: 'center', justifyContent: 'center', minHeight: t(140) }}>
                {r.items.length === 0 ? (
                  <Text style={{ fontFamily: fonts.medio, fontSize: t(12), color: colors.textoSecundario, textAlign: 'center' }}>
                    Sin movimientos en este período.{'\n'}Registra uno en Inicio para ver tu gráfico.
                  </Text>
                ) : (
                  <View style={{ width: '100%', gap: t(10) }}>
                    <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.lavanda }}>{r.tituloGrafico}</Text>
                    {r.comparando
                      ? (graficoVisible === 'barras' ? <ComparadoVertical comparado={r.comparado} /> : <ComparadoHorizontal comparado={r.comparado} />)
                      : graficoVisible === 'torta' ? <Torta items={r.agrupados} total={r.total} etiqueta={r.etiquetaCentro} />
                      : graficoVisible === 'barras' ? <BarrasVerticales items={r.agrupados} />
                      : <BarrasHorizontales items={r.items} />}
                  </View>
                )}
              </View>

              {r.items.length > 0 && <ListaDesglose items={r.items} comparando={r.comparando} />}
            </>
          )}
        </View>
      </View>
    </ScrollView>
  );
}
