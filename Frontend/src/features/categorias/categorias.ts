    import { colors } from '../../shared/theme/colors';

    export type TipoMovimiento = 'ingreso' | 'egreso';

    export type Categoria = {
    id: string;
    nombre: string;
    icono: string;      // Nombre del ícono (MaterialCommunityIcons)
    color: string;      // Color intenso: ícono y borde al seleccionar
    colorSuave: string; // Color suave: fondo del ícono
    tipo: TipoMovimiento;
    };

    // Categorías predeterminadas genéricas (Historia 003): 3 por tipo de movimiento.
    // Pendiente backend: cuando exista el módulo de categorías, se cargarán desde la API.
    export const categoriasPredeterminadas: Categoria[] = [
    // Ingresos
    { id: 'ventas', nombre: 'Ventas', icono: 'shopping-outline', color: colors.verde, colorSuave: colors.verdeSuave, tipo: 'ingreso' },
    { id: 'servicios', nombre: 'Servicios', icono: 'handshake-outline', color: colors.azul, colorSuave: colors.azulSuave, tipo: 'ingreso' },
    { id: 'pedidos', nombre: 'Pedidos', icono: 'clipboard-text-outline', color: colors.naranjo, colorSuave: colors.naranjoSuave, tipo: 'ingreso' },

    // Egresos
    { id: 'insumos', nombre: 'Insumos', icono: 'package-variant-closed', color: colors.naranjo, colorSuave: colors.naranjoSuave, tipo: 'egreso' },
    { id: 'transporte', nombre: 'Transporte', icono: 'truck-delivery', color: colors.azul, colorSuave: colors.azulSuave, tipo: 'egreso' },
    { id: 'gastos-fijos', nombre: 'Gastos fijos', icono: 'home-city-outline', color: colors.rosado, colorSuave: colors.rosadoSuave, tipo: 'egreso' },
    ];

    // Apariencia de las categorías que crea el usuario
    export function crearCategoriaPersonalizada(nombre: string, tipo: TipoMovimiento): Categoria {
    return {
        id: `personalizada-${Date.now()}`,
        nombre: nombre.trim(),
        icono: 'tag-outline',
        color: colors.lavanda,
        colorSuave: colors.lavandaSuave,
        tipo,
    };
    }