package service

import (
	"time"

	"github.com/google/uuid"
)

func generateID() string {
	return uuid.New().String()
}

func now() time.Time {
	return time.Now()
}
