# Análisis arquitectónico: módulo de préstamos de laptops

## 1. Descripción del problema

Actualmente, el sistema Biblioteca API permite gestionar un catálogo de libros y realizar préstamos a los estudiantes. La aplicación está desarrollada en Go con Gin, utiliza MySQL como sistema de persistencia y cuenta con autenticación mediante JWT.

El proyecto requiere ampliar las funcionalidades existentes mediante la incorporación de un módulo para gestionar el inventario y los préstamos de laptops.

El nuevo módulo debe permitir que los administradores registren equipos y que los estudiantes autenticados puedan consultar el inventario, solicitar préstamos de laptops disponibles y realizar devoluciones.

La nueva funcionalidad debe integrarse con la arquitectura actual sin afectar el funcionamiento de los libros, sus préstamos ni los mecanismos de autenticación existentes.

## 2. Análisis de la arquitectura actual

El proyecto utiliza una arquitectura por capas, organizada de la siguiente manera:

- **Handler:** recibe las solicitudes HTTP, valida los datos de entrada y devuelve las respuestas.
- **Service:** contiene la lógica de negocio y las reglas que deben cumplirse.
- **Repository:** se encarga de consultar y modificar la información almacenada en MySQL.
- **Models:** contiene las estructuras de datos utilizadas por las diferentes capas.

La comunicación entre las capas sigue este flujo:

Handler → Service → Repository → MySQL.

Adicionalmente, el sistema cuenta con autenticación JWT y dos roles de usuario: `student` y `admin`.

### 2.1 Archivos que componen el módulo de préstamos de libros

Durante la revisión del código se identificaron los siguientes archivos:

| Archivo | Responsabilidad |
|---|---|
| `migrations/001_init.sql` | Define las tablas originales, incluyendo `books` y `loans`. |
| `internal/models/loan.go` | Define las estructuras de los préstamos y sus solicitudes. |
| `internal/service/loan_service.go` | Contiene las reglas de negocio para solicitar y devolver libros. |
| `internal/repository/loan_repository.go` | Ejecuta las consultas SQL relacionadas con los préstamos. |
| `internal/handler/loan_handler.go` | Recibe las solicitudes HTTP y devuelve las respuestas del módulo. |
| `internal/handler/routes.go` | Registra las rutas y establece los middleware de autenticación y autorización. |

### 2.2 Identificación del acoplamiento existente

Durante la revisión del código se identificaron los siguientes puntos de dependencia relacionados con los préstamos de libros:

1. **Base de datos:** en `migrations/001_init.sql`, la tabla `loans` contiene el campo `book_id`, que es obligatorio y funciona como llave foránea hacia la tabla `books`.

2. **Modelo:** en `internal/models/loan.go`, la estructura `Loan` contiene el campo `BookID`. También existen estructuras de solicitud que utilizan `BookID` para identificar el libro que se desea prestar.

3. **Servicio:** en `internal/service/loan_service.go`, el método `Borrow()` utiliza la disponibilidad de los libros, verifica las condiciones para permitir un préstamo, calcula el vencimiento de 14 días y gestiona la disminución de las copias disponibles.

4. **Repositorio:** en `internal/repository/loan_repository.go`, las consultas SQL trabajan con las tablas `loans` y `books`. También se utilizan relaciones entre ambas tablas para obtener información como el título y el autor del libro.

5. **Detalle del préstamo:** la estructura `LoanDetail` contiene información relacionada con el libro. El JOIN que obtiene esos datos se ejecuta mediante las consultas SQL del repositorio.

6. **Handler:** en `internal/handler/loan_handler.go`, los métodos reciben las solicitudes de préstamos y devoluciones utilizando las estructuras propias del módulo de libros.

7. **Rutas:** en `internal/handler/routes.go`, los endpoints de préstamos están registrados para trabajar con el módulo de libros y utilizan el middleware de autenticación correspondiente.

### 2.3 Flujo actual de un préstamo de libro

El flujo comienza cuando un usuario autenticado solicita un préstamo mediante el endpoint:

`POST /api/loans`

El proceso sigue estas etapas:

1. **Route:** la ruta recibe la solicitud y verifica que el usuario tenga un token JWT válido.
2. **Handler:** recibe el campo `book_id` y obtiene el identificador del usuario autenticado.
3. **Service:** verifica que el libro tenga copias disponibles, valida las condiciones del préstamo y calcula la fecha de vencimiento de 14 días.
4. **Repository:** realiza las operaciones necesarias en MySQL para registrar el préstamo y actualizar la disponibilidad del libro.
5. **MySQL:** almacena la información del préstamo y los cambios en el inventario.
6. **Response:** el handler devuelve la información del préstamo registrado.

Este flujo demuestra que el préstamo actual depende directamente de las propiedades y reglas particulares de los libros.

### 2.4 Identificación del problema arquitectónico

La lógica actual de préstamos está directamente relacionada con la estructura y las reglas de los libros.

Si se intentara incorporar las laptops reutilizando o modificando directamente esta lógica, sería necesario introducir cambios en modelos, servicios, repositorios y relaciones de la base de datos que actualmente funcionan para los libros.

Esto podría aumentar el acoplamiento entre funcionalidades y generar riesgos de afectar las operaciones existentes.

Por ejemplo, los libros manejan cantidades de copias disponibles, mientras que cada laptop se identifica individualmente mediante su número de serie y tiene un estado propio.

Por esta razón, se considera conveniente evaluar una solución que permita gestionar las laptops sin depender directamente de las estructuras existentes de los libros.

## 3. Alternativas arquitectónicas evaluadas

Se identificaron tres alternativas para incorporar la nueva funcionalidad.

### Alternativa 1. Crear un módulo independiente para laptops

Consiste en desarrollar una estructura propia para gestionar las laptops y sus préstamos, incluyendo modelos, repositorios, servicios, handlers y tablas independientes en MySQL.

El módulo utilizaría la autenticación JWT y los roles existentes, pero mantendría su propia lógica de negocio.

**Ventajas:**

- No requiere modificar la lógica existente de los libros.
- Reduce el riesgo de afectar los préstamos actuales.
- Permite mantener responsabilidades separadas.
- Facilita el mantenimiento y las pruebas de cada módulo.
- Permite gestionar características propias de las laptops, como su número de serie y estado individual.

**Desventajas:**

- Se puede repetir parte de la lógica de negocio presente en los préstamos de libros.
- Se deben desarrollar y mantener nuevos componentes.
- Aumenta la cantidad de código y tablas en el proyecto.

**Riesgos:**

- Es necesario mantener correctamente las relaciones entre usuarios, laptops y préstamos.
- Se debe garantizar que una misma laptop no pueda tener dos préstamos activos simultáneamente.

### Alternativa 2. Unificar los préstamos de libros y laptops

Consiste en modificar el sistema actual para que una misma estructura de préstamos pueda gestionar tanto libros como laptops.

Para esto sería necesario modificar el modelo de datos y permitir que los préstamos se relacionen con diferentes tipos de recursos.

**Ventajas:**

- Permite reutilizar parte de la lógica existente.
- Centraliza algunas operaciones relacionadas con los préstamos.
- Puede reducir la duplicación de determinados componentes.

**Desventajas:**

- Requiere modificar la tabla `loans` y sus relaciones.
- Aumenta la complejidad de los servicios y repositorios.
- Introduce condiciones adicionales para diferenciar los tipos de recursos.
- Las reglas particulares de cada recurso deben seguir controlándose.

**Riesgos:**

- Puede afectar el funcionamiento actual de los préstamos de libros.
- Las modificaciones en las relaciones de la base de datos pueden generar incompatibilidades.
- Aumenta el acoplamiento entre funcionalidades diferentes.

### Alternativa 3. Crear un catálogo general de recursos

Consiste en diseñar una estructura general que permita registrar diferentes tipos de recursos prestables, como libros y laptops, mediante un catálogo unificado.

Los préstamos se relacionarían con un recurso general, en lugar de depender directamente de una tabla específica de libros.

**Ventajas:**

- Facilita incorporar nuevos tipos de recursos en el futuro.
- Permite centralizar la administración de los recursos.
- Ofrece una estructura más generalizada.
- Permite reutilizar una parte importante de la lógica común de los préstamos.

**Desventajas:**

- Requiere rediseñar una parte importante de la estructura actual.
- Aumenta la complejidad del modelo de datos.
- Implica modificar la lógica existente para adaptarla al nuevo catálogo.
- Es necesario definir cómo se almacenan y validan las características particulares de cada recurso.

**Riesgos:**

- Puede generar cambios importantes en las funcionalidades actuales.
- Requiere mayor esfuerzo de implementación y pruebas.
- Existe un mayor riesgo de introducir errores en los módulos existentes.

## 4. Comparación de las alternativas

| Criterio | Alternativa 1: módulo independiente | Alternativa 2: préstamos unificados | Alternativa 3: catálogo general |
|---|---|---|---|
| Modificación de la base de datos | Agrega las tablas `laptops` y `laptop_loans`, sin modificar las originales | Modifica la estructura de `loans` y sus relaciones | Introduce una tabla general de recursos y requiere adaptar las relaciones existentes |
| Reutilización | Reutiliza JWT, roles y conexión a MySQL | Reutiliza una mayor parte de la lógica común de préstamos | Reutiliza la lógica generalizada de recursos y préstamos |
| Duplicación | Repite modelos, servicios, repositorios y parte de la lógica de préstamos | Duplica menos lógica común, pero necesita condiciones específicas para cada recurso | Requiere desarrollar estructuras generales y mantener características específicas de cada recurso |
| Costo de agregar otro recurso | Alto: requiere crear otro módulo independiente | Medio o alto: requiere adaptar modelos, validaciones y persistencia | Menor después del rediseño inicial, porque el catálogo ya contempla diferentes recursos |
| Riesgo sobre lo existente | Bajo, porque no modifica los préstamos de libros | Alto, porque modifica estructuras y operaciones existentes | Alto durante la migración y adaptación de los módulos actuales |
| Complejidad inicial | Media | Alta | Alta |
| Mantenimiento | Los módulos pueden mantenerse por separado | Las reglas de los recursos se concentran en una estructura común | Requiere mantener el catálogo general y las particularidades de cada tipo de recurso |

### 4.1 Análisis de los trade-offs

Las alternativas presentan diferentes compromisos entre reutilización, extensibilidad, complejidad y riesgo.

La primera alternativa ofrece menor riesgo sobre el funcionamiento actual, pero implica mantener componentes independientes y repetir parte de la lógica.

La segunda permite reutilizar más elementos, pero requiere modificar la estructura de préstamos existente y aumenta el acoplamiento entre los recursos.

La tercera proporciona una mayor extensibilidad para incorporar recursos futuros, aunque requiere un rediseño inicial más complejo y una migración de las estructuras actuales.

Por lo tanto, la decisión depende de qué atributo tiene mayor prioridad para el proyecto: reducir el riesgo de modificar funcionalidades existentes o construir desde el principio una arquitectura general para diferentes recursos.

## 5. Decisión arquitectónica seleccionada

Después de evaluar las alternativas, se selecciona la **Alternativa 1: crear un módulo independiente para la gestión de laptops y sus préstamos**.

Esta decisión busca incorporar la nueva funcionalidad conservando la arquitectura existente y evitando modificaciones en el funcionamiento de los libros.

El nuevo módulo tendrá sus propios modelos, repositorios, servicios, handlers y tablas en MySQL.

Se reutilizarán los mecanismos de autenticación JWT y los roles de usuario que ya existen en la aplicación, evitando implementar un sistema de autenticación adicional.

No se creará un catálogo general de recursos ni se modificarán las estructuras actuales de los préstamos de libros.

La principal razón para elegir esta alternativa es reducir el riesgo de afectar el funcionamiento existente, aunque implique mantener cierta duplicación de código.

## 6. Reglas de negocio definidas

Para el nuevo módulo se establecen las siguientes reglas:

1. Cada laptop tendrá un registro individual en la base de datos.
2. Cada equipo tendrá un número de serie único para identificarlo.
3. Se almacenará información como marca, modelo, descripción y estado del equipo.
4. Los administradores serán los encargados de registrar las laptops.
5. Los usuarios autenticados podrán consultar el inventario completo de laptops.
6. Solamente los estudiantes podrán solicitar préstamos de laptops.
7. Cada préstamo tendrá una duración de 14 días, igual que los préstamos de libros.
8. Una laptop no podrá tener más de un préstamo activo simultáneamente.
9. Al realizar una devolución, la laptop volverá a estar disponible si se encuentra en condiciones de préstamo.
10. Si una laptop se devuelve con daños, pasará al estado de mantenimiento.
11. Los estudiantes podrán consultar sus propios préstamos y los administradores podrán consultar todos los préstamos.

## 7. Impacto esperado en el sistema

La incorporación del módulo independiente permitirá ampliar las funcionalidades de Biblioteca API sin alterar el comportamiento de los préstamos de libros.

Los principales cambios estarán concentrados en nuevos componentes y tablas de la base de datos:

- `laptops`: almacena el inventario individual de equipos.
- `laptop_loans`: almacena los préstamos realizados sobre las laptops.

También será necesario conectar los nuevos componentes durante la inicialización de la aplicación y registrar sus respectivas rutas.

La autenticación, los roles existentes, los endpoints de libros y las funcionalidades actuales deberán continuar funcionando sin modificaciones en su comportamiento.

En el futuro, si la biblioteca desea prestar proyectores, tablets u otros recursos, será posible aplicar una estructura independiente similar. Sin embargo, esto implicaría desarrollar y mantener nuevos componentes para cada tipo de recurso.

## 8. Conclusión

La implementación de un módulo independiente para los préstamos de laptops representa la alternativa más adecuada para este proyecto, debido a que permite incorporar una nueva funcionalidad sin comprometer la estabilidad de los componentes existentes.

Aunque esta decisión implica desarrollar y mantener cierta lógica adicional, ofrece una separación clara de responsabilidades, reduce el riesgo de afectar los préstamos de libros y facilita el mantenimiento de cada módulo.

La alternativa seleccionada prioriza la simplicidad de integración y la seguridad de los cambios sobre la generalización de toda la arquitectura.

Por lo tanto, se considera que esta alternativa ofrece un equilibrio adecuado entre riesgo, organización del código y cumplimiento de los requisitos del proyecto.