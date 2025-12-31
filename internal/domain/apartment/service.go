package apartment

type Service interface {
	CreateApartment(ownerID int, status ApartmentStatus, priceUnit PriceUnit, title string, price int, country string, city string, address string, areaM2 int, rooms int) (*Apartment, error)
	GetApartmentByID(id int) (*Apartment, error)
	GetApartmentByOwnerID(ownerID int) ([]*Apartment, error)
	GetAllApartments() ([]*Apartment, error)
	UpdateApartment(id int, status *ApartmentStatus, priceUnit *PriceUnit, title *string, price *int, country *string, city *string, address *string, areaM2 *int, rooms *int, floor *int, totalFloors *int, petsAllowed *bool) (*Apartment, error)
	DeleteApartment(id int) error
}

