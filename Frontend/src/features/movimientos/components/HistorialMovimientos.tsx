import { MaterialCommunityIcons } from '@expo/vector-icons';
import { ActivityIndicator, Pressable, Text, View } from 'react-native';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { sombras } from '../../../shared/theme/sombras';
import { useEscala } from '../../../shared/utils/escala';
import { formatearMonto } from '../../../shared/utils/formato';
import { TipoMovimiento } from '../../categorias/categorias';

// Un movimiento ya registrado (ingreso o egreso), listo para mostrar
export type Movimiento = {
  id: string;
  tipo: TipoMovimiento;
  monto: number;
  categoria: string;
  categoriaId: string | null;
  icono: string;
  color: string;
  colorSuave: string;
  metodo: string; // Vacío en los gastos: el backend aún no guarda su método de pago
  metodoId: string | null;
  nota: string;
  fecha: Date;
};

const MESES = ['ene', 'feb', 'mar', 'abr', 'may', 'jun', 'jul', 'ago', 'sep', 'oct', 'nov', 'dic'];

// Ej: "18 oct · 18:45" (se arma a mano porque Android no siempre trae los formatos en español)
function formatearFecha(f: Date) {
  const hora = `${f.getHours()}:${String(f.getMinutes()).padStart(2, '0')}`;
  return `${f.getDate()} ${MESES[f.getMonth()]} · ${hora}`;
}

type Props = {
  movimientos: Movimiento[];
  cargando?: boolean;
  titulo?: string;
  limite?: number;                       // Cuántos mostrar (sin límite = todos)
  onVerTodos?: () => void;               // Muestra el enlace "Ver todos ›"
  onPresionar?: (m: Movimiento) => void; // Tocar un movimiento: editar o eliminar
};

export default function HistorialMovimientos({ movimientos, cargando = false, titulo = 'Últimos movimientos', limite, onVerTodos, onPresionar }: Props) {
  const { e } = useEscala();
  const visibles = limite ? movimientos.slice(0, limite) : movimientos;

  return (
    <View style={{ gap: 12 }}>
      <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
        <Text style={{ fontFamily: fonts.negrita, fontSize: e(22), color: colors.lavanda }}>{titulo}</Text>
        {onVerTodos && (
          <Pressable hitSlop={10} onPress={onVerTodos}>
            <Text style={{ fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda }}>Ver todos ›</Text>
          </Pressable>
        )}
      </View>

      {cargando ? (
        <ActivityIndicator color={colors.lavanda} style={{ paddingVertical: 16 }} />
      ) : movimientos.length === 0 ? (
        <Text style={{ fontFamily: fonts.normal, fontSize: e(16), color: colors.lavanda, textAlign: 'center', paddingVertical: 16 }}>
          Aún no tienes movimientos registrados.
        </Text>
      ) : (
        <>
          {visibles.map((m) => {
            const esIngreso = m.tipo === 'ingreso';
            const detalle = [m.categoria, m.metodo, formatearFecha(m.fecha)].filter(Boolean).join(' · ');
            return (
              <Pressable
                key={`${m.tipo}-${m.id}`}
                onPress={onPresionar ? () => onPresionar(m) : undefined}
                style={({ pressed }) => ({
                  flexDirection: 'row', alignItems: 'center', gap: 14, backgroundColor: colors.superficie,
                  borderRadius: 18, padding: 14, opacity: pressed ? 0.7 : 1, ...sombras.suave,
                })}
              >
                <View style={{ width: e(52), height: e(52), borderRadius: 14, backgroundColor: m.colorSuave, alignItems: 'center', justifyContent: 'center' }}>
                  <MaterialCommunityIcons name={m.icono as any} size={e(28)} color={m.color} />
                </View>

                <View style={{ flex: 1 }}>
                  <Text numberOfLines={1} style={{ fontFamily: fonts.medio, fontSize: e(17), color: colors.lavanda }}>
                    {m.nota || m.categoria}
                  </Text>
                  <Text numberOfLines={1} style={{ fontFamily: fonts.normal, fontSize: e(14), color: colors.lavanda, opacity: 0.75 }}>
                    {detalle}
                  </Text>
                </View>

                <Text style={{ fontFamily: fonts.negrita, fontSize: e(18), color: esIngreso ? colors.ingreso : colors.egreso }}>
                  {esIngreso ? '+' : '−'}${formatearMonto(m.monto)}
                </Text>
              </Pressable>
            );
          })}
          {onPresionar && (
            <Text style={{ fontFamily: fonts.normal, fontSize: e(13), color: colors.lavanda, opacity: 0.7, textAlign: 'center' }}>
              Toca un movimiento para editarlo o eliminarlo
            </Text>
          )}
        </>
      )}
    </View>
  );
}
