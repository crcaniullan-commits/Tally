import { ApiError } from '../../shared/api/cliente';

const traducciones: Record<string, string> = {
    // Mensajes del backend en español (desde el PR #12)
    'ya existe un usuario con ese correo': 'Este correo ya está registrado',
    'ya existe un usuario con ese RUT': 'Este RUT ya está registrado',
    'recurso no encontrado': 'Correo o contraseña incorrectos',
    'El código no se puede canjear': 'La clave no es válida, ya fue utilizada o está vencida',
    'el servidor encontró un problema': 'Ocurrió un problema en el servidor. Intenta nuevamente en unos minutos',
    'no autorizado': 'Tu sesión expiró. Vuelve a iniciar sesión',
    prohibido: 'No tienes permiso para realizar esta acción',
    'no encontrado': 'No se encontró lo que buscabas',
    'la categoría tiene movimientos asociados y no se puede eliminar': 'Esta categoría tiene registros y no se puede eliminar',
    'el dígito verificador del RUT no es matemáticamente correcto': 'El RUT no es válido. Revisa el dígito verificador',
    'el formato del RUT no es válido': 'El RUT no es válido',
    'el código de recuperación no es válido': 'El código no es correcto. Revísalo e intenta de nuevo',
    'el código de recuperación todavía no fue verificado': 'Primero debes verificar el código',
    'el código de recuperación ya fue usado': 'Este código ya fue usado. Pide uno nuevo',
    'el código de recuperación venció': 'El código venció. Pide uno nuevo',
    'la fecha debe tener el formato AAAA-MM-DD': 'La fecha no es válida',
    'el ingreso no fue encontrado': 'Este ingreso ya no existe',
    'el gasto no fue encontrado': 'Este egreso ya no existe',
    // Mensajes anteriores (en inglés), por si el backend local aún no está actualizado
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
    'La categoría ya existe': 'Ya existe una categoría con ese nombre',
    forbidden: 'No tienes permiso para realizar esta acción',
    'the request was forbidden': 'No tienes permiso para realizar esta acción',
    'Correo reenviado': 'Correo reenviado',
    'Codigo desabilitado': 'Clave anulada',
};

export function traducirError(error: unknown): string {
    if (error instanceof ApiError) {
        if (error.status === 429) return 'Hiciste demasiados intentos. Espera un momento y vuelve a intentar';
        return traducciones[error.message] ?? error.message;
    }
    return 'Ocurrió un error inesperado';
}