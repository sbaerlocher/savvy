package mocks

import (
	context "context"
	services "savvy/internal/services"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockSessionServiceInterface is a mock type for the SessionServiceInterface type.
type MockSessionServiceInterface struct {
	mock.Mock
}

// ListUserSessions provides a mock function with given fields: ctx, userID, currentTokenHash
func (_m *MockSessionServiceInterface) ListUserSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) ([]services.SessionDTO, error) {
	ret := _m.Called(ctx, userID, currentTokenHash)

	var r0 []services.SessionDTO
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) []services.SessionDTO); ok {
		r0 = rf(ctx, userID, currentTokenHash)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).([]services.SessionDTO)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, string) error); ok {
		r1 = rf(ctx, userID, currentTokenHash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RevokeSession provides a mock function with given fields: ctx, userID, sessionID
func (_m *MockSessionServiceInterface) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	ret := _m.Called(ctx, userID, sessionID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, uuid.UUID) error); ok {
		r0 = rf(ctx, userID, sessionID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RevokeOtherSessions provides a mock function with given fields: ctx, userID, currentTokenHash
func (_m *MockSessionServiceInterface) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentTokenHash string) (int64, error) {
	ret := _m.Called(ctx, userID, currentTokenHash)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string) int64); ok {
		r0 = rf(ctx, userID, currentTokenHash)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, string) error); ok {
		r1 = rf(ctx, userID, currentTokenHash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RevokeAllSessions provides a mock function with given fields: ctx, userID
func (_m *MockSessionServiceInterface) RevokeAllSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	ret := _m.Called(ctx, userID)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) int64); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CleanupExpired provides a mock function with given fields: ctx
func (_m *MockSessionServiceInterface) CleanupExpired(ctx context.Context) (int64, error) {
	ret := _m.Called(ctx)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context) int64); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
