package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	userContext "rent-app/internal/context"
	domain "rent-app/internal/domain/auth"
)

type MockAuthService struct {
	ValidateAccessTokenFunc func(tokenString string) (*domain.AccessTokenClaims, error)
}

func (m *MockAuthService) GenerateToken(userID int, isLandlord, isAdmin bool) (*domain.TokenPair, error) {
	return nil, errors.New("GenerateToken not implemented")
}

func (m *MockAuthService) ValidateAccessToken(tokenString string) (*domain.AccessTokenClaims, error) {
	if m.ValidateAccessTokenFunc != nil {
		return m.ValidateAccessTokenFunc(tokenString)
	}
	return nil, errors.New("ValidateAccessToken not implemented")
}

func (m *MockAuthService) ValidateRefreshToken(tokenString string) (*domain.RefreshTokenClaims, error) {
	return nil, errors.New("ValidateRefreshToken not implemented")
}

func (m *MockAuthService) RefreshToken(refreshToken string) (*domain.TokenPair, error) {
	return nil, errors.New("RefreshToken not implemented")
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockService    *MockAuthService
		expectedStatus int
		expectedBody   string
		validate       func(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request)
	}{
		{
			name:       "successful authentication",
			authHeader: "Bearer valid_token",
			mockService: &MockAuthService{
				ValidateAccessTokenFunc: func(tokenString string) (*domain.AccessTokenClaims, error) {
					if tokenString != "valid_token" {
						t.Errorf("expected token 'valid_token', got %s", tokenString)
					}
					return &domain.AccessTokenClaims{
						UserID:     1,
						IsLandlord: false,
						IsAdmin:    false,
						TokenType:  domain.TokenTypeAccess,
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			mockService:    &MockAuthService{},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "authorization header required\n",
		},
		{
			name:           "invalid authorization header format - no space",
			authHeader:     "Bearertoken",
			mockService:    &MockAuthService{},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid authorization header format\n",
		},
		{
			name:           "invalid authorization header format - wrong prefix",
			authHeader:     "Basic token123",
			mockService:    &MockAuthService{},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid authorization header format\n",
		},
		{
			name:       "expired token",
			authHeader: "Bearer expired_token",
			mockService: &MockAuthService{
				ValidateAccessTokenFunc: func(tokenString string) (*domain.AccessTokenClaims, error) {
					return nil, errors.New("token expired")
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "token expired\n",
		},
		{
			name:       "invalid token type",
			authHeader: "Bearer refresh_token_used_as_access",
			mockService: &MockAuthService{
				ValidateAccessTokenFunc: func(tokenString string) (*domain.AccessTokenClaims, error) {
					return nil, errors.New("invalid token type")
				},
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid token type\n",
		},
		{
			name:       "sets user context with correct claims",
			authHeader: "Bearer valid_token",
			mockService: &MockAuthService{
				ValidateAccessTokenFunc: func(tokenString string) (*domain.AccessTokenClaims, error) {
					return &domain.AccessTokenClaims{
						UserID:     2,
						IsLandlord: true,
						IsAdmin:    true,
						TokenType:  domain.TokenTypeAccess,
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			validate: func(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request) {
				userInfo := domain.GetUserFromContext(req.Context())
				if userInfo == nil {
					t.Fatal("expected userInfo to be set in context")
				}
				if userInfo.UserID != 2 {
					t.Errorf("expected UserID 2, got %d", userInfo.UserID)
				}
				if !userInfo.IsLandlord {
					t.Error("expected IsLandlord to be true")
				}
				if !userInfo.IsAdmin {
					t.Error("expected IsAdmin to be true")
				}
			},
		},
		{
			name:       "successful authentication with user context",
			authHeader: "Bearer valid_token",
			mockService: &MockAuthService{
				ValidateAccessTokenFunc: func(tokenString string) (*domain.AccessTokenClaims, error) {
					return &domain.AccessTokenClaims{
						UserID:     1,
						IsLandlord: false,
						IsAdmin:    false,
						TokenType:  domain.TokenTypeAccess,
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			validate: func(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request) {
				userInfo := domain.GetUserFromContext(req.Context())
				if userInfo == nil {
					t.Fatal("expected userInfo to be set in context")
				}
				if userInfo.UserID != 1 {
					t.Errorf("expected UserID 1, got %d", userInfo.UserID)
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			var capturedReq *http.Request
			handler := AuthMiddleware(e.mockService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedReq = r
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if e.authHeader != "" {
				req.Header.Set("Authorization", e.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if e.expectedBody != "" && rr.Body.String() != e.expectedBody {
				t.Errorf("expected body %q, got %q", e.expectedBody, rr.Body.String())
			}

			if e.validate != nil && capturedReq != nil {
				e.validate(t, rr, capturedReq)
			}
		})
	}
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(req *http.Request) *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "user authenticated",
			setupContext: func(req *http.Request) *http.Request {
				userInfo := &userContext.UserInfo{
					UserID:     1,
					IsLandlord: false,
					IsAdmin:    false,
				}
				ctx := userContext.SetUserInfo(req.Context(), userInfo)
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "user not authenticated",
			setupContext: func(req *http.Request) *http.Request {
				return req
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "authentication required\n",
		},
		{
			name: "user authenticated with admin",
			setupContext: func(req *http.Request) *http.Request {
				userInfo := &userContext.UserInfo{
					UserID:     2,
					IsLandlord: true,
					IsAdmin:    true,
				}
				ctx := userContext.SetUserInfo(req.Context(), userInfo)
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			handler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = e.setupContext(req)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if rr.Body.String() != e.expectedBody {
				t.Errorf("expected body %q, got %q", e.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(req *http.Request) *http.Request
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "admin user",
			setupContext: func(req *http.Request) *http.Request {
				userInfo := &userContext.UserInfo{
					UserID:     1,
					IsLandlord: true,
					IsAdmin:    true,
				}
				ctx := userContext.SetUserInfo(req.Context(), userInfo)
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name: "non-admin user",
			setupContext: func(req *http.Request) *http.Request {
				userInfo := &userContext.UserInfo{
					UserID:     2,
					IsLandlord: true,
					IsAdmin:    false,
				}
				ctx := userContext.SetUserInfo(req.Context(), userInfo)
				return req.WithContext(ctx)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "admin access required\n",
		},
		{
			name: "user not authenticated",
			setupContext: func(req *http.Request) *http.Request {
				return req
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "authentication required\n",
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			handler := RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = e.setupContext(req)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if rr.Body.String() != e.expectedBody {
				t.Errorf("expected body %q, got %q", e.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockService    *MockAuthService
		expectedStatus int
		expectedBody   string
		validate       func(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request)
	}{
		{
			name:           "no authorization header - allows request",
			authHeader:     "",
			mockService:    &MockAuthService{},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			validate: func(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request) {
				userInfo := domain.GetUserFromContext(req.Context())
				if userInfo != nil {
					t.Error("expected userInfo to be nil when no auth header")
				}
			},
		},
		{
			name:       "valid token - sets user context",
			authHeader: "Bearer valid_token",
			mockService: &MockAuthService{
				ValidateAccessTokenFunc: func(tokenString string) (*domain.AccessTokenClaims, error) {
					return &domain.AccessTokenClaims{
						UserID:     1,
						IsLandlord: false,
						IsAdmin:    false,
						TokenType:  domain.TokenTypeAccess,
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			validate: func(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request) {
				userInfo := domain.GetUserFromContext(req.Context())
				if userInfo == nil {
					t.Fatal("expected userInfo to be set in context")
				}
				if userInfo.UserID != 1 {
					t.Errorf("expected UserID 1, got %d", userInfo.UserID)
				}
			},
		},
		{
			name:       "invalid token - allows request",
			authHeader: "Bearer invalid_token",
			mockService: &MockAuthService{
				ValidateAccessTokenFunc: func(tokenString string) (*domain.AccessTokenClaims, error) {
					return nil, errors.New("invalid token")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			validate: func(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request) {
				userInfo := domain.GetUserFromContext(req.Context())
				if userInfo != nil {
					t.Error("expected userInfo to be nil when token is invalid")
				}
			},
		},
		{
			name:       "expired token - allows request",
			authHeader: "Bearer expired_token",
			mockService: &MockAuthService{
				ValidateAccessTokenFunc: func(tokenString string) (*domain.AccessTokenClaims, error) {
					return nil, errors.New("token expired")
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			validate: func(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request) {
				userInfo := domain.GetUserFromContext(req.Context())
				if userInfo != nil {
					t.Error("expected userInfo to be nil when token is expired")
				}
			},
		},
		{
			name:       "sets user context with landlord and admin flags",
			authHeader: "Bearer valid_token",
			mockService: &MockAuthService{
				ValidateAccessTokenFunc: func(tokenString string) (*domain.AccessTokenClaims, error) {
					return &domain.AccessTokenClaims{
						UserID:     2,
						IsLandlord: true,
						IsAdmin:    true,
						TokenType:  domain.TokenTypeAccess,
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			validate: func(t *testing.T, rr *httptest.ResponseRecorder, req *http.Request) {
				userInfo := domain.GetUserFromContext(req.Context())
				if userInfo == nil {
					t.Fatal("expected userInfo to be set in context")
				}
				if userInfo.UserID != 2 {
					t.Errorf("expected UserID 2, got %d", userInfo.UserID)
				}
				if !userInfo.IsLandlord {
					t.Error("expected IsLandlord to be true")
				}
				if !userInfo.IsAdmin {
					t.Error("expected IsAdmin to be true")
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			var capturedReq *http.Request
			handler := OptionalAuthMiddleware(e.mockService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedReq = r
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if e.authHeader != "" {
				req.Header.Set("Authorization", e.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if rr.Body.String() != e.expectedBody {
				t.Errorf("expected body %q, got %q", e.expectedBody, rr.Body.String())
			}

			if e.validate != nil && capturedReq != nil {
				e.validate(t, rr, capturedReq)
			} else if e.validate != nil {
				e.validate(t, rr, req)
			}
		})
	}
}
