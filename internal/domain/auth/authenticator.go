package auth

type AuthUserInfo struct {
	ID         int
	Email      string
	IsLandlord bool
	IsAdmin    bool
}
