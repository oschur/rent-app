package apartment

import (
	"context"
	"fmt"
	"time"

	domain "rent-app/internal/domain/apartment"

	"github.com/jackc/pgx/v5/pgxpool"
)

const timeout = 3 * time.Second

type PostgresRepo struct {
	DB *pgxpool.Pool
}

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		DB: db,
	}
}

func (p *PostgresRepo) InsertApartment(a *domain.Apartment) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stmt := `
		INSERT INTO apartments (
			owner_id, status, price_unit, title, price, country, city, address,
			area_m2, rooms, floor, total_floors, pets_allowed, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`

	var id int
	now := time.Now()
	err := p.DB.QueryRow(ctx, stmt,
		a.OwnerID,
		a.Status,
		a.PriceUnit,
		a.Title,
		a.Price,
		a.Country,
		a.City,
		a.Address,
		a.AreaM2,
		a.Rooms,
		a.Floor,
		a.TotalFloors,
		a.PetsAllowed,
		now,
		now,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to insert apartment: %w", err)
	}

	a.ID = id
	a.CreatedAt = now
	a.UpdatedAt = now

	return nil
}

func (p *PostgresRepo) GetApartmentByID(id int) (*domain.Apartment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `
		SELECT 
			id, owner_id, status, price_unit, title, price, country, city, address,
			area_m2, rooms, floor, total_floors, pets_allowed, created_at, updated_at
		FROM 
			apartments
		WHERE 
			id = $1
	`

	var apartment domain.Apartment
	var floor, totalFloors *int

	err := p.DB.QueryRow(ctx, query, id).Scan(
		&apartment.ID,
		&apartment.OwnerID,
		&apartment.Status,
		&apartment.PriceUnit,
		&apartment.Title,
		&apartment.Price,
		&apartment.Country,
		&apartment.City,
		&apartment.Address,
		&apartment.AreaM2,
		&apartment.Rooms,
		&floor,
		&totalFloors,
		&apartment.PetsAllowed,
		&apartment.CreatedAt,
		&apartment.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get apartment by id: %w", err)
	}

	apartment.Floor = floor
	apartment.TotalFloors = totalFloors

	return &apartment, nil
}

func (p *PostgresRepo) GetApartmentsByOwnerID(ownerID int) ([]*domain.Apartment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `
		SELECT 
			id, owner_id, status, price_unit, title, price, country, city, address,
			area_m2, rooms, floor, total_floors, pets_allowed, created_at, updated_at
		FROM 
			apartments
		WHERE 
			owner_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.DB.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments by owner id: %w", err)
	}
	defer rows.Close()

	var apartments []*domain.Apartment
	for rows.Next() {
		var apartment domain.Apartment
		var floor, totalFloors *int

		err := rows.Scan(
			&apartment.ID,
			&apartment.OwnerID,
			&apartment.Status,
			&apartment.PriceUnit,
			&apartment.Title,
			&apartment.Price,
			&apartment.Country,
			&apartment.City,
			&apartment.Address,
			&apartment.AreaM2,
			&apartment.Rooms,
			&floor,
			&totalFloors,
			&apartment.PetsAllowed,
			&apartment.CreatedAt,
			&apartment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan apartment: %w", err)
		}

		apartment.Floor = floor
		apartment.TotalFloors = totalFloors
		apartments = append(apartments, &apartment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate apartments: %w", err)
	}

	return apartments, nil
}

func (p *PostgresRepo) GetAllApartments() ([]*domain.Apartment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	query := `
		SELECT 
			id, owner_id, status, price_unit, title, price, country, city, address,
			area_m2, rooms, floor, total_floors, pets_allowed, created_at, updated_at
		FROM 
			apartments
		ORDER BY created_at DESC
	`

	rows, err := p.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all apartments: %w", err)
	}
	defer rows.Close()

	var apartments []*domain.Apartment
	for rows.Next() {
		var apartment domain.Apartment
		var floor, totalFloors *int

		err := rows.Scan(
			&apartment.ID,
			&apartment.OwnerID,
			&apartment.Status,
			&apartment.PriceUnit,
			&apartment.Title,
			&apartment.Price,
			&apartment.Country,
			&apartment.City,
			&apartment.Address,
			&apartment.AreaM2,
			&apartment.Rooms,
			&floor,
			&totalFloors,
			&apartment.PetsAllowed,
			&apartment.CreatedAt,
			&apartment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan apartment: %w", err)
		}

		apartment.Floor = floor
		apartment.TotalFloors = totalFloors
		apartments = append(apartments, &apartment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate apartments: %w", err)
	}

	return apartments, nil
}

func (p *PostgresRepo) UpdateApartment(a *domain.Apartment) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stmt := `
		UPDATE apartments 
		SET
			status = $1,
			price_unit = $2,
			title = $3,
			price = $4,
			country = $5,
			city = $6,
			address = $7,
			area_m2 = $8,
			rooms = $9,
			floor = $10,
			total_floors = $11,
			pets_allowed = $12,
			updated_at = $13
		WHERE id = $14
	`

	result, err := p.DB.Exec(ctx, stmt,
		a.Status,
		a.PriceUnit,
		a.Title,
		a.Price,
		a.Country,
		a.City,
		a.Address,
		a.AreaM2,
		a.Rooms,
		a.Floor,
		a.TotalFloors,
		a.PetsAllowed,
		time.Now(),
		a.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update apartment: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("apartment with id %d not found", a.ID)
	}

	return nil
}

func (p *PostgresRepo) DeleteApartment(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stmt := `DELETE FROM apartments WHERE id = $1`

	result, err := p.DB.Exec(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("failed to delete apartment: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("apartment with id %d not found", id)
	}

	return nil
}
