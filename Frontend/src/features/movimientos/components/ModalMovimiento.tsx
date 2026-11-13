import { useEffect, useState } from 'react';
import { Alert, ScrollView, Text, View } from 'react-native';
import Boton from '../../../shared/components/Boton';
import CampoTexto from '../../../shared/components/CampoTexto';
import ModalFlotante from '../../../shared/components/ModalFlotante';
import { colors } from '../../../shared/theme/colors';
import { fonts } from '../../../shared/theme/fonts';
import { useEscala } from '../../../shared/utils/escala';
import { traducirError } from '../../auth/errores';
import { useSesion } from '../../auth/SesionContext';
import { Categoria } from '../../categorias/categorias';
import { obtenerIdReal } from '../../categorias/categorias.service';
import { MetodoPago, metodosPago } from '../metodosPago';
import { eliminarMovimiento, reemplazarMovimiento } from '../movimientos.service';
import Chip from './Chip';
import { Movimiento } from './HistorialMovimientos';

type Props = {
  movimiento: Movimiento | null; // null = modal cerrado
  categorias: Categoria[];
  onCerrar: () => void;
  onCambio: () => void; // Se llama después de editar o eliminar, para recargar la lista
};

// Ver, editar o eliminar un movimiento (Historia 012)
export default function ModalMovimiento({ movimiento, categorias, onCerrar, onCambio }: Props) {
  const { token, empresa, manejarError } = useSesion();
  const { e } = useEscala();
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
      await reemplazarMovimiento(movimiento, { monto: montoNumero, metodo, categoriaId: idReal, nota }, empresa.id, token);
      onCambio();
      onCerrar();
    } catch (error) {
      manejarError(error);
      Alert.alert('No se pudo guardar', traducirError(error));
    } finally {
      setGuardando(false);
    }
  }

  function eliminar() {
    Alert.alert('Eliminar movimiento', '¿Seguro que quieres eliminarlo? No se puede deshacer.', [
      { text: 'Cancelar', style: 'cancel' },
      {
        text: 'Eliminar', style: 'destructive', onPress: async () => {
          if (!token || !empresa || !movimiento) return;
          try {
            await eliminarMovimiento(movimiento.tipo, movimiento.id, empresa.id, token);
            onCambio();
            onCerrar();
          } catch (error) {
            manejarError(error);
            Alert.alert('No se pudo eliminar', traducirError(error));
          }
        },
      },
    ]);
  }

  const etiqueta = { fontFamily: fonts.medio, fontSize: e(15), color: colors.lavanda };

  return (
    <ModalFlotante visible titulo={esIngreso ? 'Editar ingreso' : 'Editar gasto'} onCerrar={onCerrar}>
      <View style={{ gap: 14 }}>
        <CampoTexto etiqueta="Monto" value={monto} onChangeText={(t) => setMonto(t.replace(/\D/g, '').slice(0, 10))}
          keyboardType="number-pad" />

        <Text style={etiqueta}>Categoría</Text>
        <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={{ gap: 8 }}>
          {opciones.map((c) => (
            <Chip key={c.id} texto={c.nombre} activo={c.id === categoriaId} onPress={() => setCategoriaId(c.id)} />
          ))}
        </ScrollView>

        {esIngreso && (
          <>
            <Text style={etiqueta}>Método de pago</Text>
            <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }}>
              {metodosPago.map((m) => (
                <Chip key={m.id} texto={m.nombre} activo={m.id === metodo} onPress={() => setMetodo(m.id)} />
              ))}
            </View>
          </>
        )}

        <CampoTexto etiqueta="Nota (opcional)" value={nota} onChangeText={setNota} maxLength={100} />

        <Boton titulo={guardando ? 'Guardando...' : 'Guardar cambios'} onPress={guardar}
          deshabilitado={incompleto || guardando} color={esIngreso ? colors.ingreso : colors.egreso} />
        <Boton titulo="Eliminar movimiento" onPress={eliminar} variante="secundario" color={colors.error} />
      </View>
    </ModalFlotante>
  );
}
