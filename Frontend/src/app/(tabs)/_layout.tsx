import { Ionicons } from '@expo/vector-icons';
import { Tabs } from 'expo-router';
import { ColorValue } from 'react-native';
import { colors } from '../../shared/theme/colors';
import { fonts } from '../../shared/theme/fonts';
import { sombras } from '../../shared/theme/sombras';

// Al iniciar sesión, la primera pestaña que se abre es Inicio
export const unstable_settings = { initialRouteName: 'inicio' };

type NombreIcono = keyof typeof Ionicons.glyphMap;

// Ícono relleno cuando la pestaña está activa y de contorno cuando no
function icono(activo: NombreIcono, inactivo: NombreIcono) {
    return ({ focused, color }: { focused: boolean; color: ColorValue }) => (
        <Ionicons name={focused ? activo : inactivo} size={28} color={color} />
    );
    }

    export default function LayoutPestanas() {
    return (
        <Tabs
        screenOptions={{
            headerShown: false,
            tabBarActiveTintColor: colors.lavanda,
            tabBarInactiveTintColor: colors.textoSecundario,
            tabBarLabelStyle: { fontFamily: fonts.medio, fontSize: 14 },
            tabBarStyle: {
            height: 80,
            paddingTop: 8,
            paddingBottom: 12,
            backgroundColor: colors.superficie,
            borderTopWidth: 0,
            borderTopLeftRadius: 28,
            borderTopRightRadius: 28,
            ...sombras.flotante,
            },
        }}
        >
        <Tabs.Screen name="inicio" options={{ title: 'Inicio', tabBarIcon: icono('home', 'home-outline') }} />
        <Tabs.Screen name="dashboard" options={{ title: 'Dashboard', tabBarIcon: icono('stats-chart', 'stats-chart-outline') }} />
        <Tabs.Screen name="configuracion" options={{ title: 'Configuración', tabBarIcon: icono('settings', 'settings-outline') }} />
        </Tabs>
    );
}