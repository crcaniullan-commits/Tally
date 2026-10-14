import { router } from 'expo-router';
import { useState } from 'react';
import { Image, useWindowDimensions, View } from 'react-native';
import Boton from '../shared/components/Boton';
import FormularioLogin from '../features/auth/components/FormularioLogin';
import FormularioRegistro, { DatosRegistro } from '../features/auth/components/FormularioRegistro';
import ModalFlotante from '../shared/components/ModalFlotante';
import { colors } from '../shared/theme/colors';
import { fonts } from '../shared/theme/fonts';
import FormularioActivacion from '../features/auth/components/FormularioActivacion';
import FondoOndas from '../shared/components/FondoOndas';
import { Alert } from 'react-native';
import FormularioRecuperar from '../features/auth/components/FormularioRecuperar';


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
  const [datosRegistro, setDatosRegistro] = useState<DatosRegistro | undefined>();

  function cerrarModal() {
    setModal(null);
    setDatosRegistro(undefined);
  }
    return (
    <View style={{ flex: 1, backgroundColor: colors.fondo }}>
      <FondoOndas />

      {/* Logo: empieza al 21% de la altura y ocupa el 67% del ancho, como en el mockup */}
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
            onExito={() => { cerrarModal(); router.replace('/inicio'); }}
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
            onExito={() => { cerrarModal(); router.replace('/inicio'); }}
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