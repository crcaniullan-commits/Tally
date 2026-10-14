    import { router } from 'expo-router';
    import { useState } from 'react';
    import { KeyboardAvoidingView, Platform, Pressable, ScrollView, Text } from 'react-native';
    import Boton from '../components/Boton';
    import CampoTexto from '../components/CampoTexto';
    import { colors } from '../constants/colors';
    import { rutEsValido } from '../utils/rut';

    type Errores = {
    nombre?: string;
    rut?: string;
    emprendimiento?: string;
    correo?: string;
    contrasena?: string;
    confirmacion?: string;
    };

    export default function Registro() {
    const [nombre, setNombre] = useState('');
    const [rut, setRut] = useState('');
    const [emprendimiento, setEmprendimiento] = useState('');
    const [correo, setCorreo] = useState('');
    const [contrasena, setContrasena] = useState('');
    const [confirmacion, setConfirmacion] = useState('');
    const [errores, setErrores] = useState<Errores>({});

    const formularioIncompleto = [nombre, rut, emprendimiento, correo, contrasena, confirmacion].some(
        (campo) => campo.trim() === ''
    );

    function registrarse() {
        const nuevos: Errores = {};

        if (nombre.trim().length > 100) nuevos.nombre = 'Máximo 100 caracteres';
        if (!rutEsValido(rut)) nuevos.rut = 'RUT inválido. Revisa el número y el dígito verificador';
        if (emprendimiento.trim().length > 100) nuevos.emprendimiento = 'Máximo 100 caracteres';
        if (!/^\S+@\S+\.\S+$/.test(correo.trim())) nuevos.correo = 'Ingresa un correo válido';
        if (contrasena.length < 8 || contrasena.length > 72) {
        nuevos.contrasena = 'La contraseña debe tener entre 8 y 72 caracteres';
        }
        if (confirmacion !== contrasena) nuevos.confirmacion = 'Las contraseñas no coinciden';

        setErrores(nuevos);

        if (Object.keys(nuevos).length === 0) {
        // Pendiente: aquí se conectará con el backend (POST /v1/auth)
        router.replace('/activacion');
        }
    }

    return (
        <KeyboardAvoidingView
        style={{ flex: 1, backgroundColor: colors.fondo }}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        >
        <ScrollView contentContainerStyle={{ padding: 24, gap: 16, paddingTop: 48 }} keyboardShouldPersistTaps="handled">
            <Text style={{ color: colors.texto, fontSize: 24, fontWeight: 'bold', textAlign: 'center' }}>
            Crear cuenta
            </Text>
            <Text style={{ color: colors.textoSecundario, textAlign: 'center', marginBottom: 8 }}>
            Paso 1 de 2: tus datos y los de tu emprendimiento
            </Text>

            <CampoTexto etiqueta="Nombre completo" placeholder="Ej: Juan Pérez"
            value={nombre} onChangeText={setNombre} error={errores.nombre} />
            <CampoTexto etiqueta="RUT" placeholder="Ej: 12.345.678-5"
            value={rut} onChangeText={setRut} autoCapitalize="characters" maxLength={12} error={errores.rut} />
            <CampoTexto etiqueta="Nombre de tu emprendimiento" placeholder="Ej: Dulces Juanita"
            value={emprendimiento} onChangeText={setEmprendimiento} error={errores.emprendimiento} />
            <CampoTexto etiqueta="Correo electrónico" placeholder="tucorreo@ejemplo.cl"
            value={correo} onChangeText={setCorreo} keyboardType="email-address" autoCapitalize="none" error={errores.correo} />
            <CampoTexto etiqueta="Contraseña" placeholder="Mínimo 8 caracteres"
            value={contrasena} onChangeText={setContrasena} secureTextEntry error={errores.contrasena} />
            <CampoTexto etiqueta="Confirmar contraseña" placeholder="Repite tu contraseña"
            value={confirmacion} onChangeText={setConfirmacion} secureTextEntry error={errores.confirmacion} />

            <Boton titulo="Continuar" onPress={registrarse} deshabilitado={formularioIncompleto} />

            <Pressable onPress={() => router.replace('/login')} style={{ alignSelf: 'center', marginTop: 8 }}>
            <Text style={{ color: colors.texto }}>
                ¿Ya tienes cuenta? <Text style={{ color: colors.primario, fontWeight: 'bold' }}>Inicia sesión</Text>
            </Text>
            </Pressable>
        </ScrollView>
        </KeyboardAvoidingView>
    );
    }