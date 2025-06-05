package main

import (
    "log"
    "net/http"

    "auth-hibrida-go/internal/handlers"
    "auth-hibrida-go/config"
	"golang.org/x/crypto/bcrypt"
    "fmt"
)

func main() {
    // Load configuration
    cfg := config.LoadConfig()

    // Print the hashed password for demonstration purposes
    hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
    fmt.Println(string(hash))

    // Set up routes
    http.HandleFunc("/login", handlers.LoginHandler)
    http.HandleFunc("/auth/google", handlers.GoogleLoginHandler)
    http.HandleFunc("/auth/google/callback", handlers.GoogleCallbackHandler)

    // Start the server
    log.Printf("Starting server on :%s", cfg.Port)
    if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
        log.Fatalf("could not start server: %v", err)
    }
}