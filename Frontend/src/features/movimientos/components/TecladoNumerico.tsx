import { Ionicons } from '@expo/vector-icons';
import { Pressable, Text, View } from 'react-native';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { sombras } from '../../../shared/theme/sombras';
import { useEscala } from '../../../shared/utils/escala';

type TecladoNumericoProps = {
  onDigito: (digito: string) => void;
  onBorrar: () => void;
  onLimpiar: () => void;
};

const filas = [
  ['1', '2', '3', 'borrar'],
  ['4', '5', '6', 'limpiar'],
  ['7', '8', '9', '0'],
];

export default function TecladoNumerico({ onDigito, onBorrar, onLimpiar }: TecladoNumericoProps) {
  const { e, alto } = useEscala();
  // Teclas grandes, pero sin pasar del 8% del alto de la pantalla (celulares bajos)
  const altoTecla = Math.min(e(64), alto * 0.08);
  return (
    <View style={{ gap: 12 }}>
      {filas.map((fila, i) => (
        <View key={i} style={{ flexDirection: 'row', gap: 12 }}>
          {fila.map((tecla) => {
            const esAccion = tecla === 'borrar' || tecla === 'limpiar';
            return (
              <Pressable
                key={tecla}
                onPress={() => (tecla === 'borrar' ? onBorrar() : tecla === 'limpiar' ? onLimpiar() : onDigito(tecla))}
                style={({ pressed }) => ({
                  flex: 1,
                  height: altoTecla,
                  borderRadius: 16,
                  alignItems: 'center',
                  justifyContent: 'center',
                  backgroundColor: esAccion ? colors.rosadoSuave : colors.lavandaSuave,
                  opacity: pressed ? 0.7 : 1,
                  ...sombras.suave,
                })}
              >
                {tecla === 'borrar' ? (
                  <Ionicons name="backspace-outline" size={e(30)} color={colors.rosado} />
                ) : (
                  <Text style={{ fontFamily: fonts.negrita, fontSize: e(30), color: esAccion ? colors.rosado : colors.lavanda }}>
                    {tecla === 'limpiar' ? 'C' : tecla}
                  </Text>
                )}
              </Pressable>
            );
          })}
        </View>
      ))}
    </View>
  );
}
