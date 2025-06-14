// internal/models/user.go
package models

import (
    "time"
    "github.com/google/uuid"
)

type Role string

const (
    ROLE_ADMINISTRADOR Role = "administrador"
    ROLE_PROFESOR     Role = "profesor"
    ROLE_ESTUDIANTE   Role = "estudiante"
)

type User struct {
    ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
    Email     string    `json:"email" gorm:"uniqueIndex"`
    Name      string    `json:"name"`
    Password  *string   `json:"-"`
    Role      Role      `json:"role"`
    GoogleID  *string   `json:"google_id,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}