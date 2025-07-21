package handlers

import (
	"component-4/config"
	"component-4/internal/auth"
	"component-4/internal/store"
	"component-4/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const UserIDKey = "user"

// UserInfo representa la información básica de un usuario para la respuesta de estado.
type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// AuthStatusResponse representa la respuesta para el estado de autenticación.
type AuthStatusResponse struct {
	User            UserInfo `json:"user"`
	IsAuthenticated bool     `json:"isAuthenticated"`
}

// RegisterNativeRequest representa el cuerpo de la solicitud para el registro de usuario nativo.
type RegisterNativeRequest struct {
	Email    string `json:"email" example:"test@example.com"`
	Password string `json:"password" example:"password123"`
	Name     string `json:"name" example:"Juan Pérez"`
	Role     string `json:"role" example:"ESTUDIANTE"`
}

// LoginInput representa el cuerpo de la solicitud para el inicio de sesión de usuario nativo.
type LoginInput struct {
	Email    string `json:"email" example:"test@example.com"`
	Password string `json:"password" example:"password123"`
}

// LinkGoogleAccountInput representa el cuerpo de la solicitud para vincular una cuenta de Google.
type LinkGoogleAccountInput struct {
	Email           string `json:"email" example:"test@example.com"`
	Password        string `json:"password" example:"password123"`
	GoogleAuthCode  string `json:"google_auth_code" example:"4/0AY0e-g7..."` // Código de autorización de Google de ejemplo
}

// TokenResponse representa el cuerpo de la respuesta que contiene un token JWT.
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// MessageResponse representa un cuerpo de respuesta de mensaje de éxito genérico.
type MessageResponse struct {
	Message string `json:"message" example:"Operación exitosa"`
}

// ErrorResponse representa un cuerpo de respuesta de mensaje de error genérico.
type ErrorResponse struct {
	Error string `json:"error" example:"Ocurrió un error"`
}

// LoginOAuthNeededResponse representa una respuesta de error cuando se sugiere iniciar sesión con OAuth.
type LoginOAuthNeededResponse struct {
	Error    string `json:"error" example:"Te registraste usando Google. Por favor, inicia sesión con Google o crea una contraseña para tu cuenta."`
	UseOAuth string `json:"use_oauth" example:"true"`
}

// GoogleLinkAccountNeededResponse represents an error response when Google account linking is required.
type GoogleLinkAccountNeededResponse struct {
	Error          string `json:"error" example:"Ya existe una cuenta con este email. Por favor, inicia sesión con tu contraseña para vincular tu cuenta de Google."`
	ActionRequired string `json:"action_required" example:"link_account"`
}

// ProtectedResponse representa la respuesta para la ruta protegida.
type ProtectedResponse struct {
	Message string `json:"message" example:"Esta es una ruta protegida"`
	UserID  string `json:"user_id" example:"some-user-id"`
}

// AuthHandler contiene las dependencias para los manejadores de autenticación.
type AuthHandler struct {
	Store  *store.UserStore
	Config *config.Config
}

// NewAuthHandler crea una nueva instancia de AuthHandler.
func NewAuthHandler(s *store.UserStore, c *config.Config) *AuthHandler {
	return &AuthHandler{Store: s, Config: c}
}

// writeJSON responde con un payload JSON y un código de estado.
func (h *AuthHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// RegisterNativeHandler godoc
// @Summary Registrar un nuevo usuario con email y contraseña
// @Description Crea una nueva cuenta de usuario utilizando su email y una contraseña elegida.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   body body RegisterNativeRequest true "Registro de Usuario"
// @Success 201 {object} MessageResponse "Usuario registrado exitosamente."
// @Failure 400 {object} ErrorResponse "Payload de solicitud inválido."
// @Failure 409 {object} ErrorResponse "El usuario ya existe."
// @Router /api/v1/register [post]
// RegisterNativeHandler maneja el registro de usuarios con email y contraseña.
func (h *AuthHandler) RegisterNativeHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterNativeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Payload de solicitud inválido."})
		return
	}

	// Validar que el rol sea válido
	if models.Role(strings.ToLower(req.Role)) != models.ROLE_ESTUDIANTE && models.Role(strings.ToLower(req.Role)) != models.ROLE_PROFESOR {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Rol inválido. Debe ser ESTUDIANTE o PROFESOR."})
		return
	}

	if _, err := h.Store.CreateNativeUser(req.Email, req.Name, req.Password, models.Role(req.Role)); err != nil {
		h.writeJSON(w, http.StatusConflict, ErrorResponse{Error: "El usuario ya existe."})
		return
	}

	h.writeJSON(w, http.StatusCreated, MessageResponse{Message: "Usuario registrado exitosamente."})
}

// LoginNativeHandler godoc
// @Summary Iniciar sesión de un usuario con email y contraseña
// @Description Autentica a un usuario y devuelve un token JWT si el inicio de sesión es exitoso.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   body body LoginInput true "Inicio de Sesión de Usuario"
// @Success 200 {object} TokenResponse "Inicio de sesión exitoso."
// @Failure 400 {object} ErrorResponse "Payload de solicitud inválido."
// @Failure 401 {object} ErrorResponse "Email o contraseña inválidos."
// @Failure 401 {object} LoginOAuthNeededResponse "Usuario registrado con Google, debe usar el inicio de sesión de Google."
// @Failure 500 {object} ErrorResponse "No se pudo generar el token."
// @Router /api/v1/login [post]
// LoginNativeHandler maneja el inicio de sesión con email y contraseña.
func (h *AuthHandler) LoginNativeHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Payload de solicitud inválido."})
		return
	}

	user, err := h.Store.FindByEmail(req.Email)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Email o contraseña inválidos."})
		return
	}

	// Verificar que el usuario tenga contraseña
	if user.Password == nil {
		h.writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Credenciales inválidas, falta la contraseña."})
		return
	}

	if err := auth.CheckPassword(user, req.Password); err != nil {
		h.writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Email o contraseña inválidos."})
		return
	}

	token, err := auth.GenerateToken(user, h.Config.JWTSecret)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "No se pudo generar el token."})
		return
	}

	// Enviar el token en el header y en la respuesta JSON
	w.Header().Set("Authorization", "Bearer "+token)
	h.writeJSON(w, http.StatusOK, TokenResponse{
		Token: token,
	})
}

// GoogleLoginHandler godoc
// @Summary Iniciar sesión con Google OAuth2
// @Description Redirige al usuario a la página de consentimiento de Google OAuth2.
// @Tags auth
// @Success 307 "Redirige a Google OAuth"
// @Router /api/v1/auth/google/login [get]
// GoogleLoginHandler redirige al usuario a la página de consentimiento de Google.
func (h *AuthHandler) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	url := auth.GetGoogleLoginURL()
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallbackHandler godoc
// @Summary Manejar el callback de Google OAuth2
// @Description Procesa el callback de Google después de la autenticación del usuario. Crea un nuevo usuario o inicia sesión en un usuario existente, devolviendo un token JWT.
// @Tags auth
// @Produce  json
// @Param   code query string true "Código de autorización de Google"
// @Success 200 {object} MessageResponse "Autenticación exitosa, token devuelto."
// @Failure 400 {object} ErrorResponse "Código no encontrado en el callback."
// @Failure 409 {object} GoogleLinkAccountNeededResponse "La cuenta ya existe, necesita vinculación."
// @Failure 500 {object} ErrorResponse "Fallo al obtener información del usuario de Google o al crear/generar token."
// @Router /api/v1/auth/google/callback [get]
// GoogleCallbackHandler maneja el callback de Google.
func (h *AuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GoogleCallbackHandler: Iniciando callback de Google")
	
	// Obtener el código de autorización
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("GoogleCallbackHandler: No se recibió código de autorización")
		http.Error(w, "No se recibió código de autorización", http.StatusBadRequest)
		return
	}
	log.Printf("GoogleCallbackHandler: Código recibido: %s", code)

	// Obtener información del usuario de Google
	userInfo, err := auth.GetGoogleUserInfo(code)
	if err != nil {
		log.Printf("GoogleCallbackHandler: Error al obtener información del usuario: %v", err)
		http.Error(w, "Error al obtener información del usuario", http.StatusInternalServerError)
		return
	}
	log.Printf("GoogleCallbackHandler: Información del usuario obtenida: %+v", userInfo)

	// Buscar o crear el usuario en la base de datos
	user, err := h.Store.UpsertGoogleUser(
		userInfo.Email,
		userInfo.Name,
		userInfo.ID,
		models.ROLE_ESTUDIANTE, // Rol por defecto
	)
	if err != nil {
		log.Printf("GoogleCallbackHandler: Error al buscar/crear usuario: %v", err)
		http.Error(w, "Error al procesar usuario", http.StatusInternalServerError)
		return
	}
	log.Printf("GoogleCallbackHandler: Usuario encontrado/creado: %+v", user)

	// Generar token JWT
	jwtToken, err := auth.GenerateToken(user, h.Config.JWTSecret)
	if err != nil {
		log.Printf("GoogleCallbackHandler: Error al generar token JWT: %v", err)
		http.Error(w, "Error al generar token", http.StatusInternalServerError)
		return
	}
	log.Printf("GoogleCallbackHandler: Token JWT generado: %s", jwtToken)

	// Redirigir al frontend con el token como parámetro de URL
	frontendURL := fmt.Sprintf("%s/auth/callback?token=%s", h.Config.FrontendURL, jwtToken)
	log.Printf("GoogleCallbackHandler: URL de redirección completa: %s", frontendURL)
	
	// Establecer headers para evitar caché
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	
	http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
}

// LinkGoogleAccountHandler godoc
// @Summary Vincular una cuenta de Google a una cuenta nativa existente
// @Description Verifica la contraseña del usuario y vincula su cuenta de Google usando un código de autorización de Google.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   body body LinkGoogleAccountInput true "Detalles para Vincular Cuenta"
// @Success 200 {object} MessageResponse "Cuenta de Google vinculada exitosamente."
// @Failure 400 {object} ErrorResponse "Solicitud inválida o el email de la cuenta de Google no coincide."
// @Failure 401 {object} ErrorResponse "Contraseña inválida."
// @Failure 404 {object} ErrorResponse "Usuario no encontrado."
// @Failure 500 {object} ErrorResponse "Fallo al verificar con Google o al vincular la cuenta."
// @Router /api/v1/auth/google/link [post]
// LinkGoogleAccountHandler vincula una cuenta de Google a una existente, verificando la contraseña.
func (h *AuthHandler) LinkGoogleAccountHandler(w http.ResponseWriter, r *http.Request) {
	var req LinkGoogleAccountInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Solicitud inválida."})
		return
	}

	user, err := h.Store.FindByEmail(req.Email)
	if err != nil {
		h.writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "Usuario no encontrado."})
		return
	}

	if err := auth.CheckPassword(user, req.Password); err != nil {
		h.writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Contraseña inválida."})
		return
	}

	googleUserInfo, err := auth.GetGoogleUserInfo(req.GoogleAuthCode)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Fallo al verificar con Google."})
		return
	}
	
	if googleUserInfo.Email != user.Email {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "El email de la cuenta de Google no coincide con el email de la cuenta."})
		return
	}

	// Usar UpsertGoogleUser para vincular la cuenta
	_, err = h.Store.UpsertGoogleUser(
		user.Email,
		user.Name, // Mantener el nombre actual
		googleUserInfo.ID,
		user.Role, // Mantener el rol actual
	)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Fallo al vincular la cuenta."})
		return
	}

	h.writeJSON(w, http.StatusOK, MessageResponse{Message: "Cuenta de Google vinculada exitosamente."})
}

// ProtectedHandler es un ejemplo de una ruta protegida.
func (h *AuthHandler) ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	h.writeJSON(w, http.StatusOK, ProtectedResponse{
		Message: "Esta es una ruta protegida.",
		UserID:  userID,
	})
}

// LogoutHandler godoc
// @Summary Cerrar sesión del usuario
// @Description Elimina el token JWT del cliente.
// @Tags auth
// @Produce json
// @Success 200 {object} MessageResponse "Sesión cerrada exitosamente."
// @Router /api/v1/logout [post]
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// No es necesario hacer nada en el servidor ya que el cliente manejará el token
	h.writeJSON(w, http.StatusOK, MessageResponse{Message: "Sesión cerrada exitosamente."})
}

// AuthStatusHandler godoc
// @Summary Obtener estado de autenticación del usuario
// @Description Retorna la información del usuario autenticado o un usuario anónimo si no hay sesión
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AuthStatusResponse "Información del usuario"
// @Router /api/v1/auth-status [get]
func (h *AuthHandler) AuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("AuthStatusHandler: Iniciando verificación de estado de autenticación")
	
	// Obtener el token del header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Printf("AuthStatusHandler: No se encontró header Authorization")
		// Si no hay token, devolver usuario anónimo
		h.writeJSON(w, http.StatusOK, AuthStatusResponse{
			User: UserInfo{
				ID:    "",
				Name:  "Anónimo",
				Email: "",
				Role:  "guest",
			},
			IsAuthenticated: false,
		})
		return
	}

	// Extraer el token del header
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
	if tokenString == "" || tokenString == "undefined" {
		log.Printf("AuthStatusHandler: Token vacío o undefined")
		h.writeJSON(w, http.StatusOK, AuthStatusResponse{
			User: UserInfo{
				ID:    "",
				Name:  "Anónimo",
				Email: "",
				Role:  "guest",
			},
			IsAuthenticated: false,
		})
		return
	}

	log.Printf("AuthStatusHandler: Token encontrado en header: %s", tokenString)

	// Validar y desencriptar el token
	log.Printf("AuthStatusHandler: Intentando validar token con secret: %s", h.Config.JWTSecret)
	claims, err := auth.ValidateToken(tokenString, h.Config.JWTSecret)
	if err != nil {
		log.Printf("AuthStatusHandler: Error al validar token: %v", err)
		// Si el token no es válido, devolver usuario anónimo
		h.writeJSON(w, http.StatusOK, AuthStatusResponse{
			User: UserInfo{
				ID:    "",
				Name:  "Anónimo",
				Email: "",
				Role:  "guest",
			},
			IsAuthenticated: false,
		})
		return
	}
	log.Printf("AuthStatusHandler: Token validado exitosamente. Claims: %+v", claims)

	// Devolver información del usuario desde el token
	response := AuthStatusResponse{
		User: UserInfo{
			ID:    claims.UserID.String(),
			Name:  claims.Name,
			Email: claims.Email,
			Role:  claims.Role,
		},
		IsAuthenticated: true,
	}
	log.Printf("AuthStatusHandler: Enviando respuesta con usuario autenticado: %+v", response)
	
	h.writeJSON(w, http.StatusOK, response)
}

// Handler para verificar si un correo ya existe
func (h *AuthHandler) UserExists(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "email required"})
		return
	}
	user, err := h.Store.FindByEmail(email)
	exists := err == nil && user != nil
	h.writeJSON(w, http.StatusOK, map[string]bool{"exists": exists})
}

