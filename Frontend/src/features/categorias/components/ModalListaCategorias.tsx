import { Pressable, ScrollView, Text, View } from 'react-native';
import { useConfirmar } from '../../../shared/components/Confirmar';
import Icono from '../../../shared/components/Icono';
import ModalFlotante from '../../../shared/components/ModalFlotante';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { Categoria, TipoMovimiento } from '../categorias';

type Props = {
  visible: boolean;
  tipo: TipoMovimiento;
  categorias: Categoria[];               // Ya filtradas por tipo
  onEditar: (c: Categoria) => void;
  onEliminar: (c: Categoria) => void;    // Revisa antes si tiene registros
  onCerrar: () => void;
};

// Recuadro flotante "Editar categorías" (Historia 003): todas las categorías del tipo elegido.
// Todas se pueden editar (nombre, color e ícono); solo las propias sin registros se pueden eliminar.
export default function ModalListaCategorias({ visible, tipo, categorias, onEditar, onEliminar, onCerrar }: Props) {
  const { t } = useEscala();
  const pedirConfirmacion = useConfirmar();

  async function confirmar(c: Categoria) {
    if (await pedirConfirmacion({ titulo: 'Eliminar categoría', mensaje: `¿Eliminar "${c.nombre}"? Solo se puede si no tiene registros.` })) onEliminar(c);
  }

  return (
    <ModalFlotante visible={visible} titulo={`Categorías de ${tipo === 'ingreso' ? 'ingresos' : 'egresos'}`} onCerrar={onCerrar}>
      <ScrollView style={{ maxHeight: t(380) }}>
        <View style={{ gap: t(8) }}>
          {categorias.map((c) => (
            <View key={c.clave} style={{
              flexDirection: 'row', alignItems: 'center', gap: t(10), padding: t(10), borderRadius: t(12),
              borderWidth: 1, borderColor: colors.lavandaOnda, backgroundColor: colors.fondoApp,
            }}>
              <View style={{ width: t(30), height: t(30), borderRadius: t(8), backgroundColor: c.colorSuave, alignItems: 'center', justifyContent: 'center' }}>
                <Icono nombre={c.icono} tamano={t(15)} color={c.color} />
              </View>
              <View style={{ flex: 1 }}>
                <Text numberOfLines={1} style={{ fontFamily: fonts.negrita, fontSize: t(12), color: colors.texto }}>{c.nombre}</Text>
                <Text style={{ fontFamily: fonts.medio, fontSize: t(10), color: colors.textoSecundario }}>{c.propia ? 'Creada por ti' : 'Predeterminada'}</Text>
              </View>
              <Pressable hitSlop={8} onPress={() => onEditar(c)} style={{ padding: t(6), borderRadius: t(8), backgroundColor: colors.lavandaSuave }}>
                <Icono nombre="pencil" tamano={t(14)} color={colors.lavanda} />
              </Pressable>
              {c.propia && (
                <Pressable hitSlop={8} onPress={() => confirmar(c)} style={{ padding: t(6), borderRadius: t(8), backgroundColor: colors.rosadoSuave }}>
                  <Icono nombre="trash" tamano={t(14)} color={colors.rosado} />
                </Pressable>
              )}
            </View>
          ))}
          <Text style={{ fontFamily: fonts.normal, fontSize: t(10), color: colors.textoSecundario, textAlign: 'center' }}>
            Las categorías con registros no se pueden eliminar, pero sí renombrar.
          </Text>
        </View>
      </ScrollView>
    </ModalFlotante>
  );
}
