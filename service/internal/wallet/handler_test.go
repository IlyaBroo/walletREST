package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Deposit(walletID string, amount int64, ctx context.Context) error {
	args := m.Called(walletID, amount, ctx)
	return args.Error(0)
}

func (m *MockRepository) Withdraw(walletID string, amount int64, ctx context.Context) error {
	args := m.Called(walletID, amount, ctx)
	return args.Error(0)
}

func (m *MockRepository) GetBalance(walletID string, ctx context.Context) (int64, error) {
	args := m.Called(walletID, ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) Close() {}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) InfoCtx(ctx context.Context, msg string) {
	m.Called(ctx, msg)
}

func (m *MockLogger) ErrorCtx(ctx context.Context, msg string) {
	m.Called(ctx, msg)
}

func (m *MockLogger) DebugCtx(ctx context.Context, msg string) {
	m.Called(ctx, msg)
}
func (m *MockLogger) FatalCtx(ctx context.Context, msg string, err error) {
	m.Called(ctx, msg, err)
}
func (m *MockLogger) WarnCtx(ctx context.Context, msg string) {
	m.Called(ctx, msg)
}

func TestHandleWalletOperation(t *testing.T) {
	mockLogger := new(MockLogger)
	mockRepo := new(MockRepository)
	handler := &Handler{repo: mockRepo, lg: mockLogger}

	tests := []struct {
		name           string
		requestBody    WalletOperationRequest
		expectedStatus int
		mockRepoFunc   func()
		mockLoggerFunc func()
	}{
		{
			name: "Successful Deposit",
			requestBody: WalletOperationRequest{
				WalletID:      "123",
				OperationType: DEPOSIT,
				Amount:        100,
			},
			expectedStatus: http.StatusOK,
			mockRepoFunc: func() {
				mockRepo.On("Deposit", "123", int64(100), mock.Anything).Return(nil)
			},
			mockLoggerFunc: func() {
				mockLogger.On("InfoCtx", mock.Anything, "wallet id = 123, operation = DEPOSIT , amount = 100 is success").Return()
			},
		},
		{
			name: "Deposit Error",
			requestBody: WalletOperationRequest{
				WalletID:      "1234",
				OperationType: DEPOSIT,
				Amount:        100,
			},
			expectedStatus: http.StatusNotFound,
			mockRepoFunc: func() {
				mockRepo.On("Deposit", "1234", int64(100), mock.Anything).Return(errWalletid)
			},
			mockLoggerFunc: func() {
				mockLogger.On("ErrorCtx", mock.Anything, "insufficient funds or walletid not found").Return()
			},
		},
		{
			name: "Invalid Operation Type",
			requestBody: WalletOperationRequest{
				WalletID:      "123",
				OperationType: "INVALID",
				Amount:        100,
			},
			expectedStatus: http.StatusBadRequest,
			mockRepoFunc:   func() {},
			mockLoggerFunc: func() {
				mockLogger.On("ErrorCtx", mock.Anything, "invalid operation type").Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepoFunc()
			tt.mockLoggerFunc()

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/wallet/operation", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.HandleWalletOperation(w, req)

			res := w.Result()
			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetWalletBalance(t *testing.T) {
	mockLogger := new(MockLogger)
	mockRepo := new(MockRepository)
	handler := &Handler{repo: mockRepo, lg: mockLogger}

	r := chi.NewRouter()
	r.Get("/wallet/balance/{id}", handler.GetWalletBalance)

	tests := []struct {
		name           string
		walletID       string
		expectedStatus int
		mockRepoFunc   func()
		mockLoggerFunc func()
	}{
		{
			name:           "Successful Get Balance",
			walletID:       "123",
			expectedStatus: http.StatusOK,
			mockRepoFunc: func() {
				mockRepo.On("GetBalance", "123", mock.Anything).Return(int64(1000), nil)
			},
			mockLoggerFunc: func() {
				mockLogger.On("DebugCtx", mock.Anything, "walletId=123").Return()
				mockLogger.On("InfoCtx", mock.Anything, "wallet id = 123, balance = 1000 is success").Return()
			},
		},
		{
			name:           "Wallet Not Found",
			walletID:       "1234",
			expectedStatus: http.StatusNotFound,
			mockRepoFunc: func() {
				mockRepo.On("GetBalance", "1234", mock.Anything).Return(int64(0), errWalletid)
			},
			mockLoggerFunc: func() {
				mockLogger.On("DebugCtx", mock.Anything, "walletId=1234").Return()
				mockLogger.On("ErrorCtx", mock.Anything, "walletid not found").Return()
			},
		},
		{
			name:           "Error Getting Balance",
			walletID:       "12345",
			expectedStatus: http.StatusInternalServerError,
			mockRepoFunc: func() {
				mockRepo.On("GetBalance", "12345", mock.Anything).Return(int64(0), errors.New("some error"))
			},
			mockLoggerFunc: func() {
				mockLogger.On("DebugCtx", mock.Anything, "walletId=12345").Return()
				mockLogger.On("ErrorCtx", mock.Anything, "error getting balance").Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepoFunc()
			tt.mockLoggerFunc()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/wallet/balance/%s", tt.walletID), nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)
			res := w.Result()
			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
