    // Cliente HTTP de Tally: todas las peticiones al backend pasan por aquí.
    // Formato del backend: éxito → { "data": ... } · error → { "error": "mensaje" }

    const API_URL = process.env.EXPO_PUBLIC_API_URL;

    export class ApiError extends Error {
    status: number;
    constructor(mensaje: string, status: number) {
        super(mensaje);
        this.status = status;
    }
    }

    type Opciones = {
    metodo?: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';
    cuerpo?: unknown;
    token?: string;
    };

    export async function peticion<T>(ruta: string, { metodo = 'GET', cuerpo, token }: Opciones = {}): Promise<T> {
    if (!API_URL) {
        throw new ApiError('Falta configurar EXPO_PUBLIC_API_URL en el archivo .env.local', 0);
    }

    const url = `${API_URL}${ruta}`;
        if (__DEV__) console.log(`[API] ${metodo} ${url}`);

        // Si el servidor no responde en 15 segundos, se cancela el pedido
        const controlador = new AbortController();
        const temporizador = setTimeout(() => controlador.abort(), 15000);

        let respuesta: Response;
        try {
            respuesta = await fetch(url, {
            method: metodo,
            headers: {
                'Content-Type': 'application/json',
                ...(token ? { Authorization: `Bearer ${token}` } : {}),
            },
            body: cuerpo !== undefined ? JSON.stringify(cuerpo) : undefined,
            signal: controlador.signal,
            });
        } catch (e) {
            if (__DEV__) console.log('[API] Error de conexión:', e);
            throw new ApiError('No se pudo conectar con el servidor. Revisa tu conexión a internet.', 0);
        } finally {
            clearTimeout(temporizador);
        }

        if (__DEV__) console.log(`[API] Respuesta ${respuesta.status}`);

    let json: { data?: T; error?: string } | null = null;
    try {
        json = await respuesta.json();
    } catch {
        // La respuesta no traía JSON (por ejemplo, un error sin cuerpo)
    }

    if (!respuesta.ok) {
        throw new ApiError(json?.error ?? 'Ocurrió un error inesperado', respuesta.status);
    }

    return json?.data as T;
    }