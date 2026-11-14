import { createContext, ReactNode, useCallback, useContext, useRef, useState } from 'react';
import { Modal, Pressable, Text, View } from 'react-native';
import { colors } from '../theme/colors';
import { fonts } from '../theme/fonts';
import { useEscala } from '../utils/escala';
import Icono from './Icono';

// Confirmación con la estética de Tally (reemplaza al Alert del sistema para eliminar cosas).
// Uso: const confirmar = useConfirmar(); if (await confirmar({ titulo, mensaje })) { ... }
// tono 'peligro' = eliminar (rosado); 'aviso' = solo informa, con un botón "Entendido" (lavanda)
type Pedido = { titulo: string; mensaje: string; textoConfirmar?: string; icono?: string; tono?: 'peligro' | 'aviso' };
const Contexto = createContext<(p: Pedido) => Promise<boolean>>(async () => false);
export const useConfirmar = () => useContext(Contexto);

// Avisos y errores con la misma estética (reemplaza a Alert.alert): alerta('No se pudo guardar', 'Revisa tu conexión')
export function useAlerta() {
  const pedir = useContext(Contexto);
  return useCallback((titulo: string, mensaje = '') => { pedir({ titulo, mensaje, tono: 'aviso' }); }, [pedir]);
}

export function ConfirmarProvider({ children }: { children: ReactNode }) {
  const { t } = useEscala();
  const [pedido, setPedido] = useState<Pedido | null>(null);
  const resolver = useRef<(v: boolean) => void>(() => {});

  const confirmar = useCallback((p: Pedido) => new Promise<boolean>((res) => {
    resolver.current = res;
    setPedido(p);
  }), []);

  const aviso = pedido?.tono === 'aviso';

  function responder(v: boolean) {
    setPedido(null);
    resolver.current(v);
  }

  return (
    <Contexto.Provider value={confirmar}>
      {children}
      <Modal visible={!!pedido} transparent animationType="fade" onRequestClose={() => responder(false)} statusBarTranslucent>
        <View style={{ flex: 1, justifyContent: 'center', padding: t(28) }}>
          <Pressable onPress={() => responder(false)} style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(60, 56, 82, 0.45)' }} />
          <View style={{ backgroundColor: colors.superficie, borderRadius: t(20), padding: t(20), alignItems: 'center', gap: t(10) }}>
            <View style={{ width: t(48), height: t(48), borderRadius: 999, backgroundColor: aviso ? colors.lavandaSuave : colors.rosadoSuave, alignItems: 'center', justifyContent: 'center' }}>
              <Icono nombre={pedido?.icono ?? (aviso ? 'info' : 'trash')} tamano={t(22)} color={aviso ? colors.lavanda : colors.rosado} />
            </View>
            <Text style={{ fontFamily: fonts.extra, fontSize: t(15), color: colors.texto, textAlign: 'center' }}>{pedido?.titulo}</Text>
            <Text style={{ fontFamily: fonts.medio, fontSize: t(12), color: colors.textoSecundario, textAlign: 'center' }}>{pedido?.mensaje}</Text>
            <View style={{ flexDirection: 'row', gap: t(8), marginTop: t(6), alignSelf: 'stretch' }}>
              {aviso ? (
                <Pressable onPress={() => responder(true)} style={{ flex: 1, alignItems: 'center', paddingVertical: t(10), borderRadius: t(12), backgroundColor: colors.lavanda }}>
                  <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.superficie }}>Entendido</Text>
                </Pressable>
              ) : (
                <>
                  <Pressable onPress={() => responder(false)} style={{ flex: 1, alignItems: 'center', paddingVertical: t(10), borderRadius: t(12), backgroundColor: colors.lavandaSuave }}>
                    <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.lavanda }}>Cancelar</Text>
                  </Pressable>
                  <Pressable onPress={() => responder(true)} style={{ flex: 1, alignItems: 'center', paddingVertical: t(10), borderRadius: t(12), backgroundColor: colors.rosado }}>
                    <Text style={{ fontFamily: fonts.extra, fontSize: t(12), color: colors.superficie }}>{pedido?.textoConfirmar ?? 'Eliminar'}</Text>
                  </Pressable>
                </>
              )}
            </View>
          </View>
        </View>
      </Modal>
    </Contexto.Provider>
  );
}
