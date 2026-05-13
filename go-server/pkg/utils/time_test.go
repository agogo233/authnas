package utils

import (
	"testing"
	"time"
)

func TestTimeUtil_NewTime(t *testing.T) {
	util := NewTime()
	if util == nil {
		t.Fatal("NewTime() returned nil")
	}
}

func TestTimeUtil_Now(t *testing.T) {
	util := NewTime()

	before := time.Now()
	result := util.Now()
	after := time.Now()

	if result.Before(before) || result.After(after) {
		t.Errorf("Now() returned time outside expected range")
	}
}

func TestTimeUtil_NowUTC(t *testing.T) {
	util := NewTime()

	before := time.Now().UTC()
	result := util.NowUTC()
	after := time.Now().UTC()

	if result.Before(before) || result.After(after) {
		t.Errorf("NowUTC() returned time outside expected range")
	}

	loc := result.Location()
	if loc != time.UTC {
		t.Errorf("NowUTC() should return UTC time, got %v", loc)
	}
}

func TestTimeUtil_AddDuration(t *testing.T) {
	util := NewTime()

	testCases := []struct {
		name     string
		duration time.Duration
	}{
		{"1 hour", time.Hour},
		{"30 minutes", 30 * time.Minute},
		{"1 day", 24 * time.Hour},
		{"negative 1 hour", -time.Hour},
	}

	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := util.AddDuration(baseTime, tc.duration)
			expected := baseTime.Add(tc.duration)

			if !result.Equal(expected) {
				t.Errorf("AddDuration() = %v, want %v", result, expected)
			}
		})
	}
}

func TestTimeUtil_ParseDuration(t *testing.T) {
	util := NewTime()

	testCases := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"1h", time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"1h30m", time.Hour + 30*time.Minute, false},
		{"1h30m15s", time.Hour + 30*time.Minute + 15*time.Second, false},
		{"invalid", 0, true},
		{"", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := util.ParseDuration(tc.input)

			if tc.hasError {
				if err == nil {
					t.Errorf("ParseDuration(%q) should return error", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseDuration(%q) returned unexpected error: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("ParseDuration(%q) = %v, want %v", tc.input, result, tc.expected)
				}
			}
		})
	}
}

func TestTimeUtil_Format(t *testing.T) {
	util := NewTime()

	testTime := time.Date(2024, 6, 15, 10, 30, 45, 0, time.UTC)
	result := util.Format(testTime)

	expected := "2024-06-15T10:30:45Z"
	if result != expected {
		t.Errorf("Format() = %q, want %q", result, expected)
	}
}

func TestTimeUtil_Format_DifferentTimes(t *testing.T) {
	util := NewTime()

	testCases := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			"midnight",
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			"2024-01-01T00:00:00Z",
		},
		{
			"end of day",
			time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			"2024-12-31T23:59:59Z",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := util.Format(tc.input)
			if result != tc.expected {
				t.Errorf("Format() = %q, want %q", result, tc.expected)
			}
		})
	}
}
