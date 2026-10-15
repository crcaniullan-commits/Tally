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
  const { token, cargando } = useSesion();

  // Mientras se revisa si hay una sesión guardada, no mostramos nada
  if (cargando) return null;

  return (
    <Stack screenOptions={{ headerShown: false }}>
      {/* Pantallas públicas: solo sin sesión */}
      <Stack.Protected guard={!token}>
        <Stack.Screen name="index" />
      </Stack.Protected>

      {/* Pantallas privadas: solo con sesión iniciada (DoD Historia 010) */}
      <Stack.Protected guard={!!token}>
        <Stack.Screen name="inicio" />
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