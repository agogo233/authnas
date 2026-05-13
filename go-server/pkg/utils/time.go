package utils

import (
	"time"
)

type TimeUtil struct{}

func NewTime() *TimeUtil {
	return &TimeUtil{}
}

func (u *TimeUtil) Now() time.Time {
	return time.Now()
}

func (u *TimeUtil) NowUTC() time.Time {
	return time.Now().UTC()
}

func (u *TimeUtil) AddDuration(t time.Time, d time.Duration) time.Time {
	return t.Add(d)
}

func (u *TimeUtil) ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

func (u *TimeUtil) Format(t time.Time) string {
	return t.Format(time.RFC3339)
}
