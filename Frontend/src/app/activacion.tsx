    import { router } from 'expo-router';
    import { useState } from 'react';
    import { KeyboardAvoidingView, Platform, Text, View } from 'react-native';
    import Boton from '../components/Boton';
    import CampoTexto from '../components/CampoTexto';
    import { colors } from '../constants/colors';

    const LARGO_CLAVE = 26;

    export default function Activacion() {
    const [clave, setClave] = useState('');
    const [error, setError] = useState<string | undefined>();

    function cambiarClave(texto: string) {
        // Quita espacios (comunes al copiar desde un correo) y pasa todo a mayúsculas
        setClave(texto.replace(/\s/g, '').toUpperCase());
        setError(undefined);
    }

    function activar() {
        if (!/^[A-Z2-7]+$/.test(clave) || clave.length !== LARGO_CLAVE) {
        setError(`La clave debe tener ${LARGO_CLAVE} caracteres. Revisa que la copiaste completa`);
        return;
        }
        // Pendiente: aquí se conectará con el backend (POST /v1/app/users/exchange/{clave})
        router.replace('/inicio');
    }

    return (
        <KeyboardAvoidingView
        style={{ flex: 1, backgroundColor: colors.fondo }}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        >
        <View style={{ flex: 1, justifyContent: 'center', padding: 24, gap: 16 }}>
            <Text style={{ color: colors.texto, fontSize: 24, fontWeight: 'bold', textAlign: 'center' }}>
            Activa tu cuenta
            </Text>
            <Text style={{ color: colors.textoSecundario, textAlign: 'center' }}>
            Paso 2 de 2
            </Text>

            <View style={{ backgroundColor: colors.azulSuave, borderRadius: 16, padding: 16 }}>
            <Text style={{ color: colors.texto, lineHeight: 20 }}>
                Tu municipalidad te envió una clave de activación a tu correo electrónico.
                Cópiala y pégala aquí. La clave se puede usar una sola vez.
            </Text>
            </View>

            <CampoTexto
            etiqueta="Clave de activación"
            placeholder="Pega aquí tu clave"
            value={clave}
            onChangeText={cambiarClave}
            autoCapitalize="characters"
            autoCorrect={false}
            maxLength={40}
            error={error}
            />
            <Text style={{ color: colors.textoSecundario, fontSize: 13, textAlign: 'right' }}>
            {clave.length}/{LARGO_CLAVE} caracteres
            </Text>

            <Boton titulo="Activar cuenta" onPress={activar} deshabilitado={clave === ''} />
        </View>
        </KeyboardAvoidingView>
    );
    }