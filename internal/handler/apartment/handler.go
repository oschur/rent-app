package apartment

import (
	"encoding/json"
	"errors"
	"net/http"
	userContext "rent-app/internal/context"
	domain "rent-app/internal/domain/apartment"
	"rent-app/internal/service/apartment"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service domain.Service
}

func NewHandler(service domain.Service) *Handler {
	return &Handler{
		service: service,
	}
}

type CreateApartmentRequest struct {
	Status      string `json:"status"`     // "active", "archived", "blocked"
	PriceUnit   string `json:"price_unit"` // "pernight", "permonth"
	Title       string `json:"title"`
	Price       int    `json:"price"`
	Country     string `json:"country"`
	City        string `json:"city"`
	Address     string `json:"address"`
	AreaM2      int    `json:"area_m2"`
	Rooms       int    `json:"rooms"`
	Floor       *int   `json:"floor,omitempty"`
	TotalFloors *int   `json:"total_floors,omitempty"`
	PetsAllowed bool   `json:"pets_allowed"`
}

type UpdateApartmentRequest struct {
	Status      *string `json:"status,omitempty"`     // "active", "archived", "blocked"
	PriceUnit   *string `json:"price_unit,omitempty"` // "pernight", "permonth"
	Title       *string `json:"title,omitempty"`
	Price       *int    `json:"price,omitempty"`
	Country     *string `json:"country,omitempty"`
	City        *string `json:"city,omitempty"`
	Address     *string `json:"address,omitempty"`
	AreaM2      *int    `json:"area_m2,omitempty"`
	Rooms       *int    `json:"rooms,omitempty"`
	Floor       *int    `json:"floor,omitempty"`
	TotalFloors *int    `json:"total_floors,omitempty"`
	PetsAllowed *bool   `json:"pets_allowed,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// CreateApartment godoc
// @Summary      Создание квартиры
// @Description  Создание новой квартиры. Доступно только для арендодателей и администраторов. Допустимые status: "active", "archived", "blocked". Допустимые price unit: "pernight", "permonth". Floor не может быть больше Totatfloors. Если установлен TotalFloors, Floor должен быть установлен, и наоборот. Все численные значения должны быть больше либо равны нулю.
// @Tags         apartments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      CreateApartmentRequest  true  "Данные квартиры"
// @Success      201      {object}  domain.Apartment
// @Failure      400      {object}  ErrorResponse  "Неверный запрос"
// @Failure      401      {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      403      {object}  ErrorResponse  "Доступ запрещен"
// @Failure      500      {object}  ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/apartments [post]
func (h *Handler) CreateApartment(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if !userInfo.IsLandlord && !userInfo.IsAdmin {
		respondError(w, http.StatusForbidden, "only landlords and admins can create apartments")
		return
	}

	var req CreateApartmentRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad request")
		return
	}

	status := domain.ApartmentStatus(req.Status)
	priceUnit := domain.PriceUnit(req.PriceUnit)

	serviceReq := domain.CreateApartmentRequest{
		OwnerID:     userInfo.UserID,
		Status:      status,
		PriceUnit:   priceUnit,
		Title:       req.Title,
		Price:       req.Price,
		Country:     req.Country,
		City:        req.City,
		Address:     req.Address,
		AreaM2:      req.AreaM2,
		Rooms:       req.Rooms,
		Floor:       req.Floor,
		TotalFloors: req.TotalFloors,
		PetsAllowed: req.PetsAllowed,
	}

	apt, err := h.service.CreateApartment(serviceReq)
	if err != nil {
		if errors.Is(err, apartment.ErrInvalidInput) {
			respondError(w, http.StatusBadRequest, "invalid input")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to create apartment")
		return
	}

	respondJSON(w, http.StatusCreated, apt)
}

// GetApartmentByID godoc
// @Summary      Получение квартиры по ID
// @Description  Получение информации о квартире по ID. Доступно для всех аутентифицированных пользователей.
// @Tags         apartments
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "ID квартиры"
// @Success      200  {object}  domain.Apartment
// @Failure      400  {object}  ErrorResponse  "Неверный запрос"
// @Failure      401  {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      404  {object}  ErrorResponse  "Квартира не найдена"
// @Router       /api/apartments/{id} [get]
func (h *Handler) GetApartmentByID(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid apartment ID")
		return
	}

	apt, err := h.service.GetApartmentByID(id)
	if err != nil {
		if errors.Is(err, apartment.ErrApartmentNotFound) {
			respondError(w, http.StatusNotFound, "apartment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get apartment")
		return
	}

	respondJSON(w, http.StatusOK, apt)
}

// GetApartmentsByOwnerID godoc
// @Summary      Получение квартир по ID владельца
// @Description  Получение списка квартир по ID владельца. Доступно для всех аутентифицированных пользователей.
// @Tags         apartments
// @Produce      json
// @Security     BearerAuth
// @Param        ownerID  path      int  true  "ID владельца"
// @Success      200      {array}   domain.Apartment
// @Failure      400      {object}  ErrorResponse  "Неверный запрос"
// @Failure      401      {object}  ErrorResponse  "Требуется аутентификация"
// @Router       /api/apartments/owner/{ownerID} [get]
func (h *Handler) GetApartmentsByOwnerID(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	ownerIDStr := chi.URLParam(r, "ownerID")
	ownerID, err := strconv.Atoi(ownerIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid owner ID")
		return
	}

	apartments, err := h.service.GetApartmentByOwnerID(ownerID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get apartments")
		return
	}

	respondJSON(w, http.StatusOK, apartments)
}

// GetAllApartments godoc
// @Summary      Получение всех квартир
// @Description  Получение списка всех квартир с возможностью фильтрации. Доступно для всех аутентифицированных пользователей.
// @Tags         apartments
// @Produce      json
// @Security     BearerAuth
// @Param        country       query     string  false  "Фильтр по стране"
// @Param        city          query     string  false  "Фильтр по городу"
// @Param        min_area_m2   query     int     false  "Минимальная площадь (м²)"
// @Param        max_area_m2   query     int     false  "Максимальная площадь (м²)"
// @Param        rooms         query     int     false  "Количество комнат"
// @Param        floor         query     int     false  "Этаж"
// @Param        pets_allowed  query     bool    false  "Разрешены ли животные"
// @Param        min_price     query     int     false  "Минимальная цена"
// @Param        max_price     query     int     false  "Максимальная цена"
// @Success      200           {array}   domain.Apartment
// @Failure      401           {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      500           {object}  ErrorResponse  "Внутренняя ошибка сервера"
// @Router       /api/apartments [get]
func (h *Handler) GetAllApartments(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	filters := parseFilters(r)

	apartments, err := h.service.GetAllApartments(filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get apartments")
		return
	}

	respondJSON(w, http.StatusOK, apartments)
}

// UpdateApartment godoc
// @Summary      Обновление квартиры
// @Description  Обновление данных квартиры. Владелец может обновлять только свои квартиры, администратор может обновлять любые. Допустимые status: "active", "archived", "blocked". Допустимые price unit: "pernight", "permonth". Floor не может быть больше Totatfloors. Если установлен TotalFloors, Floor должен быть установлен, и наоборот. Все численные значения должны быть больше либо равны нулю.
// @Tags         apartments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                  true  "ID квартиры"
// @Param        request  body      UpdateApartmentRequest  true  "Данные для обновления"
// @Success      200      {object}  domain.Apartment
// @Failure      400      {object}  ErrorResponse  "Неверный запрос"
// @Failure      401      {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      403      {object}  ErrorResponse  "Доступ запрещен"
// @Failure      404      {object}  ErrorResponse  "Квартира не найдена"
// @Router       /api/apartments/{id} [put]
func (h *Handler) UpdateApartment(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid apartment ID")
		return
	}

	apt, err := h.service.GetApartmentByID(id)
	if err != nil {
		if errors.Is(err, apartment.ErrApartmentNotFound) {
			respondError(w, http.StatusNotFound, "apartment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get apartment")
		return
	}

	// обычный пользователь может обновлять только свои квартиры
	if !userInfo.IsAdmin && apt.OwnerID != userInfo.UserID {
		respondError(w, http.StatusForbidden, "you can only update your own apartments")
		return
	}

	var req UpdateApartmentRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "bad request")
		return
	}

	serviceReq := domain.UpdateApartmentRequest{}
	if req.Status != nil {
		status := domain.ApartmentStatus(*req.Status)
		serviceReq.Status = &status
	}
	if req.PriceUnit != nil {
		priceUnit := domain.PriceUnit(*req.PriceUnit)
		serviceReq.PriceUnit = &priceUnit
	}
	if req.Title != nil {
		serviceReq.Title = req.Title
	}
	if req.Price != nil {
		serviceReq.Price = req.Price
	}
	if req.Country != nil {
		serviceReq.Country = req.Country
	}
	if req.City != nil {
		serviceReq.City = req.City
	}
	if req.Address != nil {
		serviceReq.Address = req.Address
	}
	if req.AreaM2 != nil {
		serviceReq.AreaM2 = req.AreaM2
	}
	if req.Rooms != nil {
		serviceReq.Rooms = req.Rooms
	}
	if req.Floor != nil {
		serviceReq.Floor = req.Floor
	}
	if req.TotalFloors != nil {
		serviceReq.TotalFloors = req.TotalFloors
	}
	if req.PetsAllowed != nil {
		serviceReq.PetsAllowed = req.PetsAllowed
	}

	updatedApartment, err := h.service.UpdateApartment(id, serviceReq)
	if err != nil {
		if errors.Is(err, apartment.ErrApartmentNotFound) {
			respondError(w, http.StatusNotFound, "apartment not found")
			return
		}
		if errors.Is(err, apartment.ErrInvalidInput) {
			respondError(w, http.StatusBadRequest, "invalid input")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to update apartment")
		return
	}

	respondJSON(w, http.StatusOK, updatedApartment)
}

// DeleteApartment godoc
// @Summary      Удаление квартиры
// @Description  Удаление квартиры. Владелец может удалять только свои квартиры, администратор может удалять любые.
// @Tags         apartments
// @Security     BearerAuth
// @Param        id   path      int  true  "ID квартиры"
// @Success      204  "Квартира успешно удалена"
// @Failure      400  {object}  ErrorResponse  "Неверный запрос"
// @Failure      401  {object}  ErrorResponse  "Требуется аутентификация"
// @Failure      403  {object}  ErrorResponse  "Доступ запрещен"
// @Failure      404  {object}  ErrorResponse  "Квартира не найдена"
// @Router       /api/apartments/{id} [delete]
func (h *Handler) DeleteApartment(w http.ResponseWriter, r *http.Request) {
	userInfo := userContext.GetUserInfo(r.Context())
	if userInfo == nil {
		respondError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid apartment ID")
		return
	}

	apt, err := h.service.GetApartmentByID(id)
	if err != nil {
		if errors.Is(err, apartment.ErrApartmentNotFound) {
			respondError(w, http.StatusNotFound, "apartment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get apartment")
		return
	}

	// обычный пользователь может удалять только свои квартиры
	if !userInfo.IsAdmin && apt.OwnerID != userInfo.UserID {
		respondError(w, http.StatusForbidden, "you can only delete your own apartments")
		return
	}

	err = h.service.DeleteApartment(id)
	if err != nil {
		if errors.Is(err, apartment.ErrApartmentNotFound) {
			respondError(w, http.StatusNotFound, "apartment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete apartment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseFilters(r *http.Request) *domain.ApartmentFilters {
	query := r.URL.Query()
	filters := &domain.ApartmentFilters{}

	if country := query.Get("country"); country != "" {
		filters.Country = &country
	}

	if city := query.Get("city"); city != "" {
		filters.City = &city
	}

	if minAreaStr := query.Get("min_area_m2"); minAreaStr != "" {
		if minArea, err := strconv.Atoi(minAreaStr); err == nil {
			filters.MinAreaM2 = &minArea
		}
	}

	if maxAreaStr := query.Get("max_area_m2"); maxAreaStr != "" {
		if maxArea, err := strconv.Atoi(maxAreaStr); err == nil {
			filters.MaxAreaM2 = &maxArea
		}
	}

	if roomsStr := query.Get("rooms"); roomsStr != "" {
		if rooms, err := strconv.Atoi(roomsStr); err == nil {
			filters.Rooms = &rooms
		}
	}

	if floorStr := query.Get("floor"); floorStr != "" {
		if floor, err := strconv.Atoi(floorStr); err == nil {
			filters.Floor = &floor
		}
	}

	if petsAllowedStr := query.Get("pets_allowed"); petsAllowedStr != "" {
		if petsAllowed, err := strconv.ParseBool(petsAllowedStr); err == nil {
			filters.PetsAllowed = &petsAllowed
		}
	}

	if minPriceStr := query.Get("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.Atoi(minPriceStr); err == nil {
			filters.MinPrice = &minPrice
		}
	}

	if maxPriceStr := query.Get("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.Atoi(maxPriceStr); err == nil {
			filters.MaxPrice = &maxPrice
		}
	}

	if filters.Country == nil && filters.City == nil && filters.MinAreaM2 == nil &&
		filters.MaxAreaM2 == nil && filters.Rooms == nil && filters.Floor == nil &&
		filters.PetsAllowed == nil && filters.MinPrice == nil && filters.MaxPrice == nil {
		return nil
	}

	return filters
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
