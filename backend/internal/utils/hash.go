package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// bcryptCost is the work factor. 12 is roughly 250ms on modern hardware: slow
// enough to make offline cracking expensive, fast enough that a login does not
// feel broken. The stdlib default of 10 is too cheap for a password store in
// 2026.
const bcryptCost = 12

// HashPassword returns the bcrypt hash of a plaintext password.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("utils: hashing password: %w", err)
	}
	return string(b), nil
}

// CheckPassword reports whether plain matches hash.
//
// It returns a bool rather than an error because callers must not branch on
// *why* verification failed. Distinguishing "wrong password" from "malformed
// hash" in a response is how user-enumeration oracles get built.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
