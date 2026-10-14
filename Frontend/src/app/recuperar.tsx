    import { router } from 'expo-router';
    import { useState } from 'react';
    import { Alert, KeyboardAvoidingView, Platform, Pressable, ScrollView, Text } from 'react-native';
    import Boton from '../components/Boton';
    import CampoTexto from '../components/CampoTexto';
    import { colors } from '../constants/colors';

    type Errores = { correo?: string; codigo?: string; contrasena?: string; confirmacion?: string };

    export default function Recuperar() {
    const [etapa, setEtapa] = useState<'correo' | 'nuevaContrasena'>('correo');
    const [correo, setCorreo] = useState('');
    const [codigo, setCodigo] = useState('');
    const [contrasena, setContrasena] = useState('');
    const [confirmacion, setConfirmacion] = useState('');
    const [errores, setErrores] = useState<Errores>({});

    function enviarCodigo() {
        if (!/^\S+@\S+\.\S+$/.test(correo.trim())) {
        setErrores({ correo: 'Ingresa un correo válido' });
        return;
        }
        setErrores({});
        // Pendiente backend: endpoint para solicitar código de recuperación
        setEtapa('nuevaContrasena');
    }

    function cambiarContrasena() {
        const nuevos: Errores = {};
        if (codigo.trim() === '') nuevos.codigo = 'Ingresa el código que recibiste';
        if (contrasena.length < 8 || contrasena.length > 72) {
        nuevos.contrasena = 'La contraseña debe tener entre 8 y 72 caracteres';
        }
        if (confirmacion !== contrasena) nuevos.confirmacion = 'Las contraseñas no coinciden';
        setErrores(nuevos);

        if (Object.keys(nuevos).length === 0) {
        // Pendiente backend: endpoint para confirmar código y guardar la nueva contraseña
        Alert.alert('Contraseña actualizada', 'Ya puedes iniciar sesión con tu nueva contraseña.');
        router.replace('/login');
        }
    }

    return (
        <KeyboardAvoidingView
        style={{ flex: 1, backgroundColor: colors.fondo }}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        >
        <ScrollView contentContainerStyle={{ flexGrow: 1, justifyContent: 'center', padding: 24, gap: 16 }}
            keyboardShouldPersistTaps="handled">
            <Text style={{ color: colors.texto, fontSize: 24, fontWeight: 'bold', textAlign: 'center' }}>
            Recuperar contraseña
            </Text>

            {etapa === 'correo' ? (
            <>
                <Text style={{ color: colors.textoSecundario, textAlign: 'center' }}>
                Escribe el correo con el que te registraste y te enviaremos un código de recuperación.
                </Text>
                <CampoTexto etiqueta="Correo electrónico" placeholder="tucorreo@ejemplo.cl"
                value={correo} onChangeText={setCorreo} keyboardType="email-address"
                autoCapitalize="none" error={errores.correo} />
                <Boton titulo="Enviar código" onPress={enviarCodigo} deshabilitado={correo.trim() === ''} />
            </>
            ) : (
            <>
                <Text style={{ color: colors.textoSecundario, textAlign: 'center' }}>
                Revisa tu correo <Text style={{ fontWeight: 'bold', color: colors.texto }}>{correo.trim()}</Text> e
                ingresa el código junto a tu nueva contraseña.
                </Text>
                <CampoTexto etiqueta="Código de recuperación" placeholder="Código recibido"
                value={codigo} onChangeText={setCodigo} autoCapitalize="characters"
                autoCorrect={false} error={errores.codigo} />
                <CampoTexto etiqueta="Nueva contraseña" placeholder="Mínimo 8 caracteres"
                value={contrasena} onChangeText={setContrasena} secureTextEntry error={errores.contrasena} />
                <CampoTexto etiqueta="Confirmar nueva contraseña" placeholder="Repite tu contraseña"
                value={confirmacion} onChangeText={setConfirmacion} secureTextEntry error={errores.confirmacion} />
                <Boton titulo="Cambiar contraseña" onPress={cambiarContrasena}
                deshabilitado={codigo === '' || contrasena === '' || confirmacion === ''} />
                <Pressable onPress={() => setEtapa('correo')} style={{ alignSelf: 'center' }}>
                <Text style={{ color: colors.primario, fontWeight: '600' }}>Usar otro correo</Text>
                </Pressable>
            </>
            )}

            <Pressable onPress={() => router.back()} style={{ alignSelf: 'center', marginTop: 8 }}>
            <Text style={{ color: colors.texto }}>Volver a iniciar sesión</Text>
            </Pressable>
        </ScrollView>
        </KeyboardAvoidingView>
    );
    }