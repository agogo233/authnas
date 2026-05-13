package repository

import (
	"github.com/google/uuid"
	"time"
)

func generateID() string {
	return uuid.New().String()
}

func now() time.Time {
	return time.Now()
}
