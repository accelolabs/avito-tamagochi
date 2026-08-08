package auth

import (
	"github.com/alexedwards/argon2id"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
)

var passwordParams = &argon2id.Params{
	Memory:      64 * 1024,
	Iterations:  1,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

func validPassword(password string) bool {
	return len(password) >= minPasswordLength && len(password) <= maxPasswordLength
}

func hashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, passwordParams)
}

func verifyPassword(password, encoded string) bool {
	match, err := argon2id.ComparePasswordAndHash(password, encoded)
	return err == nil && match
}
