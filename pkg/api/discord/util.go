package discord

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrRateLimit = errors.New("rate limit")

func IsRateLimit(err error) (time.Time, bool) {
	if !errors.Is(err, ErrRateLimit) {
		return time.Time{}, false
	}

	_, resetAt, found := strings.Cut(err.Error(), ":")
	if !found {
		return time.Time{}, false
	}

	resetAtInt, err := strconv.Atoi(resetAt)
	if err != nil {
		return time.Time{}, false
	}

	return time.Unix(int64(resetAtInt), 0), true
}

func wrapRateLimit(resetAt int64) error {
	return fmt.Errorf("%v:%d", resetAt, ErrRateLimit)
}
