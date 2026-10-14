import { BebasNeue_400Regular } from '@expo-google-fonts/bebas-neue';
import { GrandHotel_400Regular } from '@expo-google-fonts/grand-hotel';
import {
  HankenGrotesk_400Regular,
  HankenGrotesk_600SemiBold,
  HankenGrotesk_700Bold,
  useFonts,
} from '@expo-google-fonts/hanken-grotesk';
import { Stack } from 'expo-router';

export default function RootLayout() {
  const [fuentesCargadas] = useFonts({
    GrandHotel_400Regular,
    BebasNeue_400Regular,
    HankenGrotesk_400Regular,
    HankenGrotesk_600SemiBold,
    HankenGrotesk_700Bold,
  });

  if (!fuentesCargadas) return null;

  return <Stack screenOptions={{ headerShown: false }} />;
}