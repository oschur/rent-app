package apartment

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"rent-app/internal/context"
	domain "rent-app/internal/domain/apartment"
	"rent-app/internal/service/apartment"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

type MockApartmentService struct {
	CreateApartmentFunc       func(req domain.CreateApartmentRequest) (*domain.Apartment, error)
	GetApartmentByIDFunc      func(id int) (*domain.Apartment, error)
	GetApartmentByOwnerIDFunc func(ownerID int) ([]*domain.Apartment, error)
	GetAllApartmentsFunc      func(filters *domain.ApartmentFilters) ([]*domain.Apartment, error)
	UpdateApartmentFunc       func(id int, req domain.UpdateApartmentRequest) (*domain.Apartment, error)
	DeleteApartmentFunc       func(id int) error
}

func (m *MockApartmentService) CreateApartment(req domain.CreateApartmentRequest) (*domain.Apartment, error) {
	if m.CreateApartmentFunc != nil {
		return m.CreateApartmentFunc(req)
	}
	return nil, errors.New("CreateApartment not implemented")
}

func (m *MockApartmentService) GetApartmentByID(id int) (*domain.Apartment, error) {
	if m.GetApartmentByIDFunc != nil {
		return m.GetApartmentByIDFunc(id)
	}
	return nil, errors.New("GetApartmentByID not implemented")
}

func (m *MockApartmentService) GetApartmentByOwnerID(ownerID int) ([]*domain.Apartment, error) {
	if m.GetApartmentByOwnerIDFunc != nil {
		return m.GetApartmentByOwnerIDFunc(ownerID)
	}
	return nil, errors.New("GetApartmentByOwnerID not implemented")
}

func (m *MockApartmentService) GetAllApartments(filters *domain.ApartmentFilters) ([]*domain.Apartment, error) {
	if m.GetAllApartmentsFunc != nil {
		return m.GetAllApartmentsFunc(filters)
	}
	return nil, errors.New("GetAllApartments not implemented")
}

func (m *MockApartmentService) UpdateApartment(id int, req domain.UpdateApartmentRequest) (*domain.Apartment, error) {
	if m.UpdateApartmentFunc != nil {
		return m.UpdateApartmentFunc(id, req)
	}
	return nil, errors.New("UpdateApartment not implemented")
}

func (m *MockApartmentService) DeleteApartment(id int) error {
	if m.DeleteApartmentFunc != nil {
		return m.DeleteApartmentFunc(id)
	}
	return errors.New("DeleteApartment not implemented")
}

func jsonBytes(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func createTestApartment() *domain.Apartment {
	floor := 5
	totalFloors := 10
	return &domain.Apartment{
		ID:          1,
		OwnerID:     1,
		Status:      domain.StatusActive,
		PriceUnit:   domain.PerMonth,
		Title:       "Test Apartment",
		Price:       1000,
		Country:     "USA",
		City:        "New York",
		Address:     "123 Main St",
		AreaM2:      50,
		Rooms:       2,
		Floor:       &floor,
		TotalFloors: &totalFloors,
		PetsAllowed: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestHandler_CreateApartment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    []byte
		userInfo       *context.UserInfo
		mockService    *MockApartmentService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation by landlord",
			requestBody: jsonBytes(CreateApartmentRequest{
				Status:      "active",
				PriceUnit:   "permonth",
				Title:       "Test Apartment",
				Price:       1000,
				Country:     "USA",
				City:        "New York",
				Address:     "123 Main St",
				AreaM2:      50,
				Rooms:       2,
				PetsAllowed: true,
			}),
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				CreateApartmentFunc: func(req domain.CreateApartmentRequest) (*domain.Apartment, error) {
					apt := createTestApartment()
					apt.OwnerID = req.OwnerID
					return apt, nil
				},
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if rr.Header().Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
				}

				var response domain.Apartment
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.OwnerID != 1 {
					t.Errorf("expected OwnerID 1, got %d", response.OwnerID)
				}
				if response.Title != "Test Apartment" {
					t.Errorf("expected Title 'Test Apartment', got %s", response.Title)
				}
			},
		},
		{
			name: "successful creation by admin",
			requestBody: jsonBytes(CreateApartmentRequest{
				Status:      "active",
				PriceUnit:   "permonth",
				Title:       "Test Apartment",
				Price:       1000,
				Country:     "USA",
				City:        "New York",
				Address:     "123 Main St",
				AreaM2:      50,
				Rooms:       2,
				PetsAllowed: true,
			}),
			userInfo: &context.UserInfo{
				UserID:     2,
				IsLandlord: false,
				IsAdmin:    true,
			},
			mockService: &MockApartmentService{
				CreateApartmentFunc: func(req domain.CreateApartmentRequest) (*domain.Apartment, error) {
					apt := createTestApartment()
					apt.OwnerID = req.OwnerID
					return apt, nil
				},
			},
			expectedStatus: http.StatusCreated,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response domain.Apartment
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.OwnerID != 2 {
					t.Errorf("expected OwnerID 2, got %d", response.OwnerID)
				}
			},
		},
		{
			name:           "missing authentication",
			requestBody:    jsonBytes(CreateApartmentRequest{}),
			userInfo:       nil,
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "authentication required" {
					t.Errorf("expected error 'authentication required', got %s", response.Error)
				}
			},
		},
		{
			name: "insufficient permissions",
			requestBody: jsonBytes(CreateApartmentRequest{
				Status:    "active",
				PriceUnit: "permonth",
			}),
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusForbidden,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "only landlords and admins can create apartments" {
					t.Errorf("expected error 'only landlords and admins can create apartments', got %s", response.Error)
				}
			},
		},
		{
			name:        "invalid request body",
			requestBody: []byte("invalid json"),
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "bad request" {
					t.Errorf("expected error 'bad request', got %s", response.Error)
				}
			},
		},
		{
			name: "service returns invalid input error",
			requestBody: jsonBytes(CreateApartmentRequest{
				Status:    "active",
				PriceUnit: "permonth",
			}),
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				CreateApartmentFunc: func(req domain.CreateApartmentRequest) (*domain.Apartment, error) {
					return nil, apartment.ErrInvalidInput
				},
			},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "invalid input" {
					t.Errorf("expected error 'invalid input', got %s", response.Error)
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			handler := NewHandler(e.mockService)
			req := httptest.NewRequest(http.MethodPost, "/api/apartments", bytes.NewBuffer(e.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if e.userInfo != nil {
				ctx := context.SetUserInfo(req.Context(), e.userInfo)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.CreateApartment(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestHandler_GetApartmentByID(t *testing.T) {
	tests := []struct {
		name           string
		apartmentID    string
		userInfo       *context.UserInfo
		mockService    *MockApartmentService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:        "successful get",
			apartmentID: "1",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					return createTestApartment(), nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response domain.Apartment
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.ID != 1 {
					t.Errorf("expected ID 1, got %d", response.ID)
				}
			},
		},
		{
			name:           "missing authentication",
			apartmentID:    "1",
			userInfo:       nil,
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "authentication required" {
					t.Errorf("expected error 'authentication required', got %s", response.Error)
				}
			},
		},
		{
			name:        "invalid apartment ID",
			apartmentID: "invalid",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "invalid apartment ID" {
					t.Errorf("expected error 'invalid apartment ID', got %s", response.Error)
				}
			},
		},
		{
			name:        "apartment not found",
			apartmentID: "8",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					return nil, apartment.ErrApartmentNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "apartment not found" {
					t.Errorf("expected error 'apartment not found', got %s", response.Error)
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			handler := NewHandler(e.mockService)
			req := httptest.NewRequest(http.MethodGet, "/api/apartments/"+e.apartmentID, nil)

			if e.userInfo != nil {
				ctx := context.SetUserInfo(req.Context(), e.userInfo)
				req = req.WithContext(ctx)
			}

			r := chi.NewRouter()
			r.Get("/api/apartments/{id}", handler.GetApartmentByID)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestHandler_GetApartmentsByOwnerID(t *testing.T) {
	tests := []struct {
		name           string
		ownerID        string
		userInfo       *context.UserInfo
		mockService    *MockApartmentService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:    "successful get",
			ownerID: "1",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByOwnerIDFunc: func(ownerID int) ([]*domain.Apartment, error) {
					return []*domain.Apartment{createTestApartment()}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response []*domain.Apartment
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(response) != 1 {
					t.Errorf("expected 1 apartment, got %d", len(response))
				}
			},
		},
		{
			name:           "missing authentication",
			ownerID:        "1",
			userInfo:       nil,
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "authentication required" {
					t.Errorf("expected error 'authentication required', got %s", response.Error)
				}
			},
		},
		{
			name:    "invalid owner ID",
			ownerID: "invalid",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "invalid owner ID" {
					t.Errorf("expected error 'invalid owner ID', got %s", response.Error)
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			handler := NewHandler(e.mockService)
			req := httptest.NewRequest(http.MethodGet, "/api/apartments/owner/"+e.ownerID, nil)

			if e.userInfo != nil {
				ctx := context.SetUserInfo(req.Context(), e.userInfo)
				req = req.WithContext(ctx)
			}

			r := chi.NewRouter()
			r.Get("/api/apartments/owner/{ownerID}", handler.GetApartmentsByOwnerID)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestHandler_GetAllApartments(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		userInfo       *context.UserInfo
		mockService    *MockApartmentService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:        "successful get without filters",
			queryParams: "",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetAllApartmentsFunc: func(filters *domain.ApartmentFilters) ([]*domain.Apartment, error) {
					if filters != nil {
						t.Error("expected nil filters")
					}
					return []*domain.Apartment{createTestApartment()}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response []*domain.Apartment
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(response) != 1 {
					t.Errorf("expected 1 apartment, got %d", len(response))
				}
			},
		},
		{
			name:        "successful get with filters",
			queryParams: "?country=USA&city=NewYork&min_price=500&max_price=2000",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetAllApartmentsFunc: func(filters *domain.ApartmentFilters) ([]*domain.Apartment, error) {
					if filters == nil {
						t.Error("expected non-nil filters")
					}
					if filters.Country == nil || *filters.Country != "USA" {
						t.Error("expected country filter to be 'USA'")
					}
					if filters.City == nil || *filters.City != "NewYork" {
						t.Error("expected city filter to be 'NewYork'")
					}
					if filters.MinPrice == nil || *filters.MinPrice != 500 {
						t.Error("expected min_price filter to be 500")
					}
					if filters.MaxPrice == nil || *filters.MaxPrice != 2000 {
						t.Error("expected max_price filter to be 2000")
					}
					return []*domain.Apartment{createTestApartment()}, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response []*domain.Apartment
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
			},
		},
		{
			name:           "missing authentication",
			queryParams:    "",
			userInfo:       nil,
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "authentication required" {
					t.Errorf("expected error 'authentication required', got %s", response.Error)
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			handler := NewHandler(e.mockService)
			req := httptest.NewRequest(http.MethodGet, "/api/apartments"+e.queryParams, nil)

			if e.userInfo != nil {
				ctx := context.SetUserInfo(req.Context(), e.userInfo)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.GetAllApartments(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestHandler_UpdateApartment(t *testing.T) {
	tests := []struct {
		name           string
		apartmentID    string
		requestBody    []byte
		userInfo       *context.UserInfo
		mockService    *MockApartmentService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:        "successful update by owner",
			apartmentID: "1",
			requestBody: jsonBytes(UpdateApartmentRequest{
				Title: stringPtr("Updated Title"),
				Price: intPtr(1500),
			}),
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					apt := createTestApartment()
					apt.OwnerID = 1
					return apt, nil
				},
				UpdateApartmentFunc: func(id int, req domain.UpdateApartmentRequest) (*domain.Apartment, error) {
					apt := createTestApartment()
					if req.Title != nil {
						apt.Title = *req.Title
					}
					if req.Price != nil {
						apt.Price = *req.Price
					}
					return apt, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response domain.Apartment
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Title != "Updated Title" {
					t.Errorf("expected Title 'Updated Title', got %s", response.Title)
				}
				if response.Price != 1500 {
					t.Errorf("expected Price 1500, got %d", response.Price)
				}
			},
		},
		{
			name:        "successful update by admin",
			apartmentID: "1",
			requestBody: jsonBytes(UpdateApartmentRequest{
				Title: stringPtr("Admin Updated Title"),
			}),
			userInfo: &context.UserInfo{
				UserID:     2,
				IsLandlord: false,
				IsAdmin:    true,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					apt := createTestApartment()
					apt.OwnerID = 1
					return apt, nil
				},
				UpdateApartmentFunc: func(id int, req domain.UpdateApartmentRequest) (*domain.Apartment, error) {
					apt := createTestApartment()
					if req.Title != nil {
						apt.Title = *req.Title
					}
					return apt, nil
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response domain.Apartment
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Title != "Admin Updated Title" {
					t.Errorf("expected Title 'Admin Updated Title', got %s", response.Title)
				}
			},
		},
		{
			name:           "missing authentication",
			apartmentID:    "1",
			requestBody:    jsonBytes(UpdateApartmentRequest{}),
			userInfo:       nil,
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "authentication required" {
					t.Errorf("expected error 'authentication required', got %s", response.Error)
				}
			},
		},
		{
			name:        "invalid apartment ID",
			apartmentID: "invalid",
			requestBody: jsonBytes(UpdateApartmentRequest{}),
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "invalid apartment ID" {
					t.Errorf("expected error 'invalid apartment ID', got %s", response.Error)
				}
			},
		},
		{
			name:        "apartment not found",
			apartmentID: "8",
			requestBody: jsonBytes(UpdateApartmentRequest{}),
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					return nil, apartment.ErrApartmentNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "apartment not found" {
					t.Errorf("expected error 'apartment not found', got %s", response.Error)
				}
			},
		},
		{
			name:        "insufficient permissions - not owner and not admin",
			apartmentID: "1",
			requestBody: jsonBytes(UpdateApartmentRequest{
				Title: stringPtr("Updated Title"),
			}),
			userInfo: &context.UserInfo{
				UserID:     2,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					apt := createTestApartment()
					apt.OwnerID = 1
					return apt, nil
				},
			},
			expectedStatus: http.StatusForbidden,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "you can only update your own apartments" {
					t.Errorf("expected error 'you can only update your own apartments', got %s", response.Error)
				}
			},
		},
		{
			name:        "invalid request body",
			apartmentID: "1",
			requestBody: []byte("invalid json"),
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					return createTestApartment(), nil
				},
			},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "bad request" {
					t.Errorf("expected error 'bad request', got %s", response.Error)
				}
			},
		},
		{
			name:        "service returns invalid input error",
			apartmentID: "1",
			requestBody: jsonBytes(UpdateApartmentRequest{
				Price: intPtr(-100),
			}),
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					return createTestApartment(), nil
				},
				UpdateApartmentFunc: func(id int, req domain.UpdateApartmentRequest) (*domain.Apartment, error) {
					return nil, apartment.ErrInvalidInput
				},
			},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "invalid input" {
					t.Errorf("expected error 'invalid input', got %s", response.Error)
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			handler := NewHandler(e.mockService)
			req := httptest.NewRequest(http.MethodPut, "/api/apartments/"+e.apartmentID, bytes.NewBuffer(e.requestBody))
			req.Header.Set("Content-Type", "application/json")

			if e.userInfo != nil {
				ctx := context.SetUserInfo(req.Context(), e.userInfo)
				req = req.WithContext(ctx)
			}

			r := chi.NewRouter()
			r.Put("/api/apartments/{id}", handler.UpdateApartment)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestHandler_DeleteApartment(t *testing.T) {
	tests := []struct {
		name           string
		apartmentID    string
		userInfo       *context.UserInfo
		mockService    *MockApartmentService
		expectedStatus int
		validate       func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:        "successful delete by owner",
			apartmentID: "1",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					apt := createTestApartment()
					apt.OwnerID = 1
					return apt, nil
				},
				DeleteApartmentFunc: func(id int) error {
					return nil
				},
			},
			expectedStatus: http.StatusNoContent,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if rr.Body.Len() != 0 {
					t.Errorf("expected empty body, got %s", rr.Body.String())
				}
			},
		},
		{
			name:        "successful delete by admin",
			apartmentID: "1",
			userInfo: &context.UserInfo{
				UserID:     2,
				IsLandlord: false,
				IsAdmin:    true,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					apt := createTestApartment()
					apt.OwnerID = 1
					return apt, nil
				},
				DeleteApartmentFunc: func(id int) error {
					return nil
				},
			},
			expectedStatus: http.StatusNoContent,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if rr.Body.Len() != 0 {
					t.Errorf("expected empty body, got %s", rr.Body.String())
				}
			},
		},
		{
			name:           "missing authentication",
			apartmentID:    "1",
			userInfo:       nil,
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusUnauthorized,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "authentication required" {
					t.Errorf("expected error 'authentication required', got %s", response.Error)
				}
			},
		},
		{
			name:        "invalid apartment ID",
			apartmentID: "invalid",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService:    &MockApartmentService{},
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "invalid apartment ID" {
					t.Errorf("expected error 'invalid apartment ID', got %s", response.Error)
				}
			},
		},
		{
			name:        "apartment not found",
			apartmentID: "8",
			userInfo: &context.UserInfo{
				UserID:     1,
				IsLandlord: true,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					return nil, apartment.ErrApartmentNotFound
				},
			},
			expectedStatus: http.StatusNotFound,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "apartment not found" {
					t.Errorf("expected error 'apartment not found', got %s", response.Error)
				}
			},
		},
		{
			name:        "insufficient permissions - not owner and not admin",
			apartmentID: "1",
			userInfo: &context.UserInfo{
				UserID:     2,
				IsLandlord: false,
				IsAdmin:    false,
			},
			mockService: &MockApartmentService{
				GetApartmentByIDFunc: func(id int) (*domain.Apartment, error) {
					apt := createTestApartment()
					apt.OwnerID = 1
					return apt, nil
				},
			},
			expectedStatus: http.StatusForbidden,
			validate: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if response.Error != "you can only delete your own apartments" {
					t.Errorf("expected error 'you can only delete your own apartments', got %s", response.Error)
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			handler := NewHandler(e.mockService)
			req := httptest.NewRequest(http.MethodDelete, "/api/apartments/"+e.apartmentID, nil)

			if e.userInfo != nil {
				ctx := context.SetUserInfo(req.Context(), e.userInfo)
				req = req.WithContext(ctx)
			}

			r := chi.NewRouter()
			r.Delete("/api/apartments/{id}", handler.DeleteApartment)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != e.expectedStatus {
				t.Errorf("expected status %d, got %d", e.expectedStatus, rr.Code)
			}

			if e.validate != nil {
				e.validate(t, rr)
			}
		})
	}
}

func TestParseFilters(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedFilter func(t *testing.T, filters *domain.ApartmentFilters)
	}{
		{
			name:        "no filters",
			queryParams: "",
			expectedFilter: func(t *testing.T, filters *domain.ApartmentFilters) {
				if filters != nil {
					t.Error("expected nil filters when no query params provided")
				}
			},
		},
		{
			name:        "country filter",
			queryParams: "?country=USA",
			expectedFilter: func(t *testing.T, filters *domain.ApartmentFilters) {
				if filters == nil {
					t.Fatal("expected non-nil filters")
				}
				if filters.Country == nil || *filters.Country != "USA" {
					t.Errorf("expected country 'USA', got %v", filters.Country)
				}
			},
		},
		{
			name:        "city filter",
			queryParams: "?city=NewYork",
			expectedFilter: func(t *testing.T, filters *domain.ApartmentFilters) {
				if filters == nil {
					t.Fatal("expected non-nil filters")
				}
				if filters.City == nil || *filters.City != "NewYork" {
					t.Errorf("expected city 'NewYork', got %v", filters.City)
				}
			},
		},
		{
			name:        "price filters",
			queryParams: "?min_price=100&max_price=5000",
			expectedFilter: func(t *testing.T, filters *domain.ApartmentFilters) {
				if filters == nil {
					t.Fatal("expected non-nil filters")
				}
				if filters.MinPrice == nil || *filters.MinPrice != 100 {
					t.Errorf("expected min_price 100, got %v", filters.MinPrice)
				}
				if filters.MaxPrice == nil || *filters.MaxPrice != 5000 {
					t.Errorf("expected max_price 5000, got %v", filters.MaxPrice)
				}
			},
		},
		{
			name:        "all filters",
			queryParams: "?country=USA&city=NY&min_area_m2=30&max_area_m2=100&rooms=2&floor=5&pets_allowed=true&min_price=500&max_price=2000",
			expectedFilter: func(t *testing.T, filters *domain.ApartmentFilters) {
				if filters == nil {
					t.Fatal("expected non-nil filters")
				}
				if filters.Country == nil || *filters.Country != "USA" {
					t.Error("country filter not set correctly")
				}
				if filters.City == nil || *filters.City != "NY" {
					t.Error("city filter not set correctly")
				}
				if filters.MinAreaM2 == nil || *filters.MinAreaM2 != 30 {
					t.Error("min_area_m2 filter not set correctly")
				}
				if filters.MaxAreaM2 == nil || *filters.MaxAreaM2 != 100 {
					t.Error("max_area_m2 filter not set correctly")
				}
				if filters.Rooms == nil || *filters.Rooms != 2 {
					t.Error("rooms filter not set correctly")
				}
				if filters.Floor == nil || *filters.Floor != 5 {
					t.Error("floor filter not set correctly")
				}
				if filters.PetsAllowed == nil || *filters.PetsAllowed != true {
					t.Error("pets_allowed filter not set correctly")
				}
				if filters.MinPrice == nil || *filters.MinPrice != 500 {
					t.Error("min_price filter not set correctly")
				}
				if filters.MaxPrice == nil || *filters.MaxPrice != 2000 {
					t.Error("max_price filter not set correctly")
				}
			},
		},
		{
			name:        "invalid numeric filters are ignored",
			queryParams: "?min_price=invalid&max_price=abc&rooms=not_a_number",
			expectedFilter: func(t *testing.T, filters *domain.ApartmentFilters) {
				if filters != nil {
					t.Error("expected nil filters when all filters are invalid")
				}
			},
		},
	}

	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/apartments"+e.queryParams, nil)
			filters := parseFilters(req)
			e.expectedFilter(t, filters)
		})
	}
}

// вспомогательные функции
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
