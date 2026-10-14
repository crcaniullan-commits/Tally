    import { ReactNode } from 'react';
    import { KeyboardAvoidingView, Modal, Platform, Pressable, ScrollView, Text, View } from 'react-native';
    import { colors } from '../theme/colors';
    import { fonts } from '../theme/fonts';
    import { sombras } from '../theme/sombras';

    type ModalFlotanteProps = {
    visible: boolean;
    titulo: string;
    onCerrar: () => void;
    children: ReactNode;
    };

    export default function ModalFlotante({ visible, titulo, onCerrar, children }: ModalFlotanteProps) {
    return (
        <Modal visible={visible} transparent animationType="fade" onRequestClose={onCerrar} statusBarTranslucent>
        <KeyboardAvoidingView style={{ flex: 1 }} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
            <View style={{ flex: 1, justifyContent: 'center', padding: 20 }}>
            {/* Fondo oscurecido: al tocarlo se cierra el modal */}
            <Pressable
                onPress={onCerrar}
                style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(30, 22, 51, 0.45)' }}
            />

            {/* Tarjeta flotante */}
            <View style={{ backgroundColor: colors.superficie, borderRadius: 28, maxHeight: '90%', ...sombras.flotante }}>
                <ScrollView contentContainerStyle={{ padding: 22, gap: 12 }} keyboardShouldPersistTaps="handled">
                <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Text style={{ color: colors.lavanda, fontFamily: fonts.titulo, fontSize: 28, letterSpacing: 1 }}>
                    {titulo}
                    </Text>
                    <Pressable onPress={onCerrar} hitSlop={12}>
                    <Text style={{ color: colors.textoSecundario, fontSize: 22 }}>✕</Text>
                    </Pressable>
                </View>
                {children}
                </ScrollView>
            </View>
            </View>
        </KeyboardAvoidingView>
        </Modal>
    );
    }