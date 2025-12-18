package user

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func Test_user_passwordMathches(t *testing.T) {

	passHash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 10)

	testUser := User{
		ID:           1,
		PasswordHash: string(passHash),
	}

	var tests = []struct {
		name          string
		passwordTry   string
		expectedMatch bool
		expectedError bool
	}{
		{"valid", "correctpassword", true, false},
		{"incorrect password", "wrongpassword", false, false},
	}

	for _, e := range tests {
		match, err := testUser.PasswordMatches(e.passwordTry)
		if err != nil && e.expectedError {
			t.Errorf("%s: expected error but don't get it", e.name)
		}

		if match && !e.expectedMatch {
			t.Errorf("%s: don't expected password match but get it", e.name)
		}

		if !match && e.expectedMatch {
			t.Errorf("%s: expected password match but don't get it", e.name)
		}
	}
}
