    import { Ionicons } from '@expo/vector-icons';
    import { useState } from 'react';
    import { KeyboardAvoidingView, Platform, Pressable, ScrollView, Text, TextInput, View } from 'react-native';
    import { useSesion } from '../../features/auth/SesionContext';
    import {
    Categoria,
    categoriasPredeterminadas,
    crearCategoriaPersonalizada,
    TipoMovimiento,
    } from '../../features/categorias/categorias';
    import ModalCategorias from '../../features/categorias/components/ModalCategorias';
    import CarruselOpciones, { OpcionCarrusel } from '../../features/movimientos/components/CarruselOpciones';
    import HistorialMovimientos, { Movimiento } from '../../features/movimientos/components/HistorialMovimientos';
    import TecladoNumerico from '../../features/movimientos/components/TecladoNumerico';
    import { MetodoPago, metodosPago } from '../../features/movimientos/metodosPago';
    import Boton from '../../shared/components/Boton';
    import { datosPrueba } from '../../shared/mocks/datosPrueba';
    import { colors } from '../../shared/theme/colors';
    import { fonts } from '../../shared/theme/fonts';
    import { sombras } from '../../shared/theme/sombras';
    import { formatearMonto } from '../../shared/utils/formato';

    const MAX_DIGITOS = 10;
    const ID_OTRA = 'otra'; // Tarjeta que abre el modal para crear una categoría propia

    export default function Inicio() {
    const { cerrarSesion } = useSesion();
    // Pendiente: el nombre vendrá del usuario con sesión iniciada (backend)
    const primerNombre = datosPrueba.nombre.split(' ')[0] || 'emprendedor';

    const [tipo, setTipo] = useState<TipoMovimiento>('ingreso');
    const [monto, setMonto] = useState(0);
    const [categoriaId, setCategoriaId] = useState<string | null>(null);
    const [metodo, setMetodo] = useState<MetodoPago | null>(null);
    const [nota, setNota] = useState('');
    const [modalCategorias, setModalCategorias] = useState(false);
    // Pendiente backend: las categorías propias se guardarán en la base de datos
    const [categoriasPropias, setCategoriasPropias] = useState<Categoria[]>([]);
    // Pendiente backend: se cargará y guardará en la base de datos
    const [movimientos, setMovimientos] = useState<Movimiento[]>([]);

    const predefinidas = categoriasPredeterminadas.filter((c) => c.tipo === tipo);
    const propias = categoriasPropias.filter((c) => c.tipo === tipo);
    const colorTipo = tipo === 'ingreso' ? colors.ingreso : colors.egreso;
    const formularioIncompleto = monto === 0 || !categoriaId || !metodo;

    // Tarjetas del carrusel de categorías: predeterminadas + propias + "Otra"
    const opcionesCategorias: OpcionCarrusel[] = [
        ...predefinidas,
        ...propias,
        { id: ID_OTRA, nombre: 'Otra', icono: 'plus', color: colors.lavanda, colorSuave: colors.lavandaSuave },
    ];

    // Tarjetas del carrusel de métodos de pago (los 4 que acepta el backend)
    const opcionesMetodos: OpcionCarrusel[] = metodosPago.map((m) => ({
        ...m,
        color: colors.lavanda,
        colorSuave: colors.lavandaSuave,
    }));

    function cambiarTipo(nuevo: TipoMovimiento) {
        setTipo(nuevo);
        setCategoriaId(null); // Las categorías de ingreso y egreso son distintas
    }

    function seleccionarCategoria(id: string) {
        if (id === ID_OTRA) setModalCategorias(true);
        else setCategoriaId(id);
    }

    function agregarDigito(digito: string) {
        const nuevo = `${monto === 0 ? '' : monto}${digito}`;
        if (nuevo.length > MAX_DIGITOS) return;
        setMonto(Number(nuevo));
    }

    function agregarCategoria(nombre: string) {
        const nueva = crearCategoriaPersonalizada(nombre, tipo);
        setCategoriasPropias([...categoriasPropias, nueva]);
        setCategoriaId(nueva.id); // Queda seleccionada automáticamente
    }

    function registrar() {
        const categoria = opcionesCategorias.find((c) => c.id === categoriaId);
        const metodoPago = metodosPago.find((m) => m.id === metodo);
        if (!categoria || !metodoPago || monto === 0) return;

        const nuevo: Movimiento = {
        id: `${Date.now()}`,
        tipo,
        monto,
        categoria: categoria.nombre,
        icono: categoria.icono,
        color: categoria.color,
        colorSuave: categoria.colorSuave,
        metodo: metodoPago.nombre,
        nota: nota.trim(),
        fecha: new Date(),
        };
        setMovimientos([nuevo, ...movimientos]); // El más reciente queda arriba

        // Limpiar el formulario para el siguiente registro
        setMonto(0);
        setCategoriaId(null);
        setMetodo(null);
        setNota('');
        // Pendiente: guardar en el backend (POST /v1/app/incomes)
    }

    return (
        <KeyboardAvoidingView style={{ flex: 1, backgroundColor: colors.fondoApp }} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
        <ScrollView contentContainerStyle={{ paddingHorizontal: 20, paddingTop: 52, paddingBottom: 32, gap: 22 }} keyboardShouldPersistTaps="handled">

            {/* Encabezado */}
            <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
            <Text style={{ color: colors.lavanda, fontFamily: fonts.negrita, fontSize: 28 }}>¡Hola, {primerNombre}!</Text>
            <View style={{ flexDirection: 'row', gap: 20 }}>
                <Pressable hitSlop={10}>
                <Ionicons name="notifications-outline" size={28} color={colors.lavanda} />
                </Pressable>
                <Pressable hitSlop={10} onPress={cerrarSesion}>
                <Ionicons name="log-out-outline" size={28} color={colors.lavanda} />
                </Pressable>
            </View>
            </View>

            {/* Pasos 2 y 3: categoría y método de pago en carrusel */}
            <CarruselOpciones titulo="Categorías" opciones={opcionesCategorias} seleccionadaId={categoriaId} onSeleccionar={seleccionarCategoria} />
            <CarruselOpciones titulo="Método de pago" opciones={opcionesMetodos} seleccionadaId={metodo} onSeleccionar={(id) => setMetodo(id as MetodoPago)} />

            {/* Paso 1: calculadora (monto + ingreso/egreso + teclado) en una sola tarjeta */}
            <View style={{ backgroundColor: colors.superficie, borderRadius: 24, padding: 16, gap: 14, ...sombras.flotante }}>
            <Text style={{ fontFamily: fonts.titulo, fontSize: 60, color: colors.lavanda, textAlign: 'center', letterSpacing: 1 }}>
                $ {formatearMonto(monto)}
            </Text>

            <View style={{ flexDirection: 'row', borderRadius: 26, overflow: 'hidden', backgroundColor: colors.lavandaSuave }}>
                {(['ingreso', 'egreso'] as const).map((opcion) => {
                const activo = tipo === opcion;
                const colorOpcion = opcion === 'ingreso' ? colors.ingreso : colors.egreso;
                return (
                    <Pressable
                    key={opcion}
                    onPress={() => cambiarTipo(opcion)}
                    style={{ flex: 1, paddingVertical: 14, alignItems: 'center', justifyContent: 'center', borderRadius: 26, backgroundColor: activo ? colorOpcion : 'transparent' }}
                    >
                    <Text style={{ width: '100%', textAlign: 'center', fontFamily: fonts.medio, fontSize: 16, color: activo ? colors.superficie : colors.lavanda }}>
                        {opcion === 'ingreso' ? 'INGRESO (+)' : 'EGRESO (−)'}
                    </Text>
                    </Pressable>
                );
                })}
            </View>

            <TecladoNumerico
                onDigito={agregarDigito}
                onBorrar={() => setMonto(Math.floor(monto / 10))}
                onLimpiar={() => setMonto(0)}
            />
            </View>

            {/* Nota opcional + registrar */}
            <TextInput
            value={nota}
            onChangeText={setNota}
            maxLength={100}
            placeholder="Nota (opcional). Ej: Venta poleras diseño"
            placeholderTextColor={colors.lavanda}
            style={{ backgroundColor: colors.superficie, borderRadius: 16, padding: 16, fontFamily: fonts.normal, fontSize: 16, color: colors.lavanda, borderWidth: 1.5, borderColor: colors.lavandaSuave }}
            />

            <Boton
            titulo={tipo === 'ingreso' ? 'Registrar venta' : 'Registrar gasto'}
            color={colorTipo}
            onPress={registrar}
            deshabilitado={formularioIncompleto}
            />

            <HistorialMovimientos movimientos={movimientos} />
        </ScrollView>

        <ModalCategorias
            visible={modalCategorias}
            categorias={propias}
            nombresExistentes={[...predefinidas, ...propias].map((c) => c.nombre)}
            seleccionadaId={categoriaId}
            onSeleccionar={(id) => { setCategoriaId(id); setModalCategorias(false); }}
            onAgregar={agregarCategoria}
            onCerrar={() => setModalCategorias(false)}
        />
        </KeyboardAvoidingView>
    );
    }