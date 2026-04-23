package agent

import (
	"errors"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// IsRetryableError determines if an error is retryable.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}

	msg := strings.ToLower(err.Error())
	status := extractHTTPStatus(msg)
	if status == 429 || status == 529 {
		return true
	}
	return strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "502") ||
		strings.Contains(msg, "503") ||
		strings.Contains(msg, "504")
}

// RetryBackoff returns the backoff duration for a given retry attempt.
func RetryBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	base := 500 * time.Millisecond
	maxDelay := 32 * time.Second

	delay := base << attempt
	if delay > maxDelay {
		delay = maxDelay
	}

	// Deterministic jitter (0-25%) from attempt number keeps retries de-synced
	// across calls without introducing randomness in tests.
	jitterPct := (attempt*17 + 11) % 26
	jitter := time.Duration(int64(delay) * int64(jitterPct) / 100)
	return delay + jitter
}

func extractHTTPStatus(lowerMsg string) int {
	quotedStatus := regexp.MustCompile(`"status"\s*:\s*(\d{3})`)
	if m := quotedStatus.FindStringSubmatch(lowerMsg); len(m) == 2 {
		if status, err := strconv.Atoi(m[1]); err == nil {
			return status
		}
	}
	plainStatus := regexp.MustCompile(`\bstatus(?:\s*code)?\s*[:=]?\s*(\d{3})\b`)
	if m := plainStatus.FindStringSubmatch(lowerMsg); len(m) == 2 {
		if status, err := strconv.Atoi(m[1]); err == nil {
			return status
		}
	}
	return 0
}
