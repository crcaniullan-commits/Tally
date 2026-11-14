import React from 'react';
import { ActivityIndicator, Pressable, Text, View } from 'react-native';
import Icono from '../../../shared/components/Icono';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
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
  metodo: string; // Vacío en los egresos: el backend no guarda su método de pago
  metodoId: string | null;
  nota: string;
  fecha: Date;
};

const MESES = ['ene', 'feb', 'mar', 'abr', 'may', 'jun', 'jul', 'ago', 'sep', 'oct', 'nov', 'dic'];

// Como el mockup: "Hoy, 20:15", "Ayer, 18:45" o "12 nov, 10:30"
// (se arma a mano porque Android no siempre trae los formatos en español)
function formatearFecha(f: Date) {
  const hora = `${f.getHours()}:${String(f.getMinutes()).padStart(2, '0')}`;
  const hoy = new Date();
  const ayer = new Date(hoy.getFullYear(), hoy.getMonth(), hoy.getDate() - 1);
  const mismoDia = (a: Date, b: Date) => a.toDateString() === b.toDateString();
  if (mismoDia(f, hoy)) return `Hoy, ${hora}`;
  if (mismoDia(f, ayer)) return `Ayer, ${hora}`;
  return `${f.getDate()} ${MESES[f.getMonth()]}${f.getFullYear() !== hoy.getFullYear() ? ` ${f.getFullYear()}` : ''}, ${hora}`;
}

const DIAS = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'];
const fechaLarga = (f: Date) =>
  `${DIAS[f.getDay()]} ${f.getDate()} ${MESES[f.getMonth()]} ${f.getFullYear()} · ${f.getHours()}:${String(f.getMinutes()).padStart(2, '0')}`;

type Props = {
  movimientos: Movimiento[];
  cargando?: boolean;
  titulo?: string;
  detallada?: boolean;                   // false = abreviada (categoría y monto); true = tarjeta con todo
  derechaTitulo?: React.ReactNode;       // Ej: selector abreviada / detallada
  onEditar?: (m: Movimiento) => void;    // Lápiz de cada tarjeta
};

// Lista de movimientos de Registros: vista abreviada o detallada, con lápiz para editar
export default function HistorialMovimientos({ movimientos, cargando = false, titulo = 'Movimientos', detallada = false, derechaTitulo, onEditar }: Props) {
  const { t } = useEscala();

  const lapiz = (m: Movimiento) => onEditar && (
    <Pressable hitSlop={10} onPress={() => onEditar(m)} style={{ padding: t(4), borderRadius: t(6), backgroundColor: colors.lavandaSuave }}>
      <Icono nombre="pencil" tamano={t(11)} color={colors.lavanda} />
    </Pressable>
  );

  const fila = (etiqueta: string, valor: string) => (
    <View style={{ flexDirection: 'row', gap: t(6) }}>
      <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: colors.textoSecundario, width: t(92) }}>{etiqueta}</Text>
      <Text style={{ fontFamily: fonts.medio, fontSize: t(11), color: colors.texto, flex: 1 }}>{valor}</Text>
    </View>
  );

  return (
    <View style={{ gap: t(8), paddingTop: t(4) }}>
      <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
        <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.lavanda }}>{titulo}</Text>
        {derechaTitulo}
      </View>

      {cargando ? (
        <ActivityIndicator color={colors.lavanda} style={{ paddingVertical: t(16) }} />
      ) : movimientos.length === 0 ? (
        <Text style={{ fontFamily: fonts.medio, fontSize: t(12), color: colors.textoSecundario, textAlign: 'center', paddingVertical: t(16) }}>
          No hay movimientos para mostrar.
        </Text>
      ) : movimientos.map((m) => {
        const esIngreso = m.tipo === 'ingreso';
        const color = esIngreso ? colors.ingreso : colors.egreso;
        const monto = `${esIngreso ? '+' : '-'}$${formatearMonto(m.monto)}`;
        const tarjeta = {
          backgroundColor: colors.superficie, borderRadius: t(12), borderWidth: 1, borderColor: colors.lavandaOnda,
          boxShadow: '0px 1px 2px rgba(0, 0, 0, 0.05)',
        };

        if (!detallada) {
          return (
            <View key={`${m.tipo}-${m.id}`} style={[tarjeta, { flexDirection: 'row', alignItems: 'center', gap: t(10), paddingVertical: t(9), paddingHorizontal: t(12) }]}>
              <View style={{ width: t(28), height: t(28), borderRadius: t(8), backgroundColor: m.colorSuave, alignItems: 'center', justifyContent: 'center' }}>
                <Icono nombre={m.icono} tamano={t(14)} color={m.color} />
              </View>
              <View style={{ flex: 1 }}>
                <Text numberOfLines={1} style={{ fontFamily: fonts.negrita, fontSize: t(12), color: colors.texto }}>{m.categoria}</Text>
                <Text numberOfLines={1} style={{ fontFamily: fonts.normal, fontSize: t(10), color: colors.textoSecundario }}>
                  {[formatearFecha(m.fecha), m.nota].filter(Boolean).join(' • ')}
                </Text>
              </View>
              <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color }}>{monto}</Text>
              {lapiz(m)}
            </View>
          );
        }

        return (
          <View key={`${m.tipo}-${m.id}`} style={[tarjeta, { padding: t(12), gap: t(8) }]}>
            <View style={{ flexDirection: 'row', alignItems: 'center', gap: t(10) }}>
              <View style={{ width: t(36), height: t(36), borderRadius: t(10), backgroundColor: m.colorSuave, alignItems: 'center', justifyContent: 'center' }}>
                <Icono nombre={m.icono} tamano={t(18)} color={m.color} />
              </View>
              <View style={{ flex: 1 }}>
                <Text style={{ fontFamily: fonts.medio, fontSize: t(10), color }}>{esIngreso ? 'Ingreso' : 'Egreso'}</Text>
                <Text numberOfLines={1} style={{ fontFamily: fonts.extra, fontSize: t(14), color: colors.texto }}>{m.categoria}</Text>
              </View>
              <Text style={{ fontFamily: fonts.extra, fontSize: t(16), color }}>{monto}</Text>
              {lapiz(m)}
            </View>
            <View style={{ gap: t(3), paddingTop: t(8), borderTopWidth: 1, borderTopColor: colors.lavandaSuave }}>
              {fila('Fecha', fechaLarga(m.fecha))}
              {esIngreso && fila('Método de pago', m.metodo || '—')}
              {fila('Comentario', m.nota || 'Sin comentario')}
            </View>
          </View>
        );
      })}
    </View>
  );
}
