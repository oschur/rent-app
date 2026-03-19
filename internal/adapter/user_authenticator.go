package adapter

import (
	"errors"
	authDomain "rent-app/internal/domain/auth"
	userDomain "rent-app/internal/domain/user"
)

// этот адаптер нужен для изоляции модулей авторизации и модулей юзера друг от друга

type UserAuthService interface {
	GetUserByEmailForAuth(email string) (*userDomain.User, error)
}

type UserAuthenticatorAdapter struct {
	userService UserAuthService
}

func NewUserAuthenticatorAdapter(userService UserAuthService) *UserAuthenticatorAdapter {
	return &UserAuthenticatorAdapter{
		userService: userService,
	}
}

func (a *UserAuthenticatorAdapter) Authenticate(email, password string) (*authDomain.AuthUserInfo, error) {
	user, err := a.userService.GetUserByEmailForAuth(email)
	if err != nil {
		return nil, err
	}

	match, err := user.PasswordMatches(password)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, errors.New("invalid credentials")
	}

	return &authDomain.AuthUserInfo{
		ID:         user.ID,
		Email:      user.Email,
		IsLandlord: user.IsLandlord,
		IsAdmin:    user.IsAdmin,
	}, nil
}
