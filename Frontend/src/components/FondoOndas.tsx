import Svg, { Path } from 'react-native-svg';
import { colors } from '../constants/colors';

// Forma ondulada decorativa del fondo, según el mockup de la Bienvenida.
// Se estira para ocupar toda la pantalla, sin importar su tamaño.
export default function FondoOndas() {
    return (
        <Svg
        width="100%"
        height="100%"
        viewBox="0 0 268 560"
        preserveAspectRatio="none"
        style={{ position: 'absolute', top: 0, left: 0 }}
        >
        <Path
            d="M268,55 C225,45 195,70 185,115 C172,175 130,195 75,190 C30,186 0,205 0,245 L0,490 C30,480 55,450 70,420 C95,370 150,345 210,330 C240,322 260,300 268,275 Z"
            fill={colors.lavandaOnda}
        />
        </Svg>
    );
}