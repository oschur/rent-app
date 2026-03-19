package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	userContext "rent-app/internal/context"
	domain "rent-app/internal/domain/user"
	"rent-app/internal/service/user"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service UserService
}

type UserService interface {
	CreateUser(email, password, firstname, lastname string, isLandlord, isAdmin bool) (*domain.User, error)
	GetUserByID(id int) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	GetAllUsers() ([]*domain.User, error)
	UpdateUser(id int, email, firstname, lastname *string, isLandlord, isAdmin *bool) (*domain.User, error)
	DeleteUser(id int) error
	ResetPassword(id int, password string) error
}

func NewHandler(service UserService) *Handler {
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

// CreateUser godoc
// @Summary      Создание пользователя
// @Description  Создание нового пользователя. Доступно без аутентификации, но только админы могут создавать других админов.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      CreateUserRequest  true  "Данные пользователя"
// @Success      201      {object}  domain.User
// @Failure      400      {object}  ErrorResponse  "Неверный запрос"
// @Failure      409      {object}  ErrorResponse  "Email уже занят"
// @Failure      500      {object}  ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/users [post]
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad request")
		return
	}

	userInfo := userContext.GetUserInfo(r.Context())

	// админа может создать только админ
	if userInfo == nil || !userInfo.IsAdmin {
		req.IsAdmin = false
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
		return
	}

	respondJSON(w, http.StatusCreated, u)
}

// UpdateUser godoc
// @Summary      Обновление пользователя
// @Description  Обновление данных пользователя. Обычные пользователи могут обновлять только свой профиль.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                true  "ID пользователя"
// @Param        request  body      UpdateUserRequest  true  "Данные для обновления"
// @Success      200      {object}  domain.User
// @Failure      400      {object}  ErrorResponse  "Неверный запрос"
// @Failure      401      {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      403      {object}  ErrorResponse  "Доступ запрещен"
// @Failure      404      {object}  ErrorResponse  "Пользователь не найден"
// @Failure      409      {object}  ErrorResponse  "Email уже занят"
// @Router       /api/users/{id} [put]
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// обычный пользователь может обновлять только себя
	if !userInfo.IsAdmin && userInfo.UserID != id {
		respondError(w, http.StatusForbidden, "you can only update your own profile")
		return
	}

	var req UpdateUserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad request")
		return
	}

	if !userInfo.IsAdmin {
		req.IsAdmin = nil
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
		if errors.Is(err, user.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	respondJSON(w, http.StatusOK, u)
}

// GetUserByID godoc
// @Summary      Получение пользователя по ID
// @Description  Получение информации о пользователе по ID. Требуются права администратора.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID пользователя"
// @Success      200  {object}  domain.User
// @Failure      400  {object}  ErrorResponse  "Неверный запрос"
// @Failure      401  {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      403  {object}  ErrorResponse  "Требуются права администратора"
// @Failure      404  {object}  ErrorResponse  "Пользователь не найден"
// @Router       /api/users/{id} [get]
func (h *Handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if !userInfo.IsAdmin {
		respondError(w, http.StatusForbidden, "admin access required")
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

// GetUserByEmail godoc
// @Summary      Получение пользователя по email
// @Description  Получение информации о пользователе по email. Требуются права администратора.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        email  path      string  true  "Email пользователя"
// @Success      200    {object}  domain.User
// @Failure      400    {object}  ErrorResponse  "Неверный запрос"
// @Failure      401    {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      403    {object}  ErrorResponse  "Требуются права администратора"
// @Failure      404    {object}  ErrorResponse  "Пользователь не найден"
// @Router       /api/users/email/{email} [get]
func (h *Handler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if !userInfo.IsAdmin {
		respondError(w, http.StatusForbidden, "admin access required")
		return
	}

	emailParam := chi.URLParam(r, "email")
	if emailParam == "" {
		respondError(w, http.StatusBadRequest, "email parameter is required")
		return
	}

	// приводим email%40example.com к email@example.com
	email, err := url.PathUnescape(emailParam)
	if err != nil {
		email = emailParam
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

// GetAllUsers godoc
// @Summary      Получение всех пользователей
// @Description  Получение списка всех пользователей. Требуются права администратора.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   domain.User
// @Failure      401  {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      403  {object}  ErrorResponse  "Требуются права администратора"
// @Failure      500  {object}  ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/users [get]
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if !userInfo.IsAdmin {
		respondError(w, http.StatusForbidden, "admin access required")
		return
	}

	users, err := h.service.GetAllUsers()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get users")
		return
	}

	respondJSON(w, http.StatusOK, users)
}

// ResetPassword godoc
// @Summary      Сброс пароля
// @Description  Сброс пароля пользователя. Пользователь может сбросить только свой пароль.
// @Tags         users
// @Accept       json
// @Security     BearerAuth
// @Param        id       path      int                  true  "ID пользователя"
// @Param        request  body      ResetPasswordRequest  true  "Новый пароль"
// @Success      204      "Пароль успешно изменен"
// @Failure      400      {object}  ErrorResponse  "Неверный запрос"
// @Failure      401      {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      403      {object}  ErrorResponse  "Доступ запрещен"
// @Failure      404      {object}  ErrorResponse  "Пользователь не найден"
// @Router       /api/users/{id}/reset-password [put]
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// пользователь может сбросить только свой пароль
	if userInfo.UserID != id {
		respondError(w, http.StatusForbidden, "you can only reset your own password")
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

// DeleteUser godoc
// @Summary      Удаление пользователя
// @Description  Удаление пользователя. Обычные пользователи могут удалить только свой аккаунт.
// @Tags         users
// @Security     BearerAuth
// @Param        id   path      int  true  "ID пользователя"
// @Success      204  "Пользователь успешно удален"
// @Failure      400  {object}  ErrorResponse  "Неверный запрос"
// @Failure      401  {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      403  {object}  ErrorResponse  "Доступ запрещен"
// @Failure      404  {object}  ErrorResponse  "Пользователь не найден"
// @Router       /api/users/{id} [delete]
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	// обычный пользователь может удалить только себя
	if !userInfo.IsAdmin && userInfo.UserID != id {
		respondError(w, http.StatusForbidden, "you can only delete your own account")
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
