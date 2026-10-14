    import { ReactNode } from 'react';
    import { KeyboardAvoidingView, Platform, ScrollView, Text, View } from 'react-native';
    import { colors } from '../constants/colors';
    import { fonts } from '../constants/fonts';
    import { sombras } from '../constants/sombras';

    type FormularioFlotanteProps = {
    titulo: string;
    subtitulo?: string;
    colorTitulo?: string;
    children: ReactNode;
    };

    export default function FormularioFlotante({ titulo, subtitulo, colorTitulo = colors.lavanda, children }: FormularioFlotanteProps) {
    return (
        <KeyboardAvoidingView
        style={{ flex: 1, backgroundColor: colors.fondo }}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        >
        <ScrollView
            contentContainerStyle={{ flexGrow: 1, justifyContent: 'center', padding: 20 }}
            keyboardShouldPersistTaps="handled"
        >
            <View
            style={{
                backgroundColor: colors.superficie,
                borderRadius: 28,
                padding: 22,
                gap: 12,
                ...sombras.flotante,
            }}
            >
            <Text style={{ color: colorTitulo, fontFamily: fonts.titulo, fontSize: 38, textAlign: 'center', letterSpacing: 1 }}>
                {titulo}
            </Text>
            {subtitulo ? (
                <Text style={{ color: colors.textoSecundario, fontFamily: fonts.normal, textAlign: 'center', marginTop: -6 }}>
                {subtitulo}
                </Text>
            ) : null}
            {children}
            </View>
        </ScrollView>
        </KeyboardAvoidingView>
    );
    }