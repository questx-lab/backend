package pinata

import (
	"context"
	"io"
)

type IEndpoint interface {
	PinFile(ctx context.Context, name string, f io.Reader) (string, error)
}
