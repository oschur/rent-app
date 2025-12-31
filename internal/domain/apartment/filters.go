package apartment

type ApartmentFilters struct {
	Country     *string
	City        *string
	MinAreaM2   *int
	MaxAreaM2   *int
	Rooms       *int
	Floor       *int
	PetsAllowed *bool
}

