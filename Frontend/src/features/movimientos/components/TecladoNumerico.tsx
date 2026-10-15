    import { Ionicons } from '@expo/vector-icons';
    import { Pressable, Text, View } from 'react-native';
    import { colors } from '../../../shared/theme/colors';
    import { fonts } from '../../../shared/theme/fonts';
    import { sombras } from '../../../shared/theme/sombras';

    type TecladoNumericoProps = {
    onDigito: (digito: string) => void;
    onBorrar: () => void;
    onLimpiar: () => void;
    };

    const filas = [
    ['1', '2', '3', 'borrar'],
    ['4', '5', '6', 'limpiar'],
    ['7', '8', '9', '0'],
    ];

    export default function TecladoNumerico({ onDigito, onBorrar, onLimpiar }: TecladoNumericoProps) {
    return (
        <View style={{ gap: 10 }}>
        {filas.map((fila, i) => (
            <View key={i} style={{ flexDirection: 'row', gap: 10 }}>
            {fila.map((tecla) => {
                const esAccion = tecla === 'borrar' || tecla === 'limpiar';
                return (
                <Pressable
                    key={tecla}
                    onPress={() => (tecla === 'borrar' ? onBorrar() : tecla === 'limpiar' ? onLimpiar() : onDigito(tecla))}
                    style={({ pressed }) => ({
                    flex: 1,
                    height: 48,
                    borderRadius: 14,
                    alignItems: 'center',
                    justifyContent: 'center',
                    backgroundColor: esAccion ? colors.rosadoSuave : colors.lavandaSuave,
                    opacity: pressed ? 0.7 : 1,
                    ...sombras.suave,
                    })}
                >
                    {tecla === 'borrar' ? (
                    <Ionicons name="backspace-outline" size={22} color={colors.rosado} />
                    ) : (
                    <Text style={{ fontFamily: fonts.medio, fontSize: 20, color: esAccion ? colors.rosado : colors.lavanda }}>
                        {tecla === 'limpiar' ? 'C' : tecla}
                    </Text>
                    )}
                </Pressable>
                );
            })}
            </View>
        ))}
        </View>
    );
    }