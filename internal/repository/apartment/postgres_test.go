package apartment

import (
	"context"
	"rent-app/internal/database"
	"testing"

	domain "rent-app/internal/domain/apartment"
)

const testDSN = "postgres://admin:admin@localhost:5435/users?sslmode=disable"

func setupTestDB(t *testing.T) *PostgresRepo {
	t.Helper()
	ctx := context.Background()
	pool, err := database.Connect(ctx, testDSN)
	if err != nil {
		t.Fatalf("failed to connect test db: %v", err)
	}

	stmt := `
		CREATE TABLE IF NOT EXISTS apartments (
			id SERIAL PRIMARY KEY,
			owner_id INT NOT NULL,
			status VARCHAR(64) NOT NULL,
			price_unit VARCHAR(64) NOT NULL,
			title TEXT NOT NULL,
			price INT NOT NULL,
			country VARCHAR(64) NOT NULL,
			city VARCHAR(64) NOT NULL,
			address VARCHAR(128) NOT NULL,
			area_m2 INT NOT NULL,
			rooms INT NOT NULL,
			floor INT,
			total_floors INT,
			pets_allowed BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`
	_, err = pool.Exec(ctx, stmt)
	if err != nil {
		t.Fatalf("failed to create apartments table: %v", err)
	}

	_, err = pool.Exec(ctx, "TRUNCATE apartments RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate apartments: %v", err)
	}

	return &PostgresRepo{DB: pool}
}

func createTestApartment() *domain.Apartment {
	return &domain.Apartment{
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
		PetsAllowed: false,
	}
}

func TestInsertApartment(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	apartment := createTestApartment()

	err := repo.InsertApartment(apartment)
	if err != nil {
		t.Errorf("insert apartment returned %s", err)
	}

	if apartment.ID == 0 {
		t.Errorf("expected apartment ID to be set but got default value")
	}

	if apartment.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}

	if apartment.UpdatedAt.IsZero() {
		t.Errorf("expected UpdatedAt to be set")
	}
}

func TestGetApartmentByID(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	apartment := createTestApartment()

	_ = repo.InsertApartment(apartment)

	found, err := repo.GetApartmentByID(apartment.ID)
	if err != nil {
		t.Errorf("expected getting apartment but got err %s", err)
	}

	if found.ID != apartment.ID {
		t.Errorf("expected apartment ID to be %d but got %d", apartment.ID, found.ID)
	}

	found, err = repo.GetApartmentByID(4)
	if err == nil {
		t.Error("expected getting err but don't")
	}
}

func TestGetApartmentsByOwnerID(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	apartment1 := createTestApartment()
	apartment2 := createTestApartment()

	_ = repo.InsertApartment(apartment1)
	_ = repo.InsertApartment(apartment2)

	found, err := repo.GetApartmentsByOwnerID(apartment1.OwnerID)
	if err != nil {
		t.Errorf("expected getting apartments but got err %s", err)
	}

	if len(found) != 2 {
		t.Errorf("expected 2 apartments but got %d", len(found))
	}

	if found[0].ID != apartment2.ID {
		t.Errorf("expected apartment ID to be %d but got %d", apartment1.ID, found[1].ID)
	}

	if found[1].ID != apartment1.ID {
		t.Errorf("expected apartment ID to be %d but got %d", apartment2.ID, found[0].ID)
	}
}

func TestGetAllApartments(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	apartment1 := createTestApartment()
	apartment2 := createTestApartment()

	_ = repo.InsertApartment(apartment1)
	_ = repo.InsertApartment(apartment2)

	found, err := repo.GetAllApartments()
	if err != nil {
		t.Errorf("expected getting apartments but got err %s", err)
	}

	if len(found) != 2 {
		t.Errorf("expected 2 apartments but got %d", len(found))
	}

	if found[0].ID != apartment2.ID {
		t.Errorf("expected apartment ID to be %d but got %d", apartment2.ID, found[0].ID)
	}

	if found[1].ID != apartment1.ID {
		t.Errorf("expected apartment ID to be %d but got %d", apartment1.ID, found[1].ID)
	}
}

func TestUpdateApartment(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	apartment := createTestApartment()
	_ = repo.InsertApartment(apartment)

	apartment.Title = "Updated Apartment"
	err := repo.UpdateApartment(apartment)
	if err != nil {
		t.Errorf("expected updating apartment but got err %s", err)
	}

	found, _ := repo.GetApartmentByID(apartment.ID)
	if found.Title != "Updated Apartment" {
		t.Errorf("expected apartment title to be %s but got %s", "Updated Apartment", found.Title)
	}
}

func TestDeleteApartment(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	apartment := createTestApartment()
	_ = repo.InsertApartment(apartment)

	err := repo.DeleteApartment(apartment.ID)
	if err != nil {
		t.Errorf("expected deleting apartment but got err %s", err)
	}

	_, err = repo.GetApartmentByID(apartment.ID)
	if err == nil {
		t.Error("expected getting err but don't")
	}
}
