# Tally 

Tally es una aplicación de finanzas simples orientada a emprendedores informales, que permite registrar rápidamente ingresos y egresos con categorías personalizables, visualizar esta información en un dashboard con gráficos, definir metas de venta mensual, recibir recordatorios de registros diarios y pagos pendientes, mediante notificaciones. Las municipalidades adquieren el plan completo y activan el acceso para sus vecinos emprendedores mediante su correo y clave, sin tener visibilidad de los datos financieros de los usuarios. 

## Tecnologías usadas

- FrontEnd:       Aplicación móvil (pendiente de definición)
- BackEnd:        Servidor local o en la nube (pendiente de definición)
- Base de datos:  En local o nube, se usará el motor PostgresSQL

## Instrucciones para ejecutar el proyecto

### Clonando repositorio
Primero clona el proyecto en un directorio de su preferencia
```
git clone https://github.com/crcaniullan-commits/Tally.git
```

### Ejecutar el Backend
Entrar a la carpeta del Backend, será necesario crear un archivo .env que contenga definido las variables del archivo .env.example 
una vez creado el archivo .env se puede levantar el contenedor de la base de datos
```
docker-compose up
```
Cuando el docker compose esté listo prueba que la base de datos exista usando un motor de base de datos de su preferencia, pruebe que pueda
conectarse con el usuario definido en el .env

una vez comprobado ejecute en su terminal:
```
go install tool
```
y
```
go mod tidy
```
esto instalara las herramientas necesarias para continuar, cuando tenga esto listo ejecuta los siguientes comandos
```
go tool task migration-up
```
esto ejecutara las migraciones y dejaran la base de datos con las tablas que se usaran.

Finalmente ejecuta
```
go tool task start
```
para compilar y ejecutar el backend, cuando este ejecutado
ve a

```
http://localhost:8080/v1/swagger/
```
si todo salio bien, cargara la pagina del swagger con las rutas actualmente vigentes,
un detalle es que el puerto dependerá de lo que coloques en el .env, fíjense que sea el mismo

### Ejecutando el archivo Docker-Compose

En el Backend del proyecto habrá una carpeta docker, dentro encontrara un archivo dockerfile,
un archivo docker compose, un .env.example y un script entrypoint.sh, antes que nada debe 
crear un archivo .env en la misma carpeta use el .env.example para saber que variables debe 
definir en el .env. si ya tiene listo puede ejecutar
```
docker-compose up
```

esto creara los contenedores de la app, la base de datos y ejecutara las migraciones para 
dejar la base de datos como la app espera que este.

A partir de ahi puede ir a
```
http://localhost:8080/v1/swagger/
```

para probar que todo funcione.


## Integrantes

|  Integrante  |  Rol  |
|--------------|-------|
| Lucas Marchant | Lider de proyecto |
| Thiare Hernández | Desarrolladora Frontend UX/UI |
| Cristopher Caniullan | Desarrollador Backend y base de datos |

## Metodología

El proyecto será desarrollado usando la metodología ágil Scrum, en sprints de una semana de duración

## Arquitectura

pendiente de confirmación
