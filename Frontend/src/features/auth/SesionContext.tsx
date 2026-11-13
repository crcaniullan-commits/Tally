import * as SecureStore from 'expo-secure-store';
import { createContext, ReactNode, useCallback, useContext, useEffect, useState } from 'react';
import { ApiError } from '../../shared/api/cliente';
import { Empresa, obtenerOCrearEmpresa } from '../empresa/empresa.service';
import { detectarRol, Rol } from './auth.service';

const CLAVE_TOKEN = 'tally_token';
const CLAVE_ROL = 'tally_rol';

type Sesion = {
  token: string | null;
  rol: Rol | null;          // usuario (emprendedor) o municipal
  empresa: Empresa | null;  // Empresa activa del emprendedor: se carga sola al iniciar sesión
  cargando: boolean;
  iniciarSesion: (token: string) => Promise<void>;
  cerrarSesion: () => Promise<void>;
  /** Llamar cuando el backend responde 401: la sesión venció (el token dura 3 días) */
  manejarError: (error: unknown) => void;
  /** Actualiza la empresa en memoria (por ejemplo, al cambiarle el nombre en Configuración) */
  actualizarEmpresa: (empresa: Empresa) => void;
};

const SesionContext = createContext<Sesion | null>(null);

export function SesionProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(null);
  const [rol, setRol] = useState<Rol | null>(null);
  const [empresa, setEmpresa] = useState<Empresa | null>(null);
  const [cargando, setCargando] = useState(true);

  // Al abrir la app, recupera la sesión guardada (si existe)
  useEffect(() => {
    Promise.all([SecureStore.getItemAsync(CLAVE_TOKEN), SecureStore.getItemAsync(CLAVE_ROL)])
      .then(([t, r]) => {
        if (t && (r === 'usuario' || r === 'municipal')) {
          setToken(t);
          setRol(r);
        }
      })
      .finally(() => setCargando(false));
  }, []);

  const cerrarSesion = useCallback(async () => {
    await SecureStore.deleteItemAsync(CLAVE_TOKEN);
    await SecureStore.deleteItemAsync(CLAVE_ROL);
    setToken(null);
    setRol(null);
    setEmpresa(null);
  }, []);

  const manejarError = useCallback((error: unknown) => {
    if (error instanceof ApiError && error.status === 401) cerrarSesion();
  }, [cerrarSesion]);

  // Si es emprendedor, carga (o crea) su empresa
  useEffect(() => {
    if (!token || rol !== 'usuario') return;
    obtenerOCrearEmpresa(token).then(setEmpresa).catch(manejarError);
  }, [token, rol, manejarError]);

  async function iniciarSesion(nuevoToken: string) {
    const nuevoRol = await detectarRol(nuevoToken);
    await SecureStore.setItemAsync(CLAVE_TOKEN, nuevoToken);
    await SecureStore.setItemAsync(CLAVE_ROL, nuevoRol);
    setRol(nuevoRol);
    setToken(nuevoToken);
  }

  return (
    <SesionContext.Provider
      value={{ token, rol, empresa, cargando, iniciarSesion, cerrarSesion, manejarError, actualizarEmpresa: setEmpresa }}
    >
      {children}
    </SesionContext.Provider>
  );
}

export function useSesion() {
  const contexto = useContext(SesionContext);
  if (!contexto) throw new Error('useSesion debe usarse dentro de SesionProvider');
  return contexto;
}
