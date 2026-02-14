package mocks

import (
	"context"
	"savvy/internal/services"

	"github.com/stretchr/testify/mock"
)

// MockHealthCheckServiceInterface is a mock for services.HealthCheckServiceInterface.
type MockHealthCheckServiceInterface struct {
	mock.Mock
}

func (m *MockHealthCheckServiceInterface) CheckReadiness(ctx context.Context) (*services.ReadinessReport, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ReadinessReport), args.Error(1)
}
