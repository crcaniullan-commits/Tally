import { Ionicons } from '@expo/vector-icons';
import { router, useFocusEffect } from 'expo-router';
import { useCallback, useState } from 'react';
import { Alert, KeyboardAvoidingView, Platform, Pressable, ScrollView, Text, TextInput, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { traducirError } from '../../features/auth/errores';
import { useSesion } from '../../features/auth/SesionContext';
import { Categoria, categoriasPredeterminadas, combinarCategorias, TipoMovimiento } from '../../features/categorias/categorias';
import {
  crearCategoria,
  eliminarCategoria,
  listarCategorias,
  obtenerIdReal,
  renombrarCategoria,
} from '../../features/categorias/categorias.service';
import ModalCategorias from '../../features/categorias/components/ModalCategorias';
import CarruselOpciones, { OpcionCarrusel } from '../../features/movimientos/components/CarruselOpciones';
import HistorialMovimientos, { Movimiento } from '../../features/movimientos/components/HistorialMovimientos';
import ModalMovimiento from '../../features/movimientos/components/ModalMovimiento';
import TecladoNumerico from '../../features/movimientos/components/TecladoNumerico';
import { armarHistorial } from '../../features/movimientos/historial';
import { ID_PENDIENTE, MetodoPago, metodoPendiente, metodosPago } from '../../features/movimientos/metodosPago';
import { crearGasto, crearIngreso, listarGastos, listarIngresos } from '../../features/movimientos/movimientos.service';
import { agregarPendiente, listarPendientes } from '../../features/pendientes/pendientes';
import Boton from '../../shared/components/Boton';
import CampoTexto from '../../shared/components/CampoTexto';
import SelectorFecha from '../../shared/components/SelectorFecha';
import { colors } from '../../shared/theme/colors';
import { fonts } from '../../shared/theme/fonts';
import { sombras } from '../../shared/theme/sombras';
import { useEscala } from '../../shared/utils/escala';
import { aISO, desdeISO, diasHasta, hoy, sumarDias } from '../../shared/utils/fechas';
import { formatearMonto } from '../../shared/utils/formato';

const MAX_DIGITOS = 10;
const ID_OTRA = 'otra'; // Tarjeta que abre el modal de categorías propias

export default function Inicio() {
  const { token, empresa, cerrarSesion, manejarError } = useSesion();
  const { e } = useEscala();
  const insets = useSafeAreaInsets();

  // Formulario de registro rápido (Historia 004)
  const [tipo, setTipo] = useState<TipoMovimiento>('ingreso');
  const [monto, setMonto] = useState(0);
  const [categoriaId, setCategoriaId] = useState<string | null>(null);
  const [metodo, setMetodo] = useState<string | null>(null);
  const [nota, setNota] = useState('');
  // Solo para pagos pendientes
  const [cliente, setCliente] = useState('');
  const [vence, setVence] = useState(sumarDias(hoy(), 7));

  const [categorias, setCategorias] = useState<Categoria[]>(categoriasPredeterminadas);
  const [movimientos, setMovimientos] = useState<Movimiento[]>([]);
  const [avisos, setAvisos] = useState(0); // Pagos pendientes que vencen hoy o ya vencieron
  const [modalCategorias, setModalCategorias] = useState(false);
  const [seleccionado, setSeleccionado] = useState<Movimiento | null>(null);
  const [cargandoDatos, setCargandoDatos] = useState(true);
  const [guardando, setGuardando] = useState(false);

  // Carga categorías, historial y pagos pendientes
  const cargarDatos = useCallback(async () => {
    if (!token || !empresa) return;
    try {
      const [cats, ingresos, gastos, pendientes] = await Promise.all([
        listarCategorias(empresa.id, token),
        listarIngresos(empresa.id, token),
        listarGastos(empresa.id, token),
        listarPendientes(empresa.id),
      ]);
      const preparadas = combinarCategorias(cats);
      setCategorias(preparadas);
      setMovimientos(armarHistorial(ingresos, gastos, preparadas));
      setAvisos(pendientes.filter((p) => diasHasta(desdeISO(p.vence)) <= 0).length);
    } catch (error) {
      manejarError(error);
      Alert.alert('No se pudieron cargar tus datos', traducirError(error));
    } finally {
      setCargandoDatos(false);
    }
  }, [token, empresa, manejarError]);

  // Se recarga al volver a esta pestaña (por si se editó algo en Historial o Pendientes)
  useFocusEffect(useCallback(() => { cargarDatos(); }, [cargarDatos]));

  const esIngreso = tipo === 'ingreso';
  const esPendiente = esIngreso && metodo === ID_PENDIENTE;
  const delTipo = categorias.filter((c) => c.tipo === tipo || c.tipo === 'ambos');
  const propias = categorias.filter((c) => c.tipo === 'ambos');
  const colorTipo = esIngreso ? colors.ingreso : colors.egreso;
  const formularioIncompleto =
    monto === 0 || !categoriaId || !empresa || (esIngreso && !metodo) || (esPendiente && !cliente.trim());

  const opcionesCategorias: OpcionCarrusel[] = [
    ...delTipo,
    { id: ID_OTRA, nombre: 'Otra', icono: 'plus', color: colors.lavanda, colorSuave: colors.lavandaSuave },
  ];
  const opcionesMetodos: OpcionCarrusel[] = [
    ...metodosPago.map((m) => ({ ...m, color: colors.lavanda, colorSuave: colors.lavandaSuave })),
    { ...metodoPendiente, color: colors.naranjo, colorSuave: colors.naranjoSuave },
  ];

  function cambiarTipo(nuevo: TipoMovimiento) {
    setTipo(nuevo);
    setCategoriaId(null); // Las categorías de ingreso y egreso son distintas
  }

  function seleccionarCategoria(id: string) {
    if (id === ID_OTRA) setModalCategorias(true);
    else setCategoriaId(id);
  }

  function agregarDigito(digito: string) {
    const nuevo = `${monto === 0 ? '' : monto}${digito}`;
    if (nuevo.length > MAX_DIGITOS) return;
    setMonto(Number(nuevo));
  }

  function limpiarFormulario() {
    setMonto(0);
    setCategoriaId(null);
    setMetodo(null);
    setNota('');
    setCliente('');
    setVence(sumarDias(hoy(), 7));
  }

  // ---- Categorías propias ----
  async function agregarCategoria(nombre: string) {
    if (!token || !empresa) return;
    try {
      const creada = await crearCategoria(nombre, empresa.id, token);
      const nueva: Categoria = {
        id: creada.id, nombre: creada.nombre, tipo: 'ambos',
        icono: 'tag-outline', color: colors.lavanda, colorSuave: colors.lavandaSuave,
      };
      setCategorias((actuales) => [...actuales, nueva]);
      setCategoriaId(nueva.id);
      setModalCategorias(false);
    } catch (error) {
      manejarError(error);
      Alert.alert('No se pudo crear la categoría', traducirError(error));
    }
  }

  async function cambiarNombreCategoria(c: Categoria, nombre: string) {
    if (!token || !empresa) return;
    try {
      await renombrarCategoria(c.id, nombre, empresa.id, token);
      setCategorias((actuales) => actuales.map((x) => (x.id === c.id ? { ...x, nombre } : x)));
    } catch (error) {
      manejarError(error);
      Alert.alert('No se pudo renombrar', traducirError(error));
    }
  }

  async function borrarCategoria(c: Categoria) {
    if (!token || !empresa) return;
    try {
      await eliminarCategoria(c.id, empresa.id, token);
      setCategorias((actuales) => actuales.filter((x) => x.id !== c.id));
      if (categoriaId === c.id) setCategoriaId(null);
    } catch (error) {
      manejarError(error);
      Alert.alert('No se pudo eliminar', traducirError(error));
    }
  }

  // ---- Registrar ----
  async function registrar() {
    const categoria = categorias.find((c) => c.id === categoriaId);
    if (!token || !empresa || !categoria || formularioIncompleto) return;
    setGuardando(true);
    try {
      if (esPendiente) {
        // Venta fiada: se guarda como pendiente y se avisa el día que vence
        await agregarPendiente(empresa.id, { cliente: cliente.trim(), monto, vence: aISO(vence), nota: nota.trim(), categoria: categoria.nombre });
        Alert.alert('Pago pendiente guardado', `Te avisaremos el día que vence. Puedes verlo en Registros → Pagos pendientes.`);
      } else {
        const idReal = await obtenerIdReal(categoria, empresa.id, token);
        if (esIngreso) await crearIngreso({ monto, metodo: metodo as MetodoPago, categoriaId: idReal, nota }, empresa.id, token);
        else await crearGasto({ monto, categoriaId: idReal, nota }, empresa.id, token);
      }
      limpiarFormulario();
      await cargarDatos();
    } catch (error) {
      manejarError(error);
      Alert.alert('No se pudo registrar', traducirError(error));
    } finally {
      setGuardando(false);
    }
  }

  const textoBoton = guardando ? 'Guardando...' : esPendiente ? 'Guardar pago pendiente' : esIngreso ? 'Registrar venta' : 'Registrar gasto';

  return (
    <KeyboardAvoidingView style={{ flex: 1, backgroundColor: colors.fondoApp }} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
      <ScrollView contentContainerStyle={{ paddingHorizontal: 20, paddingTop: insets.top + 16, paddingBottom: 32, gap: 22 }} keyboardShouldPersistTaps="handled">

        {/* Encabezado: saludo, emprendimiento, pagos pendientes y salir */}
        <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
          <View style={{ flex: 1 }}>
            <Text style={{ color: colors.lavanda, fontFamily: fonts.negrita, fontSize: e(30) }}>¡Hola!</Text>
            <Text numberOfLines={1} style={{ color: colors.lavanda, fontFamily: fonts.medio, fontSize: e(17), opacity: 0.8 }}>
              {empresa?.nombre_empresa ?? 'Cargando…'}
            </Text>
          </View>
          <View style={{ flexDirection: 'row', gap: 20 }}>
            <Pressable hitSlop={10} onPress={() => router.push({ pathname: '/registros', params: { vista: 'pendientes' } })}>
              <Ionicons name="notifications-outline" size={e(30)} color={colors.lavanda} />
              {avisos > 0 && (
                <View style={{ position: 'absolute', top: -4, right: -6, minWidth: 20, height: 20, borderRadius: 10, backgroundColor: colors.rosado, alignItems: 'center', justifyContent: 'center', paddingHorizontal: 4 }}>
                  <Text style={{ color: colors.superficie, fontFamily: fonts.negrita, fontSize: 12 }}>{avisos}</Text>
                </View>
              )}
            </Pressable>
            <Pressable hitSlop={10} onPress={cerrarSesion}>
              <Ionicons name="log-out-outline" size={e(30)} color={colors.lavanda} />
            </Pressable>
          </View>
        </View>

        {/* Paso 2 y 3: categoría y método de pago (los gastos no piden método) */}
        <CarruselOpciones titulo="Categorías" opciones={opcionesCategorias} seleccionadaId={categoriaId} onSeleccionar={seleccionarCategoria} />
        {esIngreso && (
          <CarruselOpciones titulo="Método de pago" opciones={opcionesMetodos} seleccionadaId={metodo} onSeleccionar={setMetodo} />
        )}

        {/* Paso 1: calculadora */}
        <View style={{ backgroundColor: colors.superficie, borderRadius: 24, padding: 16, gap: 14, ...sombras.flotante }}>
          <Text numberOfLines={1} adjustsFontSizeToFit style={{ fontFamily: fonts.titulo, fontSize: e(64), color: colors.lavanda, textAlign: 'center', letterSpacing: 1 }}>
            $ {formatearMonto(monto)}
          </Text>

          <View style={{ flexDirection: 'row', borderRadius: 26, overflow: 'hidden', backgroundColor: colors.lavandaSuave }}>
            {(['ingreso', 'egreso'] as const).map((opcion) => {
              const activo = tipo === opcion;
              const colorOpcion = opcion === 'ingreso' ? colors.ingreso : colors.egreso;
              return (
                <Pressable
                  key={opcion}
                  onPress={() => cambiarTipo(opcion)}
                  style={{ flex: 1, paddingVertical: e(14), alignItems: 'center', justifyContent: 'center', borderRadius: 26, backgroundColor: activo ? colorOpcion : 'transparent' }}
                >
                  <Text style={{ width: '100%', textAlign: 'center', fontFamily: fonts.negrita, fontSize: e(18), color: activo ? colors.superficie : colors.lavanda }}>
                    {opcion === 'ingreso' ? 'INGRESO (+)' : 'EGRESO (−)'}
                  </Text>
                </Pressable>
              );
            })}
          </View>

          <TecladoNumerico onDigito={agregarDigito} onBorrar={() => setMonto(Math.floor(monto / 10))} onLimpiar={() => setMonto(0)} />
        </View>

        {/* Datos extra de un pago pendiente */}
        {esPendiente && (
          <View style={{ backgroundColor: colors.naranjoSuave, borderRadius: 20, padding: 16, gap: 12 }}>
            <CampoTexto etiqueta="¿Quién te debe?" placeholder="Nombre del cliente" value={cliente} onChangeText={setCliente} maxLength={60} />
            <SelectorFecha etiqueta="¿Cuándo te paga?" fecha={vence} onCambiar={setVence} />
          </View>
        )}

        {/* Nota opcional + registrar */}
        <TextInput
          value={nota}
          onChangeText={setNota}
          maxLength={100}
          placeholder="Nota (opcional). Ej: Venta poleras diseño"
          placeholderTextColor={colors.lavanda}
          style={{ backgroundColor: colors.superficie, borderRadius: 16, padding: 16, fontFamily: fonts.normal, fontSize: 16, color: colors.lavanda, borderWidth: 1.5, borderColor: colors.lavandaSuave }}
        />

        <Boton titulo={textoBoton} color={esPendiente ? colors.naranjo : colorTipo} onPress={registrar} deshabilitado={formularioIncompleto || guardando} />

        <HistorialMovimientos
          movimientos={movimientos}
          cargando={cargandoDatos}
          limite={3}
          onVerTodos={() => router.push({ pathname: '/registros', params: { vista: 'movimientos' } })}
          onPresionar={setSeleccionado}
        />
      </ScrollView>

      <ModalCategorias
        visible={modalCategorias}
        categorias={propias}
        nombresExistentes={categorias.map((c) => c.nombre)}
        seleccionadaId={categoriaId}
        onSeleccionar={(id) => { setCategoriaId(id); setModalCategorias(false); }}
        onAgregar={agregarCategoria}
        onRenombrar={cambiarNombreCategoria}
        onEliminar={borrarCategoria}
        onCerrar={() => setModalCategorias(false)}
      />

      <ModalMovimiento movimiento={seleccionado} categorias={categorias} onCerrar={() => setSeleccionado(null)} onCambio={cargarDatos} />
    </KeyboardAvoidingView>
  );
}
