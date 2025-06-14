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
    // Crear índices si no existen
    _, err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
        CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);
    `)
    if err != nil {
        panic(fmt.Sprintf("Error creating indexes: %v", err))
    }
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
        return nil, fmt.Errorf("error finding user: %w", err)
    }
    user.ID = id
    if password.Valid {
        user.Password = &password.String
    }
    user.Role = models.Role(role)
    if googleID.Valid {
        user.GoogleID = &googleID.String
    }
    return user, nil
}

func (s *UserStore) CreateNativeUser(email, name, password string, role models.Role) (*models.User, error) {
    // Iniciar transacción
    tx, err := s.db.Begin()
    if err != nil {
        return nil, fmt.Errorf("error starting transaction: %w", err)
    }
    defer tx.Rollback()

    // Verificar si el email ya existe
    var exists bool
    err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
    if err != nil {
        return nil, fmt.Errorf("error checking email existence: %w", err)
    }
    if exists {
        return nil, fmt.Errorf("email already exists")
    }

    // Generar hash de contraseña
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("error hashing password: %w", err)
    }

    id := uuid.New()
    now := time.Now()
    hashStr := string(hash)
    
    // Insertar usuario
    _, err = tx.Exec(
        `INSERT INTO users (id, email, name, password, role, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        id, email, name, hashStr, string(role), now, now,
    )
    if err != nil {
        return nil, fmt.Errorf("error creating user: %w", err)
    }

    // Commit de la transacción
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("error committing transaction: %w", err)
    }

    return &models.User{
        ID:        id,
        Email:     email,
        Name:      name,
        Password:  &hashStr,
        Role:      role,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

func (s *UserStore) CreateGoogleUser(email, name, googleID string, role models.Role) (*models.User, error) {
    // Iniciar transacción
    tx, err := s.db.Begin()
    if err != nil {
        return nil, fmt.Errorf("error starting transaction: %w", err)
    }
    defer tx.Rollback()

    // Verificar si el email ya existe
    var exists bool
    err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
    if err != nil {
        return nil, fmt.Errorf("error checking email existence: %w", err)
    }
    if exists {
        return nil, fmt.Errorf("email already exists")
    }

    id := uuid.New()
    now := time.Now()
    
    // Insertar usuario
    _, err = tx.Exec(
        `INSERT INTO users (id, email, name, role, google_id, created_at, updated_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        id, email, name, string(role), googleID, now, now,
    )
    if err != nil {
        return nil, fmt.Errorf("error creating user: %w", err)
    }

    // Commit de la transacción
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("error committing transaction: %w", err)
    }

    return &models.User{
        ID:        id,
        Email:     email,
        Name:      name,
        Role:      role,
        GoogleID:  &googleID,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

func (s *UserStore) SetPassword(email, password string) error {
    // Iniciar transacción
    tx, err := s.db.Begin()
    if err != nil {
        return fmt.Errorf("error starting transaction: %w", err)
    }
    defer tx.Rollback()

    // Verificar si el usuario existe
    var exists bool
    err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
    if err != nil {
        return fmt.Errorf("error checking user existence: %w", err)
    }
    if !exists {
        return fmt.Errorf("user not found")
    }

    // Generar hash de contraseña
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("error hashing password: %w", err)
    }

    // Actualizar contraseña
    _, err = tx.Exec(
        "UPDATE users SET password = $1, updated_at = $2 WHERE email = $3",
        string(hash), time.Now(), email,
    )
    if err != nil {
        return fmt.Errorf("error updating password: %w", err)
    }

    // Commit de la transacción
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("error committing transaction: %w", err)
    }

    return nil
}

func (s *UserStore) UpsertGoogleUser(email, name, googleID string, role models.Role) (*models.User, error) {
    // Iniciar transacción
    tx, err := s.db.Begin()
    if err != nil {
        return nil, fmt.Errorf("error starting transaction: %w", err)
    }
    defer tx.Rollback()

    user, err := s.FindByEmail(email)
    now := time.Now()
    if err != nil {
        if err.Error() == "user not found" {
            // No existe, crear nuevo usuario con Google
            return s.CreateGoogleUser(email, name, googleID, role)
        }
        return nil, fmt.Errorf("error finding user: %w", err)
    }

    // Ya existe: actualizamos GoogleID si no está seteado
    if user.GoogleID == nil {
        _, err := tx.Exec(
            `UPDATE users SET google_id = $1, name = $2, role = $3, updated_at = $4 WHERE email = $5`,
            googleID, name, string(role), now, email,
        )
        if err != nil {
            return nil, fmt.Errorf("error updating user: %w", err)
        }
        user.GoogleID = &googleID
        user.Name = name
        user.Role = role
        user.UpdatedAt = now
    } else {
        // Opcional: sincronizar nombre y rol si vienen distintos de Google
        if user.Name != name || user.Role != role {
            _, err := tx.Exec(
                `UPDATE users SET name = $1, role = $2, updated_at = $3 WHERE email = $4`,
                name, string(role), now, email,
            )
            if err != nil {
                return nil, fmt.Errorf("error updating user: %w", err)
            }
            user.Name = name
            user.Role = role
            user.UpdatedAt = now
        }
    }

    // Commit de la transacción
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("error committing transaction: %w", err)
    }

    return user, nil
}