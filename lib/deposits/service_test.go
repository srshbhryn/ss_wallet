package deposits_test

import (
	"context"
	"testing"
	"wallet/lib/core"
	"wallet/lib/deposits"
	"wallet/lib/deposits/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// --- Mocks ---

type MockDepositRepo struct{ mock.Mock }

func (m *MockDepositRepo) Create(ctx context.Context, d *repository.Deposit) error {
	return m.Called(ctx, d).Error(0)
}
func (m *MockDepositRepo) Update(ctx context.Context, d *repository.Deposit) error {
	return m.Called(ctx, d).Error(0)
}
func (m *MockDepositRepo) GetApplicableDeposits(ctx context.Context, prefix string) ([]repository.Deposit, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]repository.Deposit), args.Error(1)
}
func (m *MockDepositRepo) GetDBTransaction() *gorm.DB { return nil }
func (m *MockDepositRepo) Commit() error              { return m.Called().Error(0) }
func (m *MockDepositRepo) RollBack() error            { return m.Called().Error(0) }

type MockDepositRepoFactory struct{ mock.Mock }

func (m *MockDepositRepoFactory) New(tx *gorm.DB) repository.Repo {
	args := m.Called(tx)
	return args.Get(0).(repository.Repo)
}

type MockWalletRepo struct{ mock.Mock }

func (m *MockWalletRepo) GetOrCreateForUpdate(ctx context.Context, userID uuid.UUID) (*core.Wallet, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*core.Wallet), args.Error(1)
}
func (m *MockWalletRepo) GetOrCreate(ctx context.Context, userID uuid.UUID) (*core.Wallet, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*core.Wallet), args.Error(1)
}
func (m *MockWalletRepo) Update(ctx context.Context, wallet *core.Wallet) error {
	return m.Called(ctx, wallet).Error(0)
}

type MockTransactionRepo struct{ mock.Mock }

func (m *MockTransactionRepo) Create(ctx context.Context, trx *core.Transaction) error {
	return m.Called(ctx, trx).Error(0)
}
func (m *MockTransactionRepo) Get(ctx context.Context, userID uuid.UUID, pageNumber int, pageSize int) ([]core.Transaction, bool, error) {
	args := m.Called(ctx, userID, pageNumber, pageSize)
	return args.Get(0).([]core.Transaction), args.Bool(1), args.Error(2)
}

type MockCoreRepo struct{ mock.Mock }

func (m *MockCoreRepo) Wallet() core.WalletRepo {
	return m.Called().Get(0).(core.WalletRepo)
}
func (m *MockCoreRepo) Transaction() core.TransactionRepo {
	return m.Called().Get(0).(core.TransactionRepo)
}
func (m *MockCoreRepo) GetDBTransaction() *gorm.DB { return nil }
func (m *MockCoreRepo) Commit() error              { return m.Called().Error(0) }
func (m *MockCoreRepo) RollBack() error            { return m.Called().Error(0) }

type MockCoreRepoFactory struct{ mock.Mock }

func (m *MockCoreRepoFactory) New(tx *gorm.DB) core.Repo {
	args := m.Called(tx)
	return args.Get(0).(core.Repo)
}

// --- Tests ---

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	deposit := &repository.Deposit{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Amount:      100,
		Description: "test deposit",
	}

	depRepo := new(MockDepositRepo)
	walletRepo := new(MockWalletRepo)
	trxRepo := new(MockTransactionRepo)
	coreRepo := new(MockCoreRepo)
	depRepoFactory := new(MockDepositRepoFactory)
	coreRepoFactory := new(MockCoreRepoFactory)

	// DepositRepo mocks
	depRepo.On("Create", ctx, deposit).Return(nil)
	depRepo.On("Update", ctx, deposit).Return(nil)
	depRepo.On("Commit").Return(nil)
	depRepo.On("RollBack").Return(nil)
	depRepoFactory.On("New", (*gorm.DB)(nil)).Return(depRepo)

	// WalletRepo & CoreRepo mocks
	wallet := &core.Wallet{UserID: deposit.UserID}
	walletRepo.On("GetOrCreateForUpdate", ctx, deposit.UserID).Return(wallet, nil)
	walletRepo.On("Update", ctx, wallet).Return(nil)

	trxRepo.On("Create", ctx, mock.Anything).Run(func(args mock.Arguments) {
		trx := args.Get(1).(*core.Transaction)
		trx.ID = 1 // assign fake ID so deposit.BlockTransactionID is set
	}).Return(nil)

	coreRepo.On("Wallet").Return(walletRepo)
	coreRepo.On("Transaction").Return(trxRepo)
	coreRepoFactory.On("New", (*gorm.DB)(nil)).Return(coreRepo)

	service := deposits.New(coreRepoFactory, depRepoFactory)
	err := service.Create(ctx, deposit)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), wallet.BlockedBalance)
	assert.NotZero(t, deposit.BlockTransactionID)
}

func TestService_Apply(t *testing.T) {
	ctx := context.Background()
	deposit := &repository.Deposit{
		ID:                 uuid.New(),
		UserID:             uuid.New(),
		Amount:             200,
		Description:        "apply deposit",
		BlockTransactionID: 1,
	}

	depRepo := new(MockDepositRepo)
	walletRepo := new(MockWalletRepo)
	trxRepo := new(MockTransactionRepo)
	coreRepo := new(MockCoreRepo)
	depRepoFactory := new(MockDepositRepoFactory)
	coreRepoFactory := new(MockCoreRepoFactory)

	// DepositRepo mocks
	depRepo.On("Update", ctx, deposit).Return(nil)
	depRepo.On("Commit").Return(nil)
	depRepo.On("RollBack").Return(nil)
	depRepoFactory.On("New", (*gorm.DB)(nil)).Return(depRepo)

	// WalletRepo & CoreRepo mocks
	wallet := &core.Wallet{UserID: deposit.UserID, BlockedBalance: 200}
	walletRepo.On("GetOrCreateForUpdate", ctx, deposit.UserID).Return(wallet, nil)
	walletRepo.On("Update", ctx, wallet).Return(nil)

	trxRepo.On("Create", ctx, mock.Anything).Run(func(args mock.Arguments) {
		trx := args.Get(1).(*core.Transaction)
		trx.ID = 2 // assign fake ID so deposit.ApplyTransactionID is set
	}).Return(nil)

	coreRepo.On("Wallet").Return(walletRepo)
	coreRepo.On("Transaction").Return(trxRepo)
	coreRepoFactory.On("New", (*gorm.DB)(nil)).Return(coreRepo)

	service := deposits.New(coreRepoFactory, depRepoFactory)
	err := service.Apply(ctx, deposit)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), wallet.BlockedBalance)
	assert.Equal(t, int64(200), wallet.AvailableBalance)
	assert.NotZero(t, deposit.ApplyTransactionID)
}

func TestService_GetApplicableDeposits(t *testing.T) {
	ctx := context.Background()
	depRepo := new(MockDepositRepo)
	depRepoFactory := new(MockDepositRepoFactory)

	expected := []repository.Deposit{
		{ID: uuid.New(), Amount: 100},
	}

	depRepo.On("GetApplicableDeposits", ctx, "test").Return(expected, nil)
	depRepoFactory.On("New", (*gorm.DB)(nil)).Return(depRepo)

	service := deposits.New(nil, depRepoFactory)
	result, err := service.GetApplicableDeposits(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
