package duration

import (
	"fmt"
	"strconv"
	"time"
)

// Parse parses a duration string the same way kool timeout / for-every do:
// prefer time.ParseDuration; if that fails, treat a bare integer as seconds.
// The result must be greater than 0.
func Parse(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	dur, durErr := time.ParseDuration(s)
	if durErr == nil {
		if dur <= 0 {
			return 0, fmt.Errorf("duration must be greater than 0")
		}
		return dur, nil
	}
	d, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		if d <= 0 {
			return 0, fmt.Errorf("duration must be greater than 0")
		}
		return time.Duration(d) * time.Second, nil
	}
	return 0, fmt.Errorf("invalid duration: %v", durErr)
}
