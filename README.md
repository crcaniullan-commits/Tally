# Tally 

Tally es una aplicación de finanzas simples orientada a emprendedores informales, que permite registrar rápidamente ingresos y egresos con categorías personalizables, visualizar esta información en un dashboard con gráficos, definir metas de venta mensual, recibir recordatorios de registros diarios y pagos pendientes, mediante notificaciones. Las municipalidades adquieren el plan completo y activan el acceso para sus vecinos emprendedores mediante su correo y clave, sin tener visibilidad de los datos financieros de los usuarios. 

## Tecnologías usadas

- FrontEnd:       Aplicación móvil (pendiente de definición)
- BackEnd:        Servidor local o en la nube (pendiente de definición)
- Base de datos:  En local o nube, se usara el motor PostgresSQL

## Instrucciones para ejecutar el proyecto

### Clonando repositorio
Primero clona el proyecto en un directorio de su preferencia
```
git clone https://github.com/crcaniullan-commits/Tally.git
```

### Ejecutar el Backend
Entrar a la carpeta del Backend, sera necesario crear un archivo .env que contenga definido las variables del archivo .env.example 
una ves creado el archivo .env se puede levantar el contenedor de la base de datos
```
docker-compose up
```
Una ves este lista la base de datos prueba que exista usando un motor de base de datos de su preferencia para probar que puede
conectarse con el usuario definido en el .env

una ves comprobado ejecute en su terminal:
```
go install tool
```
y
```
go mod tidy
```
esto instalara las herramientas necesarias para continuar, una ves listo ejecuta los siguientes comandos
```
go tool task migration-up
```
esto ejecutara las migraciones y dejaran la base de datos con las tablas que se usaran
Finalmente ejecuta
```
go tool task start
```
para compilar y ejecutar el backend
ve a

```
http://localhost:8080/v1/swagger/
```
si todo salio bien, cargara la pagina del swagger con las rutas actualmente vigentes,
un detalle es que el puerto dependerá de lo que coloques en el .env, fijense que sea el mismo

### Ejecutando el archivo Docker-Compose

pendiente

## Integrantes

|  Integrante  |  Rol  |
|--------------|-------|
| Lucas Marchant | Lider de proyecto |
| Thiare Hernández | Desarrolladora Frontend UX/UI |
| Cristopher Caniullan | Desarrollador Backend y base de datos |

## Metodología

El proyecto sera desarrollado usando la metodología ágil Scrum, en sprints de una semanada de duración

## Arquitectura

pendiente de confirmación
