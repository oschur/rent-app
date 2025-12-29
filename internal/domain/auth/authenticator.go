package auth

type UserAuthenticator interface {
	Authenticate(email, password string) (*AuthUserInfo, error)
}

type AuthUserInfo struct {
	ID         int
	Email      string
	IsLandlord bool
	IsAdmin    bool
}
