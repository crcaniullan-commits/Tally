import { ApiError } from '../../shared/api/cliente';

const traducciones: Record<string, string> = {
    'a user with that email already exists': 'Este correo ya está registrado',
    'a user with that rut already exists': 'Este RUT ya está registrado',
    'resource not found': 'Correo o contraseña incorrectos',
    'crypto/bcrypt: hashedPassword is not the hash of the given password': 'Correo o contraseña incorrectos',
    'the code was already exchanged': 'Esta clave ya fue utilizada',
    'the code is expired': 'Esta clave está vencida. Solicita una nueva a tu municipalidad',
    'the code was revoked': 'Esta clave fue anulada. Solicita una nueva a tu municipalidad',
    'not found': 'La clave no es válida. Revisa que la copiaste completa',
    "The code can't be Redeemed": 'La clave no es válida o ya fue utilizada',
    'the server encountered a problem': 'Ocurrió un problema en el servidor. Intenta nuevamente en unos minutos',
    unauthorized: 'Tu sesión expiró. Vuelve a intentarlo',
};

export function traducirError(error: unknown): string {
    if (error instanceof ApiError) {
        return traducciones[error.message] ?? error.message;
    }
    return 'Ocurrió un error inesperado';
}