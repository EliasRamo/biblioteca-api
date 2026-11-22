# Verificación del sistema de préstamos de laptops

## 1. Objetivo

Verificar que la funcionalidad de préstamos de laptops se integre correctamente al sistema de biblioteca, manteniendo el funcionamiento de los préstamos de libros, la autenticación y los roles existentes.

Las pruebas se realizaron utilizando Docker Compose, PowerShell, solicitudes HTTP y Adminer para comprobar los resultados en la base de datos.

## 2. Verificación del entorno

| Elemento | Resultado esperado | Estado |
|---|---|---|
| `go build ./...` | La compilación termina sin errores | Aprobado |
| `go vet ./...` | No se presentan errores de análisis | Aprobado |
| Docker Compose | Los servicios inician correctamente | Aprobado |
| API | El contenedor inicia sin errores | Aprobado |
| MySQL | La base de datos está disponible y saludable | Aprobado |
| Adminer | El contenedor permanece activo y permite acceder a la base de datos | Aprobado |
| Migración 001 | Se crean las tablas originales del sistema | Aprobado |
| Migración 002 | Se crean las tablas de laptops y préstamos | Aprobado |
| Ejecución automática de migraciones | MySQL ejecuta ambas migraciones al inicializar una base de datos nueva | Aprobado |

**Verificación de migraciones:** Se creó un contenedor temporal e independiente de MySQL para comprobar la ejecución automática de los archivos SQL. Los registros confirmaron la ejecución de `001_init.sql` y `002_laptops.sql`, sin modificar la base de datos principal.

## 3. Verificación de funcionalidades existentes

| Prueba | Resultado obtenido | Estado |
|---|---|---|
| GET /health | Respondió con `{"status":"ok"}` | Aprobado |
| GET /ping | Respondió con `{"message":"pong"}` | Aprobado |
| Registro de usuarios | Se registró correctamente un nuevo estudiante con ID 2 | Aprobado |
| Inicio de sesión | Permitió iniciar sesión y devolvió un JWT válido | Aprobado |
| Listar libros | Devolvió los cinco libros registrados | Aprobado |
| Solicitar préstamo de libro | Se registró correctamente el préstamo ID 2 | Aprobado |
| Consultar préstamos propios | Mostró los préstamos del usuario autenticado | Aprobado |
| Devolver libro | El préstamo ID 2 se devolvió correctamente | Aprobado |
| Acceso sin token | La API respondió con HTTP 401 | Aprobado |

Las pruebas confirmaron que el registro, la autenticación y las operaciones principales del módulo de libros continúan funcionando.

## 4. Verificación de préstamos de laptops

| Prueba | Resultado obtenido | Estado |
|---|---|---|
| Registrar laptop | El administrador registró LAP-001 correctamente | Aprobado |
| Listar laptops | El estudiante consultó el inventario completo, incluyendo equipos en mantenimiento | Aprobado |
| Consultar inventario completo | El administrador visualizó las laptops registradas | Aprobado |
| Solicitar laptop | Se creó correctamente el préstamo ID 1 | Aprobado |
| Consultar préstamos propios | El estudiante consultó sus préstamos | Aprobado |
| Consultar todos los préstamos | El administrador visualizó los dos préstamos registrados | Aprobado |
| Devolver laptop en buen estado | La laptop pasó al estado `available` | Aprobado |
| Solicitar nuevamente una laptop devuelta | Se permitió crear un nuevo préstamo después de la devolución | Aprobado |
| Devolver laptop dañada | La laptop pasó al estado `maintenance` | Aprobado |
| Solicitar laptop ocupada | El sistema rechazó la solicitud porque la laptop no estaba disponible | Aprobado |
| Evitar préstamo duplicado | No se permitió otro préstamo activo para la misma laptop | Aprobado |
| Control de roles | Se impidió que un estudiante consultara el historial administrativo | Aprobado |
| Registro de laptops por estudiantes | La API rechazó la operación con error HTTP 403 | Aprobado |
| Registrar número de serie duplicado | La API rechazó el segundo registro de LAP-001 con HTTP 409 | Aprobado |

## 5. Reglas de negocio por verificar

- Cada laptop se registra como un elemento individual.
- Cada laptop tiene un número de serie único.
- Los préstamos tienen una duración de 14 días.
- Solamente los estudiantes pueden solicitar laptops.
- Los administradores pueden registrar laptops y consultar el inventario.
- Los administradores pueden consultar todos los préstamos.
- Una laptop no puede tener más de un préstamo activo.
- Al devolver una laptop en buen estado, vuelve a estar disponible.
- Al devolver una laptop dañada, pasa a estado de mantenimiento.
- La autenticación utiliza el JWT existente.
- Las rutas administrativas están protegidas mediante el control de roles.

**Observación:** Las reglas de duración, disponibilidad, devolución y control de acceso fueron comprobadas mediante pruebas funcionales. La restricción de número de serie único también fue probada intentando registrar nuevamente la laptop `LAP-001`. La base de datos rechazó el registro duplicado mediante la restricción `UNIQUE` y la API respondió correctamente con HTTP 409 (Conflict).

### Verificación adicional de casos de error

| Prueba | Resultado obtenido | Estado |
|---|---|---|
| Consultar laptops sin token | HTTP 401 | Aprobado |
| Registrar laptop con rol de estudiante | HTTP 403 | Aprobado |
| Consultar todos los préstamos con rol de estudiante | HTTP 403 | Aprobado |
| Devolver un préstamo inexistente | HTTP 404 | Aprobado |
| Solicitar una laptop que ya está prestada | La API rechazó la solicitud | Aprobado |
| Registrar una laptop con número de serie duplicado | La API rechazó el registro con HTTP 409 y un mensaje explicativo | Aprobado |

## 6. Compatibilidad con el sistema anterior

Se comprobó que las nuevas funcionalidades no afectaran las siguientes operaciones:

- El registro y el inicio de sesión de usuarios.
- La generación y validación de JWT.
- Los roles de estudiante y administrador.
- El catálogo y la consulta de libros.
- Los préstamos y devoluciones de libros.
- Los endpoints de salud y disponibilidad de la API.

Las pruebas realizadas de los préstamos de libros, la autenticación y los endpoints de salud finalizaron correctamente.

También se ejecutaron `go build ./...` y `go vet ./...` sin errores. La reconstrucción de Docker Compose finalizó correctamente y los tres contenedores quedaron activos, con MySQL en estado `healthy`.

Después de corregir el manejo de números de serie duplicados, se reconstruyó la API y se repitió la prueba, obteniendo HTTP 409 correctamente.

## 7. Resultado final

La implementación del módulo de préstamos de laptops se integró correctamente al sistema de biblioteca, respetando la arquitectura por capas y manteniendo independientes los módulos de libros y laptops.

Las pruebas realizadas permitieron comprobar el registro de equipos, la consulta del inventario completo, la solicitud de préstamos, la duración de 14 días, la prevención de préstamos simultáneos, las devoluciones y el control de acceso según los roles de usuario.

También se comprobó que una laptop devuelta en buen estado vuelve a estar disponible y que una laptop devuelta con daños pasa a mantenimiento.

La prueba de número de serie duplicado confirmó que la restricción de unicidad de la base de datos funciona correctamente. Además, se ajustó el manejo de este error en la API para devolver HTTP 409 (Conflict), en lugar de HTTP 500.

Finalmente, se verificó que el módulo original de libros, la autenticación, los endpoints de salud y las migraciones automáticas continúan funcionando correctamente.

**Estado general:** Pruebas principales de compilación, entorno, migraciones, funcionalidades y manejo de errores aprobadas.