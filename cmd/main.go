package main

import (
	"database/sql"
	"log"
	"net/http"

	"component-4/config"
	"component-4/internal/auth"
	"component-4/internal/handlers"
	"component-4/internal/store"
	"component-4/internal/models"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	// Import generated docs
	_ "component-4/docs" // Make sure this path is correct based on your go.mod module name
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title API de autenticación híbrida
// @version 1.1
// @description Esta es la documentación del servidor para un sistema de autenticación híbrido (Nativo + Google OAuth), el cual incluye registro, inicio de sesión, autenticación y vinculación de cuentas de Google.\n
// @description El sistema permite a los usuarios registrarse y autenticarse utilizando credenciales nativas, así como iniciar sesión con Google OAuth. Además, los usuarios pueden vincular sus cuentas de Google a su cuenta nativa para una experiencia de inicio de sesión unificada.\n
// @description En resumen, cubre todos los casos de uso de autenticación híbrida.\n
// @description
// @description **IMPORTANTE:**\n
// @description - Los endpoints de autenticación con Google (como `/auth/google/login`) realizan redirecciones a Google o usan su API y **no pueden ser probados directamente desde Swagger UI**. Para probar el flujo OAuth, abre la ruta en tu navegador.\n
// @description - Cuando inicies sesión exitosamente (Ya sea con login nativo o con Google), recibirás un token JWT. Para acceder a rutas protegidas desde Swagger UI, haz clic en el botón "Authorize" y escribe: `Bearer [TOKEN]` (sin los corchetes, dejando un espacio después de "Bearer").\n
// @description - Si tienes dudas sobre el uso de los endpoints, revisa las descripciones individuales o contacta al soporte.
// @termsOfService http://swagger.io/terms/

// @contact.name Soporte API
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Licencia Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Escribe Bearer [TOKEN] (sin los corchetes, dejando un espacio después de "Bearer") para acceder a las rutas protegidas
func main() {
	// Cargar configuración
	cfg := config.LoadConfig()
	if cfg.Port == "" {
		cfg.Port = "8080" // Puerto por defecto
	}
	if cfg.JWTSecret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is not set.")
	}

	// Inicializar la configuración de OAuth de Google
	auth.ConfigureGoogleOauth(cfg)

	// Inicializar la conexión a la base de datos
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error al conectar con la base de datos: %v", err)
	}
	defer db.Close()

	// Inicializar el almacén de usuarios
	userStore := store.NewUserStore(db)

	// Crear usuario administrador si no existe
	_, err = userStore.FindByEmail("rector@colegio.edu")
	if err != nil {
		if err.Error() == "user not found" {
			_, err = userStore.CreateNativeUser(
				"rector@gmail.com",
				"Rector del Colegio Luis Alberto",
				"rector123",
				models.ROLE_ADMINISTRADOR,
			)
			if err != nil {
				log.Printf("Error al crear usuario administrador: %v", err)
			} else {
				log.Println("Usuario administrador creado exitosamente")
			}
		} else {
			log.Printf("Error al buscar usuario administrador: %v", err)
		}
	}

	// Inicializar los manejadores de autenticación
	authHandler := handlers.NewAuthHandler(userStore, cfg)

	// Crear el router principal
	r := mux.NewRouter()

	// Subrouter para la API versionada
	api := r.PathPrefix("/api/v1").Subrouter()

	// Rutas de autenticación
	api.HandleFunc("/register", authHandler.RegisterNativeHandler).Methods("POST")
	api.HandleFunc("/login", authHandler.LoginNativeHandler).Methods("POST")
	api.HandleFunc("/auth/google/login", authHandler.GoogleLoginHandler).Methods("GET")
	api.HandleFunc("/auth/google/callback", authHandler.GoogleCallbackHandler).Methods("GET")
	api.HandleFunc("/auth/google/link", authHandler.LinkGoogleAccountHandler).Methods("POST")
	api.HandleFunc("/auth-status", authHandler.AuthStatusHandler).Methods("GET")
	api.HandleFunc("/logout", authHandler.LogoutHandler).Methods("POST")

	// Rutas protegidas
	protected := api.PathPrefix("/profile").Subrouter()
	protected.Use(handlers.AuthMiddleware(cfg.JWTSecret))
	protected.HandleFunc("", authHandler.ProtectedHandler).Methods("GET")

	// Swagger endpoint (fuera de /api)
	r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	// Envolver el mux con el middleware de CORS.
	handler := corsMiddleware(r)

	// Iniciar el servidor
	log.Printf("Starting server on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}

// corsMiddleware agrega encabezados CORS a las respuestas.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permitimos todos los orígenes. Puedes limitarlo a dominios específicos.
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")
		
		// Responder a solicitudes preflight
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
