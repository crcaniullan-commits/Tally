// Utilidades de fechas en hora local de Chile (sin librerías externas)

export const MESES = ['enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio', 'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'];
const DIAS = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'];

const dos = (n: number) => String(n).padStart(2, '0');

// Date → "2026-11-02" (formato del backend)
export const aISO = (f: Date) => `${f.getFullYear()}-${dos(f.getMonth() + 1)}-${dos(f.getDate())}`;

// "2026-11-02" → Date local (a medianoche, sin desfase por zona horaria)
export function desdeISO(texto: string): Date {
  const [a, m, d] = texto.slice(0, 10).split('-').map(Number);
  return new Date(a, m - 1, d);
}

// Date → "02/11/2026"
export const aDMA = (f: Date) => `${dos(f.getDate())}/${dos(f.getMonth() + 1)}/${f.getFullYear()}`;

// Date → "lunes 2 de noviembre"
export const aLargo = (f: Date) => `${DIAS[f.getDay()]} ${f.getDate()} de ${MESES[f.getMonth()]}`;

export function sumarDias(f: Date, dias: number): Date {
  return new Date(f.getFullYear(), f.getMonth(), f.getDate() + dias);
}

export const hoy = () => {
  const f = new Date();
  return new Date(f.getFullYear(), f.getMonth(), f.getDate());
};

// Días que faltan desde hoy hasta la fecha (negativo = ya pasó)
export const diasHasta = (f: Date) => Math.round((f.getTime() - hoy().getTime()) / 86400000);

// "2026-11" → clave del mes actual (para la meta mensual)
export const claveMes = (f = new Date()) => `${f.getFullYear()}-${dos(f.getMonth() + 1)}`;
