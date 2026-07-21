package repositories

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// isUniqueViolation reports whether an error is a unique-constraint violation.
//
// The precise way to do this is to type-assert *pgconn.PgError and compare
// SQLSTATE 23505. That would drag pgx's error types through the repository
// layer for one comparison, so this matches on the driver-agnostic text
// instead. The cost of a false negative is a 500 where a 409 belonged — the
// caller's fallback path is still correct, just less polite.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "unique constraint")
}

// isForeignKeyViolation reports whether an error is an FK violation (23503),
// which in this schema means a client referenced a product that does not exist.
func isForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "foreign key") || strings.Contains(msg, "23503")
}
