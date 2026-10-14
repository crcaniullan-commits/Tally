// Datos ficticios para probar la app durante el desarrollo.
// __DEV__ es true solo en desarrollo: en la app final los campos aparecerán vacíos.

const datos = {
    nombre: 'Juan Pérez',
    rut: '11.111.111-1',
    emprendimiento: 'Dulces Juanita',
    correo: 'juan.perez@ejemplo.cl',
    contrasena: 'Tally2026',
    claveActivacion: 'JBSWY3DPEHPK3PXPJBSWY3DPEH',
};

const vacios = {
    nombre: '',
    rut: '',
    emprendimiento: '',
    correo: '',
    contrasena: '',
    claveActivacion: '',
};

export const datosPrueba = __DEV__ ? datos : vacios;