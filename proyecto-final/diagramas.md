# Diagramas de arquitectura – Biblioteca API

## 1. Diagrama C4 – Contexto del sistema

Este diagrama representa la interacción entre los usuarios y el sistema Biblioteca API.

Se identifican dos actores principales:

- **Estudiante:** consulta el inventario de libros y laptops, solicita préstamos y realiza devoluciones.
- **Administrador:** registra libros y laptops, administra el inventario y consulta los préstamos realizados.

El sistema utiliza autenticación JWT para validar las solicitudes de los usuarios y controlar el acceso según sus roles.

```mermaid
C4Context
    title Diagrama de contexto - Biblioteca API

    Person(estudiante, "Estudiante", "Consulta el inventario, solicita préstamos y realiza devoluciones")
    Person(admin, "Administrador", "Registra recursos, administra inventarios y consulta préstamos")

    System(biblioteca, "Biblioteca API", "Sistema para gestionar libros, laptops y sus respectivos préstamos")

    Rel(estudiante, biblioteca, "Consulta recursos y solicita préstamos")
    Rel(admin, biblioteca, "Registra recursos y consulta préstamos")
```

## 2. Diagrama C4 – Contenedores

Este diagrama representa los principales contenedores que conforman la aplicación y su infraestructura de ejecución.

La arquitectura utiliza tres contenedores principales:

- **API REST:** desarrollada en Go con Gin. Gestiona las solicitudes, la autenticación y las reglas de negocio.
- **MySQL:** almacena usuarios, libros, préstamos, laptops y préstamos de laptops.
- **Adminer:** permite visualizar y administrar la base de datos.

Los contenedores se ejecutan mediante Docker Compose.

```mermaid
C4Container
    title Diagrama de contenedores - Biblioteca API

    Person(usuario, "Usuario", "Estudiante o administrador")

    System_Boundary(sistema, "Biblioteca API") {
        Container(api, "API REST", "Go + Gin", "Gestiona solicitudes HTTP, autenticación y lógica de negocio")
        ContainerDb(mysql, "Base de datos", "MySQL 8", "Almacena usuarios, libros, préstamos y laptops")
        Container(adminer, "Adminer", "Aplicación web", "Permite consultar y administrar la base de datos")
    }

    Rel(usuario, api, "Realiza solicitudes HTTP")
    Rel(api, mysql, "Consulta y modifica información")
    Rel(adminer, mysql, "Administra y consulta los datos")
```

## 3. Diagrama de componentes – Arquitectura interna

## 3. Diagrama de componentes – Arquitectura interna

Este diagrama representa la organización interna de Biblioteca API y la incorporación del módulo de laptops.

Se identifican dos módulos independientes:

**Módulo de libros y préstamos:**
- Handlers: reciben las solicitudes relacionadas con libros y préstamos.
- Services: contienen las reglas de negocio.
- Repositories: realizan las operaciones en MySQL.

**Módulo de laptops:**
- LaptopHandler: recibe las solicitudes del inventario y los préstamos.
- LaptopService: aplica las reglas de negocio de las laptops.
- LaptopRepository: realiza las operaciones de persistencia.

Ambos módulos mantienen sus propios componentes y responsabilidades, pero comparten la infraestructura de autenticación JWT, el control de roles y la conexión a MySQL.

```mermaid
C4Component
    title Diagrama de componentes - Biblioteca API

    Container_Boundary(api, "Biblioteca API") {

        Component(router, "Gin Router", "Gin", "Registra las rutas de la API")
        Component(middleware, "Middleware", "AuthRequired / StudentRequired / AdminRequired", "Valida JWT y controla los roles")

        Boundary(libros, "Módulo de libros y préstamos") {
            Component(bookHandler, "BookHandler / LoanHandler", "Handlers", "Reciben las solicitudes HTTP")
            Component(bookService, "BookService / LoanService", "Services", "Contienen las reglas de negocio")
            Component(bookRepository, "BookRepository / LoanRepository", "Repositories", "Gestionan los datos de libros y préstamos")
        }

        Boundary(laptops, "Módulo independiente de laptops") {
            Component(laptopHandler, "LaptopHandler", "Handler", "Recibe las solicitudes de laptops")
            Component(laptopService, "LaptopService", "Service", "Aplica las reglas de negocio")
            Component(laptopRepository, "LaptopRepository", "Repository", "Gestiona los datos de laptops y préstamos")
        }
    }

    ContainerDb(mysql, "MySQL", "MySQL 8", "Almacena los libros, préstamos, laptops y préstamos de laptops")

    Rel(router, middleware, "Aplica autenticación y autorización")

    Rel(middleware, bookHandler, "Protege las rutas de libros")
    Rel(middleware, laptopHandler, "Protege las rutas de laptops")

    Rel(bookHandler, bookService, "Invoca operaciones")
    Rel(bookService, bookRepository, "Solicita acceso a los datos")

    Rel(laptopHandler, laptopService, "Invoca operaciones")
    Rel(laptopService, laptopRepository, "Solicita acceso a los datos")

    Rel(bookRepository, mysql, "Consulta y modifica books y loans")
    Rel(laptopRepository, mysql, "Consulta y modifica laptops y laptop_loans")
```

**Relación entre los módulos:** los préstamos de libros y laptops no comparten sus modelos, servicios ni repositorios. Cada uno administra sus propias tablas, mientras que ambos aprovechan los mecanismos de autenticación y persistencia de la aplicación.

## 4. Diagrama de secuencia – Préstamo de una laptop

Este diagrama representa el proceso que ocurre cuando un estudiante solicita una laptop mediante el endpoint `POST /api/laptop-loans`.

El flujo contempla las siguientes operaciones:

1. El estudiante envía una solicitud con su token JWT y el identificador de la laptop.
2. El middleware `AuthRequired` valida el token.
3. El middleware `StudentRequired` comprueba que el usuario tenga el rol de estudiante.
4. El handler recibe la solicitud y obtiene el identificador del usuario.
5. El servicio consulta si la laptop existe y verifica que tenga el estado `available`.
6. El servicio calcula la fecha de vencimiento de 14 días.
7. El repositorio inicia una transacción y actualiza el estado de la laptop a `loaned`, únicamente si todavía está disponible.
8. Si la actualización es correcta, registra el préstamo en `laptop_loans`.
9. MySQL aplica la restricción `UNIQUE` sobre `active_laptop_id` para impedir más de un préstamo activo por laptop.
10. Se confirma la transacción y se devuelve la información del préstamo.

Si la laptop no existe o ya no está disponible, la solicitud es rechazada.

```mermaid
sequenceDiagram
    autonumber

    actor Estudiante
    participant Auth as AuthRequired
    participant Role as StudentRequired
    participant Handler as LaptopHandler
    participant Service as LaptopService
    participant Repo as LaptopRepository
    participant DB as MySQL

    Estudiante->>Auth: POST /api/laptop-loans + JWT
    Auth->>Auth: Validar token JWT
    Auth->>Role: Usuario autenticado
    Role->>Role: Verificar rol student

    Role->>Handler: Permitir solicitud
    Handler->>Service: Borrow(userID, laptopID)

    Service->>Repo: FindByID(laptopID)
    Repo->>DB: SELECT laptop WHERE id = ?
    DB-->>Repo: Resultado de búsqueda
    Repo-->>Service: Laptop encontrada

    alt Laptop inexistente
        Service-->>Handler: sql.ErrNoRows
        Handler-->>Estudiante: HTTP 404 - Laptop no encontrada
    else Laptop existente

        Service->>Service: Verificar status == available

        alt Laptop no disponible
            Service-->>Handler: Error de disponibilidad
            Handler-->>Estudiante: HTTP 400 - Laptop no disponible

        else Laptop disponible

            Service->>Service: Calcular vencimiento (14 días)
            Service->>Repo: CreateLoan(userID, laptopID, dueAt)

            Repo->>DB: BEGIN
            Repo->>DB: UPDATE laptops SET status = loaned WHERE id = ? AND status = available

            alt No se actualizó ninguna fila
                DB-->>Repo: 0 filas afectadas
                Repo-->>Service: Laptop ya no disponible
                Service-->>Handler: Error de disponibilidad
                Handler-->>Estudiante: Solicitud rechazada

            else Laptop actualizada correctamente
                DB-->>Repo: 1 fila afectada

                Repo->>DB: INSERT INTO laptop_loans
                Note over DB: UNIQUE(active_laptop_id) impide préstamos activos duplicados

                DB-->>Repo: ID del préstamo
                Repo->>DB: COMMIT

                Repo->>DB: SELECT préstamo por ID
                DB-->>Repo: Datos del préstamo

                Repo-->>Service: Préstamo registrado
                Service-->>Handler: Préstamo y fecha de vencimiento
                Handler-->>Estudiante: HTTP 201 - Préstamo creado
            end
        end
    end
```

## 5. Conclusión

Los diagramas representan la incorporación del módulo de laptops dentro de la arquitectura actual de Biblioteca API.

Se conserva la estructura general de la aplicación y se incorpora una nueva separación de responsabilidades para administrar los equipos y sus préstamos.

El módulo de laptops mantiene sus propios modelos, servicios, repositorios, handlers y tablas en MySQL. A su vez, reutiliza la autenticación JWT y los mecanismos de autorización existentes.

De esta manera, los préstamos de libros y laptops mantienen sus responsabilidades independientes, evitando modificar la lógica del módulo original de libros.