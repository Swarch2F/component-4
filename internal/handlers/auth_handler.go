package handlers

import (
	"component-4/config"
	"component-4/internal/auth"
	"component-4/internal/store"
	"encoding/json"
	"net/http"
)

// --- DTOs (Objetos de Transferencia de Datos) para la documentación de Swagger ---

// RegisterNativeRequest representa el cuerpo de la solicitud para el registro de usuario nativo.
type RegisterNativeRequest struct {
	Email    string `json:"email" example:"test@example.com"`
	Password string `json:"password" example:"password123"`
}

// LoginNativeRequest representa el cuerpo de la solicitud para el inicio de sesión de usuario nativo.
type LoginNativeRequest struct {
	Email    string `json:"email" example:"test@example.com"`
	Password string `json:"password" example:"password123"`
}

// LinkGoogleAccountRequest representa el cuerpo de la solicitud para vincular una cuenta de Google.
type LinkGoogleAccountRequest struct {
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
// @Router /register [post]
// RegisterNativeHandler maneja el registro de usuarios con email y contraseña.
func (h *AuthHandler) RegisterNativeHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterNativeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Payload de solicitud inválido."})
		return
	}

	if _, err := h.Store.CreateNativeUser(req.Email, req.Password); err != nil {
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
// @Param   body body LoginNativeRequest true "Inicio de Sesión de Usuario"
// @Success 200 {object} TokenResponse "Inicio de sesión exitoso, token devuelto"
// @Failure 400 {object} ErrorResponse "Payload de solicitud inválido."
// @Failure 401 {object} ErrorResponse "Email o contraseña inválidos."
// @Failure 401 {object} LoginOAuthNeededResponse "Usuario registrado con Google, debe usar el inicio de sesión de Google."
// @Failure 500 {object} ErrorResponse "No se pudo generar el token."
// @Router /login [post]
// LoginNativeHandler maneja el inicio de sesión con email y contraseña.
func (h *AuthHandler) LoginNativeHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginNativeRequest
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
	if user.PasswordHash == nil {
		h.writeJSON(w, http.StatusUnauthorized, LoginOAuthNeededResponse{
			Error:   "Te registraste usando Google. Por favor, inicia sesión con Google o crea una contraseña para tu cuenta.",
			UseOAuth: "true",
		})
		return
	}

	if err := auth.CheckPassword(user, req.Password); err != nil {
		h.writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "Email o contraseña inválidos."})
		return
	}

	token, err := auth.GenerateToken(user.ID, h.Config.JWTSecret)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "No se pudo generar el token."})
		return
	}

	h.writeJSON(w, http.StatusOK, TokenResponse{Token: token})
}

// GoogleLoginHandler godoc
// @Summary Iniciar sesión con Google OAuth2
// @Description Redirige al usuario a la página de consentimiento de Google OAuth2.
// @Tags auth
// @Success 307 "Redirige a Google OAuth"
// @Router /auth/google/login [get]
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
// @Success 200 {object} TokenResponse "Autenticación exitosa, token devuelto."
// @Failure 400 {object} ErrorResponse "Código no encontrado en el callback."
// @Failure 409 {object} GoogleLinkAccountNeededResponse "La cuenta ya existe, necesita vinculación."
// @Failure 500 {object} ErrorResponse "Fallo al obtener información del usuario de Google o al crear/generar token."
// @Router /auth/google/callback [get]
// GoogleCallbackHandler maneja el callback de Google.
func (h *AuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		h.writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Código no encontrado en el callback."})
		return
	}

	googleUserInfo, err := auth.GetGoogleUserInfo(code)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Fallo al obtener información del usuario de Google."})
		return
	}

	user, err := h.Store.FindByEmail(googleUserInfo.Email)
	if err != nil { // Usuario no existe, lo creamos.
		newUser, err := h.Store.CreateGoogleUser(googleUserInfo.Email, googleUserInfo.ID)
		if err != nil {
			h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Fallo al crear usuario."})
			return
		}
		// Iniciar sesión para el nuevo usuario
		token, _ := auth.GenerateToken(newUser.ID, h.Config.JWTSecret)
		h.writeJSON(w, http.StatusOK, TokenResponse{Token: token})
		return
	}

	// Usuario ya existe.
	if user.GoogleID == nil {
		// Escenario 1: Usuario nativo que inicia sesión con Google por primera vez.
		// En una API real, aquí devolverías un token de vinculación o un mensaje claro.
		// Por simplicidad, aquí lo vinculamos directamente si queremos (menos seguro) o devolvemos un error.
		// Devolveremos un error para que el frontend pueda pedir la contraseña.
		h.writeJSON(w, http.StatusConflict, GoogleLinkAccountNeededResponse{
			Error:          "Ya existe una cuenta con este email. Por favor, inicia sesión con tu contraseña para vincular tu cuenta de Google.",
			ActionRequired: "link_account",
		})
		return
	}

	// Usuario ya existe y tiene Google vinculado, iniciamos sesión.
	token, err := auth.GenerateToken(user.ID, h.Config.JWTSecret)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "No se pudo generar el token."})
		return
	}
	h.writeJSON(w, http.StatusOK, TokenResponse{Token: token})
}

// LinkGoogleAccountHandler godoc
// @Summary Vincular una cuenta de Google a una cuenta nativa existente
// @Description Verifica la contraseña del usuario y vincula su cuenta de Google usando un código de autorización de Google.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   body body LinkGoogleAccountRequest true "Detalles para Vincular Cuenta"
// @Success 200 {object} MessageResponse "Cuenta de Google vinculada exitosamente."
// @Failure 400 {object} ErrorResponse "Solicitud inválida o el email de la cuenta de Google no coincide."
// @Failure 401 {object} ErrorResponse "Contraseña inválida."
// @Failure 404 {object} ErrorResponse "Usuario no encontrado."
// @Failure 500 {object} ErrorResponse "Fallo al verificar con Google o al vincular la cuenta."
// @Router /auth/google/link [post]
// LinkGoogleAccountHandler vincula una cuenta de Google a una existente, verificando la contraseña.
func (h *AuthHandler) LinkGoogleAccountHandler(w http.ResponseWriter, r *http.Request) {
    var req LinkGoogleAccountRequest
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

    if _, err := h.Store.LinkGoogleAccount(user.Email, googleUserInfo.ID); err != nil {
        h.writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Fallo al vincular la cuenta."})
        return
    }

    h.writeJSON(w, http.StatusOK, MessageResponse{Message: "Cuenta de Google vinculada exitosamente."})
}

// ProtectedHandler godoc
// @Summary Acceder a una ruta protegida
// @Description Ejemplo de una ruta que requiere autenticación JWT.
// @Tags protected
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} ProtectedResponse "Acceso concedido."
// @Failure 401 {object} ErrorResponse "No autorizado."
// @Router /api/profile [get]
// ProtectedHandler es un ejemplo de una ruta protegida.
func (h *AuthHandler) ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	h.writeJSON(w, http.StatusOK, ProtectedResponse{
		Message: "Esta es una ruta protegida.",
		UserID:  userID,
	})
}
