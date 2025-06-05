# Auth Híbrida en Go

Este proyecto es una aplicación de autenticación híbrida en Go que permite a los usuarios iniciar sesión utilizando correo y contraseña, así como mediante OAuth con Google. 

## Estructura del Proyecto

```
auth-hibrida-go
├── cmd
│   └── main.go               # Punto de entrada de la aplicación
├── internal
│   ├── auth
│   │   ├── email.go          # Manejo de autenticación por correo
│   │   └── oauth.go          # Manejo de autenticación OAuth con Google
│   ├── handlers
│   │   ├── auth_handler.go    # Controladores para las rutas de autenticación
│   │   └── middleware.go      # Middleware para proteger rutas
│   └── models
│       └── user.go           # Modelo de usuario
├── config
│   └── config.go             # Manejo de configuración
├── go.mod                     # Módulo de Go
├── go.sum                     # Sumas de verificación de dependencias
└── README.md                  # Documentación del proyecto
```

## Requisitos

- Go 1.16 o superior
- Dependencias necesarias que se instalarán mediante `go mod`

## Instalación

1. Clona el repositorio:

   ```
   git clone <URL_DEL_REPOSITORIO>
   cd auth-hibrida-go
   ```

2. Instala las dependencias:

   ```
   go mod tidy
   ```

3. Configura las variables de entorno necesarias para la aplicación. Asegúrate de definir las variables para la autenticación de Google y cualquier otra configuración requerida.

## Ejecución

Para ejecutar la aplicación, utiliza el siguiente comando:

```
go run cmd/main.go
```

La aplicación se iniciará y estará disponible en `http://localhost:8080` (puedes cambiar el puerto en la configuración).

## Funcionalidades

- **Autenticación por correo y contraseña**: Los usuarios pueden registrarse e iniciar sesión utilizando su correo electrónico y contraseña.
- **Autenticación OAuth con Google**: Los usuarios pueden iniciar sesión utilizando su cuenta de Google, lo que simplifica el proceso de autenticación.

## Contribuciones

Las contribuciones son bienvenidas. Si deseas mejorar este proyecto, por favor abre un issue o un pull request.

## Licencia

Este proyecto está bajo la Licencia MIT. Consulta el archivo LICENSE para más detalles.