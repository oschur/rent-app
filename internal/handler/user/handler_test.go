package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	userContext "rent-app/internal/context"
	domain "rent-app/internal/domain/user"
	"rent-app/internal/service/user"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
)

type MockService struct {
	CreateUserFunc     func(email, password, firstname, lastname string, isLandlord, isAdmin bool) (*domain.User, error)
	GetUserByIDFunc    func(id int) (*domain.User, error)
	GetUserByEmailFunc func(email string) (*domain.User, error)
	GetAllUsersFunc    func() ([]*domain.User, error)
	UpdateUserFunc     func(id int, email, firstname, lastname *string, isLandlord, isAdmin *bool) (*domain.User, error)
	DeleteUserFunc     func(id int) error
	ResetPasswordFunc  func(id int, password string) error
}

func (m *MockService) CreateUser(email, password, firstname, lastname string, isLandlord, isAdmin bool) (*domain.User, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(email, password, firstname, lastname, isLandlord, isAdmin)
	}
	return nil, errors.New("CreateUser not implemented")
}

func (m *MockService) GetUserByID(id int) (*domain.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(id)
	}
	return nil, errors.New("GetUserByID not implemented")
}

func (m *MockService) GetUserByEmail(email string) (*domain.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, errors.New("GetUserByEmail not implemented")
}

func (m *MockService) GetAllUsers() ([]*domain.User, error) {
	if m.GetAllUsersFunc != nil {
		return m.GetAllUsersFunc()
	}
	return nil, errors.New("GetAllUsers not implemented")
}

func (m *MockService) UpdateUser(id int, email, firstname, lastname *string, isLandlord, isAdmin *bool) (*domain.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(id, email, firstname, lastname, isLandlord, isAdmin)
	}
	return nil, errors.New("UpdateUser not implemented")
}

func (m *MockService) DeleteUser(id int) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(id)
	}
	return errors.New("DeleteUser not implemnted")
}

func (m *MockService) ResetPassword(id int, password string) error {
	if m.ResetPasswordFunc != nil {
		return m.ResetPasswordFunc(id, password)
	}
	return errors.New("ResetPassword not implemented")
}

func jsonBytes(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name           string
		requestbody    []byte
		mockService    *MockService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation",
			requestbody: jsonBytes(CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password",
				FirstName: "George",
				LastName:  "Washington",
			}),
			mockService: &MockService{
				CreateUserFunc: func(email, password, firstname, lastname string, isLandlord, isAdmin bool) (*domain.User, error) {
					return &domain.User{
						ID:        1,
						Email:     email,
						FirstName: firstname,
						LastName:  lastname,
					}, nil
				},
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var user domain.User
				if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
					t.Fatalf("failed to unmarshal repsonce, %s", err)
				}
				if user.Email != "test@example.com" {
					t.Errorf("expected user email to be test@example.com but got %s", user.Email)
				}
				if user.PasswordHash != "" {
					t.Errorf("expected password hash to be clear but it isn't")
				}
			},
		},
		{
			name:           "invalid json",
			requestbody:    []byte("invalid json"),
			mockService:    &MockService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid input",
			requestbody: jsonBytes(CreateUserRequest{
				Email:     "",
				Password:  "password",
				FirstName: "George",
				LastName:  "Washington",
			}),
			mockService: &MockService{
				CreateUserFunc: func(email, password, firstname, lastname string, isLandlord, isAdmin bool) (*domain.User, error) {
					return nil, user.ErrInvalidInput
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "email already taken",
			requestbody: jsonBytes(CreateUserRequest{
				Email:     "existing@example.com",
				Password:  "password",
				FirstName: "George",
				LastName:  "Washington",
			}),
			mockService: &MockService{
				CreateUserFunc: func(email, password, firstname, lastname string, isLandlord, isAdmin bool) (*domain.User, error) {
					return nil, user.ErrEmailAlreadyTaken
				},
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "service error",
			requestbody: jsonBytes(CreateUserRequest{
				Email:     "test@example.com",
				Password:  "password",
				FirstName: "George",
				LastName:  "Washington",
			}),
			mockService: &MockService{
				CreateUserFunc: func(email, password, firstname, lastname string, isLandlord, isAdmin bool) (*domain.User, error) {
					return nil, errors.New("service error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(e.requestbody))
			r.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := NewHandler(e.mockService)
			handler.CreateUser(rr, r)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d but got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestUserHandler_GetUserByID(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockService    *MockService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:   "success getting user",
			userID: "1",
			mockService: &MockService{
				GetUserByIDFunc: func(id int) (*domain.User, error) {
					return &domain.User{
						ID:        id,
						Email:     "test@example.com",
						FirstName: "George",
						LastName:  "Washington",
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var user domain.User
				if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
					t.Fatalf("failed to unmarshal repsonce, %s", err)
				}
				if user.ID != 1 {
					t.Errorf("expected user id to be 1 but got %d", user.ID)
				}
				if user.Email != "test@example.com" {
					t.Errorf("expected user email to be test@example.com but got %s", user.Email)
				}
				if user.PasswordHash != "" {
					t.Errorf("expected password hash to be clear but it isn't")
				}
			},
		},
		{
			name:   "user not found",
			userID: "8",
			mockService: &MockService{
				GetUserByIDFunc: func(id int) (*domain.User, error) {
					return nil, user.ErrUserNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid ID",
			userID:         "invalid",
			mockService:    &MockService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "service error",
			userID: "1",
			mockService: &MockService{
				GetUserByIDFunc: func(id int) (*domain.User, error) {
					return nil, errors.New("service error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/users/"+e.userID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", e.userID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			r = withAdminUser(r)

			rr := httptest.NewRecorder()

			handler := NewHandler(e.mockService)
			handler.GetUserByID(rr, r)
			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d but got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestUserHandler_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name           string
		userEmail      string
		mockService    *MockService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:      "success getting user",
			userEmail: "test@example.com",
			mockService: &MockService{
				GetUserByEmailFunc: func(email string) (*domain.User, error) {
					return &domain.User{
						ID:        1,
						Email:     email,
						FirstName: "George",
						LastName:  "Washington",
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var user domain.User
				if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
					t.Fatalf("failed to unmarshal repsonce, %s", err)
				}
				if user.ID != 1 {
					t.Errorf("expected user id to be 1 but got %d", user.ID)
				}
				if user.Email != "test@example.com" {
					t.Errorf("expected user email to be test@example.com but got %s", user.Email)
				}
				if user.PasswordHash != "" {
					t.Errorf("expected password hash to be clear but it isn't")
				}
			},
		},
		{
			name:      "user not found",
			userEmail: "nonexisting@example.com",
			mockService: &MockService{
				GetUserByEmailFunc: func(email string) (*domain.User, error) {
					return nil, user.ErrUserNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty email",
			userEmail:      "",
			mockService:    &MockService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "service error",
			userEmail: "test@example.com",
			mockService: &MockService{
				GetUserByEmailFunc: func(email string) (*domain.User, error) {
					return nil, errors.New("service error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/users/"+e.userEmail, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("email", e.userEmail)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			r = withAdminUser(r)

			rr := httptest.NewRecorder()

			handler := NewHandler(e.mockService)
			handler.GetUserByEmail(rr, r)
			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d but got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	tests := []struct {
		name           string
		mockService    *MockService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "successful getting",
			mockService: &MockService{
				GetAllUsersFunc: func() ([]*domain.User, error) {
					users := []*domain.User{
						{ID: 1, Email: "user1@example.com", FirstName: "user1", LastName: "first"},
						{ID: 1, Email: "user2@example.com", FirstName: "user2", LastName: "second"},
					}
					return users, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var users []*domain.User
				if err := json.Unmarshal(rr.Body.Bytes(), &users); err != nil {
					t.Fatalf("failed to unmarshal repsonce, %s", err)
				}
				if len(users) != 2 {
					t.Errorf("expected getting 2 users but got %d", len(users))
				}
			},
		},
		{
			name: "service error",
			mockService: &MockService{
				GetAllUsersFunc: func() ([]*domain.User, error) {
					return nil, errors.New("service error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			r = withAdminUser(r)
			rr := httptest.NewRecorder()

			handler := NewHandler(e.mockService)
			handler.GetAllUsers(rr, r)
			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d but got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestbody    []byte
		mockService    *MockService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:   "successful update",
			userID: "1",
			requestbody: jsonBytes(UpdateUserRequest{
				Email:     stringPtr("new@example.com"),
				FirstName: stringPtr("NewName"),
			}),
			mockService: &MockService{
				UpdateUserFunc: func(id int, email, firstname, lastname *string, isLandlord, isAdmin *bool) (*domain.User, error) {
					user := &domain.User{
						ID:        id,
						Email:     "old@example.com",
						FirstName: "OldName",
						LastName:  "OldLastName",
					}
					if email != nil {
						user.Email = *email
					}
					if firstname != nil {
						user.FirstName = *firstname
					}
					if lastname != nil {
						user.LastName = *lastname
					}
					return user, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var user domain.User
				if err := json.Unmarshal(rr.Body.Bytes(), &user); err != nil {
					t.Fatalf("failed to unmarshal repsonce, %s", err)
				}
				if user.Email != "new@example.com" {
					t.Errorf("expected user email to be new@example.com but got %s", user.Email)
				}
				if user.FirstName != "NewName" {
					t.Errorf("expected user firstname to be NewName but got %s", user.FirstName)
				}
				if user.LastName != "OldLastName" {
					t.Errorf("expected user lastname to be OldLastName but got %s", user.LastName)
				}
			},
		},
		{
			name:        "user not found",
			userID:      "8",
			requestbody: jsonBytes(UpdateUserRequest{}), // Пустой объект для частичного обновления
			mockService: &MockService{
				UpdateUserFunc: func(id int, email, firstname, lastname *string, isLandlord, isAdmin *bool) (*domain.User, error) {
					return nil, user.ErrUserNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			userID:         "invalid",
			requestbody:    nil,
			mockService:    &MockService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			userID:         "1",
			requestbody:    []byte("invalid"),
			mockService:    &MockService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "email already taken",
			userID: "1",
			requestbody: jsonBytes(UpdateUserRequest{
				Email: stringPtr("existing@example.com"),
			}),
			mockService: &MockService{
				UpdateUserFunc: func(id int, email, firstname, lastname *string, isLandlord, isAdmin *bool) (*domain.User, error) {
					return nil, user.ErrEmailAlreadyTaken
				},
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:        "service error",
			userID:      "1",
			requestbody: jsonBytes(UpdateUserRequest{}),
			mockService: &MockService{
				UpdateUserFunc: func(id int, email, firstname, lastname *string, isLandlord, isAdmin *bool) (*domain.User, error) {
					return nil, errors.New("service error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			var bodyBuffer *bytes.Buffer
			if e.requestbody != nil {
				bodyBuffer = bytes.NewBuffer(e.requestbody)
			} else {
				bodyBuffer = bytes.NewBuffer([]byte{})
			}
			r := httptest.NewRequest(http.MethodPut, "/api/users/"+e.userID, bodyBuffer)
			r.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", e.userID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			r = withAdminUser(r)

			rr := httptest.NewRecorder()
			handler := NewHandler(e.mockService)
			handler.UpdateUser(rr, r)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d but got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockService    *MockService
		expectedStatus int
	}{
		{
			name:   "succesful delete",
			userID: "1",
			mockService: &MockService{
				DeleteUserFunc: func(id int) error {
					return nil
				},
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "user not found",
			userID: "8",
			mockService: &MockService{
				DeleteUserFunc: func(id int) error {
					return user.ErrUserNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			userID:         "invalid",
			mockService:    &MockService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "service error",
			userID: "1",
			mockService: &MockService{
				DeleteUserFunc: func(id int) error {
					return errors.New("service error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodDelete, "/api/users/"+e.userID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", e.userID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			r = withAdminUser(r)

			rr := httptest.NewRecorder()
			handler := NewHandler(e.mockService)
			handler.DeleteUser(rr, r)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d but got %d", e.expectedStatus, rr.Code)
			}
		})
	}
}

func TestUserHandler_ResetPassword(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    []byte
		mockService    *MockService
		expectedStatus int
	}{
		{
			name:   "successful reset",
			userID: "1",
			requestBody: jsonBytes(ResetPasswordRequest{
				Password: "newpassword123",
			}),
			mockService: &MockService{
				ResetPasswordFunc: func(id int, newPassword string) error {
					return nil
				},
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid ID",
			userID:         "invalid",
			requestBody:    jsonBytes(ResetPasswordRequest{Password: "newpass"}),
			mockService:    &MockService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			userID:         "1",
			requestBody:    []byte("invalid json"),
			mockService:    &MockService{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "empty password",
			userID:      "1",
			requestBody: jsonBytes(ResetPasswordRequest{Password: ""}),
			mockService: &MockService{
				ResetPasswordFunc: func(id int, newPassword string) error {
					return user.ErrInvalidInput
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "user not found",
			userID:      "8",
			requestBody: jsonBytes(ResetPasswordRequest{Password: "newpassword123"}),
			mockService: &MockService{
				ResetPasswordFunc: func(id int, newPassword string) error {
					return user.ErrUserNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "service error",
			userID:      "1",
			requestBody: jsonBytes(ResetPasswordRequest{Password: "newpassword123"}),
			mockService: &MockService{
				ResetPasswordFunc: func(id int, newPassword string) error {
					return errors.New("database error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			var bodyBuffer *bytes.Buffer
			if e.requestBody != nil {
				bodyBuffer = bytes.NewBuffer(e.requestBody)
			} else {
				bodyBuffer = bytes.NewBuffer([]byte{})
			}

			r := httptest.NewRequest(http.MethodPost, "/api/users/"+e.userID+"/reset-password", bodyBuffer)
			r.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", e.userID)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
			uID, _ := strconv.Atoi(e.userID)
			r = withRegularUser(r, uID)
			rr := httptest.NewRecorder()
			handler := NewHandler(e.mockService)
			handler.ResetPassword(rr, r)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d but got %d", e.expectedStatus, rr.Code)
			}
		})
	}
}

// вспомогательные функции
func stringPtr(s string) *string {
	return &s
}

func withAdminUser(r *http.Request) *http.Request {
	userInfo := &userContext.UserInfo{
		UserID:     1,
		IsLandlord: false,
		IsAdmin:    true,
	}
	return r.WithContext(userContext.SetUserInfo(r.Context(), userInfo))
}

func withRegularUser(r *http.Request, userID int) *http.Request {
	userInfo := &userContext.UserInfo{
		UserID:     userID,
		IsLandlord: false,
		IsAdmin:    false,
	}
	return r.WithContext(userContext.SetUserInfo(r.Context(), userInfo))
}
