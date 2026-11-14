import { useState } from 'react';
import { ActivityIndicator, Text, View } from 'react-native';
import { useSesion } from '../../features/auth/SesionContext';
import Icono from './Icono';
import { colors } from '../theme/colors';
import { fonts } from '../theme/fonts';
import { sombras } from '../theme/sombras';
import { useEscala } from '../utils/escala';
import Boton from './Boton';

// Se muestra mientras el emprendimiento no ha cargado: un indicador de carga,
// o el error con un botón para reintentar (así la pantalla nunca queda en blanco)
export default function EstadoEmpresa() {
  const { empresa, errorEmpresa, recargarEmpresa } = useSesion();
  const { e } = useEscala();
  const [reintentando, setReintentando] = useState(false);
  if (empresa) return null;

  if (!errorEmpresa) {
    return (
      <View style={{ alignItems: 'center', gap: 8, paddingVertical: 24 }}>
        <ActivityIndicator color={colors.lavanda} />
        <Text style={{ fontFamily: fonts.medio, fontSize: e(15), color: colors.lavanda }}>Cargando tu emprendimiento…</Text>
      </View>
    );
  }

  return (
    <View style={{ backgroundColor: colors.superficie, borderRadius: 18, padding: 16, gap: 10, alignItems: 'center', ...sombras.suave }}>
      <Icono nombre="wifi-off" tamano={e(32)} color={colors.rosado} />
      <Text style={{ fontFamily: fonts.negrita, fontSize: e(16), color: colors.lavanda, textAlign: 'center' }}>
        No pudimos cargar tus datos
      </Text>
      <Text style={{ fontFamily: fonts.normal, fontSize: e(14), color: colors.lavanda, textAlign: 'center' }}>{errorEmpresa}</Text>
      <Boton
        titulo={reintentando ? 'Reintentando...' : 'Reintentar'}
        deshabilitado={reintentando}
        onPress={async () => { setReintentando(true); await recargarEmpresa(); setReintentando(false); }}
      />
    </View>
  );
}
