package store

import (
	"component-4/internal/models"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserStore simula una base de datos de usuarios en memoria.
type UserStore struct {
	mu    sync.Mutex
	users map[string]*models.User
}

// NewUserStore crea una nueva instancia del almacén de usuarios.
func NewUserStore() *UserStore {
	store := &UserStore{
		users: make(map[string]*models.User),
	}
	// Añadimos un usuario de ejemplo con contraseña para pruebas.
	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	stringHash := string(hash)
	store.users["test@example.com"] = &models.User{
		ID:           uuid.NewString(),
		Email:        "test@example.com",
		PasswordHash: &stringHash,
		GoogleID:     nil,
	}
	fmt.Println("Usuario de prueba 'test@example.com' con contraseña 'password123' creado.")
	return store
}

// FindByEmail busca un usuario por su email.
func (s *UserStore) FindByEmail(email string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user, ok := s.users[email]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

// CreateNativeUser crea un nuevo usuario con email y contraseña.
func (s *UserStore) CreateNativeUser(email, password string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[email]; exists {
		return nil, fmt.Errorf("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	stringHash := string(hash)

	user := &models.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: &stringHash,
		GoogleID:     nil,
	}

	s.users[email] = user
	return user, nil
}

// CreateGoogleUser crea un nuevo usuario que se registró solo con Google.
func (s *UserStore) CreateGoogleUser(email, googleID string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[email]; exists {
		return nil, fmt.Errorf("user already exists")
	}

	user := &models.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: nil, // Sin contraseña
		GoogleID:     &googleID,
	}

	s.users[email] = user
	return user, nil
}

// LinkGoogleAccount vincula una cuenta de Google a un usuario nativo existente.
func (s *UserStore) LinkGoogleAccount(email, googleID string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[email]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}

	user.GoogleID = &googleID
	return user, nil
}

// SetPassword establece o cambia la contraseña de un usuario.
func (s *UserStore) SetPassword(email, password string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[email]
	if !ok {
		return fmt.Errorf("user not found")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	stringHash := string(hash)
	user.PasswordHash = &stringHash
	return nil
}
