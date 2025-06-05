package auth

import (
	"errors"
	"os"
	"golang.org/x/crypto/bcrypt"
	"fmt"

)

// User represents a user in the system
type User struct {
	ID          string
	Email       string
	PasswordHash string
}

// AuthenticateUser verifies the user's email and password
func AuthenticateUser(email, password string) (*User, error) {
	// Here you would typically fetch the user from the database
	// For demonstration, let's assume we have a user with the following credentials
	storedEmail := os.Getenv("USER_EMAIL")
	storedPasswordHash := os.Getenv("USER_PASSWORD_HASH")

	fmt.Println("Stored Email:", storedEmail)
	fmt.Println("Stored Password Hash:", storedPasswordHash)
	fmt.Println("Provided Email:", email)
	fmt.Println("Provided Password:", password)

	if email != storedEmail {
		// If the email does not match, return an error, and print a message
		fmt.Println("Authentication failed: email does not match")
		return nil, errors.New("invalid email or password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(storedPasswordHash), []byte(password))
	if err != nil {
		// If the password does not match, return an error, and print a message
		fmt.Println("Authentication failed: password does not match")
		return nil, errors.New("invalid email or password")
	}

	// If authentication is successful, return the user
	return &User{
		ID:    "1", // This would typically be fetched from the database
		Email: storedEmail,
	}, nil
}