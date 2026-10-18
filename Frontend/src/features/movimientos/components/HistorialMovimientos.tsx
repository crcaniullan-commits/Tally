    import { MaterialCommunityIcons } from '@expo/vector-icons';
    import { Text, View } from 'react-native';
    import { colors } from '../../../shared/theme/colors';
    import { fonts } from '../../../shared/theme/fonts';
    import { sombras } from '../../../shared/theme/sombras';
    import { formatearMonto } from '../../../shared/utils/formato';
    import { TipoMovimiento } from '../../categorias/categorias';

    // Un movimiento ya registrado (ingreso o egreso)
    export type Movimiento = {
    id: string;
    tipo: TipoMovimiento;
    monto: number;
    categoria: string;
    icono: string;
    color: string;
    colorSuave: string;
    metodo: string;
    nota: string;
    fecha: Date;
    };

    const MESES = ['ene', 'feb', 'mar', 'abr', 'may', 'jun', 'jul', 'ago', 'sep', 'oct', 'nov', 'dic'];

    // Ej: "18 oct · 18:45" (se arma a mano porque Android no siempre trae los formatos en español)
    function formatearFecha(f: Date) {
    const hora = `${f.getHours()}:${String(f.getMinutes()).padStart(2, '0')}`;
    return `${f.getDate()} ${MESES[f.getMonth()]} · ${hora}`;
    }

    export default function HistorialMovimientos({ movimientos }: { movimientos: Movimiento[] }) {
    return (
        <View style={{ gap: 12 }}>
        <Text style={{ fontFamily: fonts.negrita, fontSize: 19, color: colors.lavanda }}>Historial de movimientos</Text>

        {movimientos.length === 0 ? (
            <Text style={{ fontFamily: fonts.normal, fontSize: 15, color: colors.lavanda, textAlign: 'center', paddingVertical: 16 }}>
            Aún no tienes movimientos registrados.
            </Text>
        ) : (
            movimientos.map((m) => {
            const esIngreso = m.tipo === 'ingreso';
            return (
                <View
                key={m.id}
                style={{ flexDirection: 'row', alignItems: 'center', gap: 14, backgroundColor: colors.superficie, borderRadius: 18, padding: 14, ...sombras.suave }}
                >
                <View style={{ width: 48, height: 48, borderRadius: 14, backgroundColor: m.colorSuave, alignItems: 'center', justifyContent: 'center' }}>
                    <MaterialCommunityIcons name={m.icono as any} size={26} color={m.color} />
                </View>

                <View style={{ flex: 1 }}>
                    <Text numberOfLines={1} style={{ fontFamily: fonts.medio, fontSize: 16, color: colors.lavanda }}>
                    {m.nota || m.categoria}
                    </Text>
                    <Text numberOfLines={1} style={{ fontFamily: fonts.normal, fontSize: 13, color: colors.lavanda, opacity: 0.75 }}>
                    {m.categoria} · {m.metodo} · {formatearFecha(m.fecha)}
                    </Text>
                </View>

                <Text style={{ fontFamily: fonts.negrita, fontSize: 17, color: esIngreso ? colors.ingreso : colors.egreso }}>
                    {esIngreso ? '+' : '−'}${formatearMonto(m.monto)}
                </Text>
                </View>
            );
            })
        )}
        </View>
    );
    }