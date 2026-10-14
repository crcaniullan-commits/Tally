import { useState } from 'react';
import { Text, View } from 'react-native';
import { colors } from '../../../shared/theme/colors';
import { datosPrueba } from '../../../shared/mocks/datosPrueba';
import { fonts } from '../../../shared/theme/fonts';
import BarraPasos from '../../../shared/components/BarraPasos';
import Boton from '../../../shared/components/Boton';
import CampoTexto from '../../../shared/components/CampoTexto';

const LARGO_CLAVE = 26;

type FormularioActivacionProps = {
    onExito: () => void;
    onVolver: () => void;
};

export default function FormularioActivacion({ onExito, onVolver }: FormularioActivacionProps) {    const [clave, setClave] = useState(datosPrueba.claveActivacion);
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
        // Pendiente: conexión con el backend (POST /v1/app/users/exchange/{clave})
        onExito();
    }

    return (
        <>
        <BarraPasos actual={2} total={2} descripcion="activa tu cuenta" />

        <View style={{ backgroundColor: colors.azulSuave, borderRadius: 14, padding: 14 }}>
            <Text style={{ color: colors.texto, fontFamily: fonts.normal, lineHeight: 20 }}>
            Tu municipalidad te envió una clave de activación a tu correo. Cópiala y pégala aquí.
            La clave se puede usar una sola vez.
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
        <Text style={{ color: colors.textoSecundario, fontFamily: fonts.normal, fontSize: 13, textAlign: 'right', marginTop: -6 }}>
            {clave.length}/{LARGO_CLAVE} caracteres
        </Text>

        <Boton titulo="Activar cuenta" onPress={activar} deshabilitado={clave === ''} />
        <Boton titulo="Volver al paso 1" variante="secundario" onPress={onVolver} />
        </>
    );
}