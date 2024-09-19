package runners

import (
	"bytes"
	"context"

	"github.com/stretchr/testify/mock"
)

// MockCommandRunner is a mock implementation of the runners.Provider interface
type MockCommandRunner struct {
	mock.Mock
}

func (m *MockCommandRunner) Run(ctx context.Context, input *bytes.Reader) ([]byte, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockCommandRunner) HealthCheck(ctx context.Context) error {
	return m.Called().Error(0)
}
