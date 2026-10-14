    import { Ionicons } from '@expo/vector-icons';
    import { useState } from 'react';
    import { Pressable, Text, TextInput, TextInputProps, View } from 'react-native';
    import { colors } from '../constants/colors';
    import { fonts } from '../constants/fonts';
    import { sombras } from '../constants/sombras';

    type CampoTextoProps = TextInputProps & {
    etiqueta: string;
    error?: string;
    };

    export default function CampoTexto({ etiqueta, error, secureTextEntry, ...props }: CampoTextoProps) {
    const esContrasena = !!secureTextEntry;
    const [mostrarContrasena, setMostrarContrasena] = useState(false);

    return (
        <View style={{ width: '100%', gap: 4 }}>
        <Text style={{ color: colors.texto, fontFamily: fonts.medio, fontSize: 14 }}>{etiqueta}</Text>

        <View style={{ justifyContent: 'center' }}>
            <TextInput
            {...props}
            secureTextEntry={esContrasena && !mostrarContrasena}
            placeholderTextColor={colors.textoSecundario}
            style={{
                backgroundColor: colors.superficie,
                borderRadius: 12,
                paddingLeft: 14,
                paddingRight: esContrasena ? 46 : 14,
                paddingVertical: 10,
                fontSize: 15,
                fontFamily: fonts.normal,
                color: colors.texto,
                borderWidth: 1.5,
                borderColor: error ? colors.error : colors.lavandaSuave,
                ...sombras.suave,
            }}
            />

            {esContrasena && (
            <Pressable
                onPress={() => setMostrarContrasena(!mostrarContrasena)}
                hitSlop={10}
                style={{ position: 'absolute', right: 12 }}
            >
                <Ionicons
                name={mostrarContrasena ? 'eye-off-outline' : 'eye-outline'}
                size={22}
                color={colors.textoSecundario}
                />
            </Pressable>
            )}
        </View>

        {error ? <Text style={{ color: colors.error, fontSize: 12, fontFamily: fonts.normal }}>{error}</Text> : null}
        </View>
    );
    }