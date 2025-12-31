package apartment

import (
	"errors"
	"testing"
	"time"

	domain "rent-app/internal/domain/apartment"
)

type MockRepository struct {
	apartments        map[int]*domain.Apartment
	insertError       error
	getByIDError      error
	getByOwnerIDError error
	getAllError       error
	updateError       error
	deleteError       error
	nextID            int
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		apartments: make(map[int]*domain.Apartment),
		nextID:     1,
	}
}

func (m *MockRepository) InsertApartment(a *domain.Apartment) error {
	if m.insertError != nil {
		return m.insertError
	}
	a.ID = m.nextID
	m.nextID++
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now
	m.apartments[a.ID] = a
	return nil
}

func (m *MockRepository) GetApartmentByID(id int) (*domain.Apartment, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	apartment, exists := m.apartments[id]
	if !exists {
		return nil, errors.New("apartment not found")
	}
	copy := *apartment
	return &copy, nil
}

func (m *MockRepository) GetApartmentsByOwnerID(ownerID int) ([]*domain.Apartment, error) {
	if m.getByOwnerIDError != nil {
		return nil, m.getByOwnerIDError
	}
	var result []*domain.Apartment
	for _, apt := range m.apartments {
		if apt.OwnerID == ownerID {
			result = append(result, apt)
		}
	}
	return result, nil
}

func (m *MockRepository) GetAllApartments() ([]*domain.Apartment, error) {
	if m.getAllError != nil {
		return nil, m.getAllError
	}
	var result []*domain.Apartment
	for _, apt := range m.apartments {
		result = append(result, apt)
	}
	return result, nil
}

func (m *MockRepository) UpdateApartment(a *domain.Apartment) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, ok := m.apartments[a.ID]; !ok {
		return errors.New("apartment not found")
	}
	a.UpdatedAt = time.Now()
	m.apartments[a.ID] = a
	return nil
}

func (m *MockRepository) DeleteApartment(id int) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, ok := m.apartments[id]; !ok {
		return errors.New("apartment not found")
	}
	delete(m.apartments, id)
	return nil
}

func TestService_CreateApartment(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	t.Run("successful creation", func(t *testing.T) {
		floor := 5
		totalFloors := 10
		req := domain.CreateApartmentRequest{
			OwnerID:     1,
			Status:      domain.StatusActive,
			PriceUnit:   domain.PerMonth,
			Title:       "Test Apartment",
			Price:       100000,
			Country:     "United States",
			City:        "New York",
			Address:     "123 Main St",
			AreaM2:      50,
			Rooms:       2,
			Floor:       &floor,
			TotalFloors: &totalFloors,
			PetsAllowed: false,
		}

		apartment, err := service.CreateApartment(req)
		if err != nil {
			t.Errorf("CreateApartment error %v, expected nil", err)
			return
		}

		if apartment == nil {
			t.Fatal("expected apartment to be created, got nil")
		}

		if apartment.ID == 0 {
			t.Error("expected apartment ID to be set")
		}

		if apartment.OwnerID != req.OwnerID {
			t.Errorf("expected OwnerID %d, got %d", req.OwnerID, apartment.OwnerID)
		}

		if apartment.Status != req.Status {
			t.Errorf("expected Status %s, got %s", req.Status, apartment.Status)
		}

		if apartment.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name string
			req  domain.CreateApartmentRequest
		}{
			{
				name: "missing ownerID",
				req: domain.CreateApartmentRequest{
					Status:    domain.StatusActive,
					PriceUnit: domain.PerMonth,
					Price:     100000,
					Country:   "US",
					City:      "NYC",
					Address:   "123 St",
					AreaM2:    50,
					Rooms:     2,
				},
			},
			{
				name: "invalid status",
				req: domain.CreateApartmentRequest{
					OwnerID:   1,
					Status:    domain.ApartmentStatus("invalid"),
					PriceUnit: domain.PerMonth,
					Price:     100000,
					Country:   "US",
					City:      "NYC",
					Address:   "123 St",
					AreaM2:    50,
					Rooms:     2,
				},
			},
			{
				name: "invalid price unit",
				req: domain.CreateApartmentRequest{
					OwnerID:   1,
					Status:    domain.StatusActive,
					PriceUnit: domain.PriceUnit("invalid"),
					Price:     100000,
					Country:   "US",
					City:      "NYC",
					Address:   "123 St",
					AreaM2:    50,
					Rooms:     2,
				},
			},
			{
				name: "zero price",
				req: domain.CreateApartmentRequest{
					OwnerID:   1,
					Status:    domain.StatusActive,
					PriceUnit: domain.PerMonth,
					Price:     0,
					Country:   "US",
					City:      "NYC",
					Address:   "123 St",
					AreaM2:    50,
					Rooms:     2,
				},
			},
			{
				name: "empty country",
				req: domain.CreateApartmentRequest{
					OwnerID:   1,
					Status:    domain.StatusActive,
					PriceUnit: domain.PerMonth,
					Price:     100000,
					Country:   "",
					City:      "NYC",
					Address:   "123 St",
					AreaM2:    50,
					Rooms:     2,
				},
			},
			{
				name: "empty city",
				req: domain.CreateApartmentRequest{
					OwnerID:   1,
					Status:    domain.StatusActive,
					PriceUnit: domain.PerMonth,
					Price:     100000,
					Country:   "US",
					City:      "",
					Address:   "123 St",
					AreaM2:    50,
					Rooms:     2,
				},
			},
			{
				name: "empty address",
				req: domain.CreateApartmentRequest{
					OwnerID:   1,
					Status:    domain.StatusActive,
					PriceUnit: domain.PerMonth,
					Price:     100000,
					Country:   "US",
					City:      "NYC",
					Address:   "",
					AreaM2:    50,
					Rooms:     2,
				},
			},
			{
				name: "zero areaM2",
				req: domain.CreateApartmentRequest{
					OwnerID:   1,
					Status:    domain.StatusActive,
					PriceUnit: domain.PerMonth,
					Price:     100000,
					Country:   "US",
					City:      "NYC",
					Address:   "123 St",
					AreaM2:    0,
					Rooms:     2,
				},
			},
			{
				name: "zero rooms",
				req: domain.CreateApartmentRequest{
					OwnerID:   1,
					Status:    domain.StatusActive,
					PriceUnit: domain.PerMonth,
					Price:     100000,
					Country:   "US",
					City:      "NYC",
					Address:   "123 St",
					AreaM2:    50,
					Rooms:     0,
				},
			},
			{
				name: "floor greater than total floors",
				req: domain.CreateApartmentRequest{
					OwnerID:     1,
					Status:      domain.StatusActive,
					PriceUnit:   domain.PerMonth,
					Price:       100000,
					Country:     "US",
					City:        "NYC",
					Address:     "123 St",
					AreaM2:      50,
					Rooms:       2,
					Floor:       intPtr(10),
					TotalFloors: intPtr(5),
				},
			},
			{
				name: "floor without total floors",
				req: domain.CreateApartmentRequest{
					OwnerID:   1,
					Status:    domain.StatusActive,
					PriceUnit: domain.PerMonth,
					Price:     100000,
					Country:   "US",
					City:      "NYC",
					Address:   "123 St",
					AreaM2:    50,
					Rooms:     2,
					Floor:     intPtr(5),
				},
			},
			{
				name: "total floors without floor",
				req: domain.CreateApartmentRequest{
					OwnerID:     1,
					Status:      domain.StatusActive,
					PriceUnit:   domain.PerMonth,
					Price:       100000,
					Country:     "US",
					City:        "NYC",
					Address:     "123 St",
					AreaM2:      50,
					Rooms:       2,
					TotalFloors: intPtr(10),
				},
			},
			{
				name: "negative floor",
				req: domain.CreateApartmentRequest{
					OwnerID:     1,
					Status:      domain.StatusActive,
					PriceUnit:   domain.PerMonth,
					Price:       100000,
					Country:     "US",
					City:        "NYC",
					Address:     "123 St",
					AreaM2:      50,
					Rooms:       2,
					Floor:       intPtr(-1),
					TotalFloors: intPtr(10),
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := service.CreateApartment(tt.req)
				if err == nil {
					t.Errorf("CreateApartment expected error, got nil")
				}
				if !errors.Is(err, ErrInvalidInput) {
					t.Errorf("CreateApartment error %v, expected ErrInvalidInput", err)
				}
			})
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := NewMockRepository()
		mockRepo.insertError = errors.New("database error")
		service := NewService(mockRepo)

		req := domain.CreateApartmentRequest{
			OwnerID:   1,
			Status:    domain.StatusActive,
			PriceUnit: domain.PerMonth,
			Price:     100000,
			Country:   "US",
			City:      "NYC",
			Address:   "123 St",
			AreaM2:    50,
			Rooms:     2,
		}

		_, err := service.CreateApartment(req)
		if err == nil {
			t.Error("CreateApartment expected error, got nil")
		}
	})
}

func TestService_GetApartmentByID(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	t.Run("successful get", func(t *testing.T) {
		floor := 5
		totalFloors := 10
		req := domain.CreateApartmentRequest{
			OwnerID:     1,
			Status:      domain.StatusActive,
			PriceUnit:   domain.PerMonth,
			Title:       "Test Apartment",
			Price:       100000,
			Country:     "US",
			City:        "NYC",
			Address:     "123 St",
			AreaM2:      50,
			Rooms:       2,
			Floor:       &floor,
			TotalFloors: &totalFloors,
		}

		created, _ := service.CreateApartment(req)
		apartment, err := service.GetApartmentByID(created.ID)
		if err != nil {
			t.Errorf("GetApartmentByID error %v, expected nil", err)
			return
		}

		if apartment == nil {
			t.Fatal("expected apartment, got nil")
		}

		if apartment.ID != created.ID {
			t.Errorf("expected ID %d, got %d", created.ID, apartment.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := service.GetApartmentByID(8)
		if err == nil {
			t.Error("GetApartmentByID expected error, got nil")
			return
		}

		if !errors.Is(err, ErrApartmentNotFound) {
			t.Errorf("GetApartmentByID error %v, expected ErrApartmentNotFound", err)
		}
	})
}

func TestService_GetApartmentByOwnerID(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	floor := 5
	totalFloors := 10
	req1 := domain.CreateApartmentRequest{
		OwnerID:     1,
		Status:      domain.StatusActive,
		PriceUnit:   domain.PerMonth,
		Price:       100000,
		Country:     "US",
		City:        "NYC",
		Address:     "123 St",
		AreaM2:      50,
		Rooms:       2,
		Floor:       &floor,
		TotalFloors: &totalFloors,
	}

	req2 := domain.CreateApartmentRequest{
		OwnerID:     1,
		Status:      domain.StatusActive,
		PriceUnit:   domain.PerMonth,
		Price:       200000,
		Country:     "US",
		City:        "LA",
		Address:     "456 Ave",
		AreaM2:      100,
		Rooms:       3,
		Floor:       &floor,
		TotalFloors: &totalFloors,
	}

	req3 := domain.CreateApartmentRequest{
		OwnerID:     2,
		Status:      domain.StatusActive,
		PriceUnit:   domain.PerMonth,
		Price:       150000,
		Country:     "Russia",
		City:        "Moscow",
		Address:     "789 ulitsa",
		AreaM2:      75,
		Rooms:       2,
		Floor:       &floor,
		TotalFloors: &totalFloors,
	}

	_, _ = service.CreateApartment(req1)
	_, _ = service.CreateApartment(req2)
	_, _ = service.CreateApartment(req3)

	t.Run("successful get by owner", func(t *testing.T) {
		apartments, err := service.GetApartmentByOwnerID(1)
		if err != nil {
			t.Errorf("GetApartmentByOwnerID error %v, expected nil", err)
			return
		}

		if len(apartments) != 2 {
			t.Errorf("expected 2 apartments, got %d", len(apartments))
		}

		for _, apt := range apartments {
			if apt.OwnerID != 1 {
				t.Errorf("expected OwnerID 1, got %d", apt.OwnerID)
			}
		}
	})

	t.Run("empty result", func(t *testing.T) {
		apartments, err := service.GetApartmentByOwnerID(999)
		if err != nil {
			t.Errorf("GetApartmentByOwnerID error %v, expected nil", err)
			return
		}

		if len(apartments) != 0 {
			t.Errorf("expected 0 apartments, got %d", len(apartments))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := NewMockRepository()
		mockRepo.getByOwnerIDError = errors.New("database error")
		service := NewService(mockRepo)

		_, err := service.GetApartmentByOwnerID(1)
		if err == nil {
			t.Error("GetApartmentByOwnerID expected error, got nil")
		}
	})
}

func TestService_GetAllApartments(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	floor := 5
	totalFloors := 10
	req1 := domain.CreateApartmentRequest{
		OwnerID:     1,
		Status:      domain.StatusActive,
		PriceUnit:   domain.PerMonth,
		Price:       100000,
		Country:     "US",
		City:        "NYC",
		Address:     "123 St",
		AreaM2:      50,
		Rooms:       2,
		Floor:       &floor,
		TotalFloors: &totalFloors,
	}

	req2 := domain.CreateApartmentRequest{
		OwnerID:     2,
		Status:      domain.StatusActive,
		PriceUnit:   domain.PerMonth,
		Price:       200000,
		Country:     "US",
		City:        "LA",
		Address:     "456 Ave",
		AreaM2:      100,
		Rooms:       3,
		Floor:       &floor,
		TotalFloors: &totalFloors,
	}

	_, _ = service.CreateApartment(req1)
	_, _ = service.CreateApartment(req2)

	t.Run("successful get all", func(t *testing.T) {
		apartments, err := service.GetAllApartments()
		if err != nil {
			t.Errorf("GetAllApartments error %v, expected nil", err)
			return
		}

		if len(apartments) != 2 {
			t.Errorf("expected 2 apartments, got %d", len(apartments))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := NewMockRepository()
		mockRepo.getAllError = errors.New("database error")
		service := NewService(mockRepo)

		_, err := service.GetAllApartments()
		if err == nil {
			t.Error("GetAllApartments expected error, got nil")
		}
	})
}

func TestService_UpdateApartment(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	floor := 5
	totalFloors := 10
	req := domain.CreateApartmentRequest{
		OwnerID:     1,
		Status:      domain.StatusActive,
		PriceUnit:   domain.PerMonth,
		Title:       "Original Title",
		Price:       100000,
		Country:     "US",
		City:        "NYC",
		Address:     "123 St",
		AreaM2:      50,
		Rooms:       2,
		Floor:       &floor,
		TotalFloors: &totalFloors,
		PetsAllowed: false,
	}

	created, _ := service.CreateApartment(req)

	t.Run("successful update", func(t *testing.T) {
		newTitle := "Updated Title"
		newPrice := 150000
		newStatus := domain.StatusArchived
		updateReq := domain.UpdateApartmentRequest{
			Title:  &newTitle,
			Price:  &newPrice,
			Status: &newStatus,
		}

		apartment, err := service.UpdateApartment(created.ID, updateReq)
		if err != nil {
			t.Errorf("UpdateApartment error %v, expected nil", err)
			return
		}

		if apartment.Title != newTitle {
			t.Errorf("expected Title %s, got %s", newTitle, apartment.Title)
		}

		if apartment.Price != newPrice {
			t.Errorf("expected Price %d, got %d", newPrice, apartment.Price)
		}

		if apartment.Status != newStatus {
			t.Errorf("expected Status %s, got %s", newStatus, apartment.Status)
		}
	})

	t.Run("not found", func(t *testing.T) {
		newTitle := "Updated Title"
		updateReq := domain.UpdateApartmentRequest{
			Title: &newTitle,
		}

		_, err := service.UpdateApartment(999, updateReq)
		if err == nil {
			t.Error("UpdateApartment expected error, got nil")
			return
		}

		if !errors.Is(err, ErrApartmentNotFound) {
			t.Errorf("UpdateApartment error %v, expected ErrApartmentNotFound", err)
		}
	})

	t.Run("validation errors for floor and totalFloors when apartment has no floor/totalFloors", func(t *testing.T) {
		reqNoFloor := domain.CreateApartmentRequest{
			OwnerID:   1,
			Status:    domain.StatusActive,
			PriceUnit: domain.PerMonth,
			Price:     100000,
			Country:   "US",
			City:      "NYC",
			Address:   "123 St",
			AreaM2:    50,
			Rooms:     2,
		}
		aptNoFloor, _ := service.CreateApartment(reqNoFloor)

		t.Run("floor without total floors", func(t *testing.T) {
			updateReq := domain.UpdateApartmentRequest{
				Floor: intPtr(5),
			}
			_, err := service.UpdateApartment(aptNoFloor.ID, updateReq)
			if err == nil {
				t.Error("UpdateApartment expected error, got nil")
			}
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("UpdateApartment error %v, expected ErrInvalidInput", err)
			}
		})

		t.Run("total floors without floor", func(t *testing.T) {
			updateReq := domain.UpdateApartmentRequest{
				TotalFloors: intPtr(10),
			}
			_, err := service.UpdateApartment(aptNoFloor.ID, updateReq)
			if err == nil {
				t.Error("UpdateApartment expected error, got nil")
			}
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("UpdateApartment error %v, expected ErrInvalidInput", err)
			}
		})
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name string
			req  domain.UpdateApartmentRequest
		}{
			{
				name: "invalid status",
				req: domain.UpdateApartmentRequest{
					Status: statusPtr(domain.ApartmentStatus("invalid")),
				},
			},
			{
				name: "invalid price unit",
				req: domain.UpdateApartmentRequest{
					PriceUnit: priceUnitPtr(domain.PriceUnit("invalid")),
				},
			},
			{
				name: "zero price",
				req: domain.UpdateApartmentRequest{
					Price: intPtr(0),
				},
			},
			{
				name: "empty country",
				req: domain.UpdateApartmentRequest{
					Country: stringPtr(""),
				},
			},
			{
				name: "empty city",
				req: domain.UpdateApartmentRequest{
					City: stringPtr(""),
				},
			},
			{
				name: "empty address",
				req: domain.UpdateApartmentRequest{
					Address: stringPtr(""),
				},
			},
			{
				name: "zero areaM2",
				req: domain.UpdateApartmentRequest{
					AreaM2: intPtr(0),
				},
			},
			{
				name: "zero rooms",
				req: domain.UpdateApartmentRequest{
					Rooms: intPtr(0),
				},
			},
			{
				name: "floor greater than total floors",
				req: domain.UpdateApartmentRequest{
					Floor:       intPtr(10),
					TotalFloors: intPtr(5),
				},
			},
			{
				name: "negative floor",
				req: domain.UpdateApartmentRequest{
					Floor:       intPtr(-1),
					TotalFloors: intPtr(10),
				},
			},
		}

		for _, e := range tests {
			t.Run(e.name, func(t *testing.T) {
				_, err := service.UpdateApartment(created.ID, e.req)
				if err == nil {
					t.Errorf("UpdateApartment expected error, got nil")
				}
				if !errors.Is(err, ErrInvalidInput) {
					t.Errorf("UpdateApartmen error %v, expected ErrInvalidInput", err)
				}
			})
		}
	})

	t.Run("repository error on update", func(t *testing.T) {
		mockRepo := NewMockRepository()
		service := NewService(mockRepo)

		floor := 5
		totalFloors := 10
		req := domain.CreateApartmentRequest{
			OwnerID:     1,
			Status:      domain.StatusActive,
			PriceUnit:   domain.PerMonth,
			Price:       100000,
			Country:     "US",
			City:        "NYC",
			Address:     "123 St",
			AreaM2:      50,
			Rooms:       2,
			Floor:       &floor,
			TotalFloors: &totalFloors,
		}

		created, _ := service.CreateApartment(req)
		mockRepo.updateError = errors.New("database error")

		newTitle := "Updated Title"
		updateReq := domain.UpdateApartmentRequest{
			Title: &newTitle,
		}

		_, err := service.UpdateApartment(created.ID, updateReq)
		if err == nil {
			t.Error("UpdateApartment expected error, got nil")
		}
	})
}

func TestService_DeleteApartment(t *testing.T) {
	mockRepo := NewMockRepository()
	service := NewService(mockRepo)

	floor := 5
	totalFloors := 10
	req := domain.CreateApartmentRequest{
		OwnerID:     1,
		Status:      domain.StatusActive,
		PriceUnit:   domain.PerMonth,
		Price:       100000,
		Country:     "US",
		City:        "NYC",
		Address:     "123 St",
		AreaM2:      50,
		Rooms:       2,
		Floor:       &floor,
		TotalFloors: &totalFloors,
	}

	created, _ := service.CreateApartment(req)

	t.Run("successful delete", func(t *testing.T) {
		err := service.DeleteApartment(created.ID)
		if err != nil {
			t.Errorf("DeleteApartment error %v, expected nil", err)
			return
		}

		_, err = service.GetApartmentByID(created.ID)
		if err == nil {
			t.Error("expected apartment to be deleted")
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := service.DeleteApartment(8)
		if err == nil {
			t.Error("DeleteApartment expected error, got nil")
			return
		}

		if !errors.Is(err, ErrApartmentNotFound) {
			t.Errorf("DeleteApartment error %v, expected ErrApartmentNotFound", err)
		}
	})

	t.Run("repository error on delete", func(t *testing.T) {
		mockRepo := NewMockRepository()
		service := NewService(mockRepo)

		floor := 5
		totalFloors := 10
		req := domain.CreateApartmentRequest{
			OwnerID:     1,
			Status:      domain.StatusActive,
			PriceUnit:   domain.PerMonth,
			Price:       100000,
			Country:     "US",
			City:        "NYC",
			Address:     "123 St",
			AreaM2:      50,
			Rooms:       2,
			Floor:       &floor,
			TotalFloors: &totalFloors,
		}

		created, _ := service.CreateApartment(req)
		mockRepo.deleteError = errors.New("database error")

		err := service.DeleteApartment(created.ID)
		if err == nil {
			t.Error("DeleteApartment expected error, got nil")
		}
	})
}

// вспомогательные функции
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func statusPtr(s domain.ApartmentStatus) *domain.ApartmentStatus {
	return &s
}

func priceUnitPtr(p domain.PriceUnit) *domain.PriceUnit {
	return &p
}
