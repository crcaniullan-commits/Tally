import { Pressable, Text, View } from 'react-native';
import Icono from '../../../shared/components/Icono';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';

const MAX_DIGITOS = 8; // Hasta $99.999.999
// Teclado del mockup: 4 columnas, borrar y limpiar a la derecha
const TECLAS = ['1', '2', '3', 'borrar', '4', '5', '6', 'C', '7', '8', '9', '0'];

// Calcula el nuevo monto según la tecla presionada
export function aplicarTecla(monto: number, tecla: string): number {
  if (tecla === 'borrar') return Math.floor(monto / 10);
  if (tecla === 'C') return 0;
  const nuevo = `${monto === 0 ? '' : monto}${tecla}`;
  return nuevo.length <= MAX_DIGITOS ? Number(nuevo) : monto;
}

type Props = {
  monto: number;
  onCambiar: (monto: number) => void;
  textoBoton: string;
  colorBoton: string;
  deshabilitado: boolean; // Faltan datos obligatorios o se está guardando
  onRegistrar: () => void;
};

// Zona del pulgar: teclado numérico (crece para llenar la pantalla) y botón Registrar
export default function Calculadora({ monto, onCambiar, textoBoton, colorBoton, deshabilitado, onRegistrar }: Props) {
  const { t } = useEscala();
  return (
    <>
      <View style={{ flex: 1, gap: t(6) }}>
        {[0, 1, 2].map((fila) => (
          <View key={fila} style={{ flex: 1, flexDirection: 'row', gap: t(6) }}>
            {TECLAS.slice(fila * 4, fila * 4 + 4).map((tecla) => {
              const especial = tecla === 'borrar' || tecla === 'C';
              return (
                <Pressable key={tecla} onPress={() => onCambiar(aplicarTecla(monto, tecla))} style={({ pressed }) => ({
                  flex: 1, minHeight: t(42), maxHeight: t(70), borderRadius: t(12), alignItems: 'center', justifyContent: 'center',
                  backgroundColor: especial ? colors.rosadoSuave : pressed ? colors.lavandaOnda : colors.lavandaSuave,
                  transform: [{ scale: pressed ? 0.95 : 1 }],
                  boxShadow: pressed ? 'none' : '0px 2px 3px rgba(135, 120, 205, 0.28)',
                })}>
                  {tecla === 'borrar' ? (
                    <Icono nombre="delete" tamano={t(20)} color={colors.rosado} />
                  ) : (
                    <Text style={{ fontFamily: tecla === 'C' ? fonts.extra : fonts.negrita, fontSize: tecla === 'C' ? t(14) : t(20), color: tecla === 'C' ? colors.rosado : colors.lavanda }}>
                      {tecla}
                    </Text>
                  )}
                </Pressable>
              );
            })}
          </View>
        ))}
      </View>

      {/* Deshabilitado mientras falten datos obligatorios (Definición de Terminado, Historia 004) */}
      <Pressable onPress={onRegistrar} disabled={deshabilitado} style={({ pressed }) => ({
        flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: t(6), paddingVertical: t(10), borderRadius: t(12),
        backgroundColor: colorBoton, opacity: deshabilitado ? 0.4 : pressed ? 0.92 : 1,
        boxShadow: deshabilitado ? 'none' : '0px 1px 2px rgba(0, 0, 0, 0.08)',
      })}>
        <Icono nombre="check-circle-2" tamano={t(16)} color={colors.superficie} />
        <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.superficie }}>{textoBoton}</Text>
      </Pressable>
    </>
  );
}
