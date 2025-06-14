// internal/handlers/middleware.go
package handlers

import (
    "context"
    "net/http"
    // "os"
    "strings"
    "github.com/golang-jwt/jwt/v5"
    "github.com/gorilla/mux"
    "component-4/internal/auth"
)

func AuthMiddleware(JWTSecret string) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "No authorization header", http.StatusUnauthorized)
                return
            }

            tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
            claims := &auth.Claims{}

            token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
                return []byte(JWTSecret), nil
            })

            if err != nil || !token.Valid {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // Añadir claims al contexto
            ctx := context.WithValue(r.Context(), "user", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}