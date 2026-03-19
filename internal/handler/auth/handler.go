package auth

import (
	"encoding/json"
	"net/http"

	domain "rent-app/internal/domain/auth"
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
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	userInfo, err := h.authenticator.Authenticate(req.Email, req.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokenPair, err := h.authService.GenerateToken(userInfo.ID, userInfo.IsLandlord, userInfo.IsAdmin)
	if err != nil {
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
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	// Обновляем токены
	tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "token expired" {
			respondError(w, http.StatusUnauthorized, "token expired")
			return
		}
		if errMsg == "token has been revoked" {
			respondError(w, http.StatusUnauthorized, "token has been revoked")
			return
		}
		if errMsg == "invalid token type" {
			respondError(w, http.StatusUnauthorized, "invalid token type")
			return
		}
		respondError(w, http.StatusUnauthorized, "invalid token")
		return
	}

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
