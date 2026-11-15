    import { API } from '../../shared/api/rutas';
import { ApiError, peticion } from '../../shared/api/cliente';
    import { DatosRegistro } from './components/FormularioRegistro';

    // POST /v1/autenticacion → crea la cuenta y devuelve el token de sesión
    export function registrar(datos: DatosRegistro, clave: string): Promise<string> {
    return peticion<string>(API.registro, {
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

    // POST /v1/autenticacion/inicio-sesion → devuelve el token de sesión
    export function iniciarSesion(correo: string, contrasena: string): Promise<string> {
    return peticion<string>(API.login, {
        metodo: 'POST',
        cuerpo: { email: correo.trim().toLowerCase(), password: contrasena },
    });
    }


// Recuperar contraseña (Historia 011): 1) pedir código al correo, 2) verificarlo, 3) guardar la nueva contraseña
const correoLimpio = (correo: string) => correo.trim().toLowerCase();

export function solicitarRecuperacion(correo: string): Promise<unknown> {
  return peticion(API.recuperarSolicitar, { metodo: 'POST', cuerpo: { email: correoLimpio(correo) } });
}

export function verificarCodigoRecuperacion(correo: string, codigo: string): Promise<unknown> {
  return peticion(API.recuperarVerificar, { metodo: 'POST', cuerpo: { email: correoLimpio(correo), codigo: codigo.trim() } });
}

export function confirmarRecuperacion(correo: string, codigo: string, contrasena: string): Promise<unknown> {
  return peticion(API.recuperarConfirmar, {
    metodo: 'POST',
    cuerpo: { email: correoLimpio(correo), codigo: codigo.trim(), password: contrasena },
  });
}

export type Rol = 'usuario' | 'municipal';

// Averigua el rol de la cuenta (el token no lo trae): se consulta una ruta solo para municipalidades.
// Si responde 403 (prohibido) es un emprendedor; cualquier otra respuesta significa que sí tiene acceso municipal.
export async function detectarRol(token: string): Promise<Rol> {
  try {
    await peticion(API.usuarioPorRut('11111111-1'), { token });
    return 'municipal';
  } catch (error) {
    if (error instanceof ApiError && error.status === 403) return 'usuario';
    if (error instanceof ApiError && error.status === 401) throw error; // Token inválido
    return 'municipal'; // 404 (RUT de prueba no existe) = pasó el control de rol
  }
}
