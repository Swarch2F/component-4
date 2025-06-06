package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken crea un nuevo token JWT para un usuario.
func GenerateToken(userID, jwtSecret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID, // "sub" (subject) es el ID del usuario
		"exp": time.Now().Add(time.Hour * 72).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtSecret))
}

// ValidateToken analiza un token y devuelve el ID del usuario si es válido.
func ValidateToken(tokenString, jwtSecret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["sub"].(string)
		return userID, nil
	}

	return "", fmt.Errorf("invalid token")
}
