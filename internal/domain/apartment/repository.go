package apartment

type Repository interface {
	InsertApartment(a *Apartment) error
	GetApartmentByID(id int) (*Apartment, error)
	GetApartmentsByOwnerID(ownerID int) ([]*Apartment, error)
	GetAllApartments(filters *ApartmentFilters) ([]*Apartment, error)
	UpdateApartment(a *Apartment) error
	DeleteApartment(id int) error
}
