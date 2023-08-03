package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/questx-lab/backend/internal/common"
	"github.com/questx-lab/backend/pkg/errorx"
	"github.com/questx-lab/backend/pkg/router"
	"github.com/questx-lab/backend/pkg/xcontext"
)

func WithStartTime() router.MiddlewareFunc {
	return func(ctx context.Context) (context.Context, error) {
		now := time.Now()
		return xcontext.WithStartTime(ctx, now), nil
	}
}
func Prometheus() router.CloserFunc {
	return func(ctx context.Context) {
		startTime := xcontext.StartTime(ctx)

		req := xcontext.HTTPRequest(ctx)
		code := 0
		if err := xcontext.Error(ctx); err != nil {
			var errx errorx.Error
			if errors.As(err, &errx) {
				code = int(errx.Code)
			} else {
				code = -1
			}
		}
		path := req.URL.Path

		for key, counter := range common.PromCounters {
			switch key {
			case common.HTTPRequestTotal:
				counter.WithLabelValues(path, fmt.Sprint(code)).Inc()
			}

		}

		for key, histogram := range common.PromHistograms {
			switch key {
			case common.HTTPRequestDurationSeconds:
				histogram.WithLabelValues(path, fmt.Sprint(code)).Observe(time.Since(startTime).Seconds())
			}
		}
	}
}
