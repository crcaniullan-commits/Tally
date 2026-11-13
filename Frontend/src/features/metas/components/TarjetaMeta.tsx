import { MaterialCommunityIcons } from '@expo/vector-icons';
import { useEffect, useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import Boton from '../../../shared/components/Boton';
import CampoTexto from '../../../shared/components/CampoTexto';
import ModalFlotante from '../../../shared/components/ModalFlotante';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { sombras } from '../../../shared/theme/sombras';
import { useEscala } from '../../../shared/utils/escala';
import { MESES } from '../../../shared/utils/fechas';
import { formatearMonto } from '../../../shared/utils/formato';
import { guardarMetaDelMes, obtenerMetaDelMes } from '../metas';

type Props = { empresaId: string; ventasDelMes: number };

// Meta de venta mensual con barra de progreso (Historia 007)
export default function TarjetaMeta({ empresaId, ventasDelMes }: Props) {
  const { e } = useEscala();
  const [meta, setMeta] = useState<number | null>(null);
  const [editando, setEditando] = useState(false);
  const [texto, setTexto] = useState('');

  useEffect(() => { obtenerMetaDelMes(empresaId).then(setMeta); }, [empresaId]);

  async function guardar() {
    const monto = Number(texto.replace(/\D/g, ''));
    if (!monto) return;
    await guardarMetaDelMes(empresaId, monto);
    setMeta(monto);
    setEditando(false);
  }

  const mes = MESES[new Date().getMonth()];
  const porcentaje = meta ? Math.min(Math.round((ventasDelMes / meta) * 100), 100) : 0;
  const cumplida = !!meta && ventasDelMes >= meta;
  const faltan = meta ? Math.max(meta - ventasDelMes, 0) : 0;

  return (
    <View style={{ backgroundColor: colors.superficie, borderRadius: 20, padding: 16, gap: 12, ...sombras.suave }}>
      <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8, flex: 1 }}>
          <MaterialCommunityIcons name={cumplida ? 'trophy-outline' : 'flag-checkered'} size={e(26)} color={cumplida ? colors.naranjo : colors.lavanda} />
          <Text style={{ fontFamily: fonts.negrita, fontSize: e(20), color: colors.lavanda }}>Meta de {mes}</Text>
        </View>
        <Pressable hitSlop={10} onPress={() => { setTexto(meta ? String(meta) : ''); setEditando(true); }}>
          <Text style={{ fontFamily: fonts.medio, fontSize: e(15), color: colors.lavanda }}>{meta ? 'Cambiar' : 'Definir'}</Text>
        </Pressable>
      </View>

      {meta ? (
        <>
          <Text style={{ fontFamily: fonts.medio, fontSize: e(16), color: colors.lavanda }}>
            ${formatearMonto(ventasDelMes)} de ${formatearMonto(meta)}
          </Text>
          <View style={{ height: e(18), borderRadius: 9, backgroundColor: colors.lavandaSuave, overflow: 'hidden' }}>
            <View style={{ width: `${porcentaje}%`, height: '100%', backgroundColor: cumplida ? colors.naranjo : colors.ingreso, borderRadius: 9 }} />
          </View>
          <Text style={{ fontFamily: fonts.negrita, fontSize: e(15), color: cumplida ? colors.naranjo : colors.lavanda }}>
            {cumplida ? '¡Lo lograste! Cumpliste tu meta del mes 🎉' : `${porcentaje}% · te faltan $${formatearMonto(faltan)}`}
          </Text>
        </>
      ) : (
        <Text style={{ fontFamily: fonts.normal, fontSize: e(15), color: colors.lavanda }}>
          Define cuánto quieres vender este mes y sigue tu avance aquí.
        </Text>
      )}

      <ModalFlotante visible={editando} titulo={`Meta de ventas de ${mes}`} onCerrar={() => setEditando(false)}>
        <View style={{ gap: 14 }}>
          <CampoTexto etiqueta="¿Cuánto quieres vender este mes?" value={texto ? formatearMonto(Number(texto)) : ''}
            onChangeText={(t) => setTexto(t.replace(/\D/g, '').slice(0, 10))} keyboardType="number-pad" placeholder="Ej: 500.000" />
          <Boton titulo="Guardar meta" onPress={guardar} deshabilitado={!Number(texto)} />
        </View>
      </ModalFlotante>
    </View>
  );
}
