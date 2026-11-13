import { useLocalSearchParams } from 'expo-router';
import { useEffect, useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useSesion } from '../../features/auth/SesionContext';
import PanelMovimientos from '../../features/registros/components/PanelMovimientos';
import PanelPendientes from '../../features/registros/components/PanelPendientes';
import { colors } from '../../shared/theme/colors';
import { fonts } from '../../shared/theme/fonts';
import { useEscala } from '../../shared/utils/escala';

type Vista = 'movimientos' | 'pendientes';

const vistas: { id: Vista; nombre: string }[] = [
  { id: 'movimientos', nombre: 'Movimientos' },
  { id: 'pendientes', nombre: 'Pagos pendientes' },
];

// Registros: todo lo que el emprendedor ha anotado (movimientos con Excel y pagos pendientes)
export default function Registros() {
  const { empresa } = useSesion();
  const { e } = useEscala();
  const insets = useSafeAreaInsets();
  const params = useLocalSearchParams<{ vista?: Vista }>();
  const [vista, setVista] = useState<Vista>('movimientos');

  // La campana de Inicio abre directo la vista de pagos pendientes
  useEffect(() => {
    if (params.vista === 'pendientes' || params.vista === 'movimientos') setVista(params.vista);
  }, [params.vista]);

  return (
    <View style={{ flex: 1, backgroundColor: colors.fondoApp, paddingTop: insets.top + 16 }}>
      <View style={{ paddingHorizontal: 20, gap: 14, paddingBottom: 14 }}>
        <View>
          <Text style={{ fontFamily: fonts.negrita, fontSize: e(30), color: colors.lavanda }}>Registros</Text>
          <Text numberOfLines={1} style={{ fontFamily: fonts.medio, fontSize: e(17), color: colors.lavanda, opacity: 0.8 }}>
            {empresa?.nombre_empresa ?? ''}
          </Text>
        </View>

        {/* Selector de vista */}
        <View style={{ flexDirection: 'row', borderRadius: 26, backgroundColor: colors.lavandaSuave }}>
          {vistas.map((v) => {
            const activo = v.id === vista;
            return (
              <Pressable key={v.id} onPress={() => setVista(v.id)}
                style={{ flex: 1, paddingVertical: e(12), borderRadius: 26, alignItems: 'center', backgroundColor: activo ? colors.lavanda : 'transparent' }}>
                <Text style={{ fontFamily: fonts.negrita, fontSize: e(16), color: activo ? colors.superficie : colors.lavanda }}>{v.nombre}</Text>
              </Pressable>
            );
          })}
        </View>
      </View>

      {vista === 'movimientos' ? <PanelMovimientos /> : <PanelPendientes />}
    </View>
  );
}
