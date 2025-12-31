package apartment

import (
	"errors"
	"fmt"
	"time"

	domain "rent-app/internal/domain/apartment"
)

var (
	ErrApartmentNotFound = errors.New("apartment not found")
	ErrInvalidInput      = errors.New("invalid input")
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

var validStatuses = map[domain.ApartmentStatus]struct{}{
	domain.StatusActive:   {},
	domain.StatusArchived: {},
	domain.StatusBlocked:  {},
}

var validPriceUnits = map[domain.PriceUnit]struct{}{
	domain.PerNight: {},
	domain.PerMonth: {},
}

func (s *Service) CreateApartment(req domain.CreateApartmentRequest) (*domain.Apartment, error) {
	if req.OwnerID == 0 {
		return nil, fmt.Errorf("%w: ownerID is required", ErrInvalidInput)
	}
	if !isValidStatus(req.Status) {
		return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	if !isValidPriceUnit(req.PriceUnit) {
		return nil, fmt.Errorf("%w: invalid price unit", ErrInvalidInput)
	}
	if req.Price <= 0 {
		return nil, fmt.Errorf("%w: price must be positive", ErrInvalidInput)
	}
	if req.Country == "" {
		return nil, fmt.Errorf("%w: country is required", ErrInvalidInput)
	}
	if req.City == "" {
		return nil, fmt.Errorf("%w: city is required", ErrInvalidInput)
	}
	if req.Address == "" {
		return nil, fmt.Errorf("%w: address is required", ErrInvalidInput)
	}
	if req.AreaM2 <= 0 {
		return nil, fmt.Errorf("%w: areaM2 must be positive", ErrInvalidInput)
	}
	if req.Rooms <= 0 {
		return nil, fmt.Errorf("%w: rooms must be positive", ErrInvalidInput)
	}

	err := validateFloorAndTotalFloors(req.Floor, req.TotalFloors)
	if err != nil {
		return nil, err
	}

	apartment := &domain.Apartment{
		OwnerID:     req.OwnerID,
		Status:      req.Status,
		PriceUnit:   req.PriceUnit,
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
	err = s.repo.InsertApartment(apartment)
	if err != nil {
		return nil, fmt.Errorf("failed to insert apartment: %w", err)
	}
	return apartment, nil
}

func (s *Service) GetApartmentByID(id int) (*domain.Apartment, error) {
	apartment, err := s.repo.GetApartmentByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrApartmentNotFound)
	}
	return apartment, nil
}

func (s *Service) GetApartmentByOwnerID(ownerID int) ([]*domain.Apartment, error) {
	apartments, err := s.repo.GetApartmentsByOwnerID(ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments: %w", err)
	}
	return apartments, nil
}

func (s *Service) GetAllApartments(filters *domain.ApartmentFilters) ([]*domain.Apartment, error) {
	apartments, err := s.repo.GetAllApartments(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments: %w", err)
	}
	return apartments, nil
}

func (s *Service) UpdateApartment(id int, req domain.UpdateApartmentRequest) (*domain.Apartment, error) {
	apartment, err := s.repo.GetApartmentByID(id)
	if err != nil {
		return nil, fmt.Errorf("%w", ErrApartmentNotFound)
	}

	if req.Status != nil {
		if !isValidStatus(*req.Status) {
			return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
		}
		apartment.Status = *req.Status
	}
	if req.PriceUnit != nil {
		if !isValidPriceUnit(*req.PriceUnit) {
			return nil, fmt.Errorf("%w: invalid price unit", ErrInvalidInput)
		}
		apartment.PriceUnit = *req.PriceUnit
	}
	if req.Title != nil {
		apartment.Title = *req.Title
	}
	if req.Price != nil {
		if *req.Price <= 0 {
			return nil, fmt.Errorf("%w: price must be positive", ErrInvalidInput)
		}
		apartment.Price = *req.Price
	}
	if req.Country != nil {
		if *req.Country == "" {
			return nil, fmt.Errorf("%w: country cannot be empty", ErrInvalidInput)
		}
		apartment.Country = *req.Country
	}
	if req.City != nil {
		if *req.City == "" {
			return nil, fmt.Errorf("%w: city cannot be empty", ErrInvalidInput)
		}
		apartment.City = *req.City
	}
	if req.Address != nil {
		if *req.Address == "" {
			return nil, fmt.Errorf("%w: address cannot be empty", ErrInvalidInput)
		}
		apartment.Address = *req.Address
	}
	if req.AreaM2 != nil {
		if *req.AreaM2 <= 0 {
			return nil, fmt.Errorf("%w: areaM2 must be positive", ErrInvalidInput)
		}
		apartment.AreaM2 = *req.AreaM2
	}
	if req.Rooms != nil {
		if *req.Rooms <= 0 {
			return nil, fmt.Errorf("%w: rooms must be positive", ErrInvalidInput)
		}
		apartment.Rooms = *req.Rooms
	}
	if req.Floor != nil {
		apartment.Floor = req.Floor
	}
	if req.TotalFloors != nil {
		apartment.TotalFloors = req.TotalFloors
	}
	if req.PetsAllowed != nil {
		apartment.PetsAllowed = *req.PetsAllowed
	}

	err = validateFloorAndTotalFloors(apartment.Floor, apartment.TotalFloors)
	if err != nil {
		return nil, err
	}

	apartment.UpdatedAt = time.Now()

	err = s.repo.UpdateApartment(apartment)
	if err != nil {
		return nil, fmt.Errorf("failed to update apartment: %w", err)
	}
	return apartment, nil
}

func (s *Service) DeleteApartment(id int) error {
	_, err := s.repo.GetApartmentByID(id)
	if err != nil {
		return fmt.Errorf("%w", ErrApartmentNotFound)
	}
	err = s.repo.DeleteApartment(id)
	if err != nil {
		return fmt.Errorf("failed to delete apartment: %w", err)
	}
	return nil
}

// вспомогательные функции
func isValidStatus(status domain.ApartmentStatus) bool {
	_, ok := validStatuses[status]
	if !ok {
		return false
	}
	return true
}

func isValidPriceUnit(priceUnit domain.PriceUnit) bool {
	_, ok := validPriceUnits[priceUnit]
	if !ok {
		return false
	}
	return true
}

func validateFloorAndTotalFloors(floor *int, totalFloors *int) error {
	if floor != nil {
		if *floor <= 0 {
			return fmt.Errorf("%w: floor must be positive", ErrInvalidInput)
		}
		if totalFloors == nil {
			return fmt.Errorf("%w: expected total floors to be set when floor is set", ErrInvalidInput)
		}
	}
	if totalFloors != nil {
		if *totalFloors <= 0 {
			return fmt.Errorf("%w: total floors must be positive", ErrInvalidInput)
		}
		if floor == nil {
			return fmt.Errorf("%w: expected floor to be set when total floors are set", ErrInvalidInput)
		}
	}
	if floor != nil && totalFloors != nil {
		if *floor > *totalFloors {
			return fmt.Errorf("%w: floor cannot be greater than total floors", ErrInvalidInput)
		}
	}
	return nil
}
