import { peticion } from '../../shared/api/cliente';
import { guardarLocal, leerLocal } from '../../shared/utils/almacen';

// Backoffice municipal (Historia 009): solo cuentas con rol municipal

export type ClaveAcceso = {
  id: string;
  code: string;
  redeemed_at: string | null;
  expires_at: string | null;
  created_at: string;
};

export type ClaveEmitida = ClaveAcceso & { email: string; revocada?: boolean };

export type UsuarioMunicipal = {
  id: string;
  email: string;
  nombre: string;
  rut: { Cuerpo: number; DV: string };
  created_at: string;
  plan_expires_at: string | null;
};

// POST /v1/app/key/ → genera una clave de un solo uso y la envía al correo
export function emitirClave(email: string, token: string): Promise<ClaveAcceso> {
  return peticion<ClaveAcceso>('/app/key/', { metodo: 'POST', cuerpo: { email: email.trim().toLowerCase() }, token });
}

// POST /v1/app/key/AccessID/{id} → reenvía el correo con la clave
export function reenviarClave(id: string, token: string): Promise<string> {
  return peticion<string>(`/app/key/AccessID/${id}`, { metodo: 'POST', token });
}

// POST /v1/app/key/revoke/{code} → anula una clave que aún no se usa
export function revocarClave(code: string, token: string): Promise<string> {
  return peticion<string>(`/app/key/revoke/${code}`, { metodo: 'POST', token });
}

// GET /v1/app/users/municipal/{rut} → busca un emprendedor por RUT (sin datos financieros)
export function buscarPorRut(rut: string, token: string): Promise<UsuarioMunicipal> {
  const limpio = rut.replace(/\./g, '').trim().toUpperCase();
  return peticion<UsuarioMunicipal>(`/app/users/municipal/${encodeURIComponent(limpio)}`, { token });
}

// Historial de claves emitidas desde este celular (el backend aún no tiene un listado)
const CLAVE_LOCAL = 'claves_emitidas';
export const listarClavesEmitidas = () => leerLocal<ClaveEmitida[]>(CLAVE_LOCAL, []);
export const guardarClavesEmitidas = (lista: ClaveEmitida[]) => guardarLocal(CLAVE_LOCAL, lista);
