package wallet

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPool struct {
	mock.Mock
}

func (m *MockPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	ret := m.Called(ctx, sql, args[0], args[1])
	return ret.Get(0).(pgconn.CommandTag), ret.Error(1)
}

func (m *MockPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	ret := m.Called(ctx, sql, args[0])
	return ret.Get(0).(pgx.Row)
}

func (m *MockPool) Close() {
	m.Called()
}

type mockRow struct {
	balance int64
	err     error
}

func (r *mockRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != 1 {
		return errors.New("expected one destination")
	}
	*dest[0].(*int64) = r.balance
	return nil
}

func newMockRow(balance int64, err error) *mockRow {
	return &mockRow{
		balance: balance,
		err:     err,
	}
}

func TestRepository_Deposit(t *testing.T) {
	mockLogger := new(MockLogger)

	mockPool := new(MockPool)
	repo := &Repository{db: mockPool, lg: mockLogger, ctx: context.Background()}

	tests := []struct {
		name           string
		walletID       string
		amount         int64
		mockSetup      func()
		expectedErr    error
		mockLoggerFunc func()
	}{
		{
			name:     "Successful Deposit",
			walletID: "123",
			amount:   100,
			mockSetup: func() {
				mockPool.On("Exec", mock.Anything, "UPDATE wallets SET balance = balance + $1 WHERE id = $2", int64(100), "123").
					Return(pgconn.NewCommandTag("UPDATE 1"), nil).Once()
			},
			mockLoggerFunc: func() {

			},
			expectedErr: nil,
		},
		{
			name:     "Wallet Not Found",
			walletID: "123",
			amount:   100,
			mockSetup: func() {
				mockPool.On("Exec", mock.Anything, "UPDATE wallets SET balance = balance + $1 WHERE id = $2", int64(100), "123").
					Return(pgconn.NewCommandTag("UPDATE 0"), nil).Once()
			},
			mockLoggerFunc: func() {
				mockLogger.On("ErrorCtx", mock.Anything, "func deposit walletid not found").Return().Once()
			},
			expectedErr: errWalletid,
		},
		{
			name:     "Database Error",
			walletID: "123",
			amount:   100,
			mockSetup: func() {
				mockPool.On("Exec", mock.Anything, "UPDATE wallets SET balance = balance + $1 WHERE id = $2", int64(100), "123").
					Return(pgconn.CommandTag{}, errors.New("db error")).Once()
			},
			mockLoggerFunc: func() {
				mockLogger.On("ErrorCtx", mock.Anything, "func deposit sql query failed").Return().Once()
			},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			tt.mockLoggerFunc()

			err := repo.Deposit(tt.walletID, tt.amount, context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			mockPool.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestRepository_Withdraw(t *testing.T) {
	mockLogger := new(MockLogger)

	mockPool := new(MockPool)
	repo := &Repository{db: mockPool, lg: mockLogger, ctx: context.Background()}

	tests := []struct {
		name           string
		walletID       string
		amount         int64
		mockSetup      func()
		mockLoggerFunc func()
		expectedErr    error
	}{
		{
			name:     "Successful Withdraw",
			walletID: "123",
			amount:   50,
			mockSetup: func() {
				mockPool.On("Exec", mock.Anything, "UPDATE wallets SET balance = balance - $1 WHERE id = $2 AND balance >= $1", int64(50), "123").
					Return(pgconn.NewCommandTag("UPDATE 1"), nil).Once()
			},
			mockLoggerFunc: func() {

			},
			expectedErr: nil,
		},
		{
			name:     "Insufficient Funds",
			walletID: "123",
			amount:   50,
			mockSetup: func() {
				mockPool.On("Exec", mock.Anything, "UPDATE wallets SET balance = balance - $1 WHERE id = $2 AND balance >= $1", int64(50), "123").
					Return(pgconn.NewCommandTag("UPDATE 0"), nil).Once()
			},
			mockLoggerFunc: func() {
				mockLogger.On("ErrorCtx", mock.Anything, "func withdraw insufficient funds or walletid not found").Return().Once()
			},
			expectedErr: errWithdraw,
		},
		{
			name:     "Database Error",
			walletID: "123",
			amount:   50,
			mockSetup: func() {
				mockPool.On("Exec", mock.Anything, "UPDATE wallets SET balance = balance - $1 WHERE id = $2 AND balance >= $1", int64(50), "123").
					Return(pgconn.CommandTag{}, errors.New("db error")).Once()
			},
			mockLoggerFunc: func() {
				mockLogger.On("ErrorCtx", mock.Anything, "func withdraw sql query failed").Return().Once()
			},
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			tt.mockLoggerFunc()

			err := repo.Withdraw(tt.walletID, tt.amount, context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			mockPool.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestRepository_GetBalance(t *testing.T) {
	mockLogger := new(MockLogger)

	mockPool := new(MockPool)
	repo := &Repository{db: mockPool, lg: mockLogger, ctx: context.Background()}

	tests := []struct {
		name            string
		walletID        string
		mockSetup       func()
		expectedBalance int64
		expectedErr     error
		mockLoggerFunc  func()
	}{
		{
			name:     "Successful Get Balance",
			walletID: "123",
			mockSetup: func() {
				mockPool.On("QueryRow", mock.Anything, "SELECT balance FROM wallets WHERE id = $1", "123").
					Return(newMockRow(int64(100), nil)).Once()
			},
			mockLoggerFunc: func() {

			},
			expectedBalance: 100,
			expectedErr:     nil,
		},
		{
			name:     "Wallet Not Found",
			walletID: "123",
			mockSetup: func() {
				mockPool.On("QueryRow", mock.Anything, "SELECT balance FROM wallets WHERE id = $1", "123").
					Return(newMockRow(int64(0), pgx.ErrNoRows)).Once()
			},
			mockLoggerFunc: func() {
				mockLogger.On("ErrorCtx", mock.Anything, "func getbalance walletid not found").Return().Once()
			},
			expectedBalance: 0,
			expectedErr:     errWalletid,
		},
		{
			name:     "Database Error",
			walletID: "123",
			mockSetup: func() {
				mockPool.On("QueryRow", mock.Anything, "SELECT balance FROM wallets WHERE id = $1", "123").
					Return(newMockRow(int64(0), errors.New("db error"))).Once()
			},
			mockLoggerFunc: func() {
				mockLogger.On("ErrorCtx", mock.Anything, "Could not scan wallet").Return().Once()
			},
			expectedBalance: 0,
			expectedErr:     errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			tt.mockLoggerFunc()

			balance, err := repo.GetBalance(tt.walletID, context.Background())

			assert.Equal(t, tt.expectedBalance, balance)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			mockPool.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
