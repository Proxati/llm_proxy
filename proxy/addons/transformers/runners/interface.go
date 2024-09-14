package runners

import (
	"bytes"
	"context"
)

type Provider interface {
	Run(ctx context.Context, input *bytes.Reader) ([]byte, error)
	HealthCheck() error
}
