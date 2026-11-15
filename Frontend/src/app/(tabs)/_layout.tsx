import { Tabs } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { ColorValue } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import Icono from '../../shared/components/Icono';
import { colors } from '../../shared/theme/colors';
import { useEscala } from '../../shared/utils/escala';

// Al iniciar sesión, la primera pestaña que se abre es Inicio
export const unstable_settings = { initialRouteName: 'inicio' };

export default function LayoutPestanas() {
  // Espacio que el propio teléfono informa para su barra de botones o de gestos (0 si no hace falta)
  const insets = useSafeAreaInsets();
  const { t } = useEscala();

  // Barra inferior blanca, solo con íconos (el activo en lavanda y más grueso)
  const pestana = (titulo: string, icono: string) => ({
    title: titulo,
    tabBarIcon: ({ color, focused }: { color: ColorValue; focused: boolean }) => (
      <Icono nombre={icono} tamano={t(19)} color={String(color)} grosor={focused ? 2.5 : 1.8} />
    ),
  });

  return (
    <>
      <StatusBar style="dark" />
      <Tabs
        screenOptions={{
          headerShown: false,
          tabBarShowLabel: false,
          tabBarActiveTintColor: colors.lavanda,
          tabBarInactiveTintColor: colors.textoSecundario,
          tabBarStyle: {
            height: t(30) + insets.bottom,
            paddingTop: t(2),
            paddingBottom: insets.bottom + t(2),
            backgroundColor: colors.superficie,
            borderTopWidth: 1,
            borderTopColor: colors.lavandaOnda,
            elevation: 0,
          },
        }}
      >
        <Tabs.Screen name="inicio" options={pestana('Inicio', 'home')} />
        <Tabs.Screen name="dashboard" options={pestana('Dashboard', 'bar-chart-2')} />
        <Tabs.Screen name="registros" options={pestana('Registros', 'receipt')} />
        <Tabs.Screen name="configuracion" options={pestana('Configuración', 'settings')} />
      </Tabs>
    </>
  );
}
