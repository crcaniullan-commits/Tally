    import { MaterialCommunityIcons } from '@expo/vector-icons';
    import { Pressable, Text } from 'react-native';
    import { colors } from '../constants/colors';
    import { fonts } from '../constants/fonts';
    import { sombras } from '../constants/sombras';
    import { Categoria } from '../constants/categorias';

    type TarjetaCategoriaProps = {
    categoria: Categoria;
    onPress: () => void;
    };

    export default function TarjetaCategoria({ categoria, onPress }: TarjetaCategoriaProps) {
    return (
        <Pressable
        onPress={onPress}
        style={({ pressed }) => ({
            width: '48%',
            aspectRatio: 1.3,
            backgroundColor: categoria.color,
            borderRadius: 20,
            padding: 14,
            justifyContent: 'space-between',
            opacity: pressed ? 0.85 : 1,
            ...sombras.boton,
        })}
        >
        <MaterialCommunityIcons name={categoria.icono as any} size={34} color={colors.superficie} />
        <Text style={{ color: colors.superficie, fontFamily: fonts.medio, fontSize: 15 }}>
            {categoria.nombre}
        </Text>
        </Pressable>
    );
    }