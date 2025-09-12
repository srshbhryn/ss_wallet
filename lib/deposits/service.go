package deposits

import (
	"context"
	"wallet/lib/core"
	"wallet/lib/deposits/repository"
)

type Deposit = repository.Deposit

type Service interface {
	Create(context.Context, *Deposit) error
	Apply(context.Context, *Deposit) error
	GetApplicableDeposits(ctx context.Context, IDPrefix string) ([]Deposit, error)
}

func New(
	coreRepoFactory core.RepoFactory,
	repoFactory repository.RepoFactory,
) Service {
	return &service{
		coreRepoFactory: coreRepoFactory,
		repoFactory:     repoFactory,
	}
}

type service struct {
	coreRepoFactory core.RepoFactory
	repoFactory     repository.RepoFactory
}

func (s *service) Create(ctx context.Context, deposit *Deposit) error {
	depositRepo := s.repoFactory.New(nil)
	coreRepo := s.coreRepoFactory.New(depositRepo.GetDBTransaction())
	defer func() {
		_ = depositRepo.RollBack()
	}()
	if err := depositRepo.Create(ctx, deposit); err != nil {
		return err
	}
	wallet, err := coreRepo.Wallet().GetOrCreateForUpdate(ctx, deposit.UserID)
	if err != nil {
		return err
	}
	wallet.BlockedBalance += deposit.Amount
	if err := coreRepo.Wallet().Update(ctx, wallet); err != nil {
		return err
	}
	trx := &core.Transaction{
		WalletID:      wallet.UserID,
		BlockedAmount: deposit.Amount,
		Amount:        0,
		Reference:     deposit.ID,
		Description:   deposit.Description,
	}
	if err := coreRepo.Transaction().Create(ctx, trx); err != nil {
		return err
	}
	deposit.BlockTransactionID = trx.ID
	if err := depositRepo.Update(ctx, deposit); err != nil {
		return err
	}
	return depositRepo.Commit()
}

func (s *service) Apply(ctx context.Context, deposit *Deposit) error {
	depositRepo := s.repoFactory.New(nil)
	coreRepo := s.coreRepoFactory.New(depositRepo.GetDBTransaction())
	defer func() {
		_ = depositRepo.RollBack()
	}()
	wallet, err := coreRepo.Wallet().GetOrCreateForUpdate(ctx, deposit.UserID)
	if err != nil {
		return err
	}
	wallet.AvailableBalance += deposit.Amount
	wallet.BlockedBalance -= deposit.Amount
	if err := coreRepo.Wallet().Update(ctx, wallet); err != nil {
		return err
	}
	trx := &core.Transaction{
		WalletID:      deposit.UserID,
		Amount:        deposit.Amount,
		BlockedAmount: -deposit.Amount,
		Reference:     deposit.ID,
		Description:   deposit.Description,
	}
	if err := coreRepo.Transaction().Create(ctx, trx); err != nil {
		return err
	}
	deposit.ApplyTransactionID = trx.ID
	if err := depositRepo.Update(ctx, deposit); err != nil {
		return err
	}
	return depositRepo.Commit()
}

func (s *service) GetApplicableDeposits(ctx context.Context, IDPrefix string) ([]Deposit, error) {
	return s.repoFactory.New(nil).GetApplicableDeposits(ctx, IDPrefix)
}
