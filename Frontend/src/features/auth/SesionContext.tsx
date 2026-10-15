    import * as SecureStore from 'expo-secure-store';
    import { createContext, ReactNode, useContext, useEffect, useState } from 'react';

    const CLAVE_TOKEN = 'tally_token';

    type Sesion = {
    token: string | null;
    cargando: boolean;
    iniciarSesion: (token: string) => Promise<void>;
    cerrarSesion: () => Promise<void>;
    };

    const SesionContext = createContext<Sesion | null>(null);

    export function SesionProvider({ children }: { children: ReactNode }) {
    const [token, setToken] = useState<string | null>(null);
    const [cargando, setCargando] = useState(true);

    // Al abrir la app, recupera la sesión guardada (si existe)
    useEffect(() => {
        SecureStore.getItemAsync(CLAVE_TOKEN)
        .then((guardado) => setToken(guardado))
        .finally(() => setCargando(false));
    }, []);

    async function iniciarSesion(nuevoToken: string) {
        await SecureStore.setItemAsync(CLAVE_TOKEN, nuevoToken);
        setToken(nuevoToken);
    }

    async function cerrarSesion() {
        await SecureStore.deleteItemAsync(CLAVE_TOKEN);
        setToken(null);
    }

    return (
        <SesionContext.Provider value={{ token, cargando, iniciarSesion, cerrarSesion }}>
        {children}
        </SesionContext.Provider>
    );
    }

    export function useSesion() {
    const contexto = useContext(SesionContext);
    if (!contexto) throw new Error('useSesion debe usarse dentro de SesionProvider');
    return contexto;
    }