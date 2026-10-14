# Tally — Frontend

App móvil de Tally para emprendedores. Construida con **React Native + Expo** (Expo Router) y **TypeScript**.

## Cómo ejecutar

```bash
cd Frontend
npm install
npx expo start
```

Escanear el código QR con la app **Expo Go** (el celular y el computador deben estar en la misma red WiFi).

## Arquitectura

El frontend usa una **arquitectura modular por funcionalidad**, la misma idea que el backend (`Backend/internal/`): cada módulo de negocio tiene su propia carpeta.

```
src/
├── app/          Rutas (pantallas). Cada archivo es una pantalla (Expo Router).
├── features/     Módulos de negocio
│   ├── auth/         Registro, login, activación con clave municipal, recuperar contraseña
│   └── categorias/   Categorías de movimientos
└── shared/       Código compartido por todos los módulos
    ├── components/   Componentes de interfaz reutilizables (Boton, CampoTexto, ModalFlotante…)
    ├── theme/        Colores, fuentes y sombras
    ├── utils/        Funciones de apoyo (validación de RUT)
    └── mocks/        Datos de prueba (solo se usan en desarrollo)
```

**Regla para ubicar un archivo nuevo:**
- ¿Es una pantalla? → `app/`
- ¿Pertenece a un módulo de negocio? → `features/<modulo>/`
- ¿Lo usan varios módulos? → `shared/`

### Relación con los módulos del backend

| Frontend | Backend | Historias |
|---|---|---|
| `features/auth` | `auth`, `users`, `accessKeys` | 001, 002, 010, 011 |
| `features/categorias` | `categories` | 003 |

## Convenciones

- **Colores, fuentes y sombras** se toman siempre de `shared/theme`. No escribir códigos de color directamente en las pantallas.
- **Montos de dinero:** números enteros en pesos chilenos (CLP no usa decimales), igual que el backend.
- **Datos de prueba:** `shared/mocks/datosPrueba.ts` usa `__DEV__`, por lo que en la app final los formularios aparecen vacíos.
- Las partes que esperan un endpoint del backend están marcadas en el código con el comentario `// Pendiente`.