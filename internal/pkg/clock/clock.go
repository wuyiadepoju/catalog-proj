package clock

import "time"

// Clock provides an abstraction for time operations
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using the system clock
type RealClock struct{}

// NewRealClock creates a new real clock instance
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current system time
func (r *RealClock) Now() time.Time {
	return time.Now()
}
