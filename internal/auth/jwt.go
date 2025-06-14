package auth
// internal/auth/jwt.go

import (
	"fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "component-4/internal/models"
)

const TokenExpirySeconds = 24 * 60 * 60 // 24 horas en segundos

type Claims struct {
    UserID uuid.UUID `json:"sub"`
    Email  string    `json:"email"`
    Name   string    `json:"name"`
    Role   string    `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(user *models.User, secret string) (string, error) {
    claims := &Claims{
        UserID: user.ID,
        Email:  user.Email,
        Name:   user.Name,
        Role:   string(user.Role),
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpirySeconds * time.Second)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// ValidateToken analiza un token y devuelve el objeto *auth.Claims completo en lugar de solo el UserID.
func ValidateToken(tokenString, jwtSecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
