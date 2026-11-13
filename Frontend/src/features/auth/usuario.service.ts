import { peticion } from '../../shared/api/cliente';

// PATCH /v1/app/users/ → cambia el nombre y/o la contraseña del usuario
export function actualizarUsuario(datos: { nombre?: string; contrasena?: string }, token: string): Promise<unknown> {
  return peticion('/app/users/', {
    metodo: 'PATCH',
    cuerpo: {
      ...(datos.nombre ? { nombre: datos.nombre.trim() } : {}),
      ...(datos.contrasena ? { password: datos.contrasena } : {}),
    },
    token,
  });
}

// DELETE /v1/app/users/ → elimina la cuenta del usuario
export function eliminarCuenta(token: string): Promise<unknown> {
  return peticion('/app/users/', { metodo: 'DELETE', token });
}
