# Componente 4: Sistema de Autenticación Híbrida

Este componente implementa un sistema de autenticación híbrida que permite a los usuarios iniciar sesión tanto con credenciales tradicionales (email/contraseña) como mediante OAuth con Google. El sistema maneja inteligentemente la vinculación de cuentas y la gestión de sesiones usando JWT.

---

## Estructura del Proyecto

```
component-4
├── cmd/
│   ├── main.go                # Punto de entrada de la aplicación
│   └── migrate/
│       └── main.go            # Script de migración de base de datos
├── config/
│   └── config.go              # Configuración y variables de entorno
├── internal/
│   ├── auth/
│   │   ├── email.go           # Autenticación por email/contraseña
│   │   ├── jwt.go             # Gestión de JWT
│   │   └── oauth.go           # Integración con Google OAuth
│   ├── handlers/
│   │   ├── auth_handler.go    # Manejadores de rutas de autenticación
│   │   └── middleware.go      # Middleware de autenticación
│   ├── models/
│   │   └── user.go            # Modelo de usuario y roles
│   └── store/
│       └── user_store.go      # Operaciones de base de datos
├── migrations/                # Scripts SQL para la base de datos
├── Dockerfile                 # Configuración de Docker
├── docker-compose.yml         # Orquestación de servicios
├── go.mod                     # Dependencias de Go
└── go.sum                     # Checksums de dependencias
```

---

## Características Principales

- **Autenticación Dual**: Soporte para login tradicional y Google OAuth
- **Gestión Inteligente de Cuentas**: 
  - Vinculación automática de cuentas Google con usuarios existentes
  - Prevención de duplicación de cuentas
  - Manejo de roles (Administrador, Profesor, Estudiante)
- **Seguridad**:
  - JWT en cabeceras HTTP (no cookies por SSR)
  - Almacenamiento seguro en localStorage (frontend)
  - Hashing de contraseñas con bcrypt
- **Base de Datos**:
  - PostgreSQL con migraciones automáticas
  - Índices optimizados para búsquedas
  - Transacciones para garantizar integridad

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

---

## Implementación de Autenticación

### Flujo de Autenticación

1. **Login Tradicional**:
   - Usuario envía email/contraseña
   - Sistema verifica credenciales
   - Genera JWT si son válidas

2. **Login con Google**:
   - Usuario inicia flujo OAuth
   - Si el email existe:
     - Vincula cuenta Google si no está vinculada
     - Genera JWT
   - Si el email no existe:
     - Crea nueva cuenta
     - Genera JWT

### Seguridad

- JWT almacenados en localStorage (frontend)
- Tokens en cabeceras HTTP (no cookies por SSR)
- Validación de tokens en middleware
- Protección contra duplicación de cuentas

---

## Configuración

1. Crea un archivo `.env`:

    ```env
    PORT=8080
    GOOGLE_CLIENT_ID="tu_client_id"
    GOOGLE_CLIENT_SECRET="tu_client_secret"
    GOOGLE_REDIRECT_URL="http://localhost:8080/api/v1/auth/google/callback"
    JWT_SECRET="tu_jwt_secret"
    DB_HOST=localhost
    DB_PORT=5432
    DB_USER=authuser
    DB_PASSWORD=authpass
    DB_NAME=authdb
    DB_SSL_MODE=disable
    ```

2. Ejecuta las migraciones:
   ```bash
   go run cmd/migrate/main.go
   ```

3. Inicia el servidor:
   ```bash
   go run cmd/main.go
   ```

---

## Requisitos

- Go 1.16+
- PostgreSQL
- Docker (opcional)

---

## Licencia

Este proyecto está bajo la Licencia MIT.
