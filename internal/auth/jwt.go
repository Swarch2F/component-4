package auth
// internal/auth/jwt.go

import (
	"fmt"
    "time"
    "github.com/golang-jwt/jwt"
    "github.com/google/uuid"
)

type Claims struct {
    UserID uuid.UUID `json:"sub"`
    Email  string    `json:"email"`
    Name   string    `json:"name"`
    Role   string    `json:"role"`
    jwt.StandardClaims
}

func GenerateToken(user *models.User, secret string) (string, error) {
    claims := &Claims{
        UserID: user.ID,
        Email:  user.Email,
        Name:   user.Name,
        Role:   string(user.Role),
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
            IssuedAt:  time.Now().Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
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
