    import { MaterialCommunityIcons } from '@expo/vector-icons';
    import { DimensionValue, Pressable, Text, View } from 'react-native';
    import { colors } from '../../../shared/theme/colors';
    import { fonts } from '../../../shared/theme/fonts';
    import { sombras } from '../../../shared/theme/sombras';

    type OpcionIconoProps = {
    nombre: string;
    icono: string;
    color: string;
    colorSuave: string;
    seleccionado: boolean;
    ancho?: DimensionValue;
    onPress: () => void;
    };

    export default function OpcionIcono({ nombre, icono, color, colorSuave, seleccionado, onPress, ancho = '23%' }: OpcionIconoProps) {
    return (
    <Pressable onPress={onPress} style={{ width: ancho, alignItems: 'center', gap: 4 }}>
        <View
            style={{
            width: '100%',
            aspectRatio: 1,
            borderRadius: 16,
            backgroundColor: colorSuave,
            alignItems: 'center',
            justifyContent: 'center',
            borderWidth: 2,
            borderColor: seleccionado ? color : 'transparent',
            ...sombras.suave,
            }}
        >
            <MaterialCommunityIcons name={icono as any} size={30} color={color} />
        </View>
        <Text numberOfLines={2} style={{ fontFamily: seleccionado ? fonts.negrita : fonts.normal, fontSize: 12, color: colors.texto, textAlign: 'center' }}>
            {nombre}
        </Text>
        </Pressable>
    );
    }