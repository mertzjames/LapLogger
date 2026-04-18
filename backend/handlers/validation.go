package handlers

import (
	"time"

	"github.com/google/uuid"
)

// isValidUUID checks if a string is a valid UUID format.
func isValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// isValidDate checks if a string is a valid YYYY-MM-DD date.
func isValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// Field length limits.
const (
	maxNameLength     = 255
	maxLocationLength = 500
)
