import { BebasNeue_400Regular } from '@expo-google-fonts/bebas-neue';
import { GrandHotel_400Regular } from '@expo-google-fonts/grand-hotel';
import {
  PlusJakartaSans_400Regular,
  PlusJakartaSans_600SemiBold,
  PlusJakartaSans_700Bold,
  PlusJakartaSans_800ExtraBold,
  useFonts,
} from '@expo-google-fonts/plus-jakarta-sans';
import { Stack } from 'expo-router';
import { Image, useWindowDimensions, View } from 'react-native';
import FondoOndas from '../shared/components/FondoOndas';
import { colors } from '../shared/theme/colors';
import { SesionProvider, useSesion } from '../features/auth/SesionContext';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { ConfirmarProvider } from '../shared/components/Confirmar';
import { AvisoProvider } from '../shared/components/Toast';

// Pantalla de carga: misma estética que la Bienvenida (fondo lavanda, ondas y logo),
// se ve mientras cargan las fuentes y se revisa si hay una sesión guardada
function PantallaCarga() {
  const { width, height } = useWindowDimensions();
  return (
    <View style={{ flex: 1, backgroundColor: colors.fondo }}>
      <FondoOndas />
      <View style={{ position: 'absolute', top: height * 0.3, width: '100%', alignItems: 'center' }}>
        <Image
          source={require('../../assets/images/logo-tally.png')}
          style={{ width: width * 0.67, height: width * 0.67 * 0.92 }}
          resizeMode="contain"
        />
      </View>
    </View>
  );
}

function Navegacion() {
  const { token, rol, cargando } = useSesion();

  // Mientras se revisa si hay una sesión guardada, no mostramos nada
  if (cargando) return <PantallaCarga />;

  const esEmprendedor = !!token && rol === 'usuario';
  const esMunicipal = !!token && rol === 'municipal';

  return (
    <Stack screenOptions={{ headerShown: false }}>
      {/* Pantallas públicas: solo sin sesión */}
      <Stack.Protected guard={!token}>
        <Stack.Screen name="index" />
      </Stack.Protected>

      {/* Emprendedor: pestañas Inicio, Dashboard, Registros y Configuración (DoD Historia 010) */}
      <Stack.Protected guard={esEmprendedor}>
        <Stack.Screen name="(tabs)" />
      </Stack.Protected>

      {/* Funcionario municipal: backoffice de claves (Historia 009) */}
      <Stack.Protected guard={esMunicipal}>
        <Stack.Screen name="municipal" />
      </Stack.Protected>
    </Stack>
  );
}

export default function RootLayout() {
  const [fuentesCargadas] = useFonts({
    GrandHotel_400Regular,
    BebasNeue_400Regular,
    PlusJakartaSans_400Regular,
    PlusJakartaSans_600SemiBold,
    PlusJakartaSans_700Bold,
    PlusJakartaSans_800ExtraBold,
  });

  if (!fuentesCargadas) return <PantallaCarga />;

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
    <SesionProvider>
      {/* Avisos y confirmaciones con la estética de Tally, disponibles en todas las pantallas */}
      <AvisoProvider>
        <ConfirmarProvider>
          <Navegacion />
        </ConfirmarProvider>
      </AvisoProvider>
    </SesionProvider>
    </GestureHandlerRootView>
  );
}