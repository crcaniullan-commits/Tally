import { useEffect, useState } from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import Icono from '../../../shared/components/Icono';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { Categoria, TipoMovimiento } from '../categorias';
import { EstiloCategoria } from '../estilosPropios';
import { useCategorias } from '../useCategorias';
import ChipCategoria from './ChipCategoria';
import GrillaOrdenable from './GrillaOrdenable';
import ModalCategorias from './ModalCategorias';
import ModalListaCategorias from './ModalListaCategorias';

type Props = {
  tipo: TipoMovimiento;
  categorias: ReturnType<typeof useCategorias>;  // Datos y acciones (vienen de useCategorias)
  categoriaId: string | null;
  onElegir: (id: string | null) => void;
  onArrastrando: (activo: boolean) => void;     // Para bloquear el scroll mientras se arrastra un tag
};

// Bloque "Categoría" de Inicio: flecha (carrusel / todas), Editar, + Nueva, los tags y sus ventanas
export default function SelectorCategorias({ tipo, categorias: cats, categoriaId, onElegir, onArrastrando }: Props) {
  const { t } = useEscala();
  const [expandida, setExpandida] = useState(false);
  const [modalAbierto, setModalAbierto] = useState(false);
  const [listaAbierta, setListaAbierta] = useState(false);
  const [editando, setEditando] = useState<Categoria | null>(null);
  const delTipo = cats.delTipo(tipo);

  // La preferencia de Configuración decide si parte en carrusel o en grilla
  useEffect(() => { setExpandida(cats.vistaGrilla); }, [cats.vistaGrilla]);

  const guardarCategoria = async (nombre: string, estilo: EstiloCategoria) => {
    const guardada = await cats.guardar(editando, nombre, estilo, tipo);
    if (!guardada) return;
    if (!editando) onElegir(guardada.id); // La categoría recién creada queda marcada
    setModalAbierto(false);
  };

  const eliminarCategoria = async (c: Categoria) => {
    const eliminada = await cats.eliminar(c);
    if (eliminada && categoriaId === c.id) onElegir(delTipo.find((x) => x.id !== c.id)?.id ?? null);
  };

  const boton = (icono: string, texto: string, accion: () => void) => (
    <Pressable onPress={accion} hitSlop={6} style={{
      flexDirection: 'row', alignItems: 'center', gap: t(3), backgroundColor: colors.lavandaSuave,
      paddingHorizontal: t(10), paddingVertical: t(4), borderRadius: t(8),
    }}>
      <Icono nombre={icono} tamano={t(11)} color={colors.lavanda} />
      <Text style={{ fontFamily: fonts.negrita, fontSize: t(10), color: colors.lavanda }}>{texto}</Text>
    </Pressable>
  );
  const chip = (c: Categoria) => <ChipCategoria key={c.clave} categoria={c} activa={c.id === categoriaId} onPress={() => onElegir(c.id)} />;

  return (
    <View style={{ gap: t(10) }}>
      <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
        {/* Flecha: muestra todas las categorías o vuelve al carrusel */}
        <Pressable onPress={() => setExpandida(!expandida)} hitSlop={8} style={{ flexDirection: 'row', alignItems: 'center', gap: t(4) }}>
          <View style={{ transform: [{ rotate: expandida ? '180deg' : '0deg' }] }}>
            <Icono nombre="chevron-down" tamano={t(16)} color={colors.lavanda} />
          </View>
          <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: colors.lavanda }}>Categoría</Text>
        </Pressable>
        <View style={{ flexDirection: 'row', gap: t(8) }}>
          {boton('pencil', 'Editar', () => setListaAbierta(true))}
          {boton('plus', 'Nueva', () => { setEditando(null); setModalAbierto(true); })}
        </View>
      </View>

      {expandida ? (
        <>
          <GrillaOrdenable
            items={delTipo}
            clave={(c) => c.clave}
            renderItem={chip}
            onArrastrando={onArrastrando}
            onReordenar={(claves) => cats.reordenar(tipo, claves)}
          />
          <Text style={{ fontFamily: fonts.medio, fontSize: t(10), color: colors.textoSecundario, textAlign: 'center' }}>
            Mantén presionada una categoría y arrástrala para cambiarla de lugar
          </Text>
        </>
      ) : (
        <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={{ gap: t(6) }}>
          {delTipo.map(chip)}
        </ScrollView>
      )}

      <ModalCategorias
        visible={modalAbierto}
        tipo={tipo}
        editando={editando}
        nombresExistentes={cats.categorias.map((c) => c.nombre)}
        onGuardar={guardarCategoria}
        onCerrar={() => setModalAbierto(false)}
      />
      <ModalListaCategorias
        visible={listaAbierta}
        tipo={tipo}
        categorias={delTipo}
        onEditar={(c) => { setListaAbierta(false); setEditando(c); setModalAbierto(true); }}
        onEliminar={eliminarCategoria}
        onCerrar={() => setListaAbierta(false)}
      />
    </View>
  );
}
