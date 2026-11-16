// Package notificaciones es el módulo de notificaciones push.
//
// Esqueleto: todavía no tiene lógica. Acá va lo que registrar y enviar
// notificaciones push a los dispositivos de los usuarios (registro de tokens,
// envío, historial), con la misma estructura del resto de los módulos:
// handler.go -> service.go -> storage.go.
//
// El envío real va detrás de una interfaz (tipo la SendEmail de
// internal/clavesAcceso), para poder pluggear el proveedor (FCM, APNs, etc.)
// sin tocar este módulo.
package notificaciones
