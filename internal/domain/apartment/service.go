package apartment

type CreateApartmentRequest struct {
	OwnerID     int
	Status      ApartmentStatus
	PriceUnit   PriceUnit
	Title       string
	Price       int
	Country     string
	City        string
	Address     string
	AreaM2      int
	Rooms       int
	Floor       *int
	TotalFloors *int
	PetsAllowed bool
}

type UpdateApartmentRequest struct {
	Status      *ApartmentStatus
	PriceUnit   *PriceUnit
	Title       *string
	Price       *int
	Country     *string
	City        *string
	Address     *string
	AreaM2      *int
	Rooms       *int
	Floor       *int
	TotalFloors *int
	PetsAllowed *bool
}
