import { Ionicons } from '@expo/vector-icons';
import { useEffect, useState } from 'react';
import { Alert, KeyboardAvoidingView, Platform, ScrollView, Switch, Text, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { traducirError } from '../../features/auth/errores';
import { useSesion } from '../../features/auth/SesionContext';
import { actualizarUsuario, eliminarCuenta } from '../../features/auth/usuario.service';
import { cambiarNombreEmpresa, MAX_NOMBRE_EMPRESA } from '../../features/empresa/empresa.service';
import Chip from '../../features/movimientos/components/Chip';
import {
  cancelarRecordatorioDiario,
  notificacionesDisponibles,
  pedirPermiso,
  programarRecordatorioDiario,
} from '../../features/notificaciones/notificaciones';
import Boton from '../../shared/components/Boton';
import CampoTexto from '../../shared/components/CampoTexto';
import { colors } from '../../shared/theme/colors';
import { fonts } from '../../shared/theme/fonts';
import { sombras } from '../../shared/theme/sombras';
import { guardarLocal, leerLocal } from '../../shared/utils/almacen';
import { useEscala } from '../../shared/utils/escala';

type Recordatorio = { activo: boolean; hora: number };
const HORAS = [9, 13, 18, 20, 21];

const tarjeta = { backgroundColor: colors.superficie, borderRadius: 20, padding: 16, gap: 12, ...sombras.suave };

export default function Configuracion() {
  const { token, empresa, actualizarEmpresa, cerrarSesion, manejarError } = useSesion();
  const { e } = useEscala();
  const insets = useSafeAreaInsets();

  const [nombreEmpresa, setNombreEmpresa] = useState(empresa?.nombre_empresa ?? '');
  const [nombreUsuario, setNombreUsuario] = useState('');
  const [contrasena, setContrasena] = useState('');
  const [confirmacion, setConfirmacion] = useState('');
  const [errorContrasena, setErrorContrasena] = useState<string | undefined>();
  const [guardando, setGuardando] = useState<string | null>(null); // Qué sección se está guardando

  const [recordatorio, setRecordatorio] = useState<Recordatorio>({ activo: false, hora: 20 });

  useEffect(() => { leerLocal<Recordatorio>('recordatorio', { activo: false, hora: 20 }).then(setRecordatorio); }, []);

  // Recordatorio diario (Historia 008): notificación local a la hora elegida
  async function cambiarRecordatorio(nuevo: Recordatorio) {
    if (nuevo.activo) {
      if (!notificacionesDisponibles()) {
        Alert.alert('No disponible en Expo Go', 'El recordatorio funciona en la app instalada. En Expo Go las notificaciones están limitadas.');
        return;
      }
      const permitido = await pedirPermiso();
      if (!permitido) {
        Alert.alert('Permiso necesario', 'Activa las notificaciones de Tally en los ajustes del celular para recibir el recordatorio.');
        return;
      }
      await programarRecordatorioDiario(nuevo.hora, 0);
    } else {
      await cancelarRecordatorioDiario();
    }
    setRecordatorio(nuevo);
    await guardarLocal('recordatorio', nuevo);
  }

  // Cuando la empresa termina de cargar, se muestra su nombre en el campo
  useEffect(() => {
    if (empresa) setNombreEmpresa(empresa.nombre_empresa);
  }, [empresa]);

  const tituloSeccion = (icono: keyof typeof Ionicons.glyphMap, texto: string) => (
    <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
      <Ionicons name={icono} size={e(24)} color={colors.lavanda} />
      <Text style={{ fontFamily: fonts.negrita, fontSize: e(20), color: colors.lavanda }}>{texto}</Text>
    </View>
  );

  // Ejecuta una acción mostrando "Guardando..." en su botón y avisando el resultado
  async function guardar(seccion: string, accion: () => Promise<void>, mensaje: string) {
    setGuardando(seccion);
    try {
      await accion();
      Alert.alert('Listo', mensaje);
    } catch (error) {
      manejarError(error);
      Alert.alert('No se pudo guardar', traducirError(error));
    } finally {
      setGuardando(null);
    }
  }

  function guardarEmpresa() {
    if (!token || !empresa) return;
    guardar('empresa', async () => {
      await cambiarNombreEmpresa(empresa.id, nombreEmpresa, token);
      actualizarEmpresa({ ...empresa, nombre_empresa: nombreEmpresa.trim() });
    }, 'El nombre de tu emprendimiento se actualizó.');
  }

  function guardarNombre() {
    if (!token) return;
    guardar('nombre', async () => {
      await actualizarUsuario({ nombre: nombreUsuario }, token);
      setNombreUsuario('');
    }, 'Tu nombre se actualizó.');
  }

  function guardarContrasena() {
    if (!token) return;
    if (contrasena.length < 8 || contrasena.length > 72) {
      setErrorContrasena('Debe tener entre 8 y 72 caracteres');
      return;
    }
    if (contrasena !== confirmacion) {
      setErrorContrasena('Las contraseñas no coinciden');
      return;
    }
    setErrorContrasena(undefined);
    guardar('contrasena', async () => {
      await actualizarUsuario({ contrasena }, token);
      setContrasena('');
      setConfirmacion('');
    }, 'Tu contraseña se cambió.');
  }

  function confirmarEliminar() {
    Alert.alert(
      'Eliminar cuenta',
      'Se borrará tu cuenta y no podrás recuperarla. ¿Quieres continuar?',
      [
        { text: 'Cancelar', style: 'cancel' },
        {
          text: 'Eliminar',
          style: 'destructive',
          onPress: async () => {
            if (!token) return;
            try {
              await eliminarCuenta(token);
              await cerrarSesion();
            } catch (error) {
              manejarError(error);
              Alert.alert('No se pudo eliminar la cuenta', traducirError(error));
            }
          },
        },
      ],
    );
  }

  const nombreSinCambios = !empresa || nombreEmpresa.trim() === '' || nombreEmpresa.trim() === empresa.nombre_empresa;

  return (
    <KeyboardAvoidingView style={{ flex: 1, backgroundColor: colors.fondoApp }} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
      <ScrollView
        contentContainerStyle={{ paddingHorizontal: 20, paddingTop: insets.top + 16, paddingBottom: 32, gap: 18 }}
        keyboardShouldPersistTaps="handled"
      >
        <Text style={{ fontFamily: fonts.negrita, fontSize: e(30), color: colors.lavanda }}>Configuración</Text>

        {/* Emprendimiento */}
        <View style={tarjeta}>
          {tituloSeccion('storefront-outline', 'Mi emprendimiento')}
          <CampoTexto etiqueta="Nombre del emprendimiento" value={nombreEmpresa} onChangeText={setNombreEmpresa}
            maxLength={MAX_NOMBRE_EMPRESA} placeholder="Ej: Dulces Juanita" />
          <Boton titulo={guardando === 'empresa' ? 'Guardando...' : 'Guardar nombre'} onPress={guardarEmpresa}
            deshabilitado={nombreSinCambios || guardando !== null} />
        </View>

        {/* Recordatorio diario */}
        <View style={tarjeta}>
          <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
            {tituloSeccion('alarm-outline', 'Recordatorio diario')}
            <Switch value={recordatorio.activo} onValueChange={(v) => cambiarRecordatorio({ ...recordatorio, activo: v })}
              trackColor={{ true: colors.lavanda, false: colors.lavandaSuave }} thumbColor={colors.superficie} />
          </View>
          <Text style={{ fontFamily: fonts.normal, fontSize: e(15), color: colors.lavanda }}>
            Te avisamos todos los días para que no olvides registrar tus ventas y gastos.
          </Text>
          {recordatorio.activo && (
            <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }}>
              {HORAS.map((h) => (
                <Chip key={h} texto={`${h}:00`} activo={recordatorio.hora === h} onPress={() => cambiarRecordatorio({ activo: true, hora: h })} />
              ))}
            </View>
          )}
        </View>

        {/* Cuenta */}
        <View style={tarjeta}>
          {tituloSeccion('person-circle-outline', 'Mi cuenta')}
          <CampoTexto etiqueta="Cambiar mi nombre" value={nombreUsuario} onChangeText={setNombreUsuario}
            maxLength={100} placeholder="Nombre y apellido" />
          <Boton titulo={guardando === 'nombre' ? 'Guardando...' : 'Guardar nombre'} onPress={guardarNombre}
            deshabilitado={nombreUsuario.trim() === '' || guardando !== null} variante="secundario" />

          <CampoTexto etiqueta="Nueva contraseña" value={contrasena} onChangeText={setContrasena}
            secureTextEntry maxLength={72} error={errorContrasena} />
          <CampoTexto etiqueta="Repite la nueva contraseña" value={confirmacion} onChangeText={setConfirmacion}
            secureTextEntry maxLength={72} />
          <Boton titulo={guardando === 'contrasena' ? 'Guardando...' : 'Cambiar contraseña'} onPress={guardarContrasena}
            deshabilitado={!contrasena || !confirmacion || guardando !== null} variante="secundario" />
        </View>

        {/* Sesión */}
        <View style={tarjeta}>
          {tituloSeccion('log-out-outline', 'Sesión')}
          <Boton titulo="Cerrar sesión" onPress={cerrarSesion} />
          <Boton titulo="Eliminar mi cuenta" onPress={confirmarEliminar} variante="secundario" color={colors.error} />
        </View>

        <Text style={{ fontFamily: fonts.normal, fontSize: e(13), color: colors.lavanda, opacity: 0.7, textAlign: 'center' }}>
          Tally · Finanzas simples · versión 1.0
        </Text>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
