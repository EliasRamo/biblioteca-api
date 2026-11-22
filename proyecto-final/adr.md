# ADR-001: Implementación de un módulo independiente para préstamos de laptops

- **Estado:** Aceptado
- **Fecha:** 22 de noviembre de 2026
- **Proyecto:** Biblioteca API

## 1. Contexto

El sistema Biblioteca API permite gestionar libros y sus respectivos préstamos mediante una arquitectura por capas compuesta por handlers, services y repositories. La aplicación está desarrollada en Go con Gin, utiliza MySQL como base de datos y cuenta con autenticación JWT y control de acceso mediante roles.

Se requiere incorporar una nueva funcionalidad que permita administrar un inventario de laptops y gestionar sus préstamos y devoluciones.

La revisión de la arquitectura existente permitió identificar que la lógica actual de préstamos está directamente relacionada con los libros. El modelo Loan contiene una referencia obligatoria a BookID, el servicio de préstamos depende de BookRepository y las consultas del repositorio utilizan la tabla books.

Además, los libros manejan cantidades de copias disponibles, mientras que las laptops requieren una identificación individual mediante un número de serie y un control de su estado.

Por lo tanto, incorporar las laptops directamente sobre la estructura actual implicaría modificar componentes que ya cumplen una función específica dentro del sistema.

El objetivo es agregar la nueva funcionalidad sin alterar el comportamiento de los libros ni generar dependencias innecesarias entre ambos módulos.

Para tomar la decisión se consideraron tres criterios principales:

- **Riesgo:** evitar que los cambios afecten las funcionalidades existentes.
- **Simplicidad:** implementar la nueva funcionalidad sin realizar un rediseño innecesario.
- **Extensibilidad:** permitir incorporar nuevos tipos de recursos sin perder de vista el costo de mantenimiento.

## 2. Alternativas consideradas

### Alternativa 1: módulo independiente para laptops

Crear un módulo separado para gestionar las laptops y sus préstamos. Este módulo tendrá sus propios modelos, repositorios, servicios, handlers y tablas en MySQL.

Se conservarán los mecanismos de autenticación JWT y los roles existentes, pero la lógica de negocio de los préstamos de laptops será independiente de la lógica de los libros.

**Ventajas:**

- Mantiene intacta la funcionalidad actual de los libros.
- Reduce el riesgo de introducir errores en los préstamos existentes.
- Permite separar las responsabilidades de cada módulo.
- Facilita realizar pruebas y mantenimiento de manera independiente.
- Permite administrar características específicas de las laptops, como el número de serie y su estado individual.

**Desventajas:**

- Puede generar duplicación de lógica entre los módulos.
- Requiere desarrollar nuevos componentes.
- Aumenta la cantidad de código que debe mantenerse.

**Riesgos:**

- Es necesario mantener correctamente las relaciones entre usuarios, laptops y préstamos.
- Se debe garantizar que una misma laptop no pueda tener dos préstamos activos simultáneamente.

### Alternativa 2: unificar los préstamos de libros y laptops

Modificar el modelo y la lógica actual de préstamos para permitir que una misma estructura gestione ambos tipos de recursos.

**Ventajas:**

- Permite reutilizar parte de la lógica existente.
- Centraliza las operaciones de préstamo y devolución.
- Puede reducir la duplicación de código.

**Desventajas:**

- Requiere modificar el modelo actual de préstamos.
- Aumenta la complejidad de los servicios y repositorios.
- Introduce condiciones adicionales para diferenciar libros y laptops.

**Riesgos:**

- Puede afectar las funcionalidades actuales de los libros.
- Puede generar inconsistencias en las relaciones de la base de datos.
- Aumenta el impacto de futuros cambios sobre los módulos existentes.

### Alternativa 3: crear un catálogo general de recursos

Diseñar una estructura genérica que permita administrar diferentes tipos de recursos prestables, incluyendo libros y laptops, mediante un catálogo unificado.

**Ventajas:**

- Facilita incorporar nuevos tipos de recursos.
- Permite centralizar la administración del inventario.
- Ofrece mayor flexibilidad para futuras ampliaciones.

**Desventajas:**

- Requiere rediseñar una parte importante de la arquitectura.
- Aumenta la complejidad del modelo de datos.
- Implica adaptar las funcionalidades actuales al nuevo diseño.

**Riesgos:**

- Puede afectar el funcionamiento de los préstamos existentes.
- Requiere mayor esfuerzo de implementación y pruebas.
- Introduce cambios amplios para resolver una necesidad específica.

## 3. Decisión

Se selecciona la **Alternativa 1: implementar un módulo independiente para la gestión de laptops y sus préstamos**.

La decisión se fundamenta principalmente en reducir el riesgo de afectar las funcionalidades existentes, conservar la separación de responsabilidades y evitar un rediseño de la arquitectura para incorporar un solo tipo de recurso adicional.

Aunque esta alternativa implica duplicar parte de la lógica de préstamos, se considera un costo aceptable frente al riesgo de modificar el modelo, los servicios y las relaciones de base de datos que actualmente utilizan los libros.

El nuevo módulo tendrá una estructura propia dentro de las capas actuales de la aplicación:

- **Models:** contendrá las estructuras de datos de las laptops y sus préstamos.
- **Repository:** realizará las operaciones de consulta y persistencia en MySQL.
- **Service:** implementará las reglas de negocio relacionadas con los préstamos de laptops.
- **Handler:** expondrá los endpoints necesarios para registrar equipos, consultar el inventario y gestionar los préstamos.

También se crearán tablas independientes para almacenar la información de las laptops y sus préstamos:

- `laptops`
- `laptop_loans`

Se reutilizarán los mecanismos actuales de autenticación JWT y autorización mediante roles, sin implementar un sistema adicional.

Para integrar el nuevo módulo se agregarán sus repositorios, servicios y handlers durante la inicialización de la aplicación, así como sus respectivas rutas y controles de acceso.

No se modificarán los modelos, repositorios, servicios ni tablas existentes de los libros. Tampoco se crearán estructuras de préstamos compartidas ni un catálogo general de recursos.

## 4. Reglas de negocio asociadas a la decisión

La implementación deberá cumplir las siguientes reglas:

1. Cada laptop tendrá un registro individual en la base de datos.
2. Cada equipo tendrá un número de serie único.
3. Se almacenará información de identificación como marca, modelo, descripción y estado.
4. Los administradores podrán registrar laptops.
5. Los usuarios autenticados podrán consultar el inventario completo, incluyendo los equipos disponibles, prestados y en mantenimiento.
6. Solamente los estudiantes podrán solicitar préstamos de laptops.
7. La duración de cada préstamo será de 14 días, igual que los préstamos de libros.
8. Una laptop no podrá tener dos préstamos activos simultáneamente.
9. Al registrar una devolución, la laptop volverá a estar disponible si se encuentra en condiciones de préstamo.
10. Si una laptop se devuelve con daños, pasará al estado de mantenimiento.
11. Los estudiantes podrán consultar sus propios préstamos y los administradores podrán consultar todos los préstamos.

## 5. Consecuencias de la decisión

### Consecuencias positivas

- Se conserva el funcionamiento de los préstamos de libros.
- Se mantiene la separación de responsabilidades de la arquitectura por capas.
- Los cambios se concentran en nuevos componentes y tablas.
- Se facilita el mantenimiento independiente de los dos módulos.
- Se reduce el riesgo de afectar funcionalidades existentes.
- Se aprovecha la autenticación y autorización que ya proporciona el sistema.
- Las características particulares de las laptops pueden gestionarse sin introducir condiciones adicionales en el módulo de libros.

### Consecuencias negativas

- Existirá cierta duplicación de lógica entre los préstamos de libros y laptops.
- Será necesario mantener dos conjuntos de componentes para gestionar los préstamos.
- Aumentará la cantidad de tablas y código de la aplicación.
- Incorporar otro tipo de recurso, como proyectores o tablets, implicaría desarrollar nuevos componentes y repetir parte de la lógica.

### Consideraciones de implementación

- Se creó una nueva migración SQL para las tablas de laptops y sus préstamos.
- Se registraron los nuevos repositorios, servicios y handlers en el punto de entrada de la aplicación.
- Se agregaron nuevas rutas sin modificar el comportamiento de las rutas actuales de libros.
- Se implementaron controles de acceso mediante los roles existentes.
- Se establecieron restricciones para evitar números de serie duplicados y préstamos activos simultáneos.
- Se realizaron pruebas para comprobar que las funcionalidades existentes continúan operando correctamente.
- Se estableció que un número de serie duplicado debe producir una respuesta HTTP 409 (Conflict).

## 6. Estado de la decisión

La decisión queda **ACEPTADA**.

Se implementó un módulo independiente para la gestión de laptops y sus préstamos, manteniendo la arquitectura por capas existente y reutilizando los mecanismos de autenticación y autorización.

La solución prioriza la conservación de las funcionalidades actuales, la separación de responsabilidades y la reducción del riesgo de introducir cambios en el módulo de libros.

Aunque implica mantener cierta duplicación de código y desarrollar nuevos componentes para futuros tipos de recursos, se considera que este costo es razonable frente a la simplicidad de integración y la estabilidad del sistema existente.