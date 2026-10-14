import { Text, View } from 'react-native';
import { colors } from '../constants/colors';
import { fonts } from '../constants/fonts';

type BarraPasosProps = {
    actual: number;
    total: number;
    descripcion: string;
};

export default function BarraPasos({ actual, total, descripcion }: BarraPasosProps) {
    return (
        <View style={{ gap: 6 }}>
        <View style={{ flexDirection: 'row', gap: 6 }}>
            {Array.from({ length: total }, (_, i) => (
            <View
                key={i}
                style={{
                flex: 1,
                height: 6,
                borderRadius: 3,
                backgroundColor: i < actual ? colors.lavanda : colors.lavandaSuave,
                }}
            />
            ))}
        </View>
        <Text style={{ color: colors.textoSecundario, fontFamily: fonts.medio, fontSize: 13 }}>
            Paso {actual} de {total}: {descripcion}
        </Text>
        </View>
    );
}