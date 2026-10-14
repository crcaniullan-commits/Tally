    import { Pressable, Text } from 'react-native';
    import { colors } from '../constants/colors';
    import { fonts } from '../constants/fonts';
    import { sombras } from '../constants/sombras';

    type BotonProps = {
    titulo: string;
    onPress: () => void;
    variante?: 'primario' | 'secundario';
    deshabilitado?: boolean;
    color?: string;
    };

    export default function Boton({ titulo, onPress, variante = 'primario', deshabilitado = false, color = colors.primario }: BotonProps) {
    const esPrimario = variante === 'primario';

    return (
        <Pressable
        onPress={onPress}
        disabled={deshabilitado}
        style={({ pressed }) => ({
            width: '100%',
            paddingVertical: 13,
            borderRadius: 999,
            alignItems: 'center',
            backgroundColor: esPrimario ? color : colors.superficie,
            borderWidth: esPrimario ? 0 : 2,
            borderColor: color,
            opacity: deshabilitado ? 0.4 : pressed ? 0.8 : 1,
            ...sombras.boton,
        })}
        >
        <Text style={{ color: esPrimario ? colors.superficie : color, fontSize: 16, fontFamily: fonts.medio }}>
            {titulo}
        </Text>
        </Pressable>
    );
    }