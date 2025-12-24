package user

import (
	"errors"
	"testing"
	"time"

	domain "rent-app/internal/domain/user"
)

type MockRepository struct {
	users            map[int]*domain.User
	usersByEmail     map[string]*domain.User
	nextID           int
	insertErr        error
	getByIDErr       error
	getByEmailErr    error
	getAllErr        error
	updateErr        error
	deleteErr        error
	resetPasswordErr error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:        make(map[int]*domain.User),
		usersByEmail: make(map[string]*domain.User),
		nextID:       1,
	}
}

func (m *MockRepository) InsertUser(u *domain.User) error {
	if m.insertErr != nil {
		return m.insertErr
	}
	u.ID = m.nextID
	m.nextID++
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	m.users[u.ID] = u
	m.usersByEmail[u.Email] = u
	return nil
}

func (m *MockRepository) GetUserByID(id int) (*domain.User, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	user, exists := m.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	userCopy := *user
	return &userCopy, nil
}

func (m *MockRepository) GetUserByEmail(email string) (*domain.User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	user, exists := m.usersByEmail[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	userCopy := *user
	return &userCopy, nil
}

func (m *MockRepository) GetAllUsers() ([]*domain.User, error) {
	if m.getAllErr != nil {
		return nil, m.getAllErr
	}
	users := make([]*domain.User, 0, len(m.users))
	for _, user := range m.users {
		userCopy := *user
		users = append(users, &userCopy)
	}
	return users, nil
}

func (m *MockRepository) UpdateUser(u *domain.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, exists := m.users[u.ID]; !exists {
		return errors.New("user not found")
	}
	u.UpdatedAt = time.Now()
	m.users[u.ID] = u
	if u.Email != "" {
		for email, user := range m.usersByEmail {
			if user.ID == u.ID {
				delete(m.usersByEmail, email)
				break
			}
		}
		m.usersByEmail[u.Email] = u
	}
	return nil
}

func (m *MockRepository) DeleteUser(id int) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	user, exists := m.users[id]
	if !exists {
		return errors.New("user not found")
	}
	delete(m.users, id)
	delete(m.usersByEmail, user.Email)
	return nil
}

func (m *MockRepository) ResetPassword(id int, password string) error {
	if m.resetPasswordErr != nil {
		return m.resetPasswordErr
	}
	user, exists := m.users[id]
	if !exists {
		return errors.New("user not found")
	}
	user.PasswordHash = password // можем не хэшировать пароль посколкьу для мока нет необходимости думать о безопасности
	return nil
}

func TestService_CreateUser(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		firstname     string
		lastname      string
		isLandlord    bool
		isAdmin       bool
		setupMock     func(*MockRepository)
		expectError   bool
		expectErrType error
		validate      func(*testing.T, *domain.User)
	}{
		{
			name:        "successful creation",
			email:       "test@example.com",
			password:    "password123",
			firstname:   "George",
			lastname:    "Washington",
			setupMock:   func(m *MockRepository) {},
			expectError: false,
			validate: func(t *testing.T, u *domain.User) {
				if u.Email != "test@example.com" {
					t.Errorf("expected email 'test@example.com', got '%s'", u.Email)
				}
				if u.FirstName != "George" {
					t.Errorf("expected firstname 'George', got '%s'", u.FirstName)
				}
				if u.LastName != "Washington" {
					t.Errorf("expected lastname 'Washington', got '%s'", u.LastName)
				}
				if u.PasswordHash != "" {
					t.Error("expected password hash to be cleared")
				}
				if u.ID == 0 {
					t.Error("expected user ID to be set")
				}
			},
		},
		{
			name:          "empty email",
			email:         "",
			password:      "password123",
			firstname:     "George",
			lastname:      "Washington",
			setupMock:     func(m *MockRepository) {},
			expectError:   true,
			expectErrType: ErrInvalidInput,
		},
		{
			name:          "empty password",
			email:         "test@example.com",
			password:      "",
			firstname:     "George",
			lastname:      "Washington",
			setupMock:     func(m *MockRepository) {},
			expectError:   true,
			expectErrType: ErrInvalidInput,
		},
		{
			name:          "empty firstname",
			email:         "test@example.com",
			password:      "password123",
			firstname:     "",
			lastname:      "Washington",
			setupMock:     func(m *MockRepository) {},
			expectError:   true,
			expectErrType: ErrInvalidInput,
		},
		{
			name:          "empty lastname",
			email:         "test@example.com",
			password:      "password123",
			firstname:     "George",
			lastname:      "",
			setupMock:     func(m *MockRepository) {},
			expectError:   true,
			expectErrType: ErrInvalidInput,
		},
		{
			name:      "email already exists",
			email:     "existing@example.com",
			password:  "password123",
			firstname: "George",
			lastname:  "Washington",
			setupMock: func(m *MockRepository) {
				existingUser := &domain.User{
					ID:        1,
					Email:     "existing@example.com",
					FirstName: "Existing",
					LastName:  "User",
				}
				m.users[1] = existingUser
				m.usersByEmail["existing@example.com"] = existingUser
			},
			expectError:   true,
			expectErrType: ErrEmailAlreadyTaken,
		},
		{
			name:      "repository insert error",
			email:     "test@example.com",
			password:  "password123",
			firstname: "George",
			lastname:  "Washington",
			setupMock: func(m *MockRepository) {
				m.insertErr = errors.New("database error")
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			test.setupMock(mockRepo)
			service := NewService(mockRepo)

			user, err := service.CreateUser(test.email, test.password, test.firstname, test.lastname, test.isLandlord, test.isAdmin)

			if test.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if test.expectErrType != nil {
					if !errors.Is(err, test.expectErrType) {
						t.Errorf("expected error type %s but got %s", test.expectErrType, err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if user == nil {
					t.Fatal("expected user but got got nil")
				}
				if test.validate != nil {
					test.validate(t, user)
				}
			}
		})
	}
}

func TestService_GetUserByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		setupMock     func(*MockRepository)
		expectError   bool
		expectErrType error
		validate      func(*testing.T, *domain.User)
	}{
		{
			name:   "successful get",
			userID: 1,
			setupMock: func(m *MockRepository) {
				user := &domain.User{
					ID:           1,
					Email:        "test@example.com",
					FirstName:    "George",
					LastName:     "Washington",
					PasswordHash: "hashed_password",
				}
				m.users[1] = user
			},
			expectError: false,
			validate: func(t *testing.T, u *domain.User) {
				if u.ID != 1 {
					t.Errorf("expected ID 1, got %d", u.ID)
				}
				if u.PasswordHash != "" {
					t.Error("expected password hash to be cleared")
				}
			},
		},
		{
			name:   "user not found",
			userID: 4,
			setupMock: func(m *MockRepository) {
				m.getByIDErr = errors.New("user not found")
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			test.setupMock(mockRepo)
			service := NewService(mockRepo)

			user, err := service.GetUserByID(test.userID)

			if test.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if user == nil {
					t.Fatal("expected user but got nil")
				}
				if test.validate != nil {
					test.validate(t, user)
				}
			}
		})
	}
}

func TestService_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		email         string
		setupMock     func(*MockRepository)
		expectError   bool
		expectErrType error
		validate      func(*testing.T, *domain.User)
	}{
		{
			name:   "successful get",
			userID: 1,
			email:  "test@example.com",
			setupMock: func(m *MockRepository) {
				user := &domain.User{
					ID:           1,
					Email:        "test@example.com",
					FirstName:    "George",
					LastName:     "Washington",
					PasswordHash: "hashed_password",
				}
				m.users[1] = user
				m.usersByEmail["test@example.com"] = user
			},
			expectError: false,
			validate: func(t *testing.T, u *domain.User) {
				if u.Email != "test@example.com" {
					t.Errorf("expected test@example.com email but got %s", u.Email)
				}
				if u.PasswordHash != "" {
					t.Error("expected password hash to be cleared")
				}
			},
		},
		{
			name:   "user not found",
			userID: 4,
			setupMock: func(m *MockRepository) {
				m.getByEmailErr = errors.New("user not found")
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			test.setupMock(mockRepo)
			service := NewService(mockRepo)

			user, err := service.GetUserByEmail(test.email)

			if test.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if user == nil {
					t.Fatal("expected user but got nil")
				}
				if test.validate != nil {
					test.validate(t, user)
				}
			}
		})
	}
}

func TestService_GetAllUsers(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockRepository)
		expectError bool
		validate    func(*testing.T, []*domain.User)
	}{
		{
			name: "successful get all",
			setupMock: func(m *MockRepository) {
				user1 := &domain.User{ID: 1, Email: "user1@example.com", PasswordHash: "hash1"}
				user2 := &domain.User{ID: 2, Email: "user2@example.com", PasswordHash: "hash2"}
				m.users[1] = user1
				m.users[2] = user2
			},
			expectError: false,
			validate: func(t *testing.T, users []*domain.User) {
				if len(users) != 2 {
					t.Errorf("expected 2 users, got %d", len(users))
				}
				for _, user := range users {
					if user.PasswordHash != "" {
						t.Error("expected password hash to be cleared for all users")
					}
				}
			},
		},
		{
			name: "empty list",
			setupMock: func(m *MockRepository) {
				// ноль пользователей
			},
			expectError: false,
			validate: func(t *testing.T, users []*domain.User) {
				if len(users) != 0 {
					t.Errorf("expected 0 users but got %d", len(users))
				}
			},
		},
		{
			name: "repository error",
			setupMock: func(m *MockRepository) {
				m.getAllErr = errors.New("database error")
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			test.setupMock(mockRepo)
			service := NewService(mockRepo)

			users, err := service.GetAllUsers()

			if test.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if test.validate != nil {
					test.validate(t, users)
				}
			}
		})
	}
}

func TestService_UpdateUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		email         *string
		firstname     *string
		lastname      *string
		isLandlord    *bool
		isAdmin       *bool
		setupMock     func(*MockRepository)
		expectError   bool
		expectErrType error
		validate      func(*testing.T, *domain.User)
	}{
		{
			name:   "successful update all fields",
			userID: 1,
			setupMock: func(m *MockRepository) {
				user := &domain.User{
					ID:         1,
					Email:      "old@example.com",
					FirstName:  "Old",
					LastName:   "Name",
					IsLandlord: false,
					IsAdmin:    false,
				}
				m.users[1] = user
				m.usersByEmail["old@example.com"] = user
			},
			email:       stringPtr("new@example.com"),
			firstname:   stringPtr("New"),
			lastname:    stringPtr("Name"),
			isLandlord:  boolPtr(true),
			isAdmin:     boolPtr(true),
			expectError: false,
			validate: func(t *testing.T, u *domain.User) {
				if u.Email != "new@example.com" {
					t.Errorf("expected email 'new@example.com', got '%s'", u.Email)
				}
				if u.FirstName != "New" {
					t.Errorf("expected firstname 'New', got '%s'", u.FirstName)
				}
				if u.PasswordHash != "" {
					t.Error("expected password hash to be cleared")
				}
			},
		},
		{
			name:   "update only email",
			userID: 1,
			setupMock: func(m *MockRepository) {
				user := &domain.User{
					ID:        1,
					Email:     "old@example.com",
					FirstName: "George",
					LastName:  "Washington",
				}
				m.users[1] = user
				m.usersByEmail["old@example.com"] = user
			},
			email:       stringPtr("new@example.com"),
			expectError: false,
			validate: func(t *testing.T, u *domain.User) {
				if u.Email != "new@example.com" {
					t.Errorf("expected email 'new@example.com', got '%s'", u.Email)
				}
				if u.FirstName != "George" {
					t.Errorf("expected firstname to remain 'George', got '%s'", u.FirstName)
				}
			},
		},
		{
			name:   "user not found",
			userID: 7,
			setupMock: func(m *MockRepository) {
				m.getByIDErr = errors.New("user not found")
			},
			expectError:   true,
			expectErrType: ErrUserNotFound,
		},
		{
			name:   "email already taken",
			userID: 1,
			setupMock: func(m *MockRepository) {
				user1 := &domain.User{ID: 1, Email: "user1@example.com"}
				user2 := &domain.User{ID: 2, Email: "user2@example.com"}
				m.users[1] = user1
				m.users[2] = user2
				m.usersByEmail["user1@example.com"] = user1
				m.usersByEmail["user2@example.com"] = user2
			},
			email:         stringPtr("user2@example.com"),
			expectError:   true,
			expectErrType: ErrEmailAlreadyTaken,
		},
		{
			name:   "email not changed (same email)",
			userID: 1,
			setupMock: func(m *MockRepository) {
				user := &domain.User{ID: 1, Email: "test@example.com", FirstName: "George"}
				m.users[1] = user
				m.usersByEmail["test@example.com"] = user
			},
			email:       stringPtr("test@example.com"),
			expectError: false,
			validate: func(t *testing.T, u *domain.User) {
				if u.Email != "test@example.com" {
					t.Errorf("expected email 'test@example.com', got '%s'", u.Email)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			test.setupMock(mockRepo)
			service := NewService(mockRepo)

			user, err := service.UpdateUser(test.userID, test.email, test.firstname, test.lastname, test.isLandlord, test.isAdmin)

			if test.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if test.expectErrType != nil {
					if !errors.Is(err, test.expectErrType) {
						t.Errorf("expected error type %s, got %s", test.expectErrType, err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if user == nil {
					t.Fatal("expected user, got nil")
				}
				if test.validate != nil {
					test.validate(t, user)
				}
			}
		})
	}
}

func TestService_DeleteUser(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		setupMock     func(*MockRepository)
		expectError   bool
		expectErrType error
	}{
		{
			name:   "successful delete",
			userID: 1,
			setupMock: func(m *MockRepository) {
				user := &domain.User{ID: 1, Email: "test@example.com"}
				m.users[1] = user
				m.usersByEmail["test@example.com"] = user
			},
			expectError: false,
		},
		{
			name:   "user not found",
			userID: 4,
			setupMock: func(m *MockRepository) {
				m.getByIDErr = errors.New("user not found")
			},
			expectError:   true,
			expectErrType: ErrUserNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			test.setupMock(mockRepo)
			service := NewService(mockRepo)

			err := service.DeleteUser(test.userID)

			if test.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if test.expectErrType != nil {
					if !errors.Is(err, test.expectErrType) {
						t.Errorf("expected error type %s, got %s", test.expectErrType, err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}
		})
	}
}

func TestService_ResetPassword(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		newPassword   string
		setupMock     func(*MockRepository)
		expectError   bool
		expectErrType error
	}{
		{
			name:        "successful password reset",
			userID:      1,
			newPassword: "newpassword123",
			setupMock: func(m *MockRepository) {
				user := &domain.User{ID: 1, Email: "test@example.com"}
				m.users[1] = user
			},
			expectError: false,
		},
		{
			name:        "empty password",
			userID:      1,
			newPassword: "",
			setupMock: func(m *MockRepository) {
				user := &domain.User{ID: 1, Email: "test@example.com"}
				m.users[1] = user
			},
			expectError:   true,
			expectErrType: ErrInvalidInput,
		},
		{
			name:        "user not found",
			userID:      999,
			newPassword: "newpassword123",
			setupMock: func(m *MockRepository) {
				m.getByIDErr = errors.New("user not found")
			},
			expectError:   true,
			expectErrType: ErrUserNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := NewMockRepository()
			test.setupMock(mockRepo)
			service := NewService(mockRepo)

			err := service.ResetPassword(test.userID, test.newPassword)

			if test.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if test.expectErrType != nil {
					if !errors.Is(err, test.expectErrType) {
						t.Errorf("expected error type %s but got %s", test.expectErrType, err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			}
		})
	}
}

// Вспомогательные функции
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
