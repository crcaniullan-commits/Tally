    import { MaterialCommunityIcons } from '@expo/vector-icons';
    import { Pressable, ScrollView, Text, View } from 'react-native';
    import { colors } from '../../../shared/theme/colors';
    import { fonts } from '../../../shared/theme/fonts';
    import { sombras } from '../../../shared/theme/sombras';

    // Cada tarjeta del carrusel
    export type OpcionCarrusel = {
    id: string;
    nombre: string;
    icono: string;      // Nombre del ícono (MaterialCommunityIcons)
    color: string;      // Color del ícono y del borde al seleccionar
    colorSuave: string; // Fondo del ícono
    };

    type CarruselOpcionesProps = {
    titulo: string;
    opciones: OpcionCarrusel[];
    seleccionadaId: string | null;
    onSeleccionar: (id: string) => void;
    };

    const ANCHO_TARJETA = 96;

    export default function CarruselOpciones({ titulo, opciones, seleccionadaId, onSeleccionar }: CarruselOpcionesProps) {
    return (
        <View style={{ gap: 10 }}>
        <Text style={{ fontFamily: fonts.negrita, fontSize: 19, color: colors.lavanda }}>{titulo}</Text>

        {/* Fila deslizable. Los márgenes negativos dejan que las tarjetas lleguen al borde de la pantalla */}
        <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            style={{ marginHorizontal: -20 }}
            contentContainerStyle={{ paddingHorizontal: 20, paddingVertical: 4, gap: 12 }}
        >
            {opciones.map((o) => {
            const seleccionada = seleccionadaId === o.id;
            return (
                <Pressable
                key={o.id}
                onPress={() => onSeleccionar(o.id)}
                style={{
                    width: ANCHO_TARJETA,
                    padding: 8,
                    gap: 6,
                    alignItems: 'center',
                    backgroundColor: colors.superficie,
                    borderRadius: 18,
                    borderWidth: 2,
                    borderColor: seleccionada ? o.color : 'transparent',
                    ...sombras.suave,
                }}
                >
                <View style={{ width: '100%', height: 54, borderRadius: 14, backgroundColor: o.colorSuave, alignItems: 'center', justifyContent: 'center' }}>
                    <MaterialCommunityIcons name={o.icono as any} size={30} color={o.color} />
                </View>
                <Text
                    numberOfLines={1}
                    style={{ fontFamily: seleccionada ? fonts.negrita : fonts.medio, fontSize: 13, color: colors.lavanda }}
                >
                    {o.nombre}
                </Text>
                </Pressable>
            );
            })}
        </ScrollView>
        </View>
    );
    }