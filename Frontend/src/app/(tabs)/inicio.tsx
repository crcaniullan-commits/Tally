import { useState } from 'react';
import { KeyboardAvoidingView, Platform, Pressable, ScrollView, Text, TextInput, View } from 'react-native';
import { traducirError } from '../../features/auth/errores';
import { useSesion } from '../../features/auth/SesionContext';
import { categoriasPredeterminadas, TipoMovimiento } from '../../features/categorias/categorias';
import SelectorCategorias from '../../features/categorias/components/SelectorCategorias';
import { useCategorias } from '../../features/categorias/useCategorias';
import Calculadora from '../../features/movimientos/components/Calculadora';
import SelectorMetodo from '../../features/movimientos/components/SelectorMetodo';
import { ID_PENDIENTE } from '../../features/movimientos/metodosPago';
import { registrarMovimiento, seCumplioLaMeta } from '../../features/movimientos/registrarMovimiento';
import DatosPagoPendiente from '../../features/pendientes/components/DatosPagoPendiente';
import BarraTitulo from '../../shared/components/BarraTitulo';
import { useAlerta } from '../../shared/components/Confirmar';
import EstadoEmpresa from '../../shared/components/EstadoEmpresa';
import Icono from '../../shared/components/Icono';
import { useAviso } from '../../shared/components/Toast';
import { colors } from '../../shared/theme/colors';
import { fonts } from '../../shared/theme/fonts';
import { useEscala } from '../../shared/utils/escala';
import { aISO, hoy, sumarDias } from '../../shared/utils/fechas';
import { formatearMonto } from '../../shared/utils/formato';

const MESES_CORTOS = ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun', 'Jul', 'Ago', 'Sep', 'Oct', 'Nov', 'Dic'];

// Inicio: registro rápido de ingresos, egresos y pagos pendientes (Historias 004 y 008).
// La lógica vive en features/: categorías (useCategorias), registro (registrarMovimiento) y componentes.
export default function Inicio() {
  const { token, empresa, manejarError } = useSesion();
  const { t } = useEscala();
  const alerta = useAlerta();
  const avisar = useAviso();
  const cats = useCategorias();

  // Formulario
  const [tipo, setTipo] = useState<TipoMovimiento>('ingreso');
  const [monto, setMonto] = useState(0);
  const [categoriaId, setCategoriaId] = useState<string | null>(categoriasPredeterminadas[0].id);
  const [metodo, setMetodo] = useState<string>('efectivo');
  const [nota, setNota] = useState('');
  const [cliente, setCliente] = useState('');
  const [vence, setVence] = useState(sumarDias(hoy(), 7));
  const [guardando, setGuardando] = useState(false);
  const [arrastrando, setArrastrando] = useState(false); // Mientras se arrastra un tag, la pantalla no hace scroll

  const esIngreso = tipo === 'ingreso';
  const esPendiente = esIngreso && metodo === ID_PENDIENTE;
  const categoria = cats.categorias.find((c) => c.id === categoriaId);
  // Botón deshabilitado si falta un dato obligatorio (Definición de Terminado, Historia 004)
  const incompleto = monto === 0 || !categoria || (esPendiente && !cliente.trim());
  const fechaHoy = new Date();

  function cambiarTipo(nuevo: TipoMovimiento) {
    setTipo(nuevo);
    // Las categorías de ingreso y egreso son distintas: se elige la primera del nuevo tipo
    setCategoriaId(cats.delTipo(nuevo)[0]?.id ?? null);
    if (nuevo === 'egreso' && metodo === ID_PENDIENTE) setMetodo('efectivo');
  }

  function limpiarFormulario() {
    setMonto(0);
    setNota('');
    setCliente('');
    setVence(sumarDias(hoy(), 7));
    if (metodo === ID_PENDIENTE) setMetodo('efectivo');
  }

  async function registrar() {
    if (incompleto || !token || !empresa || !categoria) return;
    setGuardando(true);
    try {
      const r = await registrarMovimiento(
        { tipo, monto, categoria, metodo, nota, pendiente: esPendiente ? { cliente, vence: aISO(vence) } : null },
        empresa.id, token,
      );
      avisar(r.mensaje, r.icono);
      if (r.categoriaCreada) await cats.cargar();
      if (esIngreso && !esPendiente && await seCumplioLaMeta(monto, empresa.id, token)) avisar('¡Cumpliste tu meta de venta! 🎉', 'trophy');
      limpiarFormulario();
    } catch (error) {
      manejarError(error);
      alerta('No se pudo registrar', traducirError(error));
    } finally {
      setGuardando(false);
    }
  }

  // ---- Estilos del mockup ----
  const tarjeta = {
    backgroundColor: colors.superficie, borderRadius: t(16), padding: t(12), gap: t(10),
    borderWidth: 1, borderColor: colors.lavandaOnda, boxShadow: '0px 1px 2px rgba(0, 0, 0, 0.05)',
  };
  const textoBoton = guardando ? 'Guardando...' : esPendiente ? 'Guardar pago pendiente' : esIngreso ? 'Registrar venta' : 'Registrar egreso';

  return (
    <KeyboardAvoidingView style={{ flex: 1, backgroundColor: colors.fondoApp }} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
      {/* Todo se ajusta al alto de la pantalla: el teclado crece para llenar el espacio.
          Solo hace scroll si el contenido no cabe (por ejemplo, con el pago pendiente abierto). */}
      <ScrollView
        scrollEnabled={!arrastrando}
        contentContainerStyle={{ flexGrow: 1, paddingBottom: t(10) }}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        <BarraTitulo
          titulo={empresa?.nombre_empresa ?? 'Cargando…'}
          derecha={
            <View style={{ flexDirection: 'row', alignItems: 'center', gap: t(4), backgroundColor: 'rgba(255,255,255,0.22)', paddingHorizontal: t(10), paddingVertical: t(4), borderRadius: 999 }}>
              <View style={{ width: t(6), height: t(6), borderRadius: 999, backgroundColor: colors.verde }} />
              <Text style={{ fontFamily: fonts.negrita, fontSize: t(10), color: colors.superficie }}>
                Hoy, {fechaHoy.getDate()} {MESES_CORTOS[fechaHoy.getMonth()]}
              </Text>
            </View>
          }
        />
        <View style={{ flex: 1, paddingHorizontal: t(16), paddingTop: t(10), gap: t(8) }}>
          <EstadoEmpresa />

          {/* 1. Tipo (ingreso / egreso) y monto */}
          <View style={tarjeta}>
            <View style={{ flexDirection: 'row', gap: t(4), backgroundColor: colors.lavandaSuave, padding: t(4), borderRadius: t(12) }}>
              {(['ingreso', 'egreso'] as const).map((opcion) => {
                const activo = tipo === opcion;
                return (
                  <Pressable key={opcion} onPress={() => cambiarTipo(opcion)} style={{
                    flex: 1, paddingVertical: t(4), borderRadius: t(8), alignItems: 'center',
                    backgroundColor: activo ? (opcion === 'ingreso' ? colors.ingreso : colors.egreso) : 'transparent',
                  }}>
                    <Text style={{ fontFamily: activo ? fonts.extra : fonts.negrita, fontSize: t(11), color: activo ? colors.superficie : colors.lavanda }}>
                      {opcion === 'ingreso' ? 'INGRESO (+)' : 'EGRESO (−)'}
                    </Text>
                  </Pressable>
                );
              })}
            </View>
            <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', paddingHorizontal: t(8), paddingVertical: t(2) }}>
              <Text style={{ fontFamily: fonts.negrita, fontSize: t(12), color: colors.textoSecundario }}>Monto</Text>
              <Text numberOfLines={1} adjustsFontSizeToFit style={{ flex: 1, textAlign: 'right', fontFamily: fonts.extra, fontSize: t(30), color: colors.lavanda, letterSpacing: -0.5 }}>
                $ {formatearMonto(monto)}
              </Text>
            </View>
          </View>

          {/* 2. Categoría, método de pago y nota */}
          <View style={[tarjeta, { gap: t(14) }]}>
            <SelectorCategorias tipo={tipo} categorias={cats} categoriaId={categoriaId} onElegir={setCategoriaId} onArrastrando={setArrastrando} />

            {/* Método de pago: solo ingresos (el backend no guarda método en los egresos) */}
            {esIngreso && <SelectorMetodo metodo={metodo} onCambiar={setMetodo} />}

            {esPendiente && <DatosPagoPendiente cliente={cliente} onCliente={setCliente} vence={vence} onVence={setVence} />}

            {/* Nota con ícono de lápiz */}
            <View>
              <View style={{ position: 'absolute', left: t(12), top: t(9), zIndex: 1 }}>
                <Icono nombre="pencil" tamano={t(14)} color={colors.textoSecundario} />
              </View>
              <TextInput value={nota} onChangeText={setNota} maxLength={100} multiline
                placeholder="Nota opcional (Ej: Venta poleras diseño)" placeholderTextColor={colors.textoSecundario}
                style={{
                  paddingVertical: t(7), paddingHorizontal: t(12), paddingLeft: t(32), textAlignVertical: 'top',
                  fontSize: t(12), fontFamily: fonts.normal, color: colors.texto,
                  backgroundColor: colors.fondoApp, borderWidth: 1, borderColor: colors.lavandaOnda, borderRadius: t(12),
                }} />
            </View>
          </View>

          {/* 3. Zona del pulgar: el teclado ocupa el espacio que queda */}
          <View style={[tarjeta, { flex: 1, minHeight: t(230) }]}>
            <Calculadora
              monto={monto}
              onCambiar={setMonto}
              textoBoton={textoBoton}
              colorBoton={esPendiente ? colors.naranjo : esIngreso ? colors.ingreso : colors.egreso}
              deshabilitado={incompleto || guardando}
              onRegistrar={registrar}
            />
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
