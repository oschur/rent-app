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

// метод POST
// кладем в /api/apartments
// может только landlord или admin, владелец квартиры устанавливается из контекста
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

// кладем в /api/apartments/{id}
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

// кладем в /api/apartments/owner/{ownerID}
// любой аутентифицированный пользователь может получить квартиры любого владельца
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

// кладем в /api/apartments
// любой аутентифицированный пользователь может получить список всех квартир
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

// метод PUT
// кладем в /api/apartments/{id}
// владелец может обновлять только свои квартиры, админ может обновлять любые
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

// метод DELETE
// кладем в /api/apartments/{id}
// владелец может удалять только свои квартиры, админ может удалять любые
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
