# Auth Híbrida en Go

Este proyecto es una aplicación de autenticación híbrida en Go que permite a los usuarios iniciar sesión utilizando correo y contraseña, así como mediante OAuth con Google. Incluye lógica avanzada para vinculación de cuentas y flujos seguros, enfocándose en una excelente experiencia de usuario.

---

## Estructura del Proyecto

```
component-4
├── cmd/
│   └── main.go                # Punto de entrada de la aplicación
├── config/
│   └── config.go              # Manejo de configuración y variables de entorno
├── docs/
│   ├── docs.go                # Inicialización de documentación Swagger
│   ├── swagger.json           # Documentación de la API en formato JSON
│   └── swagger.yaml           # Documentación de la API en formato YAML
├── internal/
│   ├── auth/
│   │   ├── email.go           # Lógica de autenticación por correo y contraseña
│   │   ├── jwt.go             # Generación y validación de JWT
│   │   └── oauth.go           # Lógica de autenticación OAuth con Google
│   ├── handlers/
│   │   ├── auth_handler.go    # Controladores para rutas de autenticación
│   │   └── middleware.go      # Middleware para proteger rutas
│   ├── models/
│   │   └── user.go            # Modelo de usuario
│   └── store/
│       └── user_store.go      # Acceso y gestión de usuarios en la base de datos o almacenamiento
├── Dockerfile                 # Imagen Docker para la aplicación
├── docker-compose.yml         # Orquestación de servicios (por ejemplo, base de datos)
├── go.mod                     # Módulo de Go
├── go.sum                     # Sumas de verificación de dependencias
├── README.md                  # Documentación del proyecto
└── README-original.md         # Versión previa del README
```

---

## Funcionalidades Clave

- **Registro y Login Nativo**: Creación de cuentas y autenticación tradicional usando email y contraseña.
- **Autenticación con Google**: Inicio de sesión y registro simplificado mediante una cuenta de Google.
- **Vinculación de Cuentas**: Un usuario registrado con contraseña puede vincular su cuenta de Google para iniciar sesión con ambos métodos.
- **Manejo Inteligente de Flujos**: El sistema identifica si un usuario se registró solo con Google y le impide iniciar sesión con contraseña (a menos que la cree).
- **API Segura con JWT**: Las rutas protegidas utilizan JSON Web Tokens (JWT) para la autorización.
- **Arquitectura Limpia**: El código está organizado por responsabilidades (configuración, handlers, modelos, store, auth).
- **Documentación Swagger**: Documentación interactiva de la API disponible en la carpeta `docs/`.

---

## Endpoints de la API

| Método | Endpoint                                 | Descripción                                                                                      | Autenticación |
|--------|------------------------------------------|--------------------------------------------------------------------------------------------------|---------------|
| POST   | `/api/v1/register`                       | Registra un nuevo usuario con `{"email": "...", "password": "..."}`                              | No            |
| POST   | `/api/v1/login`                          | Inicia sesión con `{"email": "...", "password": "..."}`. Devuelve un JWT                         | No            |
| GET    | `/api/v1/auth/google/login`              | Redirige al usuario a la autenticación de Google                                                 | No            |
| GET    | `/api/v1/auth/google/callback`           | Endpoint al que Google redirige tras la autenticación. Maneja la creación/login y devuelve un JWT| No            |
| POST   | `/api/v1/auth/google/link`               | Vincula una cuenta de Google a un usuario existente. Requiere `{"email": "...", "password": "...", "google_auth_code": "..."}` | Sí (JWT)      |
| GET    | `/api/v1/profile`                        | Ruta protegida que requiere `Authorization: Bearer <token>` en la cabecera                       | Sí (JWT)      |
| GET    | `/swagger`                               | Interfaz interactiva de documentación Swagger                                                    | No            |
---

## Documentación Interactiva (Swagger)

La API cuenta con documentación interactiva generada con Swagger. Puedes explorar y probar los endpoints directamente desde la interfaz web disponible en:

[http://localhost:8080/swagger/](http://localhost:8080/swagger/)

La documentación se encuentra en la carpeta `docs/` y se actualiza automáticamente con los cambios en la API.

---

## Requisitos

- Go 1.16 o superior
- Docker (opcional, para despliegue y pruebas)

---

## Configuración

1. Crea un archivo `.env` en la raíz del proyecto.
2. Añade las siguientes variables de entorno:

    ```env
    # Puerto de la aplicación
    PORT=8080

    # Credenciales de Google OAuth 2.0
    GOOGLE_CLIENT_ID="TU_CLIENT_ID_DE_GOOGLE"
    GOOGLE_CLIENT_SECRET="TU_CLIENT_SECRET_DE_GOOGLE"
    GOOGLE_REDIRECT_URL="http://localhost:8080/auth/google/callback"

    # Secreto para firmar los JWT (usa un valor largo y aleatorio)
    JWT_SECRET="un-secreto-muy-largo-y-seguro-aqui"
    ```

---

## Instalación y Ejecución

1. Clona el repositorio:

    ```bash
    git clone <URL_DEL_REPOSITORIO>
    cd component-4
    ```

2. Instala las dependencias:

    ```bash
    go mod tidy
    ```

3. Ejecuta la aplicación:

    ```bash
    go run ./cmd/main.go
    ```

   O bien, usando Docker:

    ```bash
    docker-compose up --build
    ```

La API estará disponible en `http://localhost:8080`.

---

## Contribuciones

Las contribuciones son bienvenidas. Si deseas mejorar este proyecto, por favor abre un issue o un pull request.

---

## Licencia

Este proyecto está bajo la Licencia MIT. Consulta el archivo LICENSE para más detalles.