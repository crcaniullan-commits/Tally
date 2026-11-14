import { useFocusEffect } from 'expo-router';
import { useCallback, useState } from 'react';
import { traducirError } from '../auth/errores';
import { useSesion } from '../auth/SesionContext';
import { categoriasPredeterminadas, combinarCategorias } from '../categorias/categorias';
import { listarCategorias } from '../categorias/categorias.service';
import { leerEstilos } from '../categorias/estilosPropios';
import { gastosDesde, ingresosDesde } from '../movimientos/movimientos.service';
import { DatosDashboard, fechaDesde } from './calcularDashboard';
import { Periodo } from './resumen';

// Carga los datos del Dashboard cada vez que se abre la pestaña o cambia el período
// (así el gráfico se actualiza solo después de registrar un movimiento)
export function useDatosDashboard(periodo: Periodo) {
  const { token, empresa, manejarError } = useSesion();
  const [datos, setDatos] = useState<DatosDashboard | null>(null);
  const [recargas, setRecargas] = useState(0); // Avisa a la tarjeta de la meta que recargue
  const [cargando, setCargando] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const cargar = useCallback(async () => {
    if (!token || !empresa) return;
    setCargando(true);
    setError(null);
    try {
      const desde = fechaDesde(periodo);
      const [ingresos, gastos, cats] = await Promise.all([
        ingresosDesde(desde, empresa.id, token),
        gastosDesde(desde, empresa.id, token),
        listarCategorias(empresa.id, token).catch(() => []),
      ]);
      const categorias = cats.length ? combinarCategorias(cats, await leerEstilos(empresa.id)) : categoriasPredeterminadas;
      setDatos({ ingresos, gastos, categorias });
      setRecargas((n) => n + 1);
    } catch (err) {
      manejarError(err);
      setError(traducirError(err));
    } finally {
      setCargando(false);
    }
  }, [token, empresa, periodo, manejarError]);

  useFocusEffect(useCallback(() => { cargar(); }, [cargar]));

  return { datos, cargando, error, cargar, recargas };
}
