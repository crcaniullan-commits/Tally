import { useEffect, useState } from 'react';
import { Pressable, Text, TextInput, View } from 'react-native';
import { useAlerta } from '../../../shared/components/Confirmar';
import Calendario from '../../../shared/components/Calendario';
import Icono from '../../../shared/components/Icono';
import ModalFlotante from '../../../shared/components/ModalFlotante';
import { useAviso } from '../../../shared/components/Toast';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { aDMA, aISO, desdeISO, hoy } from '../../../shared/utils/fechas';
import { traducirError } from '../../auth/errores';
import { useSesion } from '../../auth/SesionContext';
import { actualizarPendiente, agregarPendiente, marcarPagado, MAX_CLIENTE, Pendiente } from '../../pendientes/pendientes';

type Props = { pendiente: Pendiente | null; onCerrar: () => void; onCambio: () => void };

// Editar un pago pendiente: cliente, monto, fecha (con calendario) y comentario
export default function ModalPendiente({ pendiente, onCerrar, onCambio }: Props) {
  const alerta = useAlerta();
  const { token, empresa, manejarError } = useSesion();
  const { t } = useEscala();
  const avisar = useAviso();
  const [cliente, setCliente] = useState('');
  const [monto, setMonto] = useState('');
  const [vence, setVence] = useState(hoy());
  const [nota, setNota] = useState('');
  const [calendario, setCalendario] = useState(false);
  const [guardando, setGuardando] = useState(false);

  useEffect(() => {
    if (!pendiente) return;
    setCliente(pendiente.cliente);
    setMonto(String(pendiente.monto));
    setVence(desdeISO(pendiente.vence));
    setNota(pendiente.nota);
    setCalendario(false);
  }, [pendiente]);

  if (!pendiente) return null;
  const montoNumero = Number(monto);
  const incompleto = !cliente.trim() || !montoNumero;

  async function guardar() {
    if (!token || !empresa || !pendiente || incompleto) return;
    setGuardando(true);
    try {
      if (nota.trim() !== pendiente.nota.trim()) {
        // El backend aún no permite editar el comentario: se crea el pago corregido y se quita el original
        await agregarPendiente(empresa.id, token, { cliente, monto: montoNumero, vence: aISO(vence), nota });
        await marcarPagado(empresa.id, token, pendiente);
      } else {
        await actualizarPendiente(empresa.id, token, pendiente.id, { cliente, monto: montoNumero, vence: aISO(vence) });
      }
      avisar('¡Pago pendiente actualizado!', 'clock');
      onCambio();
      onCerrar();
    } catch (error) {
      manejarError(error);
      alerta('No se pudo guardar', traducirError(error));
    } finally {
      setGuardando(false);
    }
  }

  const etiqueta = { fontFamily: fonts.negrita, fontSize: t(11), color: colors.lavanda };
  const campo = {
    paddingHorizontal: t(12), paddingVertical: t(8), fontSize: t(12), fontFamily: fonts.medio, color: colors.texto,
    backgroundColor: colors.fondoApp, borderWidth: 1, borderColor: colors.lavandaOnda, borderRadius: t(12),
  };
  return (
    <ModalFlotante visible titulo="Editar pago pendiente" onCerrar={onCerrar}>
      <View style={{ gap: t(10) }}>
        <Text style={etiqueta}>¿Quién te debe?</Text>
        <TextInput value={cliente} onChangeText={setCliente} maxLength={MAX_CLIENTE} style={campo} />

        <Text style={etiqueta}>Monto</Text>
        <TextInput value={monto ? `$ ${montoNumero.toLocaleString('es-CL')}` : ''} keyboardType="number-pad"
          onChangeText={(v) => setMonto(v.replace(/\D/g, '').slice(0, 10))}
          style={[campo, { fontFamily: fonts.extra, fontSize: t(18), color: colors.naranjo }]} />

        <Text style={etiqueta}>¿Cuándo te paga?</Text>
        <Pressable onPress={() => setCalendario(!calendario)} style={[campo, { flexDirection: 'row', alignItems: 'center', gap: t(8) }]}>
          <Icono nombre="calendar" tamano={t(14)} color={colors.naranjo} />
          <Text style={{ flex: 1, fontFamily: fonts.extra, fontSize: t(13), color: colors.texto }}>{aDMA(vence)}</Text>
          <Icono nombre="chevron-down" tamano={t(14)} color={colors.textoSecundario} />
        </Pressable>
        {calendario && <Calendario fecha={vence} onElegir={(f) => { setVence(f); setCalendario(false); }} />}

        <Text style={etiqueta}>Comentario</Text>
        <TextInput value={nota} onChangeText={setNota} maxLength={100} multiline placeholder="Ej: paga con transferencia"
          placeholderTextColor={colors.textoSecundario} style={[campo, { textAlignVertical: 'top' }]} />

        <Pressable onPress={guardar} disabled={incompleto || guardando} style={({ pressed }) => ({
          flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: t(6), paddingVertical: t(10), borderRadius: t(12),
          backgroundColor: colors.naranjo, opacity: incompleto || guardando ? 0.5 : pressed ? 0.9 : 1, marginTop: t(4),
        })}>
          <Icono nombre="check-circle-2" tamano={t(16)} color={colors.superficie} />
          <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.superficie }}>{guardando ? 'Guardando...' : 'Guardar cambios'}</Text>
        </Pressable>
      </View>
    </ModalFlotante>
  );
}
