package espeon

import (
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Backoff is an interface that defines any method to be used when implementing backoff.
type Backoff interface {
	// NextInterval defines the next interval for backoff.
	NextInterval(order int) time.Duration
}

const (
	// DefaultBackoffInterval is the default value for any backoff interval.
	DefaultBackoffInterval = 500 * time.Millisecond
	// DefaultJitterInterval is the default value for any jitter interval.
	DefaultJitterInterval = 200 * time.Millisecond
	// DefaultMultiplier is the default value factor for exponential backoff.
	DefaultMultiplier = 2
	// DefaultMaxInterval is the maximum interval for backoff.
	DefaultMaxInterval = 3 * time.Second
)

// ConstantBackoff implements Backoff using constant interval.
type ConstantBackoff struct {
	// BackoffInterval defines how long the next backoff will be compared to the previous one.
	BackoffInterval time.Duration
	// JitterInterval defines the randomness additional value for interval.
	JitterInterval time.Duration
	// MaxInterval defines the maximum interval allowed.
	MaxInterval time.Duration
}

// NewConstantBackoff creates an instance of ConstantBackoff with default values.
// See Constants for the default values.
func NewConstantBackoff() *ConstantBackoff {
	return &ConstantBackoff{
		BackoffInterval: DefaultBackoffInterval,
		JitterInterval:  DefaultJitterInterval,
		MaxInterval:     DefaultMaxInterval,
	}
}

// NextInterval returns next interval for backoff.
func (c *ConstantBackoff) NextInterval(order int) time.Duration {
	if order <= 0 {
		return 0 * time.Millisecond
	}

	// just in case backoff interval exceeds max interval
	backoffInterval := math.Min(float64(c.BackoffInterval), float64(c.MaxInterval))
	jitterInterval := rand.Int63n(int64(c.JitterInterval))

	return time.Duration(backoffInterval + float64(jitterInterval))
}

// ExponentialBackoff implements Backoff using exponential interval.
type ExponentialBackoff struct {
	// BackoffInterval defines how long the next backoff will be compared to the previous one.
	BackoffInterval time.Duration
	// JitterInterval defines the randomness additional value for interval.
	JitterInterval time.Duration
	// MaxInterval defines the maximum interval allowed.
	MaxInterval time.Duration
	// Multiplier defines exponential factor.
	Multiplier int64
}

// NewExponentialBackoff creates an instance of ExponentialBackoff with default values.
// See Constants for the default values.
func NewExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		BackoffInterval: DefaultBackoffInterval,
		JitterInterval:  DefaultJitterInterval,
		MaxInterval:     DefaultMaxInterval,
		Multiplier:      DefaultMultiplier,
	}
}

// NextInterval returns the next order-th interval for backoff.
func (e *ExponentialBackoff) NextInterval(order int) time.Duration {
	if order <= 0 {
		return 0 * time.Millisecond
	}

	exponent := math.Pow(float64(e.Multiplier), float64(order-1))
	backoffInterval := float64(e.BackoffInterval) * exponent
	// prevent backoff to exceed max interval
	backoffInterval = math.Min(backoffInterval, float64(e.MaxInterval))
	jitterInterval := rand.Int63n(int64(e.JitterInterval))

	return time.Duration(backoffInterval + float64(jitterInterval))
}

// NoBackoff implements backoff without any interval.
type NoBackoff struct {
}

// NewNoBackoff creates an instance of NoBackoff
func NewNoBackoff() *NoBackoff {
	return &NoBackoff{}
}

// NextInterval returns the next order-th interval for backoff.
func (n *NoBackoff) NextInterval(_ int) time.Duration {
	return 0
}
