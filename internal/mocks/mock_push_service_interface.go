package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockPushServiceInterface is a mock type for the PushServiceInterface type.
type MockPushServiceInterface struct {
	mock.Mock
}

// Subscribe provides a mock function with given fields: ctx, userID, endpoint, p256dh, auth, userAgent
func (_m *MockPushServiceInterface) Subscribe(ctx context.Context, userID uuid.UUID, endpoint string, p256dh string, auth string, userAgent string) error {
	ret := _m.Called(ctx, userID, endpoint, p256dh, auth, userAgent)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, string, string, string) error); ok {
		r0 = rf(ctx, userID, endpoint, p256dh, auth, userAgent)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unsubscribe provides a mock function with given fields: ctx, endpoint
func (_m *MockPushServiceInterface) Unsubscribe(ctx context.Context, endpoint string) error {
	ret := _m.Called(ctx, endpoint)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, endpoint)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendPushToUser provides a mock function with given fields: ctx, userID, title, body, url
func (_m *MockPushServiceInterface) SendPushToUser(ctx context.Context, userID uuid.UUID, title string, body string, url string) error {
	ret := _m.Called(ctx, userID, title, body, url)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, string, string, string) error); ok {
		r0 = rf(ctx, userID, title, body, url)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendTestPush provides a mock function with given fields: ctx, userID
func (_m *MockPushServiceInterface) SendTestPush(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetVAPIDPublicKey provides a mock function
func (_m *MockPushServiceInterface) GetVAPIDPublicKey() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// IsEnabled provides a mock function
func (_m *MockPushServiceInterface) IsEnabled() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
