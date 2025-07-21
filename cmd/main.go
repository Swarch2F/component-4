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
	// _ "component-4/docs" // Make sure this path is correct based on your go.mod module name
	// httpSwagger "github.com/swaggo/http-swagger"
	"component-4/internal/migrate"
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

	// Ejecutar migraciones antes de inicializar el store
	err = migrate.RunMigrations(db, "./migrations")
	if err != nil {
		log.Fatalf("Error ejecutando migraciones: %v", err)
	}

	// Inicializar el store
	store, err := store.NewUserStore(cfg)
	if err != nil {
		log.Fatalf("Error creating store: %v", err)
	}

	// Crear usuario administrador si no existe
	_, err = store.FindByEmail("rector@colegio.edu")
	if err != nil {
		if err.Error() == "user not found" {
			_, err = store.CreateNativeUser(
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
	authHandler := handlers.NewAuthHandler(store, cfg)

	// Crear el router principal
	r := mux.NewRouter()

	// Configurar CORS
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3001")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Expose-Headers", "Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})

	// Subrouter para la API versionada
	api := r.PathPrefix("/api/v1").Subrouter()

	// Rutas de autenticación
	api.HandleFunc("/register", authHandler.RegisterNativeHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/login", authHandler.LoginNativeHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/google/login", authHandler.GoogleLoginHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/google/callback", authHandler.GoogleCallbackHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/google/link", authHandler.LinkGoogleAccountHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth-status", authHandler.AuthStatusHandler).Methods("GET", "OPTIONS")
	api.HandleFunc("/logout", authHandler.LogoutHandler).Methods("POST", "OPTIONS")
	// Nueva ruta para verificar si un correo existe
	api.HandleFunc("/users/exists", authHandler.UserExists).Methods("GET", "OPTIONS")
	// Ruta simple de health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Servidor funcionando correctamente!"))
	}).Methods("GET")

	// Rutas protegidas
	protected := api.PathPrefix("/profile").Subrouter()
	protected.Use(handlers.AuthMiddleware(cfg.JWTSecret))
	protected.HandleFunc("", authHandler.ProtectedHandler).Methods("GET", "OPTIONS")


	// Swagger endpoint (fuera de /api)
	// r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	// Iniciar el servidor
	log.Printf("Starting server on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
