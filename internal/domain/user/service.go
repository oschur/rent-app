package user

type Service interface {
	CreateUser(email, password, firstname, lastname string, isLandlord, isAdmin bool) (*User, error)
	GetUserByID(id int) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByEmailForAuth(email string) (*User, error) // Возвращает пользователя с паролем для аутентификации
	GetAllUsers() ([]*User, error)
	UpdateUser(id int, email, firstname, lastname *string, isLandlord, isAdmin *bool) (*User, error)
	DeleteUser(id int) error
	ResetPassword(id int, password string) error
}
