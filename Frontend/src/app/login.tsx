import { router } from 'expo-router';
import { useState } from 'react';
import { Image, Pressable, Text } from 'react-native';
import Boton from '../components/Boton';
import CampoTexto from '../components/CampoTexto';
import FormularioFlotante from '../components/FormularioFlotante';
import { colors } from '../constants/colors';
import { fonts } from '../constants/fonts';

export default function Login() {
    const [correo, setCorreo] = useState('');
    const [contrasena, setContrasena] = useState('');
    const [errores, setErrores] = useState<{ correo?: string; contrasena?: string }>({});

    const formularioIncompleto = correo.trim() === '' || contrasena === '';

    function iniciarSesion() {
        const nuevosErrores: { correo?: string; contrasena?: string } = {};

        if (!/^\S+@\S+\.\S+$/.test(correo.trim())) {
        nuevosErrores.correo = 'Ingresa un correo válido';
        }
        if (contrasena.length < 8) {
        nuevosErrores.contrasena = 'La contraseña debe tener al menos 8 caracteres';
        }

        setErrores(nuevosErrores);

        if (Object.keys(nuevosErrores).length === 0) {
        // Pendiente: aquí se conectará con el backend (POST /v1/auth/login)
        router.replace('/inicio');
        }
    }

    return (
        <FormularioFlotante titulo="Iniciar sesión" colorTitulo={colors.azul}>
        <Image
            source={require('../../assets/images/logo-tally.png')}
            style={{ width: 110, height: 70, alignSelf: 'center' }}
            resizeMode="contain"
        />

        <CampoTexto etiqueta="Correo electrónico" placeholder="tucorreo@ejemplo.cl"
            value={correo} onChangeText={setCorreo} keyboardType="email-address"
            autoCapitalize="none" error={errores.correo} />
        <CampoTexto etiqueta="Contraseña" placeholder="Mínimo 8 caracteres"
            value={contrasena} onChangeText={setContrasena} secureTextEntry error={errores.contrasena} />

        <Pressable onPress={() => router.push('/recuperar')} style={{ alignSelf: 'flex-end' }}>
            <Text style={{ color: colors.azul, fontFamily: fonts.medio }}>¿Olvidaste tu contraseña?</Text>
        </Pressable>

        <Boton titulo="Ingresar" color={colors.azul} onPress={iniciarSesion} deshabilitado={formularioIncompleto} />

        <Pressable onPress={() => router.push('/registro')} style={{ alignSelf: 'center' }}>
            <Text style={{ color: colors.texto, fontFamily: fonts.normal }}>
            ¿No tienes cuenta? <Text style={{ color: colors.azul, fontFamily: fonts.negrita }}>Crear cuenta</Text>
            </Text>
        </Pressable>
        </FormularioFlotante>
    );
}