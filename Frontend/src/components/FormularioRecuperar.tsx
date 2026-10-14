    import { useState } from 'react';
    import { Pressable, Text } from 'react-native';
    import { colors } from '../constants/colors';
    import { datosPrueba } from '../constants/datosPrueba';
    import { fonts } from '../constants/fonts';
    import Boton from './Boton';
    import CampoTexto from './CampoTexto';

    type Errores = { correo?: string; codigo?: string; contrasena?: string; confirmacion?: string };

    type FormularioRecuperarProps = {
    onExito: () => void;
    onVolver: () => void;
    };

    export default function FormularioRecuperar({ onExito, onVolver }: FormularioRecuperarProps) {
    const [etapa, setEtapa] = useState<'correo' | 'nuevaContrasena'>('correo');
    const [correo, setCorreo] = useState(datosPrueba.correo);
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
        if (contrasena.length < 8 || contrasena.length > 72) nuevos.contrasena = 'Debe tener entre 8 y 72 caracteres';
        if (confirmacion !== contrasena) nuevos.confirmacion = 'Las contraseñas no coinciden';
        setErrores(nuevos);

        if (Object.keys(nuevos).length === 0) {
        // Pendiente backend: endpoint para confirmar código y guardar la nueva contraseña
        onExito();
        }
    }

    if (etapa === 'correo') {
        return (
        <>
            <Text style={{ color: colors.textoSecundario, fontFamily: fonts.normal }}>
            Escribe el correo con el que te registraste y te enviaremos un código de recuperación.
            </Text>
            <CampoTexto etiqueta="Correo electrónico" placeholder="tucorreo@ejemplo.cl"
            value={correo} onChangeText={setCorreo} keyboardType="email-address"
            autoCapitalize="none" error={errores.correo} />
            <Boton titulo="Enviar código" onPress={enviarCodigo} deshabilitado={correo.trim() === ''} />
            <Boton titulo="Volver a iniciar sesión" variante="secundario" onPress={onVolver} />
        </>
        );
    }

    return (
        <>
        <Text style={{ color: colors.textoSecundario, fontFamily: fonts.normal }}>
            Revisa tu correo <Text style={{ fontFamily: fonts.negrita, color: colors.texto }}>{correo.trim()}</Text> e
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
            <Text style={{ color: colors.lavanda, fontFamily: fonts.medio }}>Usar otro correo</Text>
        </Pressable>
        </>
    );
    }