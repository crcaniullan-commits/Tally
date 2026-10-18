import { Text, View } from 'react-native';
import { colors } from '../../shared/theme/colors';
import { fonts } from '../../shared/theme/fonts';

// Pantalla provisoria: solo para ver la barra de navegación
export default function Pantalla() {
    return (
        <View style={{ flex: 1, backgroundColor: colors.fondoApp, alignItems: 'center', justifyContent: 'center', gap: 8 }}>
        <Text style={{ fontFamily: fonts.negrita, fontSize: 28, color: colors.lavanda }}>Dashboard</Text>
        <Text style={{ fontFamily: fonts.normal, fontSize: 16, color: colors.lavanda }}>Épica 006 · semanas 9-11</Text>
        </View>
    );
}