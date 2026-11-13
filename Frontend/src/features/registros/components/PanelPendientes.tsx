import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useFocusEffect } from 'expo-router';
import { useCallback, useState } from 'react';
import { ActivityIndicator, Alert, ScrollView, Text, View } from 'react-native';
import { traducirError } from '../../auth/errores';
import { useSesion } from '../../auth/SesionContext';
import { categoriasPredeterminadas, combinarCategorias, normalizar } from '../../categorias/categorias';
import { listarCategorias, obtenerIdReal } from '../../categorias/categorias.service';
import Chip from '../../movimientos/components/Chip';
import { MetodoPago, metodosPago } from '../../movimientos/metodosPago';
import { crearIngreso } from '../../movimientos/movimientos.service';
import { listarPendientes, Pendiente, quitarPendiente } from '../../pendientes/pendientes';
import Boton from '../../../shared/components/Boton';
import ModalFlotante from '../../../shared/components/ModalFlotante';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { sombras } from '../../../shared/theme/sombras';
import { useEscala } from '../../../shared/utils/escala';
import { aDMA, desdeISO, diasHasta } from '../../../shared/utils/fechas';
import { formatearMonto } from '../../../shared/utils/formato';

// Estado de un pago según su fecha de vencimiento
function estado(p: Pendiente): { texto: string; color: string; fondo: string } {
  const dias = diasHasta(desdeISO(p.vence));
  if (dias < 0) return { texto: `Venció hace ${-dias} día${dias === -1 ? '' : 's'}`, color: colors.rosado, fondo: colors.rosadoSuave };
  if (dias === 0) return { texto: 'Vence hoy', color: colors.naranjo, fondo: colors.naranjoSuave };
  return { texto: `Vence en ${dias} día${dias === 1 ? '' : 's'}`, color: colors.azul, fondo: colors.azulSuave };
}

// Pagos pendientes / fiado (Historia 008)
export default function PanelPendientes() {
  const { token, empresa, manejarError } = useSesion();
  const { e } = useEscala();
  const [lista, setLista] = useState<Pendiente[]>([]);
  const [cargando, setCargando] = useState(true);
  const [cobrando, setCobrando] = useState<Pendiente | null>(null); // Pago que se está marcando como pagado
  const [metodo, setMetodo] = useState<MetodoPago>('efectivo');
  const [guardando, setGuardando] = useState(false);

  const cargar = useCallback(async () => {
    if (!empresa) return;
    setCargando(true);
    setLista(await listarPendientes(empresa.id));
    setCargando(false);
  }, [empresa]);

  useFocusEffect(useCallback(() => { cargar(); }, [cargar]));

  const total = lista.reduce((s, p) => s + p.monto, 0);

  // Marcar como pagado: se registra como ingreso real con la fecha de hoy y sale de la lista
  async function confirmarPago() {
    if (!token || !empresa || !cobrando) return;
    setGuardando(true);
    try {
      const cats = combinarCategorias(await listarCategorias(empresa.id, token).catch(() => []));
      const categoria =
        cats.find((c) => normalizar(c.nombre) === normalizar(cobrando.categoria)) ??
        cats.find((c) => c.nombre === 'Ventas') ?? categoriasPredeterminadas[0];
      const idReal = await obtenerIdReal(categoria, empresa.id, token);
      const nota = [`Pago de ${cobrando.cliente}`, cobrando.nota].filter(Boolean).join(' · ').slice(0, 100);
      await crearIngreso({ monto: cobrando.monto, metodo, categoriaId: idReal, nota }, empresa.id, token);
      await quitarPendiente(empresa.id, cobrando);
      setCobrando(null);
      await cargar();
      Alert.alert('¡Pago registrado!', 'Se agregó a tus ingresos de hoy.');
    } catch (error) {
      manejarError(error);
      Alert.alert('No se pudo registrar el pago', traducirError(error));
    } finally {
      setGuardando(false);
    }
  }

  function eliminar(p: Pendiente) {
    Alert.alert('Eliminar pago pendiente', `¿Eliminar el pago de ${p.cliente}? No se registrará como ingreso.`, [
      { text: 'Cancelar', style: 'cancel' },
      { text: 'Eliminar', style: 'destructive', onPress: async () => { if (empresa) { await quitarPendiente(empresa.id, p); cargar(); } } },
    ]);
  }

  return (
    <View style={{ flex: 1, backgroundColor: colors.fondoApp }}>
      <ScrollView contentContainerStyle={{ paddingHorizontal: 20, paddingTop: 4, paddingBottom: 32, gap: 18 }}>
        <View style={{ backgroundColor: colors.superficie, borderRadius: 20, padding: 16, alignItems: 'center', gap: 2, ...sombras.suave }}>
          <Text style={{ fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda }}>Total por cobrar</Text>
          <Text numberOfLines={1} adjustsFontSizeToFit style={{ fontFamily: fonts.titulo, fontSize: e(44), color: colors.naranjo }}>${formatearMonto(total)}</Text>
        </View>

        {cargando ? (
          <ActivityIndicator color={colors.lavanda} />
        ) : lista.length === 0 ? (
          <View style={{ alignItems: 'center', gap: 8, paddingVertical: 24 }}>
            <MaterialCommunityIcons name="check-circle-outline" size={e(48)} color={colors.verde} />
            <Text style={{ fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda, textAlign: 'center' }}>
              No tienes pagos pendientes.{'\n'}Para agregar uno, en Inicio elige el método de pago "Pendiente".
            </Text>
          </View>
        ) : (
          lista.map((p) => {
            const est = estado(p);
            return (
              <View key={p.id} style={{ backgroundColor: colors.superficie, borderRadius: 20, padding: 16, gap: 10, ...sombras.suave }}>
                <View style={{ flexDirection: 'row', alignItems: 'center', gap: 12 }}>
                  <View style={{ width: e(48), height: e(48), borderRadius: 14, backgroundColor: est.fondo, alignItems: 'center', justifyContent: 'center' }}>
                    <MaterialCommunityIcons name="clock-outline" size={e(26)} color={est.color} />
                  </View>
                  <View style={{ flex: 1 }}>
                    <Text numberOfLines={1} style={{ fontFamily: fonts.negrita, fontSize: e(17), color: colors.lavanda }}>{p.cliente}</Text>
                    <Text style={{ fontFamily: fonts.medio, fontSize: e(14), color: est.color }}>{est.texto} · {aDMA(desdeISO(p.vence))}</Text>
                    {p.nota ? <Text numberOfLines={1} style={{ fontFamily: fonts.normal, fontSize: e(13), color: colors.lavanda, opacity: 0.8 }}>{p.nota}</Text> : null}
                  </View>
                  <Text style={{ fontFamily: fonts.negrita, fontSize: e(18), color: colors.naranjo }}>${formatearMonto(p.monto)}</Text>
                </View>
                <View style={{ flexDirection: 'row', gap: 10 }}>
                  <View style={{ flex: 1 }}>
                    <Boton titulo="Marcar pagado" color={colors.ingreso} onPress={() => { setMetodo('efectivo'); setCobrando(p); }} />
                  </View>
                  <View style={{ flex: 1 }}>
                    <Boton titulo="Eliminar" variante="secundario" color={colors.error} onPress={() => eliminar(p)} />
                  </View>
                </View>
              </View>
            );
          })
        )}
      </ScrollView>

      <ModalFlotante visible={!!cobrando} titulo="¿Cómo te pagó?" onCerrar={() => setCobrando(null)}>
        <View style={{ gap: 16 }}>
          <Text style={{ fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda }}>
            {cobrando?.cliente} · ${formatearMonto(cobrando?.monto ?? 0)}
          </Text>
          <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }}>
            {metodosPago.map((m) => <Chip key={m.id} texto={m.nombre} activo={metodo === m.id} onPress={() => setMetodo(m.id)} />)}
          </View>
          <Boton titulo={guardando ? 'Guardando...' : 'Registrar pago'} color={colors.ingreso} onPress={confirmarPago} deshabilitado={guardando} />
        </View>
      </ModalFlotante>
    </View>
  );
}
