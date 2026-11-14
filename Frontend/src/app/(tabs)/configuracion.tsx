import { useEffect, useState } from 'react';
import { KeyboardAvoidingView, Platform, Pressable, ScrollView, Switch, Text, TextInput, View } from 'react-native';
import { traducirError } from '../../features/auth/errores';
import { useSesion } from '../../features/auth/SesionContext';
import { actualizarUsuario, eliminarCuenta } from '../../features/auth/usuario.service';
import { cambiarNombreEmpresa, MAX_NOMBRE_EMPRESA } from '../../features/empresa/empresa.service';
import {
  cancelarRecordatorioDiario,
  notificacionesDisponibles,
  pedirPermiso,
  programarRecordatorioDiario,
} from '../../features/notificaciones/notificaciones';
import BarraTitulo from '../../shared/components/BarraTitulo';
import Segmentado from '../../shared/components/Segmentado';
import { guardarVistaCategorias, leerVistaCategorias, VistaCategorias } from '../../features/categorias/ordenCategorias';
import Icono from '../../shared/components/Icono';
import { useAviso } from '../../shared/components/Toast';
import { useConfirmar, useAlerta } from '../../shared/components/Confirmar';
import { ConfigAvisosPendientes, guardarConfigAvisos, leerConfigAvisos, reprogramarAvisos } from '../../features/pendientes/pendientes';
import { colors } from '../../shared/theme/colors';
import { fonts } from '../../shared/theme/fonts';
import { guardarLocal, leerLocal } from '../../shared/utils/almacen';
import { useEscala } from '../../shared/utils/escala';

type Recordatorio = { activo: boolean; hora: number };
const HORAS = [9, 13, 18, 20, 21];

export default function Configuracion() {
  const alerta = useAlerta();
  const { token, empresa, actualizarEmpresa, cerrarSesion, manejarError } = useSesion();
  const { t } = useEscala();
  const avisar = useAviso();
  const confirmar = useConfirmar();
  const [avisosPendientes, setAvisosPendientes] = useState<ConfigAvisosPendientes>({ activo: true, diasAntes: 0 });
  useEffect(() => { leerConfigAvisos().then(setAvisosPendientes); }, []);
  const [vistaCategorias, setVistaCategorias] = useState<VistaCategorias>('carrusel');
  useEffect(() => { leerVistaCategorias().then(setVistaCategorias); }, []);

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
        alerta('No disponible en Expo Go', 'El recordatorio funciona en la app instalada. En Expo Go las notificaciones están limitadas.');
        return;
      }
      const permitido = await pedirPermiso();
      if (!permitido) {
        alerta('Permiso necesario', 'Activa las notificaciones de Tally en los ajustes del celular para recibir el recordatorio.');
        return;
      }
      await programarRecordatorioDiario(nuevo.hora, 0);
    } else {
      await cancelarRecordatorioDiario();
    }
    setRecordatorio(nuevo);
    await guardarLocal('recordatorio', nuevo);
  }

  // Recordatorio de pagos pendientes (Historia 008): activar y elegir cuántos días antes avisar
  async function cambiarAvisosPendientes(nuevo: ConfigAvisosPendientes) {
    if (nuevo.activo && !notificacionesDisponibles()) {
      alerta('No disponible en Expo Go', 'Los avisos funcionan en la app instalada. En Expo Go las notificaciones están limitadas.');
    } else if (nuevo.activo && !(await pedirPermiso())) {
      alerta('Permiso necesario', 'Activa las notificaciones de Tally en los ajustes del celular.');
      return;
    }
    setAvisosPendientes(nuevo);
    await guardarConfigAvisos(nuevo);
    if (token && empresa) await reprogramarAvisos(empresa.id, token).catch(() => {});
    avisar(nuevo.activo ? 'Avisos de pagos pendientes guardados' : 'Avisos de pagos pendientes desactivados', 'bell');
  }

  // Cuando la empresa termina de cargar, se muestra su nombre en el campo
  useEffect(() => {
    if (empresa) setNombreEmpresa(empresa.nombre_empresa);
  }, [empresa]);

  // Ejecuta una acción mostrando "Guardando..." en su botón y avisando el resultado
  async function guardar(seccion: string, accion: () => Promise<void>, mensaje: string) {
    setGuardando(seccion);
    try {
      await accion();
      avisar(mensaje);
    } catch (error) {
      manejarError(error);
      alerta('No se pudo guardar', traducirError(error));
    } finally {
      setGuardando(null);
    }
  }

  function guardarEmpresa() {
    if (!token || !empresa) return;
    guardar('empresa', async () => {
      await cambiarNombreEmpresa(empresa.id, nombreEmpresa, token);
      actualizarEmpresa({ ...empresa, nombre_empresa: nombreEmpresa.trim() });
    }, '¡Nombre actualizado!');
  }

  // Mi cuenta (mockup): un solo botón guarda el nombre y/o la nueva contraseña
  function guardarCuenta() {
    if (!token) return;
    const nombre = nombreUsuario.trim();
    if (contrasena) {
      if (contrasena.length < 8 || contrasena.length > 72) return setErrorContrasena('Debe tener entre 8 y 72 caracteres');
      if (contrasena !== confirmacion) return setErrorContrasena('Las contraseñas no coinciden');
    }
    setErrorContrasena(undefined);
    guardar('cuenta', async () => {
      await actualizarUsuario({ nombre: nombre || undefined, contrasena: contrasena || undefined }, token);
      setNombreUsuario('');
      setContrasena('');
      setConfirmacion('');
    }, '¡Datos actualizados!');
  }

  async function confirmarEliminar() {
    const ok = await confirmar({
      titulo: 'Eliminar cuenta', icono: 'user-x', textoConfirmar: 'Eliminar cuenta',
      mensaje: 'Se borrará tu cuenta y no podrás recuperarla. ¿Quieres continuar?',
    });
    if (!ok || !token) return;
    try {
      await eliminarCuenta(token);
      await cerrarSesion();
    } catch (error) {
      manejarError(error);
      alerta('No se pudo eliminar la cuenta', traducirError(error));
    }
  }

  const nombreSinCambios = !empresa || nombreEmpresa.trim() === '' || nombreEmpresa.trim() === empresa.nombre_empresa;
  const cuentaSinCambios = !nombreUsuario.trim() && !contrasena;

  // ---- Estilos del mockup ----
  const tarjeta = {
    backgroundColor: colors.superficie, padding: t(14), borderRadius: t(16), gap: t(10),
    borderWidth: 1, borderColor: colors.lavandaOnda, boxShadow: '0px 1px 2px rgba(0, 0, 0, 0.05)',
  };
  const campo = {
    paddingHorizontal: t(12), paddingVertical: t(8), fontSize: t(12), fontFamily: fonts.medio, color: colors.texto,
    backgroundColor: colors.fondoApp, borderWidth: 1, borderColor: colors.lavandaOnda, borderRadius: t(12),
  };
  const tituloSeccion = (icono: string, texto: string) => (
    <View style={{ flexDirection: 'row', alignItems: 'center', gap: t(8) }}>
      <Icono nombre={icono} tamano={t(16)} color={colors.lavanda} />
      <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.lavanda }}>{texto}</Text>
    </View>
  );
  const boton = (texto: string, accion: () => void, deshabilitado: boolean, relleno: boolean) => (
    <Pressable onPress={accion} disabled={deshabilitado} style={({ pressed }) => ({
      paddingVertical: t(8), borderRadius: t(12), alignItems: 'center', borderWidth: 1, borderColor: colors.lavanda,
      backgroundColor: relleno ? colors.lavanda : pressed ? colors.lavandaSuave : 'transparent',
      opacity: deshabilitado ? 0.45 : pressed && relleno ? 0.9 : 1,
    })}>
      <Text style={{ fontFamily: fonts.negrita, fontSize: t(12), color: relleno ? colors.superficie : colors.lavanda }}>{texto}</Text>
    </Pressable>
  );

  return (
    <KeyboardAvoidingView style={{ flex: 1, backgroundColor: colors.fondoApp }} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
      <ScrollView contentContainerStyle={{ paddingBottom: t(20) }} keyboardShouldPersistTaps="handled" showsVerticalScrollIndicator={false}>
      <BarraTitulo titulo="Configuración" />
      <View style={{ paddingHorizontal: t(16), paddingTop: t(12), gap: t(12) }}>
        {/* Mi emprendimiento */}
        <View style={tarjeta}>
          {tituloSeccion('store', 'Mi emprendimiento')}
          <View>
            <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: colors.texto, marginBottom: t(4) }}>Nombre del emprendimiento</Text>
            <TextInput value={nombreEmpresa} onChangeText={setNombreEmpresa} maxLength={MAX_NOMBRE_EMPRESA}
              placeholder="Ej: Dulces Juanita" placeholderTextColor={colors.textoSecundario} style={campo} />
          </View>
          {boton(guardando === 'empresa' ? 'Guardando...' : 'Guardar nombre', guardarEmpresa, nombreSinCambios || guardando !== null, true)}
        </View>

        {/* Recordatorio diario (Historia 008) */}
        <View style={tarjeta}>
          <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
            <View style={{ flex: 1, paddingRight: t(12) }}>
              {tituloSeccion('alarm-clock', 'Recordatorio diario')}
              <Text style={{ fontFamily: fonts.normal, fontSize: t(11), color: colors.textoSecundario, marginTop: t(2), lineHeight: t(14) }}>
                Te avisamos todos los días para que no olvides registrar tus ingresos y egresos.
              </Text>
            </View>
            <Switch value={recordatorio.activo} onValueChange={(v) => cambiarRecordatorio({ ...recordatorio, activo: v })}
              trackColor={{ true: colors.lavanda, false: colors.lavandaOnda }} thumbColor={colors.superficie} />
          </View>
          {/* La hora no está en el mockup: aparece solo con el recordatorio activo */}
          {recordatorio.activo && (
            <View style={{ flexDirection: 'row', gap: t(4) }}>
              {HORAS.map((h) => {
                const activo = recordatorio.hora === h;
                return (
                  <Pressable key={h} onPress={() => cambiarRecordatorio({ activo: true, hora: h })} style={{
                    flex: 1, alignItems: 'center', paddingVertical: t(4), borderRadius: t(8), borderWidth: 1,
                    backgroundColor: activo ? colors.lavanda : colors.fondoApp, borderColor: activo ? colors.lavanda : colors.lavandaOnda,
                  }}>
                    <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: activo ? colors.superficie : colors.texto }}>{h}:00</Text>
                  </Pressable>
                );
              })}
            </View>
          )}
        </View>

        {/* Recordatorio de pagos pendientes */}
        <View style={tarjeta}>
          <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
            <View style={{ flex: 1, paddingRight: t(12) }}>
              {tituloSeccion('clock', 'Pagos pendientes')}
              <Text style={{ fontFamily: fonts.normal, fontSize: t(11), color: colors.textoSecundario, marginTop: t(2), lineHeight: t(14) }}>
                Te recordamos cuándo vence lo que te deben (a las 10:00).
              </Text>
            </View>
            <Switch value={avisosPendientes.activo} onValueChange={(v) => cambiarAvisosPendientes({ ...avisosPendientes, activo: v })}
              trackColor={{ true: colors.naranjo, false: colors.lavandaOnda }} thumbColor={colors.superficie} />
          </View>
          {avisosPendientes.activo && (
            <View style={{ flexDirection: 'row', gap: t(4) }}>
              {[{ d: 0, texto: 'El mismo día' }, { d: 1, texto: '1 día antes' }, { d: 3, texto: '3 días antes' }].map(({ d, texto }) => {
                const activo = avisosPendientes.diasAntes === d;
                return (
                  <Pressable key={d} onPress={() => cambiarAvisosPendientes({ activo: true, diasAntes: d })} style={{
                    flex: 1, alignItems: 'center', paddingVertical: t(5), borderRadius: t(8), borderWidth: 1,
                    backgroundColor: activo ? colors.naranjo : colors.fondoApp, borderColor: activo ? colors.naranjo : colors.lavandaOnda,
                  }}>
                    <Text style={{ fontFamily: fonts.negrita, fontSize: t(10), color: activo ? colors.superficie : colors.texto }}>{texto}</Text>
                  </Pressable>
                );
              })}
            </View>
          )}
        </View>

        {/* Cómo se muestran las categorías en Inicio */}
        <View style={tarjeta}>
          {tituloSeccion('tag', 'Categorías en Inicio')}
          <Text style={{ fontFamily: fonts.normal, fontSize: t(11), color: colors.textoSecundario, lineHeight: t(14) }}>
            En grilla se ven todas siempre, pero si tienes muchas tendrás que bajar para llegar a la calculadora.
          </Text>
          <Segmentado
            opciones={[{ id: 'carrusel', nombre: 'Carrusel (una fila)' }, { id: 'grilla', nombre: 'Grilla (todas)' }]}
            valor={vistaCategorias}
            onCambiar={(v) => { setVistaCategorias(v); guardarVistaCategorias(v); avisar('Preferencia guardada', 'tag'); }}
          />
        </View>

        {/* Mi cuenta */}
        <View style={tarjeta}>
          {tituloSeccion('user', 'Mi cuenta')}
          <TextInput value={nombreUsuario} onChangeText={setNombreUsuario} maxLength={100}
            placeholder="Nombre y apellido" placeholderTextColor={colors.textoSecundario} style={campo} />
          <TextInput value={contrasena} onChangeText={(v) => { setContrasena(v); setErrorContrasena(undefined); }} maxLength={72} secureTextEntry
            placeholder="Nueva contraseña" placeholderTextColor={colors.textoSecundario}
            style={[campo, errorContrasena ? { borderColor: colors.error } : null]} />
          {contrasena.length > 0 && (
            <TextInput value={confirmacion} onChangeText={setConfirmacion} maxLength={72} secureTextEntry
              placeholder="Repite la nueva contraseña" placeholderTextColor={colors.textoSecundario} style={campo} />
          )}
          {errorContrasena && <Text style={{ fontFamily: fonts.medio, fontSize: t(11), color: colors.error }}>{errorContrasena}</Text>}
          {boton(guardando === 'cuenta' ? 'Guardando...' : 'Guardar cambios', guardarCuenta, cuentaSinCambios || guardando !== null, false)}
        </View>

        {/* Cerrar sesión */}
        <Pressable onPress={cerrarSesion} style={({ pressed }) => ({
          flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: t(6), paddingVertical: t(10), borderRadius: t(12),
          backgroundColor: colors.rosadoSuave, borderWidth: 1, borderColor: 'rgba(248, 47, 143, 0.3)', opacity: pressed ? 0.85 : 1,
        })}>
          <Icono nombre="log-out" tamano={t(16)} color={colors.rosado} />
          <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.rosado }}>Cerrar sesión</Text>
        </Pressable>

        {/* Eliminar cuenta: no está en el mockup, queda como enlace discreto */}
        <Pressable onPress={confirmarEliminar} hitSlop={8} style={{ alignSelf: 'center', paddingVertical: t(4) }}>
          <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: colors.textoSecundario, textDecorationLine: 'underline' }}>Eliminar mi cuenta</Text>
        </Pressable>

        <Text style={{ fontFamily: fonts.normal, fontSize: t(10), color: colors.textoSecundario, textAlign: 'center' }}>
          Tally · Finanzas simples · versión 1.0
        </Text>
      </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
