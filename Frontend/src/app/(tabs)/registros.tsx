import { useLocalSearchParams } from 'expo-router';
import { useEffect, useState } from 'react';
import { View } from 'react-native';
import { useSesion } from '../../features/auth/SesionContext';
import PanelMovimientos from '../../features/registros/components/PanelMovimientos';
import PanelPendientes from '../../features/registros/components/PanelPendientes';
import ResumenGeneral from '../../features/registros/components/ResumenGeneral';
import BarraTitulo from '../../shared/components/BarraTitulo';
import EstadoEmpresa from '../../shared/components/EstadoEmpresa';
import Segmentado from '../../shared/components/Segmentado';
import { colors } from '../../shared/theme/colors';
import { useEscala } from '../../shared/utils/escala';

type Vista = 'movimientos' | 'pendientes';

// Registros: todo lo que el emprendedor ha anotado (movimientos con Excel y pagos pendientes)
export default function Registros() {
  const { empresa } = useSesion();
  const { t } = useEscala();
  const params = useLocalSearchParams<{ vista?: Vista }>();
  const [vista, setVista] = useState<Vista>('movimientos');
  const [version, setVersion] = useState(0); // Sube cuando se edita algo: el resumen general se recalcula

  // El botón "Ver detalle de movimientos" del Dashboard abre directo la vista de movimientos
  useEffect(() => {
    if (params.vista === 'pendientes' || params.vista === 'movimientos') setVista(params.vista);
  }, [params.vista]);

  // Resumen general arriba (no depende de los filtros) y debajo el selector de vista
  const selector = (
    <View style={{ gap: t(12) }}>
    <ResumenGeneral version={version} />
    <Segmentado
      opciones={[{ id: 'movimientos', nombre: 'Movimientos' }, { id: 'pendientes', nombre: 'Pagos pendientes' }]}
      valor={vista}
      onCambiar={setVista}
      tamano={12}
    />
    </View>
  );

  return (
    <View style={{ flex: 1, backgroundColor: colors.fondoApp }}>
      {!empresa ? (
        <>
          <BarraTitulo titulo="Registros" />
          <View style={{ padding: t(16) }}><EstadoEmpresa /></View>
        </>
      ) : vista === 'movimientos' ? (
        <PanelMovimientos selector={selector} onCambio={() => setVersion((v) => v + 1)} />
      ) : (
        <PanelPendientes selector={selector} onCambio={() => setVersion((v) => v + 1)} />
      )}
    </View>
  );
}
