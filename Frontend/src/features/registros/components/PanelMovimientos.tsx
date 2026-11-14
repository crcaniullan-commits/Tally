import { useFocusEffect } from 'expo-router';
import { ReactNode, useCallback, useMemo, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import { useAlerta } from '../../../shared/components/Confirmar';
import BarraTitulo from '../../../shared/components/BarraTitulo';
import Icono from '../../../shared/components/Icono';
import { useAviso } from '../../../shared/components/Toast';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { aISO } from '../../../shared/utils/fechas';
import { formatearMonto } from '../../../shared/utils/formato';
import { traducirError } from '../../auth/errores';
import { useSesion } from '../../auth/SesionContext';
import { Categoria, categoriasPredeterminadas, combinarCategorias, PREFIJO_LOCAL } from '../../categorias/categorias';
import { leerEstilos } from '../../categorias/estilosPropios';
import { listarCategorias } from '../../categorias/categorias.service';
import { inicioPeriodo, Periodo } from '../../dashboard/resumen';
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
  { id: 'egreso', nombre: 'Egresos' },
];
const periodos: { id: FiltroPeriodo; nombre: string }[] = [
  { id: 'semana', nombre: 'Esta semana' },
  { id: 'mes', nombre: 'Este mes' },
  { id: 'anio', nombre: 'Este año' },
  { id: 'todo', nombre: 'Todo' },
];

// Todos los movimientos con filtros, totales y exportación a Excel (Historias 005 y 012)
export default function PanelMovimientos({ selector, onCambio }: { selector: ReactNode; onCambio: () => void }) {
  const alerta = useAlerta();
  const { token, empresa, manejarError } = useSesion();
  const { t } = useEscala();
  const avisar = useAviso();
  const [detallada, setDetallada] = useState(false);

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
      const lista = cats.length ? combinarCategorias(cats, await leerEstilos(empresa.id)) : categoriasPredeterminadas;
      setCategorias(lista);
      setTodos(armarHistorial(ingresos, gastos, lista));
    } catch (error) {
      manejarError(error);
      alerta('No se pudieron cargar los movimientos', traducirError(error));
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

  async function exportar() {
    setExportando(true);
    try {
      const texto = periodos.find((p) => p.id === periodo)?.nombre ?? '';
      const compartido = await exportarExcel(filtrados, `Tally_movimientos_${aISO(new Date())}`, empresa?.nombre_empresa ?? 'Mi emprendimiento', texto);
      if (compartido) avisar('¡Excel descargado!', 'download');
    } catch (error) {
      alerta('No se pudo exportar', traducirError(error));
    } finally {
      setExportando(false);
    }
  }

  // ---- Estilos del mockup ----
  const caja = {
    backgroundColor: colors.superficie, padding: t(10), borderRadius: t(12), borderWidth: 1, borderColor: colors.lavandaOnda,
    boxShadow: '0px 1px 2px rgba(0, 0, 0, 0.05)',
  };
  const etiqueta = { fontFamily: fonts.extra, fontSize: t(11), color: colors.lavanda };
  const chip = (texto: string, activo: boolean, accion: () => void, clave?: string) => (
    <Pressable key={clave ?? texto} onPress={accion} style={{
      paddingHorizontal: t(10), paddingVertical: t(4), borderRadius: t(8), borderWidth: 1,
      backgroundColor: activo ? colors.lavanda : colors.fondoApp, borderColor: activo ? colors.lavanda : colors.lavandaOnda,
    }}>
      <Text style={{ fontFamily: activo ? fonts.negrita : fonts.medio, fontSize: t(11), color: activo ? colors.superficie : colors.texto }}>{texto}</Text>
    </Pressable>
  );

  return (
    <View style={{ flex: 1 }}>
      <ScrollView contentContainerStyle={{ paddingBottom: t(20) }} showsVerticalScrollIndicator={false}>
      <BarraTitulo
        titulo="Registros"
       
        derecha={
          <Pressable onPress={exportar} disabled={exportando} style={({ pressed }) => ({
            flexDirection: 'row', alignItems: 'center', gap: t(6), paddingHorizontal: t(12), paddingVertical: t(7), borderRadius: t(12),
            backgroundColor: colors.superficie, opacity: exportando ? 0.6 : pressed ? 0.85 : 1,
          })}>
            {exportando ? <ActivityIndicator size="small" color={colors.verde} /> : <Icono nombre="download" tamano={t(14)} color={colors.verde} />}
            <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.verde }}>Descargar Excel</Text>
          </Pressable>
        }
      />
      <View style={{ paddingHorizontal: t(16), paddingTop: t(12), gap: t(12) }}>
        {selector}

        {/* Filtros: Tipo → Categoría (solo si se eligió Ingresos o Egresos) → Período */}
        <View style={{ gap: t(8) }}>
          <View style={[caja, { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }]}>
            <Text style={etiqueta}>Tipo</Text>
            <View style={{ flexDirection: 'row', gap: t(4) }}>
              {tipos.map((x) => chip(x.nombre, tipo === x.id, () => { setTipo(x.id); setCategoriaId(null); }))}
            </View>
          </View>

          {tipo !== 'todos' && (
            <View style={[caja, { gap: t(6) }]}>
              <Text style={etiqueta}>Categoría de {tipo === 'ingreso' ? 'ingresos' : 'egresos'}</Text>
              <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={{ gap: t(6) }}>
                {chip('Todas', !categoriaId, () => setCategoriaId(null))}
                {categorias.filter((c) => !c.id.startsWith(PREFIJO_LOCAL) && (c.tipo === tipo || c.tipo === 'ambos')).map((c) =>
                  chip(c.nombre, categoriaId === c.id, () => setCategoriaId(c.id), c.clave))}
              </ScrollView>
            </View>
          )}

          <View style={[caja, { gap: t(6) }]}>
            <Text style={etiqueta}>Período</Text>
            <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={{ gap: t(6) }}>
              {periodos.map((x) => chip(x.nombre, periodo === x.id, () => setPeriodo(x.id)))}
            </ScrollView>
          </View>
        </View>

        <HistorialMovimientos
          titulo={`${filtrados.length} movimiento${filtrados.length === 1 ? '' : 's'}`}
          movimientos={filtrados}
          cargando={cargando}
          detallada={detallada}
          onEditar={setSeleccionado}
          derechaTitulo={
            <View style={{ flexDirection: 'row', backgroundColor: colors.lavandaSuave, padding: t(3), borderRadius: t(10) }}>
              {([[false, 'Abreviada', 'list'], [true, 'Detallada', 'layout-list']] as const).map(([valor, texto, icono]) => {
                const activo = detallada === valor;
                return (
                  <Pressable key={texto} onPress={() => setDetallada(valor)} style={{
                    flexDirection: 'row', alignItems: 'center', gap: t(4), paddingHorizontal: t(8), paddingVertical: t(4), borderRadius: t(7),
                    backgroundColor: activo ? colors.lavanda : 'transparent',
                  }}>
                    <Icono nombre={icono} tamano={t(12)} color={activo ? colors.superficie : colors.lavanda} />
                    <Text style={{ fontFamily: fonts.negrita, fontSize: t(10), color: activo ? colors.superficie : colors.lavanda }}>{texto}</Text>
                  </Pressable>
                );
              })}
            </View>
          }
        />
      </View>
      </ScrollView>

      <ModalMovimiento movimiento={seleccionado} categorias={categorias} onCerrar={() => setSeleccionado(null)} onCambio={() => { cargar(); onCambio(); }} />
    </View>
  );
}
