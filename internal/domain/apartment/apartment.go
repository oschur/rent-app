package apartment

import "time"

type ApartmentStatus string

const (
	StatusActive   ApartmentStatus = "active"
	StatusArchived ApartmentStatus = "archived"
	StatusBlocked  ApartmentStatus = "blocked"
)

type PriceUnit string

const (
	PerNight PriceUnit = "pernight"
	PerMonth PriceUnit = "permonth"
)

type Apartment struct {
	ID          int
	OwnerID     int
	Status      ApartmentStatus // active, archived or blocked
	PriceUnit   PriceUnit       // per night or per month
	Title       string
	Price       int
	Country     string
	City        string
	Address     string
	AreaM2      int
	Rooms       int
	Floor       *int // указатель на случай если сдается дом
	TotalFloors *int
	PetsAllowed bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
