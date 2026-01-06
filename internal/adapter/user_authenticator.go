package adapter

import (
	"errors"
	domain "rent-app/internal/domain/auth"
	userDomain "rent-app/internal/domain/user"
)

// этот адаптер нужен для изоляции модулей авторизации и модулей юзера друг от друга

type UserAuthenticatorAdapter struct {
	userService userDomain.Service
}

func NewUserAuthenticatorAdapter(userService userDomain.Service) *UserAuthenticatorAdapter {
	return &UserAuthenticatorAdapter{
		userService: userService,
	}
}

func (a *UserAuthenticatorAdapter) Authenticate(email, password string) (*domain.AuthUserInfo, error) {
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

	return &domain.AuthUserInfo{
		ID:         user.ID,
		Email:      user.Email,
		IsLandlord: user.IsLandlord,
		IsAdmin:    user.IsAdmin,
	}, nil
}
