package user

type Repository interface {
	InsertUser(u *User) error
	GetUserByID(id int) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetAllUsers() ([]*User, error)
	UpdateUser(u *User) error
	DeleteUser(id int) error
	ResetPassword(id int, password string) error
}
