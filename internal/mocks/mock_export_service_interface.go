package mocks

import (
	context "context"
	services "savvy/internal/services"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockExportServiceInterface is a mock type for the ExportServiceInterface type.
type MockExportServiceInterface struct {
	mock.Mock
}

// ExportUserData provides a mock function with given fields: ctx, userID
func (_m *MockExportServiceInterface) ExportUserData(ctx context.Context, userID uuid.UUID) (*services.ExportData, error) {
	ret := _m.Called(ctx, userID)

	var r0 *services.ExportData
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *services.ExportData); ok {
		r0 = rf(ctx, userID)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*services.ExportData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExportCardsByIDs provides a mock function with given fields: ctx, userID, ids
func (_m *MockExportServiceInterface) ExportCardsByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*services.BatchExportData, error) {
	ret := _m.Called(ctx, userID, ids)

	var r0 *services.BatchExportData
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, []uuid.UUID) *services.BatchExportData); ok {
		r0 = rf(ctx, userID, ids)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*services.BatchExportData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, []uuid.UUID) error); ok {
		r1 = rf(ctx, userID, ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExportVouchersByIDs provides a mock function with given fields: ctx, userID, ids
func (_m *MockExportServiceInterface) ExportVouchersByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*services.BatchExportData, error) {
	ret := _m.Called(ctx, userID, ids)

	var r0 *services.BatchExportData
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, []uuid.UUID) *services.BatchExportData); ok {
		r0 = rf(ctx, userID, ids)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*services.BatchExportData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, []uuid.UUID) error); ok {
		r1 = rf(ctx, userID, ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExportGiftCardsByIDs provides a mock function with given fields: ctx, userID, ids
func (_m *MockExportServiceInterface) ExportGiftCardsByIDs(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) (*services.BatchExportData, error) {
	ret := _m.Called(ctx, userID, ids)

	var r0 *services.BatchExportData
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, []uuid.UUID) *services.BatchExportData); ok {
		r0 = rf(ctx, userID, ids)
	} else if ret.Get(0) != nil {
		r0 = ret.Get(0).(*services.BatchExportData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, []uuid.UUID) error); ok {
		r1 = rf(ctx, userID, ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
