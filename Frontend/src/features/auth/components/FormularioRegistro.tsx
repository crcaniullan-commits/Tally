    import { useState } from 'react';
    import { Pressable, Text } from 'react-native';
    import { colors } from '../../../shared/theme/colors';
    import { fonts } from '../../../shared/theme/fonts';
    import { rutEsValido } from '../../../shared/utils/rut';
    import Boton from '../../../shared/components/Boton';
    import CampoTexto from '../../../shared/components/CampoTexto';
    import { datosPrueba } from '../../../shared/mocks/datosPrueba';
    import BarraPasos from '../../../shared/components/BarraPasos';

    type Errores = {
    nombre?: string; rut?: string; emprendimiento?: string;
    correo?: string; contrasena?: string; confirmacion?: string;
    };

    export type DatosRegistro = {
        nombre: string;
        rut: string;
        emprendimiento: string;
        correo: string;
        contrasena: string;
        };

        type FormularioRegistroProps = {
        valoresIniciales?: DatosRegistro;
        onExito: (datos: DatosRegistro) => void;
        onIrALogin: () => void;
        };

        export default function FormularioRegistro({ valoresIniciales, onExito, onIrALogin }: FormularioRegistroProps) {
        const inicial = valoresIniciales ?? datosPrueba;

        const [nombre, setNombre] = useState(inicial.nombre);
        const [rut, setRut] = useState(inicial.rut);
        const [emprendimiento, setEmprendimiento] = useState(inicial.emprendimiento);
        const [correo, setCorreo] = useState(inicial.correo);
        const [contrasena, setContrasena] = useState(inicial.contrasena);
        const [confirmacion, setConfirmacion] = useState(inicial.contrasena);
        const [errores, setErrores] = useState<Errores>({});

    const formularioIncompleto = [nombre, rut, emprendimiento, correo, contrasena, confirmacion]
        .some((campo) => campo.trim() === '');

    function registrarse() {
        const nuevos: Errores = {};
        if (nombre.trim().length > 100) nuevos.nombre = 'Máximo 100 caracteres';
        if (!rutEsValido(rut)) nuevos.rut = 'RUT inválido. Revisa el número y el dígito verificador';
        if (emprendimiento.trim().length > 100) nuevos.emprendimiento = 'Máximo 100 caracteres';
        if (!/^\S+@\S+\.\S+$/.test(correo.trim())) nuevos.correo = 'Ingresa un correo válido';
        if (contrasena.length < 8 || contrasena.length > 72) nuevos.contrasena = 'Debe tener entre 8 y 72 caracteres';
        if (confirmacion !== contrasena) nuevos.confirmacion = 'Las contraseñas no coinciden';
        setErrores(nuevos);

        if (Object.keys(nuevos).length === 0) {
        // Pendiente: conexión con el backend (POST /v1/auth)
            onExito({ nombre, rut, emprendimiento, correo, contrasena });
        }
    }

    return (
        <>
        <BarraPasos actual={1} total={2} descripcion="tus datos" />

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

        <Pressable onPress={onIrALogin} style={{ alignSelf: 'center' }}>
            <Text style={{ color: colors.texto, fontFamily: fonts.normal }}>
            ¿Ya tienes cuenta? <Text style={{ color: colors.lavanda, fontFamily: fonts.negrita }}>Inicia sesión</Text>
            </Text>
        </Pressable>
        </>
    );
    }