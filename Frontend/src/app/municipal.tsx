import { useEffect, useState } from 'react';
import { KeyboardAvoidingView, Platform, Pressable, ScrollView, Text, View } from 'react-native';
import { useAlerta, useConfirmar } from '../shared/components/Confirmar';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { traducirError } from '../features/auth/errores';
import { useSesion } from '../features/auth/SesionContext';
import {
  buscarPorRut,
  ClaveEmitida,
  emitirClave,
  guardarClavesEmitidas,
  listarClavesEmitidas,
  reenviarClave,
  revocarClave,
  UsuarioMunicipal,
} from '../features/municipal/municipal.service';
import { exportarClaves } from '../features/municipal/exportarClaves';
import { ApiError } from '../shared/api/cliente';
import Boton from '../shared/components/Boton';
import CampoTexto from '../shared/components/CampoTexto';
import Encabezado from '../shared/components/Encabezado';
import Icono from '../shared/components/Icono';
import { colors } from '../shared/theme/colors';
import { fonts } from '../shared/theme/fonts';
import { sombras } from '../shared/theme/sombras';
import { useEscala } from '../shared/utils/escala';
import { aDMA } from '../shared/utils/fechas';
import { rutEsValido } from '../shared/utils/rut';

const tarjeta = { backgroundColor: colors.superficie, borderRadius: 20, padding: 16, gap: 12, ...sombras.suave };
const formatoRut = (r: UsuarioMunicipal['rut']) => `${String(r.Cuerpo).replace(/\B(?=(\d{3})+(?!\d))/g, '.')}-${r.DV}`;
const fecha = (iso: string | null) => (iso ? aDMA(new Date(iso)) : '—');

// Backoffice municipal (Historia 009): generar, reenviar y anular claves, y buscar emprendedores.
// La municipalidad NO ve datos financieros: solo gestiona el acceso.
export default function Municipal() {
  const alerta = useAlerta();
  const confirmar = useConfirmar();
  const { token, cerrarSesion, manejarError } = useSesion();
  const { e } = useEscala();
  const insets = useSafeAreaInsets();

  const [correo, setCorreo] = useState('');
  const [errorCorreo, setErrorCorreo] = useState<string | undefined>();
  const [claves, setClaves] = useState<ClaveEmitida[]>([]);
  const [generadas, setGeneradas] = useState<ClaveEmitida[]>([]); // Resultado del último lote
  const [progreso, setProgreso] = useState<string | null>(null);
  const [rut, setRut] = useState('');
  const [errorRut, setErrorRut] = useState<string | undefined>();
  const [encontrado, setEncontrado] = useState<UsuarioMunicipal | null>(null);
  const [ocupado, setOcupado] = useState<string | null>(null);

  useEffect(() => { listarClavesEmitidas().then(setClaves); }, []);

  async function actualizarLista(nueva: ClaveEmitida[]) {
    setClaves(nueva);
    await guardarClavesEmitidas(nueva);
  }

  // Generación masiva (Historia 009): uno o varios correos, separados por coma, espacio o salto de línea
  async function generar() {
    if (!token) return;
    const correos = [...new Set(correo.split(/[\s,;]+/).map((c) => c.trim().toLowerCase()).filter(Boolean))];
    const invalidos = correos.filter((c) => !/^\S+@\S+\.\S+$/.test(c));
    if (correos.length === 0) return setErrorCorreo('Ingresa al menos un correo');
    if (invalidos.length > 0) return setErrorCorreo(`Correo inválido: ${invalidos.join(', ')}`);
    setErrorCorreo(undefined);
    setOcupado('generar');
    const nuevas: ClaveEmitida[] = [];
    const fallidos: string[] = [];
    for (let i = 0; i < correos.length; i++) {
      setProgreso(`Generando ${i + 1} de ${correos.length}...`);
      try {
        const clave = await emitirClave(correos[i], token);
        nuevas.push({ ...clave, email: correos[i] });
      } catch (error) {
        manejarError(error);
        fallidos.push(`${correos[i]}: ${traducirError(error)}`);
      }
    }
    setProgreso(null);
    setGeneradas(nuevas);
    await actualizarLista([...nuevas, ...claves]);
    if (nuevas.length > 0) setCorreo('');
    setOcupado(null);
    if (fallidos.length > 0) alerta('Algunas claves no se generaron', fallidos.join('\n'));
  }

  async function reenviar(c: ClaveEmitida) {
    if (!token) return;
    setOcupado(c.id);
    try {
      await reenviarClave(c.id, token);
      alerta('Correo reenviado', `Se volvió a enviar la clave a ${c.email}.`);
    } catch (error) {
      manejarError(error);
      alerta('No se pudo reenviar', traducirError(error));
    } finally {
      setOcupado(null);
    }
  }

  async function anular(c: ClaveEmitida) {
    const ok = await confirmar({ titulo: 'Anular clave', mensaje: `La clave ${c.code} dejará de funcionar. ¿Continuar?`, textoConfirmar: 'Anular', icono: 'key' });
    if (!ok || !token) return;
    try {
      await revocarClave(c.code, token);
      await actualizarLista(claves.map((x) => (x.id === c.id ? { ...x, revocada: true } : x)));
    } catch (error) {
      manejarError(error);
      alerta('No se pudo anular', traducirError(error));
    }
  }

  async function buscar() {
    if (!token) return;
    if (!rutEsValido(rut)) return setErrorRut('RUT inválido. Revisa el dígito verificador');
    setErrorRut(undefined);
    setOcupado('buscar');
    setEncontrado(null);
    try {
      setEncontrado(await buscarPorRut(rut, token));
    } catch (error) {
      manejarError(error);
      const noExiste = error instanceof ApiError && error.status === 404;
      alerta(noExiste ? 'Sin resultados' : 'No se pudo buscar', noExiste ? 'No hay un emprendedor registrado con ese RUT.' : traducirError(error));
    } finally {
      setOcupado(null);
    }
  }

  const texto = { fontFamily: fonts.normal, fontSize: e(15), color: colors.lavanda };
  const titulo = (icono: string, t: string) => (
    <View style={{ flexDirection: 'row', alignItems: 'center', gap: 8 }}>
      <Icono nombre={icono} tamano={e(24)} color={colors.lavanda} />
      <Text style={{ fontFamily: fonts.negrita, fontSize: e(20), color: colors.lavanda }}>{t}</Text>
    </View>
  );

  return (
    <KeyboardAvoidingView style={{ flex: 1, backgroundColor: colors.fondoApp }} behavior={Platform.OS === 'ios' ? 'padding' : 'height'}>
      <ScrollView contentContainerStyle={{ paddingHorizontal: 20, paddingTop: insets.top + 16, paddingBottom: insets.bottom + 32, gap: 18 }} keyboardShouldPersistTaps="handled">
        <Encabezado
          titulo="Panel municipal"
          subtitulo="Activación de emprendedores"
          volver={false}
          derecha={<Pressable hitSlop={10} onPress={cerrarSesion}><Icono nombre="log-out" tamano={e(30)} color={colors.lavanda} /></Pressable>}
        />

        {/* Generar clave */}
        <View style={tarjeta}>
          {titulo('key', 'Generar claves de acceso')}
          <Text style={texto}>Cada clave es de un solo uso, vence en un año y se envía al correo del emprendedor. Puedes generar varias a la vez: escribe un correo por línea o sepáralos con comas.</Text>
          <CampoTexto etiqueta="Correos de los emprendedores" value={correo} onChangeText={setCorreo} keyboardType="email-address"
            autoCapitalize="none" placeholder={'correo1@ejemplo.cl\ncorreo2@ejemplo.cl'} error={errorCorreo}
            multiline style={{ minHeight: 90, textAlignVertical: 'top' }} />
          <Boton titulo={progreso ?? 'Generar y enviar claves'} onPress={generar} deshabilitado={!correo.trim() || ocupado !== null} />
          {generadas.length > 0 && (
            <View style={{ backgroundColor: colors.verdeSuave, borderRadius: 16, padding: 14, gap: 8 }}>
              <Text style={{ ...texto, fontFamily: fonts.negrita }}>
                {generadas.length === 1 ? 'Clave generada' : `${generadas.length} claves generadas`} y enviadas por correo:
              </Text>
              {generadas.map((g) => (
                <View key={g.id} style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', gap: 8 }}>
                  <Text numberOfLines={1} style={{ ...texto, flex: 1 }}>{g.email}</Text>
                  <Text selectable style={{ fontFamily: fonts.titulo, fontSize: e(28), color: colors.verde, letterSpacing: 3 }}>{g.code}</Text>
                </View>
              ))}
              <Text style={{ ...texto, fontSize: e(13) }}>Mantén presionada una clave para copiarla</Text>
            </View>
          )}
        </View>

        {/* Claves emitidas */}
        <View style={tarjeta}>
          {titulo('list', 'Claves emitidas')}
          {/* Exportar a Excel para que el municipio las distribuya (Historia 009) */}
          {claves.length > 0 && (
            <Boton
              titulo={ocupado === 'exportar' ? 'Generando Excel...' : 'Exportar claves (Excel)'}
              variante="secundario"
              deshabilitado={ocupado !== null}
              onPress={async () => {
                setOcupado('exportar');
                try {
                  await exportarClaves(claves);
                } catch (error) {
                  alerta('No se pudo exportar', traducirError(error));
                } finally {
                  setOcupado(null);
                }
              }}
            />
          )}
          {claves.length === 0 ? (
            <Text style={texto}>Aún no has generado claves desde este dispositivo.</Text>
          ) : claves.map((c) => (
            <View key={c.id} style={{ borderRadius: 14, borderWidth: 1.5, borderColor: colors.lavandaSuave, padding: 12, gap: 8 }}>
              <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
                <Text selectable style={{ fontFamily: fonts.negrita, fontSize: e(18), color: c.revocada ? colors.textoSecundario : colors.lavanda, letterSpacing: 2 }}>{c.code}</Text>
                <Text style={{ fontFamily: fonts.medio, fontSize: e(13), color: c.revocada ? colors.rosado : colors.verde }}>
                  {c.revocada ? 'Anulada' : `Vence ${fecha(c.expires_at)}`}
                </Text>
              </View>
              <Text numberOfLines={1} style={texto}>{c.email}</Text>
              {!c.revocada && (
                <View style={{ flexDirection: 'row', gap: 10 }}>
                  <View style={{ flex: 1 }}>
                    <Boton titulo={ocupado === c.id ? 'Enviando...' : 'Reenviar'} variante="secundario" onPress={() => reenviar(c)} deshabilitado={ocupado !== null} />
                  </View>
                  <View style={{ flex: 1 }}>
                    <Boton titulo="Anular" variante="secundario" color={colors.error} onPress={() => anular(c)} />
                  </View>
                </View>
              )}
            </View>
          ))}
        </View>

        {/* Buscar emprendedor */}
        <View style={tarjeta}>
          {titulo('search', 'Buscar emprendedor')}
          <CampoTexto etiqueta="RUT" value={rut} onChangeText={setRut} placeholder="12.345.678-5" autoCapitalize="characters" error={errorRut} maxLength={12} />
          <Boton titulo={ocupado === 'buscar' ? 'Buscando...' : 'Buscar'} onPress={buscar} deshabilitado={!rut.trim() || ocupado !== null} variante="secundario" />
          {encontrado && (
            <View style={{ backgroundColor: colors.lavandaSuave, borderRadius: 16, padding: 14, gap: 4 }}>
              <Text style={{ fontFamily: fonts.negrita, fontSize: e(18), color: colors.lavanda }}>{encontrado.nombre}</Text>
              <Text style={texto}>RUT: {formatoRut(encontrado.rut)}</Text>
              <Text style={texto}>Correo: {encontrado.email}</Text>
              <Text style={texto}>Registrado el {fecha(encontrado.created_at)}</Text>
              <Text style={texto}>Acceso vigente hasta {fecha(encontrado.plan_expires_at)}</Text>
            </View>
          )}
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
