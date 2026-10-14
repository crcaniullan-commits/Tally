    import { peticion } from '../../shared/api/cliente';
    import { DatosRegistro } from './components/FormularioRegistro';

    // POST /v1/auth → crea la cuenta y devuelve el token de sesión
    export function registrar(datos: DatosRegistro): Promise<string> {
    return peticion<string>('/auth', {
        metodo: 'POST',
        cuerpo: {
        email: datos.correo.trim().toLowerCase(),
        password: datos.contrasena,
        nombre: datos.nombre.trim(),
        rut: datos.rut.replace(/\./g, '').trim().toUpperCase(), // Sin puntos: el backend acepta máximo 10 caracteres
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

    // POST /v1/app/users/exchange/{clave} → canjea la clave municipal (requiere sesión)
    export function canjearClave(clave: string, token: string): Promise<unknown> {
    return peticion(`/app/users/exchange/${clave}`, { metodo: 'POST', token });
    }