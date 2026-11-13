import { MaterialCommunityIcons } from '@expo/vector-icons';
import { Pressable, ScrollView, Text, View } from 'react-native';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { sombras } from '../../../shared/theme/sombras';
import { useEscala } from '../../../shared/utils/escala';

// Cada tarjeta del carrusel
export type OpcionCarrusel = {
  id: string;
  nombre: string;
  icono: string;      // Nombre del ícono (MaterialCommunityIcons)
  color: string;      // Color del ícono y del borde al seleccionar
  colorSuave: string; // Fondo del ícono
};

type CarruselOpcionesProps = {
  titulo: string;
  opciones: OpcionCarrusel[];
  seleccionadaId: string | null;
  onSeleccionar: (id: string) => void;
};


export default function CarruselOpciones({ titulo, opciones, seleccionadaId, onSeleccionar }: CarruselOpcionesProps) {
  const { e, ancho } = useEscala();
  // Caben ~3 tarjetas y un pedazo de la cuarta: así el usuario nota que puede deslizar
  const anchoTarjeta = Math.min(Math.max((ancho - 40 - 24) / 3.3, 92), 130);
  return (
    <View style={{ gap: 10 }}>
      <Text style={{ fontFamily: fonts.negrita, fontSize: e(22), color: colors.lavanda }}>{titulo}</Text>

      {/* Fila deslizable. Los márgenes negativos dejan que las tarjetas lleguen al borde de la pantalla */}
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        style={{ marginHorizontal: -20 }}
        contentContainerStyle={{ paddingHorizontal: 20, paddingVertical: 4, gap: 12 }}
      >
        {opciones.map((o) => {
          const seleccionada = seleccionadaId === o.id;
          return (
            <Pressable
              key={o.id}
              onPress={() => onSeleccionar(o.id)}
              style={{
                width: anchoTarjeta,
                padding: 8,
                gap: 6,
                alignItems: 'center',
                backgroundColor: colors.superficie,
                borderRadius: 18,
                borderWidth: 2,
                borderColor: seleccionada ? o.color : 'transparent',
                ...sombras.suave,
              }}
            >
              <View style={{ width: '100%', height: e(56), borderRadius: 14, backgroundColor: o.colorSuave, alignItems: 'center', justifyContent: 'center' }}>
                <MaterialCommunityIcons name={o.icono as any} size={e(32)} color={o.color} />
              </View>
              <Text
                numberOfLines={1}
                adjustsFontSizeToFit
                style={{ fontFamily: seleccionada ? fonts.negrita : fonts.medio, fontSize: e(15), color: colors.lavanda }}
              >
                {o.nombre}
              </Text>
            </Pressable>
          );
        })}
      </ScrollView>
    </View>
  );
}
