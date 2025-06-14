package auth

import (
	"component-4/internal/models"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

// CheckPassword compara la contraseña proporcionada con el hash almacenado.
func CheckPassword(user *models.User, password string) error {
	if user.Password == nil {
		return errors.New("user registered via OAuth and has no password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password))
	if err != nil {
		return errors.New("invalid password")
	}
	return nil
}
