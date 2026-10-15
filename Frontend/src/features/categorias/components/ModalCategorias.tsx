    import { MaterialCommunityIcons } from '@expo/vector-icons';
    import { useState } from 'react';
    import { Pressable, Text, View } from 'react-native';
    import Boton from '../../../shared/components/Boton';
    import CampoTexto from '../../../shared/components/CampoTexto';
    import ModalFlotante from '../../../shared/components/ModalFlotante';
    import { colors } from '../../../shared/theme/colors';
    import { fonts } from '../../../shared/theme/fonts';
    import { Categoria } from '../categorias';

    type ModalCategoriasProps = {
    visible: boolean;
    categorias: Categoria[];       // Categorías propias del usuario (del tipo actual)
    nombresExistentes: string[];   // Para evitar nombres repetidos
    seleccionadaId: string | null;
    onSeleccionar: (id: string) => void;
    onAgregar: (nombre: string) => void;
    onCerrar: () => void;
    };

    export default function ModalCategorias({
    visible, categorias, nombresExistentes, seleccionadaId, onSeleccionar, onAgregar, onCerrar,
    }: ModalCategoriasProps) {
    const [nombreNueva, setNombreNueva] = useState('');
    const [error, setError] = useState<string | undefined>();

    function agregar() {
        const nombre = nombreNueva.trim();
        if (nombresExistentes.some((n) => n.toLowerCase() === nombre.toLowerCase())) {
        setError('Ya existe una categoría con ese nombre');
        return;
        }
        onAgregar(nombre);
        setNombreNueva('');
        setError(undefined);
    }

    return (
        <ModalFlotante visible={visible} titulo="Mis categorías" onCerrar={onCerrar}>
        {categorias.length === 0 ? (
            <Text style={{ color: colors.textoSecundario, fontFamily: fonts.normal }}>
            Aún no tienes categorías propias. Crea la primera aquí abajo.
            </Text>
        ) : (
            <View style={{ gap: 8 }}>
            {categorias.map((c) => {
                const seleccionada = c.id === seleccionadaId;
                return (
                <Pressable
                    key={c.id}
                    onPress={() => onSeleccionar(c.id)}
                    style={{
                    flexDirection: 'row', alignItems: 'center', gap: 10, padding: 12, borderRadius: 14,
                    backgroundColor: c.colorSuave, borderWidth: 2, borderColor: seleccionada ? c.color : 'transparent',
                    }}
                >
                    <MaterialCommunityIcons name={c.icono as any} size={22} color={c.color} />
                    <Text style={{ flex: 1, color: colors.texto, fontFamily: seleccionada ? fonts.negrita : fonts.medio }}>
                    {c.nombre}
                    </Text>
                    {seleccionada && <MaterialCommunityIcons name="check" size={20} color={c.color} />}
                </Pressable>
                );
            })}
            </View>
        )}

        <CampoTexto
            etiqueta="Nueva categoría"
            placeholder="Ej: Ferias"
            value={nombreNueva}
            onChangeText={(t) => { setNombreNueva(t); setError(undefined); }}
            maxLength={100}
            error={error}
        />
        {/* Pendiente backend: guardar la categoría en la base de datos (hoy queda solo en memoria) */}
        <Boton titulo="Agregar categoría" variante="secundario" onPress={agregar} deshabilitado={nombreNueva.trim() === ''} />
        </ModalFlotante>
    );
    }