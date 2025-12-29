package auth

import (
	"encoding/json"
	"net/http"

	domain "rent-app/internal/domain/auth"
)

type Handler struct {
	authService   domain.Service
	authenticator domain.UserAuthenticator
}

func NewHandler(authService domain.Service, authenticator domain.UserAuthenticator) *Handler {
	return &Handler{
		authService:   authService,
		authenticator: authenticator,
	}
}

// метод POST
// кладем в /api/auth/login
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

// метод POST
// кладем в /api/auth/refresh
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
