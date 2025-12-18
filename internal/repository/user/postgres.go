package user

import (
	"context"
	"fmt"
	"time"

	domain "rent-app/internal/domain/user"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const timeout = 3 * time.Second

type PostgresRepo struct {
	DB *pgxpool.Pool
}

func (p *PostgresRepo) InsertUser(u *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	stmt := `
		INSERT INTO users (email, first_name, last_name, passwordhash, is_landlord, is_admin, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int
	now := time.Now()
	err := p.DB.QueryRow(ctx, stmt,
		u.Email,
		u.FirstName,
		u.LastName,
		u.PasswordHash,
		u.IsLandlord,
		u.IsAdmin,
		now,
		now,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	u.ID = id
	u.CreatedAt = now
	u.UpdatedAt = now

	return nil
}

func (p *PostgresRepo) GetUserByEmail(email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.password_hash, u.is_landlord,u.is_admin, u.created_at, u.updated_at
		FROM 
			users u
		WHERE 
		    u.email = $1`

	var user domain.User

	row := p.DB.QueryRow(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&user.IsLandlord,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *PostgresRepo) GetUserByID(id int) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.password_hash, u.is_landlord,u.is_admin, u.created_at, u.updated_at
		FROM 
			users u
		WHERE 
		    u.id = $1`

	var user domain.User

	row := p.DB.QueryRow(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&user.IsLandlord,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *PostgresRepo) UpdateUser(u *domain.User) error {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stmt := `
		UPDATE users 
		SET
			email = $1,
			first_name = $2,
			last_name = $3,
			is_landlord = $4,
			is_admin = $5,
			updated_at = $6
		WHERE id = $7
		`
	result, err := p.DB.Exec(ctx, stmt,
		u.Email,
		u.FirstName,
		u.LastName,
		u.IsLandlord,
		u.IsAdmin,
		time.Now(),
		u.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", u.ID)
	}
	return nil
}

func (p *PostgresRepo) DeleteUser(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stmt := `DELETE FROM users WHERE id = $1`

	_, err := p.DB.Exec(ctx, stmt, id)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresRepo) ResetPassword(id int, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `UPDATE users SET passwordhash = $1 WHERE id = $2`

	result, err := p.DB.Exec(ctx, stmt, passwordHash, id)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}
	return nil
}
