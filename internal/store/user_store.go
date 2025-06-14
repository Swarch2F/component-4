package store

import (
    "component-4/internal/models"
    "database/sql"
    "fmt"
    "time"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

type UserStore struct {
    db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
    return &UserStore{db: db}
}

func (s *UserStore) FindByEmail(email string) (*models.User, error) {
    user := &models.User{}
    row := s.db.QueryRow(
        `SELECT id, email, name, password, role, google_id, created_at, updated_at
         FROM users WHERE email = $1`, email)
    var id uuid.UUID
    var password, googleID sql.NullString
    var role string
    err := row.Scan(&id, &user.Email, &user.Name, &password, &role, &googleID, &user.CreatedAt, &user.UpdatedAt)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("user not found")
        }
        return nil, err
    }
    user.ID = id
    user.Password = password.String
    user.Role = models.Role(role)
    user.GoogleID = googleID.String
    return user, nil
}

func (s *UserStore) CreateNativeUser(email, name, password string, role models.Role) (*models.User, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    id := uuid.New()
    now := time.Now()
    _, err = s.db.Exec(
        `INSERT INTO users (id, email, name, password, role, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        id, email, name, string(hash), string(role), now, now,
    )
    if err != nil {
        return nil, err
    }
    return &models.User{
        ID:        id,
        Email:     email,
        Name:      name,
        Password:  string(hash),
        Role:      role,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

func (s *UserStore) CreateGoogleUser(email, name, googleID string, role models.Role) (*models.User, error) {
    id := uuid.New()
    now := time.Now()
    _, err := s.db.Exec(
        `INSERT INTO users (id, email, name, role, google_id, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        id, email, name, string(role), googleID, now, now,
    )
    if err != nil {
        return nil, err
    }
    return &models.User{
        ID:        id,
        Email:     email,
        Name:      name,
        Role:      role,
        GoogleID:  googleID,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

func (s *UserStore) SetPassword(email, password string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    _, err = s.db.Exec(
        "UPDATE users SET password = $1, updated_at = $2 WHERE email = $3",
        string(hash), time.Now(), email,
    )
    return err
}

func (s *UserStore) UpsertGoogleUser(email, name, googleID string, role models.Role) (*models.User, error) {
    user, err := s.FindByEmail(email)
    now := time.Now()
    if err != nil {
        // No existe, crear nuevo usuario con Google
        return s.CreateGoogleUser(email, name, googleID, role)
    }

    // Ya existe: actualizamos GoogleID si no está seteado
    if user.GoogleID == "" {
        _, err := s.db.Exec(
            `UPDATE users SET google_id = $1, name = $2, role = $3, updated_at = $4 WHERE email = $5`,
            googleID, name, string(role), now, email,
        )
        if err != nil {
            return nil, err
        }
        user.GoogleID = googleID
        user.Name = name
        user.Role = role
        user.UpdatedAt = now
    } else {
        // Opcional: sincronizar nombre y rol si vienen distintos de Google
        if user.Name != name || user.Role != role {
            _, err := s.db.Exec(
                `UPDATE users SET name = $1, role = $2, updated_at = $3 WHERE email = $4`,
                name, string(role), now, email,
            )
            if err != nil {
                return nil, err
            }
            user.Name = name
            user.Role = role
            user.UpdatedAt = now
        }
    }
    return user, nil
}