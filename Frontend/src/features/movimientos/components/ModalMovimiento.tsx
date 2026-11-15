import { useEffect, useState } from 'react';
import { Pressable, ScrollView, Text, TextInput, View } from 'react-native';
import Icono from '../../../shared/components/Icono';
import ModalFlotante from '../../../shared/components/ModalFlotante';
import { useAviso } from '../../../shared/components/Toast';
import { useConfirmar, useAlerta } from '../../../shared/components/Confirmar';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { traducirError } from '../../auth/errores';
import { useSesion } from '../../auth/SesionContext';
import { Categoria } from '../../categorias/categorias';
import { obtenerIdReal } from '../../categorias/categorias.service';
import { MetodoPago, metodosPago } from '../metodosPago';
import { actualizarMovimiento, eliminarMovimiento } from '../movimientos.service';
import { Movimiento } from './HistorialMovimientos';

type Props = {
  movimiento: Movimiento | null; // null = modal cerrado
  categorias: Categoria[];
  onCerrar: () => void;
  onCambio: () => void; // Se llama después de editar o eliminar, para recargar la lista
};

// Ver, editar o eliminar un movimiento (Historia 012)
export default function ModalMovimiento({ movimiento, categorias, onCerrar, onCambio }: Props) {
  const alerta = useAlerta();
  const { token, empresa, manejarError } = useSesion();
  const { t } = useEscala();
  const avisar = useAviso();
  const confirmar = useConfirmar();
  const [monto, setMonto] = useState('');
  const [categoriaId, setCategoriaId] = useState<string | null>(null);
  const [metodo, setMetodo] = useState<MetodoPago | null>(null);
  const [nota, setNota] = useState('');
  const [guardando, setGuardando] = useState(false);

  // Al abrir, el formulario se llena con los datos del movimiento
  useEffect(() => {
    if (!movimiento) return;
    setMonto(String(movimiento.monto));
    setCategoriaId(movimiento.categoriaId);
    setMetodo((movimiento.metodoId as MetodoPago) ?? null);
    setNota(movimiento.nota);
  }, [movimiento]);

  if (!movimiento) return null;
  const esIngreso = movimiento.tipo === 'ingreso';
  const opciones = categorias.filter((c) => c.tipo === movimiento.tipo || c.tipo === 'ambos');
  const montoNumero = Number(monto.replace(/\D/g, ''));
  const incompleto = !montoNumero || !categoriaId || (esIngreso && !metodo);

  async function guardar() {
    const categoria = categorias.find((c) => c.id === categoriaId);
    if (!token || !empresa || !movimiento || !categoria) return;
    setGuardando(true);
    try {
      const idReal = await obtenerIdReal(categoria, empresa.id, token);
      await actualizarMovimiento(movimiento, { monto: montoNumero, metodo, categoriaId: idReal, nota }, empresa.id, token);
      avisar(esIngreso ? '¡Ingreso actualizado!' : '¡Egreso actualizado!');
      onCambio();
      onCerrar();
    } catch (error) {
      manejarError(error);
      alerta('No se pudo guardar', traducirError(error));
    } finally {
      setGuardando(false);
    }
  }

  async function eliminar() {
    const ok = await confirmar({ titulo: 'Eliminar movimiento', mensaje: '¿Seguro que quieres eliminarlo? No se puede deshacer.' });
    if (!ok || !token || !empresa || !movimiento) return;
    try {
      await eliminarMovimiento(movimiento.tipo, movimiento.id, empresa.id, token);
      avisar('Movimiento eliminado', 'trash');
      onCambio();
      onCerrar();
    } catch (error) {
      manejarError(error);
      alerta('No se pudo eliminar', traducirError(error));
    }
  }

  // ---- Estilos del mockup ----
  const etiqueta = { fontFamily: fonts.negrita, fontSize: t(11), color: colors.lavanda };
  const campo = {
    paddingHorizontal: t(12), paddingVertical: t(8), fontSize: t(12), fontFamily: fonts.medio, color: colors.texto,
    backgroundColor: colors.fondoApp, borderWidth: 1, borderColor: colors.lavandaOnda, borderRadius: t(12),
  };
  const chip = (clave: string, texto: string, activo: boolean, accion: () => void, icono?: string, colorIcono?: string) => (
    <Pressable key={clave} onPress={accion} style={{
      flexDirection: 'row', alignItems: 'center', gap: t(4), paddingHorizontal: t(10), paddingVertical: t(6), borderRadius: t(10),
      borderWidth: activo ? 2 : 1, borderColor: activo ? colors.lavanda : colors.lavandaOnda, margin: activo ? 0 : 1,
      backgroundColor: activo ? colors.lavandaSuave : colors.superficie,
    }}>
      {icono && <Icono nombre={icono} tamano={t(13)} color={colorIcono ?? colors.lavanda} />}
      <Text style={{ fontFamily: activo ? fonts.negrita : fonts.medio, fontSize: t(11), color: colors.texto }}>{texto}</Text>
    </Pressable>
  );
  const colorTipo = esIngreso ? colors.ingreso : colors.egreso;

  return (
    <ModalFlotante visible titulo={esIngreso ? 'Editar ingreso' : 'Editar egreso'} onCerrar={onCerrar}>
      <View style={{ gap: t(10) }}>
        <Text style={etiqueta}>Monto</Text>
        <TextInput value={monto ? `$ ${Number(monto).toLocaleString('es-CL')}` : ''} keyboardType="number-pad"
          onChangeText={(v) => setMonto(v.replace(/\D/g, '').slice(0, 10))}
          style={[campo, { fontFamily: fonts.extra, fontSize: t(18), color: colorTipo }]} />

        <Text style={etiqueta}>Categoría</Text>
        <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={{ gap: t(6) }}>
          {opciones.map((c) => chip(c.clave, c.nombre, c.id === categoriaId, () => setCategoriaId(c.id), c.icono, c.color))}
        </ScrollView>

        {esIngreso && (
          <>
            <Text style={etiqueta}>Método de pago</Text>
            <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: t(6) }}>
              {metodosPago.map((m) => chip(m.id, m.nombre, m.id === metodo, () => setMetodo(m.id), m.icono))}
            </View>
          </>
        )}

        <Text style={etiqueta}>Comentario</Text>
        <TextInput value={nota} onChangeText={setNota} maxLength={100} placeholder="Nota opcional" multiline
          placeholderTextColor={colors.textoSecundario} style={[campo, { textAlignVertical: 'top' }]} />

        <Pressable onPress={guardar} disabled={incompleto || guardando} style={({ pressed }) => ({
          flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: t(6), paddingVertical: t(10), borderRadius: t(12),
          backgroundColor: colorTipo, opacity: incompleto || guardando ? 0.5 : pressed ? 0.9 : 1, marginTop: t(4),
        })}>
          <Icono nombre="check-circle-2" tamano={t(16)} color={colors.superficie} />
          <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.superficie }}>{guardando ? 'Guardando...' : 'Guardar cambios'}</Text>
        </Pressable>
        <Pressable onPress={eliminar} hitSlop={8} style={{ flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: t(6), paddingVertical: t(4) }}>
          <Icono nombre="trash" tamano={t(14)} color={colors.rosado} />
          <Text style={{ fontFamily: fonts.negrita, fontSize: t(12), color: colors.rosado }}>Eliminar movimiento</Text>
        </Pressable>
      </View>
    </ModalFlotante>
  );
}
