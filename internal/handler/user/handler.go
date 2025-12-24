package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"rent-app/internal/service/user"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *user.Service
}

func NewHandler(service *user.Service) *Handler {
	return &Handler{
		service: service,
	}
}

type CreateUserRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FirstName  string `json:"firstname"`
	LastName   string `json:"lastname"`
	IsLandlord bool   `json:"islandlord"`
	IsAdmin    bool   `json:"isadmin"`
}

type UpdateUserRequest struct {
	Email      *string `json:"email,omitempty"`
	FirstName  *string `json:"firstname,omitempty"`
	LastName   *string `json:"lastname,omitempty"`
	IsLandlord *bool   `json:"islandlord,omitempty"`
	IsAdmin    *bool   `json:"isadmin,omitempty"`
}

type ResetPasswordRequest struct {
	Password string `json:"password"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// метод POST(не PUT т.к. операция не идемпотентная)
// кладем в /api/users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad request")
		return
	}

	u, err := h.service.CreateUser(
		req.Email,
		req.Password,
		req.FirstName,
		req.LastName,
		req.IsLandlord,
		req.IsAdmin,
	)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyTaken) {
			respondError(w, http.StatusConflict, "email already taken")
			return
		}
		if errors.Is(err, user.ErrInvalidInput) {
			respondError(w, http.StatusBadRequest, "invalid input")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create user")
	}

	respondJSON(w, http.StatusCreated, u)
}

// тут нужен метод PUT т.к. эта операция идемпотентная
// кладем в /api/users/{id}
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req UpdateUserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad request")
		return
	}

	u, err := h.service.UpdateUser(
		id,
		req.Email,
		req.FirstName,
		req.LastName,
		req.IsLandlord,
		req.IsAdmin,
	)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyTaken) {
			respondError(w, http.StatusConflict, "email already taken")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	respondJSON(w, http.StatusOK, u)
}

// кладем в /api/users/{id}
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	u, err := h.service.GetUserByID(id)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, u)
}

// кладем в /api/users/{email}
func (h *Handler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if email == "" {
		respondError(w, http.StatusBadRequest, "email parameter is required")
		return
	}

	u, err := h.service.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, u)
}

// кладем в /api/users
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get users")
		return
	}

	respondJSON(w, http.StatusOK, users)
}

// метод POST
// кладем в /api/users/{id}/reset-password
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req ResetPasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad request")
		return
	}

	err = h.service.ResetPassword(id, req.Password)
	if err != nil {
		if errors.Is(err, user.ErrInvalidInput) {
			respondError(w, http.StatusBadRequest, "invalid password")
			return
		}
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to reset password")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// метод DELETE
// кладем в /api/users/{id}
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	err = h.service.DeleteUser(id)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondJSON(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
