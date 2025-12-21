package user

import (
	"context"
	"rent-app/internal/database"
	domain "rent-app/internal/domain/user"
	"testing"

	"golang.org/x/crypto/bcrypt"
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
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			password_hash TEXT NOT NULL,
			is_landlord BOOLEAN DEFAULT FALSE,
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`
	_, err = pool.Exec(ctx, stmt)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	_, err = pool.Exec(ctx, "TRUNCATE users RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate users: %v", err)
	}

	return &PostgresRepo{DB: pool}
}

func createTestUser() *domain.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	return &domain.User{
		Email:        "test@example.com",
		FirstName:    "George",
		LastName:     "Washington",
		PasswordHash: string(hash),
		IsLandlord:   false,
		IsAdmin:      false,
	}
}

func TestInsertUser(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	user := createTestUser()

	err := repo.InsertUser(user)
	if err != nil {
		t.Errorf("insert user returned %s", err)
	}

	if user.ID == 0 {
		t.Errorf("expected user ID to be set but hot default value")
	}

	if user.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}

	if user.UpdatedAt.IsZero() {
		t.Errorf("expected UpdatedAt to be set")
	}
}

func TestGetUserByEmail(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	user := createTestUser()

	_ = repo.InsertUser(user)

	_, err := repo.GetUserByEmail(user.Email)
	if err != nil {
		t.Errorf("expected getting user but got err %s", err)
	}

	_, err = repo.GetUserByEmail("invalid@example.com")
	if err == nil {
		t.Error("expected getting err but don't")
	}
}

func TestGetUserByID(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	user := createTestUser()

	_ = repo.InsertUser(user)

	_, err := repo.GetUserByID(user.ID)
	if err != nil {
		t.Errorf("expected getting user but got err %s", err)
	}

	_, err = repo.GetUserByID(4)
	if err == nil {
		t.Error("expected getting err but don't")
	}
}

func TestGelAllUsers(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	users, err := repo.GetAllUsers()
	if err != nil {
		t.Error("unexpected error:", err)
	}

	if len(users) != 0 {
		t.Errorf("error getting quantity of users: expected 0 but got %d", len(users))
	}

	user := createTestUser()
	_ = repo.InsertUser(user)

	users, err = repo.GetAllUsers()
	if err != nil {
		t.Error("unexpected error:", err)
	}

	if len(users) != 1 {
		t.Errorf("error getting quantity of users: expected 1 but got %d", len(users))
	}
}

func TestUpdateUser(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	user := createTestUser()

	err := repo.UpdateUser(user)
	if err == nil {
		t.Errorf("expected error but don't get it")
	}

	_ = repo.InsertUser(user)

	user.Email = "another@email.fr"
	user.FirstName = "Maximilien"
	user.LastName = "Robespierre"
	user.IsAdmin = true
	user.IsLandlord = true

	err = repo.UpdateUser(user)
	if err != nil {
		t.Error("got err but should not:", err)
	}

	if user.Email != "another@email.fr" {
		t.Error("expected get email another@email.fr but got", user.Email)
	}

	if user.FirstName != "Maximilien" {
		t.Error("expected get first name Maximilien but got", user.FirstName)
	}

	if user.LastName != "Robespierre" {
		t.Error("expected get last name Robespierre but got", user.LastName)
	}

	if !user.IsAdmin {
		t.Error("expected user to be Admin but it doesn't")
	}

	if !user.IsLandlord {
		t.Error("expected user to be Landlord but it doesn't")
	}
}

func TestDeleteUser(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	user := createTestUser()
	_ = repo.InsertUser(user)

	err := repo.DeleteUser(user.ID)
	if err != nil {
		t.Error("got error but shouldn't", err)
	}

	_, err = repo.GetUserByID(user.ID)
	if err == nil {
		t.Error("expected err but don't get it")
	}
}

func TestResetPassword(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.DB.Close()

	user := createTestUser()
	_ = repo.InsertUser(user)

	newpassword := "qwerty"

	err := repo.ResetPassword(user.ID, newpassword)
	if err != nil {
		t.Error("got error but shouldn't", err)
	}

	updatedUser, _ := repo.GetUserByID(user.ID)

	err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte(newpassword))
	if err != nil {
		t.Error("password hash does not match new password", err)
	}

	err = repo.ResetPassword(4, newpassword)
	if err == nil {
		t.Error("expected error but don't het it")
	}
}
