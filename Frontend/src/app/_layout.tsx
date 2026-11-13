import { BebasNeue_400Regular } from '@expo-google-fonts/bebas-neue';
import { GrandHotel_400Regular } from '@expo-google-fonts/grand-hotel';
import {
  HankenGrotesk_400Regular,
  HankenGrotesk_600SemiBold,
  HankenGrotesk_700Bold,
  useFonts,
} from '@expo-google-fonts/hanken-grotesk';
import { Stack } from 'expo-router';
import { SesionProvider, useSesion } from '../features/auth/SesionContext';

function Navegacion() {
  const { token, rol, cargando } = useSesion();

  // Mientras se revisa si hay una sesión guardada, no mostramos nada
  if (cargando) return null;

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
    HankenGrotesk_400Regular,
    HankenGrotesk_600SemiBold,
    HankenGrotesk_700Bold,
  });

  if (!fuentesCargadas) return null;

  return (
    <SesionProvider>
      <Navegacion />
    </SesionProvider>
  );
}