    import { ApiError, peticion } from '../../shared/api/cliente';
    import { DatosRegistro } from './components/FormularioRegistro';

    // POST /v1/auth → crea la cuenta y devuelve el token de sesión
    export function registrar(datos: DatosRegistro, clave: string): Promise<string> {
    return peticion<string>('/auth', {
        metodo: 'POST',
        cuerpo: {
        email: datos.correo.trim().toLowerCase(),
        password: datos.contrasena,
        nombre: datos.nombre.trim(),
        rut: datos.rut.replace(/\./g, '').trim().toUpperCase(), // Sin puntos: el backend acepta máximo 10 caracteres
        code: clave,
        // Pendiente backend: "nombre del emprendimiento" (el backend rechaza campos que no conoce)
        },
    });
    }

    // POST /v1/auth/login → devuelve el token de sesión
    export function iniciarSesion(correo: string, contrasena: string): Promise<string> {
    return peticion<string>('/auth/login', {
        metodo: 'POST',
        cuerpo: { email: correo.trim().toLowerCase(), password: contrasena },
    });
    }


export type Rol = 'usuario' | 'municipal';

// Averigua el rol de la cuenta (el token no lo trae): se consulta una ruta solo para municipalidades.
// Si responde 403 (prohibido) es un emprendedor; cualquier otra respuesta significa que sí tiene acceso municipal.
export async function detectarRol(token: string): Promise<Rol> {
  try {
    await peticion('/app/users/municipal/11111111-1', { token });
    return 'municipal';
  } catch (error) {
    if (error instanceof ApiError && error.status === 403) return 'usuario';
    if (error instanceof ApiError && error.status === 401) throw error; // Token inválido
    return 'municipal'; // 404 (RUT de prueba no existe) = pasó el control de rol
  }
}
