import Constants, { ExecutionEnvironment } from 'expo-constants';
import { Platform } from 'react-native';

// Notificaciones LOCALES (las programa el mismo celular, no necesitan backend) — Historia 008.
//
// expo-notifications se carga "a pedido" y dentro de try/catch: en Expo Go (SDK 57) le faltan
// módulos nativos y fallaría al importarlo, botando toda la app. Así, si no está disponible,
// la app sigue funcionando y solo esta función queda desactivada. En la app instalada
// (development build o APK) funciona completa.
type ModuloNotificaciones = typeof import('expo-notifications');

let modulo: ModuloNotificaciones | null | undefined;

function cargar(): ModuloNotificaciones | null {
  if (modulo !== undefined) return modulo;
  // En Expo Go ni siquiera se intenta cargar (evita el error rojo en la terminal)
  if (Constants.executionEnvironment === ExecutionEnvironment.StoreClient) {
    modulo = null;
    return modulo;
  }
  try {
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    modulo = require('expo-notifications') as ModuloNotificaciones;
    modulo.setNotificationHandler({
      handleNotification: async () => ({
        shouldShowBanner: true,
        shouldShowList: true,
        shouldPlaySound: true,
        shouldSetBadge: false,
      }),
    });
  } catch {
    modulo = null;
  }
  return modulo;
}

const ID_RECORDATORIO = 'recordatorio-diario';

// true si el celular puede programar notificaciones (en Expo Go puede no estar disponible)
export const notificacionesDisponibles = () => cargar() !== null;

// Pide permiso al usuario (Android 13+ e iPhone lo exigen). Devuelve true si lo concedió.
export async function pedirPermiso(): Promise<boolean> {
  const N = cargar();
  if (!N) return false;
  try {
    if (Platform.OS === 'android') {
      await N.setNotificationChannelAsync('default', {
        name: 'Tally',
        importance: N.AndroidImportance.HIGH,
        lightColor: '#8778cd',
      });
    }
    const actual = await N.getPermissionsAsync();
    if (actual.granted) return true;
    return (await N.requestPermissionsAsync()).granted;
  } catch {
    return false;
  }
}

// Recordatorio todos los días a la hora elegida
export async function programarRecordatorioDiario(hora: number, minuto: number): Promise<void> {
  const N = cargar();
  if (!N) return;
  await cancelarRecordatorioDiario();
  await N.scheduleNotificationAsync({
    identifier: ID_RECORDATORIO,
    content: { title: 'Tally', body: '¿Registraste tus ventas y gastos de hoy? Te toma menos de un minuto 💜' },
    trigger: { type: N.SchedulableTriggerInputTypes.DAILY, hour: hora, minute: minuto, channelId: 'default' },
  });
}

export async function cancelarRecordatorioDiario(): Promise<void> {
  const N = cargar();
  if (N) await N.cancelScheduledNotificationAsync(ID_RECORDATORIO).catch(() => {});
}

// Aviso el día que vence un pago pendiente (a las 10:00). Devuelve el id para poder cancelarlo.
export async function programarAvisoPendiente(fecha: Date, cliente: string, monto: string): Promise<string | undefined> {
  const N = cargar();
  if (!N) return undefined;
  const aviso = new Date(fecha.getFullYear(), fecha.getMonth(), fecha.getDate(), 10, 0);
  if (aviso.getTime() <= Date.now()) return undefined; // Ya pasó: no se programa
  try {
    return await N.scheduleNotificationAsync({
      content: { title: 'Pago pendiente', body: `Hoy vence el pago de ${cliente}: $${monto}` },
      trigger: { type: N.SchedulableTriggerInputTypes.DATE, date: aviso, channelId: 'default' },
    });
  } catch {
    return undefined;
  }
}

export async function cancelarAviso(id?: string): Promise<void> {
  const N = cargar();
  if (N && id) await N.cancelScheduledNotificationAsync(id).catch(() => {});
}
