package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	domain "rent-app/internal/domain/auth"
	serviceAuth "rent-app/internal/service/auth"
	"testing"
)

type MockAuthService struct {
	GenerateTokenFunc func(userID int, isLandLord bool, isAdmin bool) (*domain.TokenPair, error)
	RefreshTokenFunc  func(refreshTokenString string) (*domain.TokenPair, error)
}

func (m *MockAuthService) GenerateToken(userID int, isLandLord bool, isAdmin bool) (*domain.TokenPair, error) {
	if m.GenerateTokenFunc != nil {
		return m.GenerateTokenFunc(userID, isLandLord, isAdmin)
	}
	return nil, errors.New("GenerateToken not implemented")
}

func (m *MockAuthService) RefreshToken(refreshTokenString string) (*domain.TokenPair, error) {
	if m.RefreshTokenFunc != nil {
		return m.RefreshTokenFunc(refreshTokenString)
	}
	return nil, errors.New("RefreshToken not implemented")
}

func (m *MockAuthService) ValidateAccessToken(tokenString string) (*domain.AccessTokenClaims, error) {
	return nil, errors.New("ValidateAccessToken not implemented")
}

func (m *MockAuthService) ValidateRefreshToken(tokenString string) (*domain.RefreshTokenClaims, error) {
	return nil, errors.New("ValidateRefreshToken not implemented")
}

type MockUserAuthenticator struct {
	AuthenticateFunc func(email, passwod string) (*domain.AuthUserInfo, error)
}

func (m *MockUserAuthenticator) Authenticate(email, passwod string) (*domain.AuthUserInfo, error) {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(email, passwod)
	}
	return nil, errors.New("Authenticate not implemented")
}

func jsonBytes(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    []byte
		mockAuth       *MockAuthService
		mockAuthn      *MockUserAuthenticator
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			requestBody: jsonBytes(domain.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			}),
			mockAuth: &MockAuthService{
				GenerateTokenFunc: func(userID int, isLandLord, isAdmin bool) (*domain.TokenPair, error) {
					return &domain.TokenPair{
						AccessToken:  "access_token",
						RefreshToken: "refresh_token",
						ExpiresIn:    7200,
					}, nil
				},
			},
			mockAuthn: &MockUserAuthenticator{
				AuthenticateFunc: func(email, passwod string) (*domain.AuthUserInfo, error) {
					return &domain.AuthUserInfo{
						ID:         1,
						Email:      "test@example.com",
						IsLandlord: false,
						IsAdmin:    false,
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
				}

				var response domain.LoginResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.AccessToken != "access_token" {
					t.Errorf("expected AccessToken 'access_token', got %s", response.AccessToken)
				}

				if response.RefreshToken != "refresh_token" {
					t.Errorf("expected RefreshToken 'refresh_token', got %s", response.RefreshToken)
				}

				if response.User.ID != 1 {
					t.Errorf("expected User.ID 1, got %d", response.User.ID)
				}

				if response.User.Email != "test@example.com" {
					t.Errorf("expected User.Email 'test@example.com', got %s", response.User.Email)
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    []byte("invalid json"),
			mockAuth:       &MockAuthService{},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "invalid request body" {
					t.Errorf("expected error 'invalid request body', got %s", response["error"])
				}
			},
		},
		{
			name:           "missing email",
			requestBody:    jsonBytes(domain.LoginRequest{Password: "password123"}),
			mockAuth:       &MockAuthService{},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "email and password are required" {
					t.Errorf("expected error 'email and password are required', got %s", response["error"])
				}
			},
		},
		{
			name:           "missing password",
			requestBody:    jsonBytes(domain.LoginRequest{Email: "test@example.com"}),
			mockAuth:       &MockAuthService{},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "email and password are required" {
					t.Errorf("expected error 'email and password are required', got %s", response["error"])
				}
			},
		},
		{
			name: "invalid credentials",
			requestBody: jsonBytes(domain.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			}),
			mockAuth: &MockAuthService{},
			mockAuthn: &MockUserAuthenticator{
				AuthenticateFunc: func(email, password string) (*domain.AuthUserInfo, error) {
					return nil, errors.New("invalid credentials")
				},
			},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "invalid credentials" {
					t.Errorf("expected error 'invalid credentials', got %s", response["error"])
				}
			},
		},
		{
			name: "user not found",
			requestBody: jsonBytes(domain.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			}),
			mockAuth: &MockAuthService{},
			mockAuthn: &MockUserAuthenticator{
				AuthenticateFunc: func(email, password string) (*domain.AuthUserInfo, error) {
					return nil, errors.New("user not found")
				},
			},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "invalid credentials" {
					t.Errorf("expected error 'invalid credentials', got %s", response["error"])
				}
			},
		},
		{
			name: "token generation failure",
			requestBody: jsonBytes(domain.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			}),
			mockAuth: &MockAuthService{
				GenerateTokenFunc: func(userID int, isLandlord, isAdmin bool) (*domain.TokenPair, error) {
					return nil, errors.New("token generation failed")
				},
			},
			mockAuthn: &MockUserAuthenticator{
				AuthenticateFunc: func(email, password string) (*domain.AuthUserInfo, error) {
					return &domain.AuthUserInfo{
						ID:         1,
						Email:      "test@example.com",
						IsLandlord: false,
						IsAdmin:    false,
					}, nil
				},
			},
			expectedStatus: http.StatusInternalServerError,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "failed to generate tokens" {
					t.Errorf("expected error 'failed to generate tokens', got %s", response["error"])
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			handler := NewHandler(e.mockAuth, e.mockAuthn)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(e.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Login(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    []byte
		mockAuth       *MockAuthService
		mockAuthn      *MockUserAuthenticator
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "successful token refresh",
			requestBody: jsonBytes(map[string]string{
				"refresh_token": "valid_refresh_token",
			}),
			mockAuth: &MockAuthService{
				RefreshTokenFunc: func(refreshToken string) (*domain.TokenPair, error) {
					return &domain.TokenPair{
						AccessToken:  "new_access_token",
						RefreshToken: "new_refresh_token",
						ExpiresIn:    7200,
					}, nil
				},
			},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
				}

				var response domain.TokenPair
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response.AccessToken != "new_access_token" {
					t.Errorf("expected AccessToken 'new_access_token', got %s", response.AccessToken)
				}

				if response.RefreshToken != "new_refresh_token" {
					t.Errorf("expected RefreshToken 'new_refresh_token', got %s", response.RefreshToken)
				}

				if response.ExpiresIn != 7200 {
					t.Errorf("expected ExpiresIn 7200, got %d", response.ExpiresIn)
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    []byte("invalid json"),
			mockAuth:       &MockAuthService{},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "invalid request body" {
					t.Errorf("expected error 'invalid request body', got %s", response["error"])
				}
			},
		},
		{
			name:           "missing refresh_token",
			requestBody:    jsonBytes(map[string]string{}),
			mockAuth:       &MockAuthService{},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "refresh_token is required" {
					t.Errorf("expected error 'refresh_token is required', got %s", response["error"])
				}
			},
		},
		{
			name: "expired token",
			requestBody: jsonBytes(map[string]string{
				"refresh_token": "expired_token",
			}),
			mockAuth: &MockAuthService{
				RefreshTokenFunc: func(refreshToken string) (*domain.TokenPair, error) {
					return nil, serviceAuth.ErrTokenExpired
				},
			},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "token expired" {
					t.Errorf("expected error 'token expired', got %s", response["error"])
				}
			},
		},
		{
			name: "revoked token",
			requestBody: jsonBytes(map[string]string{
				"refresh_token": "revoked_token",
			}),
			mockAuth: &MockAuthService{
				RefreshTokenFunc: func(refreshToken string) (*domain.TokenPair, error) {
					return nil, serviceAuth.ErrTokenBlacklisted
				},
			},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "token has been revoked" {
					t.Errorf("expected error 'token has been revoked', got %s", response["error"])
				}
			},
		},
		{
			name: "invalid token type",
			requestBody: jsonBytes(map[string]string{
				"refresh_token": "access_token_used_as_refresh",
			}),
			mockAuth: &MockAuthService{
				RefreshTokenFunc: func(refreshToken string) (*domain.TokenPair, error) {
					return nil, serviceAuth.ErrInvalidTokenType
				},
			},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "invalid token type" {
					t.Errorf("expected error 'invalid token type', got %s", response["error"])
				}
			},
		},
		{
			name: "invalid token",
			requestBody: jsonBytes(map[string]string{
				"refresh_token": "invalid_token_string",
			}),
			mockAuth: &MockAuthService{
				RefreshTokenFunc: func(refreshToken string) (*domain.TokenPair, error) {
					return nil, serviceAuth.ErrInvalidToken
				},
			},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "invalid token" {
					t.Errorf("expected error 'invalid token', got %s", response["error"])
				}
			},
		},
		{
			name: "internal refresh error",
			requestBody: jsonBytes(map[string]string{
				"refresh_token": "valid_refresh_token",
			}),
			mockAuth: &MockAuthService{
				RefreshTokenFunc: func(refreshToken string) (*domain.TokenPair, error) {
					return nil, errors.New("db unavailable")
				},
			},
			mockAuthn:      &MockUserAuthenticator{},
			expectedStatus: http.StatusInternalServerError,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if response["error"] != "failed to refresh token" {
					t.Errorf("expected error 'failed to refresh token', got %s", response["error"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(tt.mockAuth, tt.mockAuthn)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewBuffer(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Refresh(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.validate != nil {
				tt.validate(t, rr)
			}
		})
	}
}
