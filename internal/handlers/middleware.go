package handlers

import (
	"net/http"
	"strings"
)

// AuthMiddleware verifica si el usuario está autenticado.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Aquí se debe implementar la lógica para verificar la autenticación del usuario.
		// Por ejemplo, se puede verificar un token en las cabeceras de la solicitud.

		// Si el usuario no está autenticado, se puede devolver un error 401.
		if !isAuthenticated(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Si el usuario está autenticado, se llama al siguiente manejador.
		next.ServeHTTP(w, r)
	})
}

// isAuthenticated es una función auxiliar que verifica si el usuario está autenticado.
func isAuthenticated(r *http.Request) bool {
	// Aquí se debe implementar la lógica para verificar la autenticación.
	// Por ejemplo, se puede comprobar si hay un token en las cabeceras.
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		// Aquí se debe validar el token.
		return validateToken(token)
	}
	return false
}

// validateToken es una función que valida el token de autenticación.
func validateToken(token string) bool {
	// Implementar la lógica de validación del token.
	// Esto puede incluir la verificación de la firma del token y su expiración.
	return true // Cambiar esto según la lógica de validación real.
}