// Datos ficticios para probar la app durante el desarrollo.
// __DEV__ es true solo en desarrollo: en la app final los campos aparecerán vacíos.

const datos = {
    nombre: 'Juan Pérez',
    rut: '12.345.678-5',
    emprendimiento: 'Dulces Juanita',
    correo: 'thiare@prueba.cl',
    contrasena: 'Tally2026',
    claveActivacion: '',
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