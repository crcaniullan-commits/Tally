import { Ionicons } from '@expo/vector-icons';
import { useFocusEffect } from 'expo-router';
import { useCallback, useMemo, useState } from 'react';
import { ActivityIndicator, Alert, Pressable, ScrollView, Text, View } from 'react-native';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { sombras } from '../../../shared/theme/sombras';
import { useEscala } from '../../../shared/utils/escala';
import { aISO } from '../../../shared/utils/fechas';
import { formatearMonto } from '../../../shared/utils/formato';
import { traducirError } from '../../auth/errores';
import { useSesion } from '../../auth/SesionContext';
import { Categoria, categoriasPredeterminadas, combinarCategorias, PREFIJO_LOCAL } from '../../categorias/categorias';
import { listarCategorias } from '../../categorias/categorias.service';
import { inicioPeriodo, Periodo } from '../../dashboard/resumen';
import Chip from '../../movimientos/components/Chip';
import HistorialMovimientos, { Movimiento } from '../../movimientos/components/HistorialMovimientos';
import ModalMovimiento from '../../movimientos/components/ModalMovimiento';
import { exportarExcel } from '../../movimientos/exportarExcel';
import { armarHistorial } from '../../movimientos/historial';
import { gastosDesde, ingresosDesde } from '../../movimientos/movimientos.service';

type FiltroTipo = 'todos' | 'ingreso' | 'egreso';
type FiltroPeriodo = Periodo | 'todo';

const tipos: { id: FiltroTipo; nombre: string }[] = [
  { id: 'todos', nombre: 'Todos' },
  { id: 'ingreso', nombre: 'Ingresos' },
  { id: 'egreso', nombre: 'Gastos' },
];
const periodos: { id: FiltroPeriodo; nombre: string }[] = [
  { id: 'semana', nombre: 'Esta semana' },
  { id: 'mes', nombre: 'Este mes' },
  { id: 'anio', nombre: 'Este año' },
  { id: 'todo', nombre: 'Todo' },
];

// Todos los movimientos con filtros, totales y exportación a Excel (Historias 005 y 012)
export default function PanelMovimientos() {
  const { token, empresa, manejarError } = useSesion();
  const { e } = useEscala();

  const [tipo, setTipo] = useState<FiltroTipo>('todos');
  const [periodo, setPeriodo] = useState<FiltroPeriodo>('mes');
  const [categoriaId, setCategoriaId] = useState<string | null>(null);
  const [todos, setTodos] = useState<Movimiento[]>([]);
  const [categorias, setCategorias] = useState<Categoria[]>(categoriasPredeterminadas);
  const [seleccionado, setSeleccionado] = useState<Movimiento | null>(null);
  const [cargando, setCargando] = useState(true);
  const [exportando, setExportando] = useState(false);

  const cargar = useCallback(async () => {
    if (!token || !empresa) return;
    setCargando(true);
    try {
      const desde = periodo === 'todo' ? '2000-01-01' : aISO(inicioPeriodo(periodo));
      const [ingresos, gastos, cats] = await Promise.all([
        ingresosDesde(desde, empresa.id, token),
        gastosDesde(desde, empresa.id, token),
        listarCategorias(empresa.id, token).catch(() => []),
      ]);
      const lista = cats.length ? combinarCategorias(cats) : categoriasPredeterminadas;
      setCategorias(lista);
      setTodos(armarHistorial(ingresos, gastos, lista));
    } catch (error) {
      manejarError(error);
      Alert.alert('No se pudieron cargar los movimientos', traducirError(error));
    } finally {
      setCargando(false);
    }
  }, [token, empresa, periodo, manejarError]);

  useFocusEffect(useCallback(() => { cargar(); }, [cargar]));

  // El período se aplica al pedir los datos; tipo y categoría se filtran aquí
  const filtrados = useMemo(
    () => todos.filter((m) => (tipo === 'todos' || m.tipo === tipo) && (!categoriaId || m.categoriaId === categoriaId)),
    [todos, tipo, categoriaId],
  );
  const totalIngresos = filtrados.filter((m) => m.tipo === 'ingreso').reduce((s, m) => s + m.monto, 0);
  const totalGastos = filtrados.filter((m) => m.tipo === 'egreso').reduce((s, m) => s + m.monto, 0);

  async function exportar() {
    if (filtrados.length === 0) return Alert.alert('Sin movimientos', 'No hay movimientos con estos filtros.');
    setExportando(true);
    try {
      await exportarExcel(filtrados, `Tally_movimientos_${aISO(new Date())}`);
    } catch (error) {
      Alert.alert('No se pudo exportar', traducirError(error));
    } finally {
      setExportando(false);
    }
  }

  const etiqueta = { fontFamily: fonts.medio, fontSize: e(15), color: colors.lavanda };

  return (
    <View style={{ flex: 1 }}>
      <ScrollView contentContainerStyle={{ paddingHorizontal: 20, paddingTop: 4, paddingBottom: 32, gap: 18 }}>
        {/* Filtros */}
        <View style={{ gap: 8 }}>
          <Text style={etiqueta}>Tipo</Text>
          <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }}>
            {tipos.map((t) => <Chip key={t.id} texto={t.nombre} activo={tipo === t.id} onPress={() => setTipo(t.id)} />)}
          </View>
          <Text style={etiqueta}>Período</Text>
          <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }}>
            {periodos.map((p) => <Chip key={p.id} texto={p.nombre} activo={periodo === p.id} onPress={() => setPeriodo(p.id)} />)}
          </View>
          <Text style={etiqueta}>Categoría</Text>
          <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={{ gap: 8 }}>
            <Chip texto="Todas" activo={!categoriaId} onPress={() => setCategoriaId(null)} />
            {categorias.filter((c) => !c.id.startsWith(PREFIJO_LOCAL)).map((c) => (
              <Chip key={c.id} texto={c.nombre} activo={categoriaId === c.id} onPress={() => setCategoriaId(c.id)} />
            ))}
          </ScrollView>
        </View>

        {/* Totales de lo filtrado */}
        <View style={{ flexDirection: 'row', gap: 12 }}>
          {[
            { t: 'Ingresos', v: totalIngresos, c: colors.ingreso },
            { t: 'Gastos', v: totalGastos, c: colors.egreso },
            { t: 'Balance', v: totalIngresos - totalGastos, c: totalIngresos - totalGastos >= 0 ? colors.ingreso : colors.egreso },
          ].map((x) => (
            <View key={x.t} style={{ flex: 1, backgroundColor: colors.superficie, borderRadius: 18, padding: 12, gap: 2, ...sombras.suave }}>
              <Text style={{ fontFamily: fonts.medio, fontSize: e(14), color: colors.lavanda }}>{x.t}</Text>
              <Text numberOfLines={1} adjustsFontSizeToFit style={{ fontFamily: fonts.negrita, fontSize: e(19), color: x.c }}>
                {x.v < 0 ? '−' : ''}${formatearMonto(Math.abs(x.v))}
              </Text>
            </View>
          ))}
        </View>

        {/* Descargar Excel con lo filtrado */}
        <Pressable
          onPress={exportar}
          disabled={exportando}
          style={({ pressed }) => ({
            flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 8, paddingVertical: e(13),
            borderRadius: 999, borderWidth: 2, borderColor: colors.verde, backgroundColor: colors.verdeSuave,
            opacity: exportando ? 0.5 : pressed ? 0.8 : 1,
          })}
        >
          {exportando ? <ActivityIndicator color={colors.verde} /> : <Ionicons name="download-outline" size={e(22)} color={colors.verde} />}
          <Text style={{ fontFamily: fonts.negrita, fontSize: e(16), color: colors.verde }}>Descargar Excel</Text>
        </Pressable>

        <HistorialMovimientos
          titulo={`${filtrados.length} movimiento${filtrados.length === 1 ? '' : 's'}`}
          movimientos={filtrados}
          cargando={cargando}
          onPresionar={setSeleccionado}
        />
      </ScrollView>

      <ModalMovimiento movimiento={seleccionado} categorias={categorias} onCerrar={() => setSeleccionado(null)} onCambio={cargar} />
    </View>
  );
}
