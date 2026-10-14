    import { colors } from './colors';

    export type Categoria = {
    id: string;
    nombre: string;
    icono: string;        // Nombre del ícono (MaterialCommunityIcons)
    color: string;        // Color intenso: fondo del bloque
    tipo: 'ingreso' | 'egreso';
    };

    // Categorías predeterminadas (Historia 003).
    // Pendiente: cuando el backend tenga el módulo de categorías, se cargarán desde la API.
    export const categoriasPredeterminadas: Categoria[] = [
    { id: 'logistica', nombre: 'Logística y envíos', icono: 'truck-delivery', color: colors.azul, tipo: 'egreso' },
    { id: 'insumos', nombre: 'Mercadería e insumos', icono: 'package-variant-closed', color: colors.naranjo, tipo: 'egreso' },
    { id: 'marketing', nombre: 'Marketing y publicidad', icono: 'bullhorn', color: colors.rosado, tipo: 'egreso' },
    { id: 'sueldo', nombre: 'Sueldo del dueño', icono: 'account-cash', color: colors.verde, tipo: 'egreso' },
    ];