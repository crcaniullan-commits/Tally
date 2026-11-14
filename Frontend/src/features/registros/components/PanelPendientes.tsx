import { ReactNode } from 'react';
import { useFocusEffect } from 'expo-router';
import { useCallback, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { traducirError } from '../../auth/errores';
import { useSesion } from '../../auth/SesionContext';
import { categoriasPredeterminadas, combinarCategorias, normalizar } from '../../categorias/categorias';
import { listarCategorias, obtenerIdReal } from '../../categorias/categorias.service';
import Chip from '../../movimientos/components/Chip';
import { MetodoPago, metodosPago } from '../../movimientos/metodosPago';
import { crearIngreso } from '../../movimientos/movimientos.service';
import { listarPendientes, marcarPagado, Pendiente } from '../../pendientes/pendientes';
import Boton from '../../../shared/components/Boton';
import BarraTitulo from '../../../shared/components/BarraTitulo';
import Icono from '../../../shared/components/Icono';
import { useAviso } from '../../../shared/components/Toast';
import { useConfirmar, useAlerta } from '../../../shared/components/Confirmar';
import ModalPendiente from './ModalPendiente';
import ModalFlotante from '../../../shared/components/ModalFlotante';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
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
export default function PanelPendientes({ selector, onCambio }: { selector: ReactNode; onCambio: () => void }) {
  const alerta = useAlerta();
  const { token, empresa, manejarError } = useSesion();
  const { e, t } = useEscala();
  const avisar = useAviso();
  const confirmar = useConfirmar();
  const [editando, setEditando] = useState<Pendiente | null>(null);
  const [lista, setLista] = useState<Pendiente[]>([]);
  const [cargando, setCargando] = useState(true);
  const [cobrando, setCobrando] = useState<Pendiente | null>(null); // Pago que se está marcando como pagado
  const [metodo, setMetodo] = useState<MetodoPago>('efectivo');
  const [guardando, setGuardando] = useState(false);

  const cargar = useCallback(async () => {
    if (!empresa || !token) return;
    setCargando(true);
    try {
      setLista(await listarPendientes(empresa.id, token));
    } catch (error) {
      manejarError(error);
      alerta('No se pudieron cargar los pagos pendientes', traducirError(error));
    } finally {
      setCargando(false);
    }
  }, [empresa, token, manejarError]);

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
      await marcarPagado(empresa.id, token, cobrando);
      setCobrando(null);
      await cargar();
      avisar('¡Pago registrado como ingreso!');
      onCambio();
    } catch (error) {
      manejarError(error);
      alerta('No se pudo registrar el pago', traducirError(error));
    } finally {
      setGuardando(false);
    }
  }

  async function eliminar(p: Pendiente) {
    const ok = await confirmar({
      titulo: 'Quitar pago pendiente',
      mensaje: `¿Quitar el pago de ${p.cliente} de la lista? No se registrará como ingreso.`,
      textoConfirmar: 'Quitar',
    });
    if (!ok || !empresa || !token) return;
    try {
      await marcarPagado(empresa.id, token, p);
      avisar('Pago pendiente quitado', 'trash');
      cargar();
    } catch (error) {
      manejarError(error);
      alerta('No se pudo quitar', traducirError(error));
    }
  }

  const tarjeta = { backgroundColor: colors.superficie, borderRadius: t(12), borderWidth: 1, borderColor: colors.lavandaOnda, boxShadow: '0px 1px 2px rgba(0, 0, 0, 0.05)' };
  const botonChico = (texto: string, color: string, fondo: string, accion: () => void) => (
    <Pressable onPress={accion} style={({ pressed }) => ({ flex: 1, alignItems: 'center', paddingVertical: t(6), borderRadius: t(8), backgroundColor: fondo, opacity: pressed ? 0.8 : 1 })}>
      <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color }}>{texto}</Text>
    </Pressable>
  );

  return (
    <View style={{ flex: 1, backgroundColor: colors.fondoApp }}>
      <ScrollView contentContainerStyle={{ paddingBottom: t(20) }} showsVerticalScrollIndicator={false}>
      <BarraTitulo titulo="Registros" />
      <View style={{ paddingHorizontal: t(16), paddingTop: t(12), gap: t(12) }}>
        {selector}

        <View style={[tarjeta, { padding: t(12), flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }]}>
          <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: colors.textoSecundario }}>Total por cobrar</Text>
          <Text numberOfLines={1} style={{ fontFamily: fonts.extra, fontSize: t(18), color: colors.naranjo }}>${formatearMonto(total)}</Text>
        </View>

        {cargando ? (
          <ActivityIndicator color={colors.lavanda} />
        ) : lista.length === 0 ? (
          <View style={{ alignItems: 'center', gap: t(8), paddingVertical: t(24) }}>
            <Icono nombre="check-circle" tamano={t(36)} color={colors.verde} />
            <Text style={{ fontFamily: fonts.medio, fontSize: t(12), color: colors.textoSecundario, textAlign: 'center' }}>
              No tienes pagos pendientes.{'\n'}Para agregar uno, en Inicio toca "Pago pendiente".
            </Text>
          </View>
        ) : (
          <>
            <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.lavanda }}>
              {lista.length} pago{lista.length === 1 ? '' : 's'} pendiente{lista.length === 1 ? '' : 's'}
            </Text>
            {lista.map((p) => {
              const est = estado(p);
              return (
                <View key={p.id} style={[tarjeta, { padding: t(12), gap: t(10) }]}>
                  <View style={{ flexDirection: 'row', alignItems: 'center', gap: t(10) }}>
                    <View style={{ width: t(32), height: t(32), borderRadius: t(8), backgroundColor: est.fondo, alignItems: 'center', justifyContent: 'center' }}>
                      <Icono nombre="clock" tamano={t(16)} color={est.color} />
                    </View>
                    <View style={{ flex: 1 }}>
                      <Text numberOfLines={1} style={{ fontFamily: fonts.negrita, fontSize: t(12), color: colors.texto }}>{p.cliente}</Text>
                      <Text numberOfLines={1} style={{ fontFamily: fonts.medio, fontSize: t(10), color: est.color }}>
                        {est.texto} • {aDMA(desdeISO(p.vence))}{p.nota ? ` • ${p.nota}` : ''}
                      </Text>
                    </View>
                    <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.naranjo }}>${formatearMonto(p.monto)}</Text>
                    <Pressable hitSlop={10} onPress={() => setEditando(p)} style={{ padding: t(4), borderRadius: t(6), backgroundColor: colors.lavandaSuave }}>
                      <Icono nombre="pencil" tamano={t(11)} color={colors.lavanda} />
                    </Pressable>
                  </View>
                  <View style={{ flexDirection: 'row', gap: t(8) }}>
                    {botonChico('Marcar pagado', colors.verde, colors.verdeSuave, () => { setMetodo('efectivo'); setCobrando(p); })}
                    {botonChico('Quitar', colors.rosado, colors.rosadoSuave, () => eliminar(p))}
                  </View>
                </View>
              );
            })}
          </>
        )}
      </View>
      </ScrollView>

      <ModalPendiente pendiente={editando} onCerrar={() => setEditando(null)} onCambio={cargar} />

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
