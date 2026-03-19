package auth

import (
	"net/http"
	"strings"

	domain "rent-app/internal/domain/auth"
)

type AuthService interface {
	ValidateAccessToken(tokenString string) (*domain.AccessTokenClaims, error)
}

func AuthMiddleware(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			claims, err := authService.ValidateAccessToken(tokenString)
			if err != nil {
				errMsg := err.Error()
				if errMsg == "token expired" {
					http.Error(w, "token expired", http.StatusUnauthorized)
					return
				}
				if errMsg == "invalid token type" {
					http.Error(w, "invalid token type", http.StatusUnauthorized)
					return
				}
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := domain.SetUserContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo := domain.GetUserFromContext(r.Context())
		if userInfo == nil {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userInfo := domain.GetUserFromContext(r.Context())
		if userInfo == nil {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		if !userInfo.IsAdmin {
			http.Error(w, "admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// пока что используем это только для CreateUser
func OptionalAuthMiddleware(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]
			claims, err := authService.ValidateAccessToken(tokenString)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := domain.SetUserContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
