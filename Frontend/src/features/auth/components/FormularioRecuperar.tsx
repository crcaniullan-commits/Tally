import { useState } from 'react';
import { Pressable, Text } from 'react-native';
import Boton from '../../../shared/components/Boton';
import CampoTexto from '../../../shared/components/CampoTexto';
import { datosPrueba } from '../../../shared/mocks/datosPrueba';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { confirmarRecuperacion, solicitarRecuperacion, verificarCodigoRecuperacion } from '../auth.service';
import { traducirError } from '../errores';

type Errores = { correo?: string; codigo?: string; contrasena?: string; confirmacion?: string };
type Etapa = 'correo' | 'codigo' | 'nuevaContrasena';

type FormularioRecuperarProps = {
  onExito: () => void;
  onVolver: () => void;
};

// Recuperar contraseña (Historia 011), en 3 pasos:
// 1) correo → el backend envía un código de 6 números, 2) verificar el código, 3) nueva contraseña
export default function FormularioRecuperar({ onExito, onVolver }: FormularioRecuperarProps) {
  const [etapa, setEtapa] = useState<Etapa>('correo');
  const [correo, setCorreo] = useState(datosPrueba.correo);
  const [codigo, setCodigo] = useState('');
  const [contrasena, setContrasena] = useState('');
  const [confirmacion, setConfirmacion] = useState('');
  const [errores, setErrores] = useState<Errores>({});
  const [cargando, setCargando] = useState(false);

  // Ejecuta una llamada al backend mostrando "cargando" y el error en el campo indicado
  async function ejecutar(accion: () => Promise<unknown>, campoError: keyof Errores, alTerminar: () => void) {
    setCargando(true);
    setErrores({});
    try {
      await accion();
      alTerminar();
    } catch (error) {
      setErrores({ [campoError]: traducirError(error) });
    } finally {
      setCargando(false);
    }
  }

  function enviarCodigo() {
    if (!/^\S+@\S+\.\S+$/.test(correo.trim())) return setErrores({ correo: 'Ingresa un correo válido' });
    ejecutar(() => solicitarRecuperacion(correo), 'correo', () => setEtapa('codigo'));
  }

  function verificarCodigo() {
    if (!/^\d{6}$/.test(codigo.trim())) return setErrores({ codigo: 'El código tiene 6 números' });
    ejecutar(() => verificarCodigoRecuperacion(correo, codigo), 'codigo', () => setEtapa('nuevaContrasena'));
  }

  function cambiarContrasena() {
    const nuevos: Errores = {};
    if (contrasena.length < 8 || contrasena.length > 72) nuevos.contrasena = 'Debe tener entre 8 y 72 caracteres';
    if (confirmacion !== contrasena) nuevos.confirmacion = 'Las contraseñas no coinciden';
    if (Object.keys(nuevos).length > 0) return setErrores(nuevos);
    ejecutar(() => confirmarRecuperacion(correo, codigo, contrasena), 'contrasena', onExito);
  }

  const textoAyuda = { color: colors.textoSecundario, fontFamily: fonts.normal };
  const correoNegrita = <Text style={{ fontFamily: fonts.negrita, color: colors.texto }}>{correo.trim()}</Text>;

  if (etapa === 'correo') {
    return (
      <>
        <Text style={textoAyuda}>Escribe el correo con el que te registraste y te enviaremos un código de recuperación.</Text>
        <CampoTexto etiqueta="Correo electrónico" placeholder="tucorreo@ejemplo.cl"
          value={correo} onChangeText={setCorreo} keyboardType="email-address"
          autoCapitalize="none" error={errores.correo} />
        <Boton titulo={cargando ? 'Enviando...' : 'Enviar código'} onPress={enviarCodigo} deshabilitado={correo.trim() === '' || cargando} />
        <Boton titulo="Volver a iniciar sesión" variante="secundario" onPress={onVolver} />
      </>
    );
  }

  if (etapa === 'codigo') {
    return (
      <>
        <Text style={textoAyuda}>Revisa tu correo {correoNegrita} e ingresa el código de 6 números que te enviamos.</Text>
        <CampoTexto etiqueta="Código de recuperación" placeholder="123456"
          value={codigo} onChangeText={(v) => setCodigo(v.replace(/\D/g, '').slice(0, 6))}
          keyboardType="number-pad" error={errores.codigo} />
        <Boton titulo={cargando ? 'Verificando...' : 'Verificar código'} onPress={verificarCodigo} deshabilitado={codigo.length !== 6 || cargando} />
        <Pressable onPress={() => ejecutar(() => solicitarRecuperacion(correo), 'codigo', () => setCodigo(''))} style={{ alignSelf: 'center' }}>
          <Text style={{ color: colors.lavanda, fontFamily: fonts.medio }}>Reenviar código</Text>
        </Pressable>
        <Pressable onPress={() => setEtapa('correo')} style={{ alignSelf: 'center' }}>
          <Text style={{ color: colors.lavanda, fontFamily: fonts.medio }}>Usar otro correo</Text>
        </Pressable>
      </>
    );
  }

  return (
    <>
      <Text style={textoAyuda}>Código verificado. Ahora crea tu nueva contraseña.</Text>
      <CampoTexto etiqueta="Nueva contraseña" placeholder="Mínimo 8 caracteres"
        value={contrasena} onChangeText={setContrasena} secureTextEntry error={errores.contrasena} />
      <CampoTexto etiqueta="Confirmar nueva contraseña" placeholder="Repite tu contraseña"
        value={confirmacion} onChangeText={setConfirmacion} secureTextEntry error={errores.confirmacion} />
      <Boton titulo={cargando ? 'Guardando...' : 'Cambiar contraseña'} onPress={cambiarContrasena}
        deshabilitado={contrasena === '' || confirmacion === '' || cargando} />
    </>
  );
}
