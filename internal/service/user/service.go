package user

import (
	"errors"
	"fmt"
	domain "rent-app/internal/domain/user"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyTaken = errors.New("email already taken")
	ErrInvalidInput      = errors.New("invalid input")
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func clearPassword(user *domain.User) *domain.User {
	user.PasswordHash = ""
	return user
}

func (s *Service) CreateUser(email, password, firstname, lastname string, isLandlord, isAdmin bool) (*domain.User, error) {
	if email == "" || password == "" || firstname == "" || lastname == "" {
		return nil, fmt.Errorf("%w: all fields are required", ErrInvalidInput)
	}

	existedUser, _ := s.repo.GetUserByEmail(email)
	if existedUser != nil {
		return nil, fmt.Errorf("%w: user with this email already exists", ErrEmailAlreadyTaken)
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		FirstName:    firstname,
		LastName:     lastname,
		PasswordHash: string(passHash),
		IsLandlord:   isLandlord,
		IsAdmin:      isAdmin,
	}

	err = s.repo.InsertUser(user)
	if err != nil {
		return nil, fmt.Errorf("error creating new user: %w", err)
	}

	return clearPassword(user), nil
}

func (s *Service) GetUserByID(id int) (*domain.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return clearPassword(user), nil
}

func (s *Service) GetUserByEmail(email string) (*domain.User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return clearPassword(user), nil
}

// GetUserByEmailForAuth возвращает пользователя с паролем для аутентификации
// в отличие от GetUserByEmail, этот метод не очищает хэш пароля
func (s *Service) GetUserByEmailForAuth(email string) (*domain.User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return user, nil
}

func (s *Service) GetAllUsers() ([]*domain.User, error) {
	users, err := s.repo.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("error getting users: %w", err)
	}

	for _, user := range users {
		clearPassword(user)
	}

	return users, nil
}

// передаем в функцию указатели на поля структуры User, поскольку мы должны разделять пустую строку и false как дефолтные значения полей от переданных значений полей
func (s *Service) UpdateUser(id int, email, firstname, lastname *string, isLandlord, isAdmin *bool) (*domain.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrUserNotFound)
	}

	if email != nil {
		// Проверяем, не занят ли email другим пользователем
		if *email != user.Email {
			existedUser, _ := s.repo.GetUserByEmail(*email)
			if existedUser != nil {
				return nil, fmt.Errorf("%w", ErrEmailAlreadyTaken)
			}
		}
		user.Email = *email
	}

	if firstname != nil {
		user.FirstName = *firstname
	}

	if lastname != nil {
		user.LastName = *lastname
	}

	if isLandlord != nil {
		user.IsLandlord = *isLandlord
	}

	if isAdmin != nil {
		user.IsAdmin = *isAdmin
	}

	err = s.repo.UpdateUser(user)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return clearPassword(user), nil
}

func (s *Service) DeleteUser(id int) error {
	_, err := s.repo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("%w", ErrUserNotFound)
	}

	err = s.repo.DeleteUser(id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

func (s *Service) ResetPassword(id int, newPassword string) error {
	if newPassword == "" {
		return fmt.Errorf("%w: password required", ErrInvalidInput)
	}

	_, err := s.repo.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("%w", ErrUserNotFound)
	}

	err = s.repo.ResetPassword(id, newPassword)
	if err != nil {
		return fmt.Errorf("error resetting password: %w", err)
	}

	return nil
}
