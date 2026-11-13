import { colors } from '../../shared/theme/colors';
import { Categoria } from '../categorias/categorias';
import { Movimiento } from './components/HistorialMovimientos';
import { metodosPago } from './metodosPago';
import { GastoApi, IngresoApi } from './movimientos.service';

export // Une ingresos y gastos del backend en una sola lista, del más reciente al más antiguo
function armarHistorial(ingresos: IngresoApi[], gastos: GastoApi[], categorias: Categoria[]): Movimiento[] {
  const buscar = (id: string | null) => categorias.find((c) => c.id === id);
  const estilo = (id: string | null) => {
    const c = buscar(id);
    return {
      categoria: c?.nombre ?? 'Sin categoría',
      icono: c?.icono ?? 'tag-outline',
      color: c?.color ?? colors.lavanda,
      colorSuave: c?.colorSuave ?? colors.lavandaSuave,
    };
  };

  const lista: Movimiento[] = [
    ...ingresos.map((i) => ({
      id: i.id, tipo: 'ingreso' as const, monto: i.monto, nota: i.descripcion ?? '', fecha: new Date(i.created_at),
      metodo: metodosPago.find((m) => m.id === i.payment_method)?.nombre ?? '', metodoId: i.payment_method,
      categoriaId: i.category_id, ...estilo(i.category_id),
    })),
    ...gastos.map((g) => ({
      id: g.id, tipo: 'egreso' as const, monto: g.monto, nota: g.descripcion ?? '', fecha: new Date(g.created_at),
      metodo: '', metodoId: null, categoriaId: g.category_id, ...estilo(g.category_id),
    })),
  ];
  return lista.sort((a, b) => b.fecha.getTime() - a.fecha.getTime());
}
