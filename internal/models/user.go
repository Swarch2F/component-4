package models

// User representa el modelo de usuario en el sistema.
// Se usan punteros para PasswordHash y GoogleID para poder representar valores nulos,
// lo que nos permite saber si un usuario tiene o no una contraseña o una cuenta de Google vinculada.
type User struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	PasswordHash *string `json:"-"` // Puntero para que pueda ser nulo
	GoogleID     *string `json:"-"` // Puntero para que pueda ser nulo
}
