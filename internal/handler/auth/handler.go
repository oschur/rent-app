package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	domain "rent-app/internal/domain/auth"
	serviceAuth "rent-app/internal/service/auth"

	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	authService   AuthService
	authenticator UserAuthenticator
}

type AuthService interface {
	GenerateToken(userID int, isLandLord bool, isAdmin bool) (*domain.TokenPair, error)
	RefreshToken(refreshTokenString string) (*domain.TokenPair, error)
}

type UserAuthenticator interface {
	Authenticate(email, password string) (*domain.AuthUserInfo, error)
}

func NewHandler(authService AuthService, authenticator UserAuthenticator) *Handler {
	return &Handler{
		authService:   authService,
		authenticator: authenticator,
	}
}

// Login godoc
// @Summary      Вход в систему
// @Description  Аутентификация пользователя и выдача JWT токенов (access и refresh)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      domain.LoginRequest  true  "Данные для входа"
// @Success      200      {object}  domain.LoginResponse
// @Failure      400      {object}  ErrorResponse  "Неверный запрос"
// @Failure      401      {object}  ErrorResponse  "Неверные учетные данные"
// @Failure      500      {object}  ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "auth.Login"
	log := slog.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("invalid request body", slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		log.Error("email and password are required")
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	userInfo, err := h.authenticator.Authenticate(req.Email, req.Password)
	if err != nil {
		log.Error("invalid credentials", slog.String("error", err.Error()))
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokenPair, err := h.authService.GenerateToken(userInfo.ID, userInfo.IsLandlord, userInfo.IsAdmin)
	if err != nil {
		log.Error("failed to generate tokens", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	response := domain.LoginResponse{
		TokenPair: *tokenPair,
		User: struct {
			ID         int    `json:"id"`
			Email      string `json:"email"`
			FirstName  string `json:"firstname"`
			LastName   string `json:"lastname"`
			IsLandlord bool   `json:"islandlord"`
			IsAdmin    bool   `json:"isadmin"`
		}{
			ID:         userInfo.ID,
			Email:      userInfo.Email,
			IsLandlord: userInfo.IsLandlord,
			IsAdmin:    userInfo.IsAdmin,
		},
	}

	log.Info("login successful", slog.Int("user_id", userInfo.ID))
	respondJSON(w, http.StatusOK, response)
}

// Refresh godoc
// @Summary      Обновление токена
// @Description  Обновление access токена с помощью refresh токена
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Refresh token"  SchemaExample({"refresh_token": "string"})
// @Success      200      {object}  domain.TokenPair
// @Failure      400      {object}  ErrorResponse  "Неверный запрос"
// @Failure      401      {object}  ErrorResponse  "Неверный или истекший токен"
// @Router       /api/auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	const op = "auth.Refresh"
	log := slog.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("invalid request body", slog.String("error", err.Error()))
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		log.Error("refresh_token is required")
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	// Обновляем токены
	tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, serviceAuth.ErrTokenExpired) {
			log.Error("token expired")
			respondError(w, http.StatusUnauthorized, "token expired")
			return
		}
		if errors.Is(err, serviceAuth.ErrTokenBlacklisted) {
			log.Error("token has been revoked")
			respondError(w, http.StatusUnauthorized, "token has been revoked")
			return
		}
		if errors.Is(err, serviceAuth.ErrInvalidTokenType) {
			log.Error("invalid token type")
			respondError(w, http.StatusUnauthorized, "invalid token type")
			return
		}
		if errors.Is(err, serviceAuth.ErrInvalidToken) {
			log.Error("invalid token")
			respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		log.Error("failed to refresh token", slog.String("error", err.Error()))
		respondError(w, http.StatusInternalServerError, "failed to refresh token")
		return
	}

	log.Info("token refreshed")
	respondJSON(w, http.StatusOK, tokenPair)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

type ErrorResponse struct {
	Error string `json:"error"`
}
