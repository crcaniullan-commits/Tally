    import { useState } from 'react';
    import { Pressable, Text } from 'react-native';
    import { colors } from '../constants/colors';
    import { fonts } from '../constants/fonts';
    import Boton from './Boton';
    import CampoTexto from './CampoTexto';
    import { datosPrueba } from '../constants/datosPrueba';

    type FormularioLoginProps = {
    onExito: () => void;
    onOlvido: () => void;
    onCrearCuenta: () => void;
    };

    export default function FormularioLogin({ onExito, onOlvido, onCrearCuenta }: FormularioLoginProps) {
    const [correo, setCorreo] = useState(datosPrueba.correo);
    const [contrasena, setContrasena] = useState(datosPrueba.contrasena);
    const [errores, setErrores] = useState<{ correo?: string; contrasena?: string }>({});

    const formularioIncompleto = correo.trim() === '' || contrasena === '';

    function iniciarSesion() {
        const nuevos: { correo?: string; contrasena?: string } = {};
        if (!/^\S+@\S+\.\S+$/.test(correo.trim())) nuevos.correo = 'Ingresa un correo válido';
        if (contrasena.length < 8) nuevos.contrasena = 'La contraseña debe tener al menos 8 caracteres';
        setErrores(nuevos);

        if (Object.keys(nuevos).length === 0) {
        // Pendiente: conexión con el backend (POST /v1/auth/login)
        onExito();
        }
    }

    return (
        <>
        <CampoTexto etiqueta="Correo electrónico" placeholder="tucorreo@ejemplo.cl"
            value={correo} onChangeText={setCorreo} keyboardType="email-address"
            autoCapitalize="none" error={errores.correo} />
        <CampoTexto etiqueta="Contraseña" placeholder="Mínimo 8 caracteres"
            value={contrasena} onChangeText={setContrasena} secureTextEntry error={errores.contrasena} />

        <Pressable onPress={onOlvido} style={{ alignSelf: 'flex-end' }}>
            <Text style={{ color: colors.lavanda, fontFamily: fonts.medio }}>¿Olvidaste tu contraseña?</Text>
        </Pressable>

        <Boton titulo="Ingresar" onPress={iniciarSesion} deshabilitado={formularioIncompleto} />

        <Pressable onPress={onCrearCuenta} style={{ alignSelf: 'center' }}>
            <Text style={{ color: colors.texto, fontFamily: fonts.normal }}>
            ¿No tienes cuenta? <Text style={{ color: colors.lavanda, fontFamily: fonts.negrita }}>Crear cuenta</Text>
            </Text>
        </Pressable>
        </>
    );
    }