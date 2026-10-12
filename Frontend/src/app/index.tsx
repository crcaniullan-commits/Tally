import { Text, View } from 'react-native';
import { colors } from '../constants/colors';

export default function Index() {
  return (
    <View
      style={{
        flex: 1,
        backgroundColor: colors.fondo,
        justifyContent: 'center',
        alignItems: 'center',
      }}
    >
      <Text style={{ color: colors.primario, fontSize: 40, fontWeight: 'bold' }}>
        Tally
      </Text>
      <Text style={{ color: colors.textoSecundario, marginTop: 8 }}>
        Finanzas simples para emprendedores
      </Text>
    </View>
  );
}