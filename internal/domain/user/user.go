package user

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	IsLandlord   bool
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) PasswordMatches(text string) (bool, error) {

	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(text))

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}
