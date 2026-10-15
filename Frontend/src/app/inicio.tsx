    import { Ionicons } from '@expo/vector-icons';
    import { useState } from 'react';
    import { KeyboardAvoidingView, Platform, Pressable, ScrollView, Text, TextInput, View } from 'react-native';
    import { useSesion } from '../features/auth/SesionContext';
    import {
    Categoria,
    categoriasPredeterminadas,
    crearCategoriaPersonalizada,
    TipoMovimiento,
    } from '../features/categorias/categorias';
    import ModalCategorias from '../features/categorias/components/ModalCategorias';
    import OpcionIcono from '../features/movimientos/components/OpcionIcono';
    import TecladoNumerico from '../features/movimientos/components/TecladoNumerico';
    import { MetodoPago, metodosPago } from '../features/movimientos/metodosPago';
    import Boton from '../shared/components/Boton';
    import { datosPrueba } from '../shared/mocks/datosPrueba';
    import { colors } from '../shared/theme/colors';
    import { fonts } from '../shared/theme/fonts';
    import { sombras } from '../shared/theme/sombras';
    import { formatearMonto } from '../shared/utils/formato';

    const MAX_DIGITOS = 10;
    const ANCHO_OPCION = '47%'; // Dos opciones por fila dentro de cada columna

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

    const predefinidas = categoriasPredeterminadas.filter((c) => c.tipo === tipo);
    const propias = categoriasPropias.filter((c) => c.tipo === tipo);
    const propiaSeleccionada = propias.find((c) => c.id === categoriaId);
    const colorTipo = tipo === 'ingreso' ? colors.ingreso : colors.egreso;
    const formularioIncompleto = monto === 0 || !categoriaId || !metodo;

    function cambiarTipo(nuevo: TipoMovimiento) {
        setTipo(nuevo);
        setCategoriaId(null); // Las categorías de ingreso y egreso son distintas
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
        // Pendiente (siguiente paso): guardar en el backend (POST /v1/app/incomes)
    }

    const tituloSeccion = { fontFamily: fonts.medio, fontSize: 15, color: colors.texto };

    return (
        <KeyboardAvoidingView style={{ flex: 1, backgroundColor: colors.fondoApp }} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
        <ScrollView contentContainerStyle={{ paddingHorizontal: 20, paddingTop: 52, paddingBottom: 32, gap: 16 }} keyboardShouldPersistTaps="handled">

            {/* Encabezado */}
            <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
            <Text style={{ color: colors.textoSecundario, fontFamily: fonts.medio, fontSize: 15 }}>¡Hola, {primerNombre}!</Text>
            <View style={{ flexDirection: 'row', gap: 18 }}>
                <Pressable hitSlop={10}>
                <Ionicons name="notifications-outline" size={24} color={colors.texto} />
                </Pressable>
                <Pressable hitSlop={10} onPress={cerrarSesion}>
                <Ionicons name="log-out-outline" size={24} color={colors.texto} />
                </Pressable>
            </View>
            </View>

            {/* Monto + selector Ingreso / Egreso */}
            <View style={{ backgroundColor: colors.superficie, borderRadius: 22, overflow: 'hidden', ...sombras.flotante }}>
            <Text style={{ fontFamily: fonts.titulo, fontSize: 52, color: colors.lavanda, textAlign: 'center', paddingVertical: 14, letterSpacing: 1 }}>
                $ {formatearMonto(monto)}
            </Text>
            <View style={{ flexDirection: 'row' }}>
                {(['ingreso', 'egreso'] as const).map((opcion) => {
                const activo = tipo === opcion;
                const colorOpcion = opcion === 'ingreso' ? colors.ingreso : colors.egreso;
                return (
                    <Pressable
                    key={opcion}
                    onPress={() => cambiarTipo(opcion)}
                    style={{ flex: 1, paddingVertical: 10, alignItems: 'center', backgroundColor: activo ? colorOpcion : colors.lavandaSuave }}
                    >
                    <Text style={{ fontFamily: fonts.medio, color: activo ? colors.superficie : colors.textoSecundario }}>
                        {opcion === 'ingreso' ? 'INGRESO (+)' : 'EGRESO (−)'}
                    </Text>
                    </Pressable>
                );
                })}
            </View>
            </View>

            {/* Paso 1: monto */}
            <TecladoNumerico
            onDigito={agregarDigito}
            onBorrar={() => setMonto(Math.floor(monto / 10))}
            onLimpiar={() => setMonto(0)}
            />

            {/* Pasos 2 y 3: categoría (izquierda) y método de pago (derecha) */}
            <View style={{ flexDirection: 'row', gap: 14 }}>
            <View style={{ flex: 1, gap: 8 }}>
                <Text style={tituloSeccion}>Categoría</Text>
                <View style={{ flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'space-between', rowGap: 10 }}>
                {predefinidas.map((c) => (
                    <OpcionIcono key={c.id} ancho={ANCHO_OPCION} nombre={c.nombre} icono={c.icono} color={c.color}
                    colorSuave={c.colorSuave} seleccionado={categoriaId === c.id} onPress={() => setCategoriaId(c.id)} />
                ))}
                <OpcionIcono
                    ancho={ANCHO_OPCION}
                    nombre={propiaSeleccionada ? propiaSeleccionada.nombre : 'Otra categoría'}
                    icono={propiaSeleccionada ? propiaSeleccionada.icono : 'dots-horizontal'}
                    color={colors.lavanda}
                    colorSuave={colors.lavandaSuave}
                    seleccionado={!!propiaSeleccionada}
                    onPress={() => setModalCategorias(true)}
                />
                </View>
            </View>

            <View style={{ flex: 1, gap: 8 }}>
                <Text style={tituloSeccion}>Método de pago</Text>
                <View style={{ flexDirection: 'row', flexWrap: 'wrap', justifyContent: 'space-between', rowGap: 10 }}>
                {metodosPago.map((m) => (
                    <OpcionIcono key={m.id} ancho={ANCHO_OPCION} nombre={m.nombre} icono={m.icono} color={colors.lavanda}
                    colorSuave={colors.lavandaSuave} seleccionado={metodo === m.id} onPress={() => setMetodo(m.id)} />
                ))}
                </View>
            </View>
            </View>

            {/* Nota opcional */}
            <Text style={tituloSeccion}>Nota (opcional)</Text>
            <TextInput
            value={nota}
            onChangeText={setNota}
            maxLength={100}
            placeholder="Ej: Venta poleras diseño"
            placeholderTextColor={colors.textoSecundario}
            style={{ backgroundColor: colors.superficie, borderRadius: 14, padding: 14, fontFamily: fonts.normal, fontSize: 15, color: colors.texto, borderWidth: 1.5, borderColor: colors.lavandaSuave, ...sombras.suave }}
            />

            <Boton
            titulo={tipo === 'ingreso' ? 'Registrar venta' : 'Registrar gasto'}
            color={colorTipo}
            onPress={registrar}
            deshabilitado={formularioIncompleto}
            />
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