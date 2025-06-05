package handlers

import (
    "net/http"
    "github.com/gorilla/mux"
    "auth-hibrida-go/internal/auth"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")
    password := r.FormValue("password")

    _, err := auth.AuthenticateUser(email, password)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Login successful"))
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
    // Ahora solo llamamos la función, que maneja la redirección internamente
    auth.AuthenticateWithGoogle(w, r)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Google authentication callback"))
}

func RegisterRoutes(r *mux.Router) {
    r.HandleFunc("/login", LoginHandler).Methods("POST")
    r.HandleFunc("/auth/google", GoogleLoginHandler).Methods("GET")
    r.HandleFunc("/auth/google/callback", GoogleCallbackHandler).Methods("GET")
}