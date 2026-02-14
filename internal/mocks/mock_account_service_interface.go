package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockAccountServiceInterface is a mock type for the AccountServiceInterface type.
type MockAccountServiceInterface struct {
	mock.Mock
}

// DeleteAccount provides a mock function with given fields: ctx, userID
func (_m *MockAccountServiceInterface) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
