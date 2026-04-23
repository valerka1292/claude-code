package agent

import (
	"errors"
	"net"
	"strings"
	"time"
)

// IsRetryableError determines if an error is retryable.
func IsRetryableError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "504")
}

// RetryBackoff returns the backoff duration for a given retry attempt.
func RetryBackoff(attempt int) time.Duration {
	steps := []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
	}
	if attempt < 0 {
		return steps[0]
	}
	if attempt >= len(steps) {
		return steps[len(steps)-1]
	}
	return steps[attempt]
}
