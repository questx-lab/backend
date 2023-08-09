package common

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/questx-lab/backend/pkg/xcontext"
)

func MapKeys[K comparable, V any](m map[K]V) []K {
	var result []K
	for k := range m {
		result = append(result, k)
	}
	return result
}

// Batch splits a into two slices. DO NOT write on the returned value.
func Batch[T any](a *[]T, n int) []T {
	if len(*a) > n {
		batch := (*a)[:n]
		*a = (*a)[n:]
		return batch
	}

	b := (*a)
	*a = (*a)[:0]
	return b
}

// DetectBottleneck is used to write a warning message when the pending queue
// size is too large for the number of processed elements.
func DetectBottleneck[T any](ctx context.Context, processed, queue []T, reason string) {
	processedSize := len(processed)
	queueSize := len(queue)
	if queueSize > 5*processedSize {
		xcontext.Logger(ctx).Warnf("Bottleneck detected when %s, ratio=%d", reason, queueSize/processedSize)
	}
}

// DetectBottleneckCount likes DetectBottleneck, but it only writes the warning
// if detected time equals to given number.
func DetectBottleneckCount[T any](ctx context.Context, processed, queue []T, reason string, number int, count *int) {
	processedSize := len(processed)
	queueSize := len(queue)
	if queueSize > 5*processedSize {
		(*count)++
	}

	if *count >= number {
		xcontext.Logger(ctx).Warnf("Bottleneck detected when %s, ratio=%d", reason, queueSize/processedSize)
		*count = 0
	}
}

const BucketDuration = time.Hour * 24 * 10

func ParseInviteDiscordURL(rawURL string) (string, error) {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return "", err
	}

	if u.Scheme != "https" {
		return "", errors.New("invalid scheme")
	}

	if u.Host != "discord.gg" {
		return "", errors.New("invalid domain")
	}

	return strings.TrimLeft(u.Path, "/"), nil
}
