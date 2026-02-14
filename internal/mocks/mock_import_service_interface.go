package mocks

import (
	context "context"
	io "io"
	services "savvy/internal/services"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockImportServiceInterface is a mock type for the ImportServiceInterface type.
type MockImportServiceInterface struct {
	mock.Mock
}

// PreviewJSON provides a mock function with given fields: ctx, data
func (_m *MockImportServiceInterface) PreviewJSON(ctx context.Context, data *services.ExportData) (*services.ImportPreview, error) {
	ret := _m.Called(ctx, data)

	var r0 *services.ImportPreview
	if rf, ok := ret.Get(0).(func(context.Context, *services.ExportData) *services.ImportPreview); ok {
		r0 = rf(ctx, data)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*services.ImportPreview)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *services.ExportData) error); ok {
		r1 = rf(ctx, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ImportJSON provides a mock function with given fields: ctx, userID, data
func (_m *MockImportServiceInterface) ImportJSON(ctx context.Context, userID uuid.UUID, data *services.ExportData) (*services.ImportResult, error) {
	ret := _m.Called(ctx, userID, data)

	var r0 *services.ImportResult
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, *services.ExportData) *services.ImportResult); ok {
		r0 = rf(ctx, userID, data)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*services.ImportResult)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, *services.ExportData) error); ok {
		r1 = rf(ctx, userID, data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ImportCardsCSV provides a mock function with given fields: ctx, userID, reader
func (_m *MockImportServiceInterface) ImportCardsCSV(ctx context.Context, userID uuid.UUID, reader io.Reader) (*services.ImportResult, error) {
	ret := _m.Called(ctx, userID, reader)

	var r0 *services.ImportResult
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, io.Reader) *services.ImportResult); ok {
		r0 = rf(ctx, userID, reader)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*services.ImportResult)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, io.Reader) error); ok {
		r1 = rf(ctx, userID, reader)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ImportVouchersCSV provides a mock function with given fields: ctx, userID, reader
func (_m *MockImportServiceInterface) ImportVouchersCSV(ctx context.Context, userID uuid.UUID, reader io.Reader) (*services.ImportResult, error) {
	ret := _m.Called(ctx, userID, reader)

	var r0 *services.ImportResult
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, io.Reader) *services.ImportResult); ok {
		r0 = rf(ctx, userID, reader)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*services.ImportResult)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, io.Reader) error); ok {
		r1 = rf(ctx, userID, reader)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ImportGiftCardsCSV provides a mock function with given fields: ctx, userID, reader
func (_m *MockImportServiceInterface) ImportGiftCardsCSV(ctx context.Context, userID uuid.UUID, reader io.Reader) (*services.ImportResult, error) {
	ret := _m.Called(ctx, userID, reader)

	var r0 *services.ImportResult
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, io.Reader) *services.ImportResult); ok {
		r0 = rf(ctx, userID, reader)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*services.ImportResult)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, io.Reader) error); ok {
		r1 = rf(ctx, userID, reader)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
