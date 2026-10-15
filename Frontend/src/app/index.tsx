
import { useState } from 'react';
import { Alert, Image, useWindowDimensions, View } from 'react-native';
import Boton from '../shared/components/Boton';
import FormularioLogin from '../features/auth/components/FormularioLogin';
import FormularioRegistro, { DatosRegistro } from '../features/auth/components/FormularioRegistro';
import ModalFlotante from '../shared/components/ModalFlotante';
import { colors } from '../shared/theme/colors';
import { fonts } from '../shared/theme/fonts';
import FormularioActivacion from '../features/auth/components/FormularioActivacion';
import FondoOndas from '../shared/components/FondoOndas';
import FormularioRecuperar from '../features/auth/components/FormularioRecuperar';
import * as authService from '../features/auth/auth.service';
import { useSesion } from '../features/auth/SesionContext';


type ModalAbierto = 'login' | 'registro' | 'activacion' | 'recuperar' | null;

const titulos = {
  login: 'Iniciar sesión',
  registro: 'Crear cuenta',
  activacion: 'Crear cuenta',
  recuperar: 'Recuperar contraseña',
};

export default function Index() {
  const { width, height } = useWindowDimensions();
  const [modal, setModal] = useState<ModalAbierto>(null);
  const { iniciarSesion } = useSesion();
  const [datosRegistro, setDatosRegistro] = useState<DatosRegistro | undefined>();
    const [tokenRegistro, setTokenRegistro] = useState<string | undefined>();

  // Paso 2 del registro: crea la cuenta (solo la primera vez) y canjea la clave municipal
  async function activarCuenta(clave: string) {
    if (!datosRegistro) return;

    let token = tokenRegistro;
    if (!token) {
      token = await authService.registrar(datosRegistro);
      setTokenRegistro(token); // Si la clave falla, al reintentar no se vuelve a crear la cuenta
    }

    await authService.canjearClave(clave, token);
    cerrarModal();
    await iniciarSesion(token);
    }

  function cerrarModal() {
    setModal(null);
    setDatosRegistro(undefined);
    setTokenRegistro(undefined);
  }
    return (
    <View style={{ flex: 1, backgroundColor: colors.fondo }}>
      <FondoOndas />

      {/* Logo: empieza al 21% de la altura y ocupa el 67% del ancho */}
      <View style={{ position: 'absolute', top: height * 0.30, width: '100%', alignItems: 'center' }}>
        <Image
          source={require('../../assets/images/logo-tally.png')}
          style={{ width: width * 0.67, height: width * 0.67 * 0.92 }}
          resizeMode="contain"
        />
      </View>

      {/* Botones: empiezan al 58% de la altura y ocupan el 90% del ancho */}
      <View style={{ position: 'absolute', top: height * 0.71, width: '100%', alignItems: 'center' }}>
        <View style={{ width: '90%', gap: 17 }}>
          <Boton titulo="Crear cuenta" onPress={() => setModal('registro')} />
          <Boton titulo="Iniciar sesión" variante="secundario" onPress={() => setModal('login')} />
        </View>
      </View>


      <ModalFlotante
        visible={modal !== null}
        titulo={modal ? titulos[modal] : ''}
        onCerrar={cerrarModal}
      >
        {modal === 'login' && (
          <FormularioLogin
            onExito={(token) => { cerrarModal(); iniciarSesion(token); }}
            onOlvido={() => setModal('recuperar')}
            onCrearCuenta={() => setModal('registro')}
          />
        )}
        {modal === 'registro' && (
          <FormularioRegistro
            valoresIniciales={datosRegistro}
            onExito={(datos) => { setDatosRegistro(datos); setModal('activacion'); }}
            onIrALogin={() => setModal('login')}
          />
        )}
        {modal === 'activacion' && (
          <FormularioActivacion
            onActivar={activarCuenta}
            onVolver={() => setModal('registro')}
          />
        )}
                {modal === 'recuperar' && (
          <FormularioRecuperar
            onExito={() => {
              setModal('login');
              Alert.alert('Contraseña actualizada', 'Ya puedes iniciar sesión con tu nueva contraseña.');
            }}
            onVolver={() => setModal('login')}
          />
        )}
      </ModalFlotante>
    </View>
  );
}