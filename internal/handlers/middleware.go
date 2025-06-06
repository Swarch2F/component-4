package handlers

import (
	"component-4/internal/auth"
	"context"
	"net/http"
	"strings"
)

// ContextKey es un tipo personalizado para usar como clave en el contexto.
type ContextKey string

const UserIDKey ContextKey = "userID"

// AuthMiddleware protege las rutas verificando un token JWT.
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			userID, err := auth.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Pasa el userID al siguiente handler a través del contexto.
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
