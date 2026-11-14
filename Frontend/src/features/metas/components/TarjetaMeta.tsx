import { useCallback, useEffect, useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import { useAlerta } from '../../../shared/components/Confirmar';
import Boton from '../../../shared/components/Boton';
import CampoTexto from '../../../shared/components/CampoTexto';
import Icono from '../../../shared/components/Icono';
import ModalFlotante from '../../../shared/components/ModalFlotante';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { MESES } from '../../../shared/utils/fechas';
import { formatearMonto } from '../../../shared/utils/formato';
import { traducirError } from '../../auth/errores';
import { guardarMetaMensual, obtenerMetaMensual } from '../metas';

type Props = { empresaId: string; token: string; recargar?: number };

// Meta de venta mensual con barra de progreso (Historia 007). Avance calculado por el backend.
export default function TarjetaMeta({ empresaId, token, recargar }: Props) {
  const alerta = useAlerta();
  const { t } = useEscala();
  const [meta, setMeta] = useState<number | null>(null);
  const [ventasDelMes, setVentasDelMes] = useState(0);
  const [editando, setEditando] = useState(false);
  const [guardando, setGuardando] = useState(false);
  const [texto, setTexto] = useState('');

  const cargar = useCallback(async () => {
    try {
      const datos = await obtenerMetaMensual(empresaId, token);
      setMeta(datos?.meta.monto_meta ?? null);
      setVentasDelMes(datos?.estado.monto_actual ?? 0);
    } catch {
      // Si falla, la tarjeta queda invitando a definir la meta
    }
  }, [empresaId, token]);

  useEffect(() => { cargar(); }, [cargar, recargar]);

  async function guardar() {
    const monto = Number(texto.replace(/\D/g, ''));
    if (!monto) return;
    setGuardando(true);
    try {
      await guardarMetaMensual(empresaId, monto, token);
      setEditando(false);
      await cargar();
    } catch (error) {
      alerta('No se pudo guardar la meta', traducirError(error));
    } finally {
      setGuardando(false);
    }
  }

  const mes = MESES[new Date().getMonth()];
  const porcentaje = meta ? Math.min(Math.round((ventasDelMes / meta) * 100), 100) : 0;
  const cumplida = !!meta && ventasDelMes >= meta;
  const faltan = meta ? Math.max(meta - ventasDelMes, 0) : 0;

  return (
    <View style={{ backgroundColor: colors.superficie, borderRadius: t(16), padding: t(14), gap: t(8), borderWidth: 1, borderColor: colors.lavandaOnda }}>
      <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' }}>
        <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8, flex: 1 }}>
          <Icono nombre={cumplida ? 'trophy' : 'flag'} tamano={t(16)} color={cumplida ? colors.naranjo : colors.lavanda} />
          <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.lavanda }}>Meta de {mes}</Text>
        </View>
        <Pressable hitSlop={10} onPress={() => { setTexto(meta ? String(meta) : ''); setEditando(true); }}>
          <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: colors.lavanda, backgroundColor: colors.lavandaSuave, paddingHorizontal: t(8), paddingVertical: t(2), borderRadius: t(6), overflow: 'hidden' }}>{meta ? 'Cambiar' : 'Definir'}</Text>
        </Pressable>
      </View>

      {meta ? (
        <>
          <Text style={{ fontFamily: fonts.extra, fontSize: t(14), color: colors.texto }}>
            ${formatearMonto(ventasDelMes)} de ${formatearMonto(meta)}
          </Text>
          <View style={{ height: t(10), borderRadius: 999, backgroundColor: colors.lavandaSuave, overflow: 'hidden' }}>
            <View style={{ width: `${porcentaje}%`, height: '100%', backgroundColor: cumplida ? colors.naranjo : colors.ingreso, borderRadius: 999 }} />
          </View>
          <Text style={{ fontFamily: fonts.negrita, fontSize: t(11), color: cumplida ? colors.naranjo : colors.lavanda }}>
            {cumplida ? '¡Lo lograste! Cumpliste tu meta del mes 🎉' : `${porcentaje}% · te faltan $${formatearMonto(faltan)}`}
          </Text>
        </>
      ) : (
        <Text style={{ fontFamily: fonts.medio, fontSize: t(11), color: colors.textoSecundario }}>
          Define cuánto quieres vender este mes y sigue tu avance aquí.
        </Text>
      )}

      <ModalFlotante visible={editando} titulo={`Meta de ventas de ${mes}`} onCerrar={() => setEditando(false)}>
        <View style={{ gap: 14 }}>
          <CampoTexto etiqueta="¿Cuánto quieres vender este mes?" value={texto ? formatearMonto(Number(texto)) : ''}
            onChangeText={(t) => setTexto(t.replace(/\D/g, '').slice(0, 10))} keyboardType="number-pad" placeholder="Ej: 500.000" />
          <Boton titulo={guardando ? 'Guardando...' : 'Guardar meta'} onPress={guardar} deshabilitado={!Number(texto) || guardando} />
        </View>
      </ModalFlotante>
    </View>
  );
}
