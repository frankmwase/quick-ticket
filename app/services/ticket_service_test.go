package services

import (
	"context"
	"reflect"
	"testing"

	"quick-ticket/domain"

	"go.uber.org/mock/gomock"
)

// Mock representations
type MockTicketRepository struct {
	ctrl     *gomock.Controller
	recorder *MockTicketRepositoryMockRecorder
}

type MockTicketRepositoryMockRecorder struct {
	mock *MockTicketRepository
}

func NewMockTicketRepository(ctrl *gomock.Controller) *MockTicketRepository {
	mock := &MockTicketRepository{ctrl: ctrl}
	mock.recorder = &MockTicketRepositoryMockRecorder{mock}
	return mock
}

func (m *MockTicketRepository) EXPECT() *MockTicketRepositoryMockRecorder {
	return m.recorder
}

func (m *MockTicketRepository) SaveBulk(ctx context.Context, tickets []*domain.Ticket) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveBulk", ctx, tickets)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTicketRepositoryMockRecorder) SaveBulk(ctx, tickets interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveBulk", reflect.TypeOf((*MockTicketRepository)(nil).SaveBulk), ctx, tickets)
}

func (m *MockTicketRepository) FindByID(ctx context.Context, id string) (*domain.Ticket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", ctx, id)
	ret0, _ := ret[0].(*domain.Ticket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTicketRepositoryMockRecorder) FindByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockTicketRepository)(nil).FindByID), ctx, id)
}

func (m *MockTicketRepository) FindByToken(ctx context.Context, token string) (*domain.Ticket, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByToken", ctx, token)
	ret0, _ := ret[0].(*domain.Ticket)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTicketRepositoryMockRecorder) FindByToken(ctx, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByToken", reflect.TypeOf((*MockTicketRepository)(nil).FindByToken), ctx, token)
}

func (m *MockTicketRepository) Update(ctx context.Context, ticket *domain.Ticket) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, ticket)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTicketRepositoryMockRecorder) Update(ctx, ticket interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockTicketRepository)(nil).Update), ctx, ticket)
}

type MockBatchRepository struct {
	ctrl     *gomock.Controller
	recorder *MockBatchRepositoryMockRecorder
}

type MockBatchRepositoryMockRecorder struct {
	mock *MockBatchRepository
}

func NewMockBatchRepository(ctrl *gomock.Controller) *MockBatchRepository {
	mock := &MockBatchRepository{ctrl: ctrl}
	mock.recorder = &MockBatchRepositoryMockRecorder{mock}
	return mock
}

func (m *MockBatchRepository) EXPECT() *MockBatchRepositoryMockRecorder {
	return m.recorder
}

func (m *MockBatchRepository) FindByID(ctx context.Context, id string) (*domain.TicketBatch, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByID", ctx, id)
	ret0, _ := ret[0].(*domain.TicketBatch)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockBatchRepositoryMockRecorder) FindByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByID", reflect.TypeOf((*MockBatchRepository)(nil).FindByID), ctx, id)
}

func (m *MockBatchRepository) Save(ctx context.Context, batch *domain.TicketBatch) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, batch)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockBatchRepositoryMockRecorder) Save(ctx, batch interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockBatchRepository)(nil).Save), ctx, batch)
}

func TestGenerateBulk(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTicketRepo := NewMockTicketRepository(ctrl)
	mockBatchRepo := NewMockBatchRepository(ctrl)

	mockBatchRepo.EXPECT().FindByID(gomock.Any(), "batch-1").Return(&domain.TicketBatch{ID: "batch-1"}, nil)
	mockTicketRepo.EXPECT().SaveBulk(gomock.Any(), gomock.Len(10)).Return(nil)

	svc := NewTicketService(mockTicketRepo, mockBatchRepo, nil, nil)

	req := domain.BulkGenerateRequest{
		TenantID: "tenant-1",
		BatchID:  "batch-1",
		Count:    10,
	}

	tickets, err := svc.GenerateBulk(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(tickets) != 10 {
		t.Fatalf("expected 10 tickets, got %d", len(tickets))
	}
}

func TestVerifyTicket_Geofence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTicketRepo := NewMockTicketRepository(ctrl)
	mockBatchRepo := NewMockBatchRepository(ctrl)

	// Festival coordinates: Lilongwe, roughly -13.98, 33.78
	geofence := &domain.Geofence{
		Lat:          -13.98,
		Lng:          33.78,
		RadiusMeters: 500, // 500 meters radius
	}

	batch := &domain.TicketBatch{
		ID:       "batch-1",
		Geofence: geofence,
	}

	ticket := &domain.Ticket{
		ID:          "tick-1",
		BatchID:     "batch-1",
		SecureToken: "token-1",
		Status:      domain.StatusIssued,
	}

	mockTicketRepo.EXPECT().FindByToken(gomock.Any(), "token-1").Return(ticket, nil).AnyTimes()
	mockBatchRepo.EXPECT().FindByID(gomock.Any(), "batch-1").Return(batch, nil).AnyTimes()

	svc := NewTicketService(mockTicketRepo, mockBatchRepo, nil, nil)

	// Inside radius
	res, err := svc.VerifyTicket(context.Background(), domain.VerificationContext{
		Token:     "token-1",
		Latitude:  -13.9801,
		Longitude: 33.7801,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !res.Valid {
		t.Fatalf("expected valid true, got message: %v", res.Message)
	}

	// Outside radius
	resOutside, err := svc.VerifyTicket(context.Background(), domain.VerificationContext{
		Token:     "token-1",
		Latitude:  -14.00,
		Longitude: 33.80,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resOutside.Valid {
		t.Fatalf("expected valid false")
	}
}
