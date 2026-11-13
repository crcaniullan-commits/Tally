import { Ionicons, MaterialCommunityIcons } from '@expo/vector-icons';
import { useState } from 'react';
import { Alert, Pressable, Text, TextInput, View } from 'react-native';
import Boton from '../../../shared/components/Boton';
import CampoTexto from '../../../shared/components/CampoTexto';
import ModalFlotante from '../../../shared/components/ModalFlotante';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { Categoria, MAX_NOMBRE_CATEGORIA, normalizar } from '../categorias';

type ModalCategoriasProps = {
  visible: boolean;
  categorias: Categoria[];       // Categorías propias del usuario
  nombresExistentes: string[];   // Para evitar nombres repetidos
  seleccionadaId: string | null;
  onSeleccionar: (id: string) => void;
  onAgregar: (nombre: string) => void;
  onRenombrar: (categoria: Categoria, nombre: string) => void;
  onEliminar: (categoria: Categoria) => void;
  onCerrar: () => void;
};

// Mis categorías (Historia 003): elegir, crear, renombrar y eliminar categorías propias
export default function ModalCategorias({
  visible, categorias, nombresExistentes, seleccionadaId, onSeleccionar, onAgregar, onRenombrar, onEliminar, onCerrar,
}: ModalCategoriasProps) {
  const { e } = useEscala();
  const [nombreNueva, setNombreNueva] = useState('');
  const [error, setError] = useState<string | undefined>();
  const [editandoId, setEditandoId] = useState<string | null>(null);
  const [textoEdicion, setTextoEdicion] = useState('');

  const repetido = (nombre: string, exceptoId?: string) =>
    nombresExistentes.some((n) => normalizar(n) === normalizar(nombre)) &&
    !categorias.some((c) => c.id === exceptoId && normalizar(c.nombre) === normalizar(nombre));

  function agregar() {
    const nombre = nombreNueva.trim();
    if (repetido(nombre)) {
      setError('Ya existe una categoría con ese nombre');
      return;
    }
    onAgregar(nombre);
    setNombreNueva('');
    setError(undefined);
  }

  function guardarEdicion(c: Categoria) {
    const nombre = textoEdicion.trim();
    if (!nombre || nombre === c.nombre) return setEditandoId(null);
    if (repetido(nombre, c.id)) return Alert.alert('Nombre repetido', 'Ya existe una categoría con ese nombre.');
    onRenombrar(c, nombre);
    setEditandoId(null);
  }

  function confirmarEliminar(c: Categoria) {
    Alert.alert('Eliminar categoría', `¿Eliminar "${c.nombre}"? Los movimientos ya registrados no se borran.`, [
      { text: 'Cancelar', style: 'cancel' },
      { text: 'Eliminar', style: 'destructive', onPress: () => onEliminar(c) },
    ]);
  }

  return (
    <ModalFlotante visible={visible} titulo="Mis categorías" onCerrar={onCerrar}>
      {categorias.length === 0 ? (
        <Text style={{ color: colors.lavanda, fontFamily: fonts.normal, fontSize: e(15) }}>
          Aún no tienes categorías propias. Crea la primera aquí abajo.
        </Text>
      ) : (
        <View style={{ gap: 8 }}>
          {categorias.map((c) => {
            const seleccionada = c.id === seleccionadaId;
            const editando = c.id === editandoId;
            return (
              <View
                key={c.id}
                style={{
                  flexDirection: 'row', alignItems: 'center', gap: 10, padding: 12, borderRadius: 14,
                  backgroundColor: c.colorSuave, borderWidth: 2, borderColor: seleccionada ? c.color : 'transparent',
                }}
              >
                <MaterialCommunityIcons name={c.icono as any} size={e(22)} color={c.color} />
                {editando ? (
                  <TextInput
                    value={textoEdicion} onChangeText={setTextoEdicion} autoFocus maxLength={MAX_NOMBRE_CATEGORIA}
                    onSubmitEditing={() => guardarEdicion(c)}
                    style={{ flex: 1, fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda, padding: 0 }}
                  />
                ) : (
                  <Pressable style={{ flex: 1 }} onPress={() => onSeleccionar(c.id)}>
                    <Text style={{ color: colors.lavanda, fontSize: e(16), fontFamily: seleccionada ? fonts.negrita : fonts.medio }}>
                      {c.nombre}
                    </Text>
                  </Pressable>
                )}
                {editando ? (
                  <Pressable hitSlop={8} onPress={() => guardarEdicion(c)}>
                    <Ionicons name="checkmark-circle" size={e(26)} color={colors.verde} />
                  </Pressable>
                ) : (
                  <>
                    <Pressable hitSlop={8} onPress={() => { setEditandoId(c.id); setTextoEdicion(c.nombre); }}>
                      <Ionicons name="pencil" size={e(20)} color={colors.lavanda} />
                    </Pressable>
                    <Pressable hitSlop={8} onPress={() => confirmarEliminar(c)}>
                      <Ionicons name="trash-outline" size={e(20)} color={colors.rosado} />
                    </Pressable>
                  </>
                )}
              </View>
            );
          })}
        </View>
      )}

      <View style={{ gap: 10, marginTop: 16 }}>
        <CampoTexto
          etiqueta="Nueva categoría"
          placeholder="Ej: Publicidad"
          value={nombreNueva}
          onChangeText={(t) => { setNombreNueva(t); setError(undefined); }}
          maxLength={MAX_NOMBRE_CATEGORIA}
          error={error}
        />
        <Boton titulo="Agregar categoría" onPress={agregar} deshabilitado={!nombreNueva.trim()} />
      </View>
    </ModalFlotante>
  );
}
