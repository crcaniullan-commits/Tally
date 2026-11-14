import { useEffect, useState } from 'react';
import { KeyboardAvoidingView, Modal, Platform, Pressable, ScrollView, Text, TextInput, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import Icono from '../../../shared/components/Icono';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { Categoria, MAX_NOMBRE_CATEGORIA, normalizar, TipoMovimiento } from '../categorias';
import { EstiloCategoria, OPCIONES_COLOR, OPCIONES_ICONO } from '../estilosPropios';

type Props = {
  visible: boolean;
  tipo: TipoMovimiento;            // Para el título: "de ingreso" o "de egreso"
  editando: Categoria | null;      // null = nueva categoría; con valor = editar (propia o predeterminada)
  nombresExistentes: string[];     // Para evitar nombres repetidos
  onGuardar: (nombre: string, estilo: EstiloCategoria) => void;
  onCerrar: () => void;
};

// Hoja inferior del mockup "Nueva categoría de venta" (Historia 003).
// También se usa para editar el nombre, color e ícono de cualquier categoría (desde "Editar").
export default function ModalCategorias({ visible, tipo, editando, nombresExistentes, onGuardar, onCerrar }: Props) {
  const { t } = useEscala();
  const insets = useSafeAreaInsets();
  const [nombre, setNombre] = useState('');
  const [colorIdx, setColorIdx] = useState(0);
  const [icono, setIcono] = useState(OPCIONES_ICONO[0]);
  const [error, setError] = useState<string | undefined>();

  // Cada vez que se abre, parte vacío o con los datos de la categoría a editar
  useEffect(() => {
    if (!visible) return;
    setNombre(editando?.nombre ?? '');
    const idx = editando ? OPCIONES_COLOR.findIndex((c) => c.color === editando.color) : 0;
    setColorIdx(idx >= 0 ? idx : 0);
    setIcono(editando?.icono && OPCIONES_ICONO.includes(editando.icono) ? editando.icono : OPCIONES_ICONO[0]);
    setError(undefined);
  }, [visible, editando]);

  function guardar() {
    const limpio = nombre.trim();
    if (!limpio) return setError('Escribe un nombre');
    const repetido = nombresExistentes.some((n) => normalizar(n) === normalizar(limpio))
      && normalizar(limpio) !== normalizar(editando?.nombre ?? '');
    if (repetido) return setError('Ya existe una categoría con ese nombre');
    onGuardar(limpio, { icono, ...OPCIONES_COLOR[colorIdx], tipo });
  }

  const etiqueta = { fontFamily: fonts.negrita, fontSize: t(11), color: colors.texto, marginBottom: t(6) };
  const titulo = editando ? 'Editar categoría' : `Nueva categoría de ${tipo === 'ingreso' ? 'ingreso' : 'egreso'}`;

  return (
    <Modal visible={visible} transparent animationType="slide" onRequestClose={onCerrar} statusBarTranslucent>
      <KeyboardAvoidingView style={{ flex: 1 }} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
        <Pressable onPress={onCerrar} style={{ flex: 1, backgroundColor: 'rgba(60, 56, 82, 0.4)' }} />

        <View style={{
          backgroundColor: colors.superficie, borderTopLeftRadius: t(24), borderTopRightRadius: t(24),
          padding: t(20), paddingBottom: t(20) + insets.bottom, gap: t(16),
        }}>
          <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
            <Text style={{ fontFamily: fonts.extra, fontSize: t(14), color: colors.lavanda }}>{titulo}</Text>
            <Pressable onPress={onCerrar} hitSlop={10} style={{ padding: t(4), borderRadius: 999, backgroundColor: colors.lavandaSuave }}>
              <Icono nombre="x" tamano={t(16)} color={colors.lavanda} />
            </Pressable>
          </View>

          <View>
            <Text style={etiqueta}>Nombre de categoría</Text>
            <TextInput
              value={nombre}
              onChangeText={(v) => { setNombre(v); setError(undefined); }}
              maxLength={MAX_NOMBRE_CATEGORIA}
              placeholder="Ej: Repostería"
              placeholderTextColor={colors.textoSecundario}
              style={{
                paddingHorizontal: t(12), paddingVertical: t(8), fontSize: t(12), fontFamily: fonts.normal, color: colors.texto,
                backgroundColor: colors.fondoApp, borderWidth: 1, borderColor: error ? colors.error : colors.lavandaOnda, borderRadius: t(12),
              }}
            />
            {error && <Text style={{ fontFamily: fonts.medio, fontSize: t(11), color: colors.error, marginTop: t(4) }}>{error}</Text>}
          </View>

          <View>
            <Text style={etiqueta}>Elige un color</Text>
            <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: t(8) }}>
              {OPCIONES_COLOR.map((c, i) => (
                <Pressable key={c.color} onPress={() => setColorIdx(i)} style={{
                  width: t(36), height: t(36), borderRadius: 999, alignItems: 'center', justifyContent: 'center',
                  borderWidth: 2, borderColor: i === colorIdx ? colors.lavanda : 'transparent',
                }}>
                  <View style={{ width: t(28), height: t(28), borderRadius: 999, backgroundColor: c.color }} />
                </Pressable>
              ))}
            </View>
          </View>

          <View>
            <Text style={etiqueta}>Elige un ícono</Text>
            <ScrollView style={{ maxHeight: t(180) }} nestedScrollEnabled>
            <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: t(8) }}>
              {OPCIONES_ICONO.map((nombreIcono) => {
                const activo = nombreIcono === icono;
                return (
                  <Pressable key={nombreIcono} onPress={() => setIcono(nombreIcono)} style={{
                    width: '14%', flexGrow: 1, paddingVertical: t(10), borderRadius: t(12), alignItems: 'center', borderWidth: 1,
                    backgroundColor: activo ? colors.lavandaSuave : colors.fondoApp,
                    borderColor: activo ? colors.lavanda : colors.lavandaOnda,
                  }}>
                    <Icono nombre={nombreIcono} tamano={t(16)} color={activo ? OPCIONES_COLOR[colorIdx].color : colors.textoSecundario} />
                  </Pressable>
                );
              })}
            </View>
            </ScrollView>
          </View>

          <Pressable onPress={guardar} style={({ pressed }) => ({
            paddingVertical: t(10), borderRadius: t(12), alignItems: 'center', backgroundColor: colors.lavanda, opacity: pressed ? 0.85 : 1,
          })}>
            <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.superficie }}>
              {editando ? 'Guardar cambios' : 'Guardar categoría'}
            </Text>
          </Pressable>

        </View>
      </KeyboardAvoidingView>
    </Modal>
  );
}
