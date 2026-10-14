    import { Ionicons } from '@expo/vector-icons';
    import { Pressable, ScrollView, Text, View } from 'react-native';
    import TarjetaCategoria from '../features/categorias/components/TarjetaCategoria';
    import { categoriasPredeterminadas } from '../features/categorias/categorias';
    import { colors } from '../shared/theme/colors';
    import { datosPrueba } from '../shared/mocks/datosPrueba';
    import { fonts } from '../shared/theme/fonts';

    export default function Inicio() {
    // Pendiente: el nombre vendrá del usuario con sesión iniciada (backend)
    const primerNombre = datosPrueba.nombre.split(' ')[0] || 'emprendedor';

    return (
        <ScrollView
        style={{ flex: 1, backgroundColor: colors.fondo }}
        contentContainerStyle={{ paddingHorizontal: 20, paddingTop: 56, paddingBottom: 32, gap: 20 }}
        >
        {/* Encabezado: saludo y notificaciones */}
        <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
            <View>
            <Text style={{ color: colors.textoSecundario, fontFamily: fonts.normal, fontSize: 15 }}>
                ¡Hola, {primerNombre}!
            </Text>
            <Text style={{ color: colors.lavanda, fontFamily: fonts.titulo, fontSize: 32, letterSpacing: 1 }}>
                Inicio rápido
            </Text>
            </View>
            <Pressable hitSlop={10}>
            <Ionicons name="notifications-outline" size={26} color={colors.texto} />
            </Pressable>
        </View>

        {/* Categorías */}
        <View style={{ flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'space-between', rowGap: 14 }}>
            {categoriasPredeterminadas.map((categoria) => (
            <TarjetaCategoria
                key={categoria.id}
                categoria={categoria}
                onPress={() => {
                // Próximo paso: abrir el registro rápido con esta categoría
                }}
            />
            ))}
        </View>
        </ScrollView>
    );
    }