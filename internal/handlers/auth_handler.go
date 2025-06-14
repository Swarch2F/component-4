package handlers

import (
	"component-4/config"
	"component-4/internal/auth"
	"component-4/internal/store"
	"component-4/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
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
// @Success 200 {object} MessageResponse "Inicio de sesión exitoso."
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

	// Escenario 2: Usuario registrado con Google intenta iniciar sesión con contraseña.
	if user.Password == nil {
		h.writeJSON(w, http.StatusUnauthorized, LoginOAuthNeededResponse{
			Error:    "Te registraste usando Google. Por favor, inicia sesión con Google o crea una contraseña para tu cuenta.",
			UseOAuth: "true",
		})
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

	// Establecer el token JWT en una cookie HttpOnly
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false,    // Permitir acceso desde JavaScript en desarrollo
		Secure:   false,   // No requerir HTTPS en desarrollo
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.TokenExpirySeconds),
	})

	h.writeJSON(w, http.StatusOK, MessageResponse{Message: "Inicio de sesión exitoso."})
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
	code := r.URL.Query().Get("code")
	if code == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Código no encontrado en el callback."})
		return
	}

	googleUserInfo, err := auth.GetGoogleUserInfo(code)
	if err != nil {
		log.Printf("Error al obtener información del usuario de Google: %v", err)
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Fallo al obtener información del usuario de Google."})
		return
	}

	// Buscar si el usuario existe
	user, err := h.Store.FindByEmail(googleUserInfo.Email)

	if err != nil {
		// Si el usuario no existe, se rechaza la solicitud (según el requisito del usuario)
		if err.Error() == "user not found" {
			h.writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error: "Credenciales inválidas, este correo no está vinculado a una cuenta nativa.",
			})
			return
		}
		// Si es otro tipo de error al buscar al usuario
		log.Printf("Error al buscar usuario por email en GoogleCallback: %v", err)
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Error interno al buscar usuario."})
		return
	}

	// Si el usuario existe, usar UpsertGoogleUser para asegurar que el google_id esté vinculado
	// y obtener el usuario actualizado para generar el token.
	upsertedUser, err := h.Store.UpsertGoogleUser(
		googleUserInfo.Email,
		user.Name, // Usamos el nombre existente del usuario ya que GoogleUserInfo no proporciona 'Name' directamente
		googleUserInfo.ID,
		user.Role, // Usar el rol existente del usuario
	)
	if err != nil {
		log.Printf("Error al procesar usuario de Google (UpsertGoogleUser): %v", err)
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Fallo al procesar el usuario de Google (vincular/actualizar)."})
		return
	}

	token, err := auth.GenerateToken(upsertedUser, h.Config.JWTSecret)
	if err != nil {
		log.Printf("Error al generar el token JWT en GoogleCallback: %v", err)
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "No se pudo generar el token."})
		return
	}

	// Establecer el token JWT en una cookie HttpOnly
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Path:     "/",
		HttpOnly: false,    // Permitir acceso desde JavaScript en desarrollo
		Secure:   false,   // No requerir HTTPS en desarrollo
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.TokenExpirySeconds),
	})

	h.writeJSON(w, http.StatusOK, MessageResponse{Message: "Inicio de sesión exitoso con Google."})
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
// @Description Elimina el token JWT de las cookies del navegador, cerrando la sesión del usuario.
// @Tags auth
// @Produce json
// @Success 200 {object} MessageResponse "Sesión cerrada exitosamente."
// @Router /api/v1/logout [post]
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Eliminar la cookie estableciendo su expiración en el pasado
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Path:     "/",
		HttpOnly: false, // Debe coincidir con la configuración al establecer la cookie
		Secure:   false, // Debe coincidir con la configuración al establecer la cookie
		Expires:  time.Unix(0, 0), // Fecha en el pasado
		MaxAge:   -1,              // Eliminar inmediatamente
		SameSite: http.SameSiteLaxMode,
	})

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
	
	// Obtener el token de las cookies
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		log.Printf("AuthStatusHandler: No se encontró cookie jwt_token: %v", err)
		// Si no hay cookie, devolver usuario anónimo
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
	log.Printf("AuthStatusHandler: Cookie jwt_token encontrada: %s", cookie.Value)

	// Validar y desencriptar el token
	log.Printf("AuthStatusHandler: Intentando validar token con secret: %s", h.Config.JWTSecret)
	claims, err := auth.ValidateToken(cookie.Value, h.Config.JWTSecret)
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
			ID:    claims.UserID.String(), // Convertir UUID a string
			Name:  claims.Name,
			Email: claims.Email,
			Role:  claims.Role,
		},
		IsAuthenticated: true,
	}
	log.Printf("AuthStatusHandler: Enviando respuesta con usuario autenticado: %+v", response)
	
	h.writeJSON(w, http.StatusOK, response)
}

