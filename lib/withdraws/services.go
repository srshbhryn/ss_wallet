package withdraws

import (
	"context"
	"wallet/lib/core"
	"wallet/lib/withdraws/enums"
	"wallet/lib/withdraws/repository"

	"github.com/google/uuid"
)

type service struct {
	coreRepoFactory     core.RepoFactory
	withdrawRepoFactory repository.RepoFactory
}

func (s *service) Create(ctx context.Context, withdraw *Withdrawal) error {
	if withdraw.ID != uuid.Nil {
		return ErrInvalidState
	}
	withdrawRepo := s.withdrawRepoFactory.New(nil)
	coreRepo := s.coreRepoFactory.New(withdrawRepo.GetDBTransaction())
	defer func() {
		_ = withdrawRepo.RollBack()
	}()
	wallet, err := coreRepo.Wallet().GetOrCreateForUpdate(ctx, withdraw.WalletID)
	if err != nil {
		return err
	}
	if wallet.AvailableBalance < withdraw.Amount {
		return ErrInsufficientBalance
	}
	wallet.AvailableBalance -= withdraw.Amount
	wallet.BlockedBalance += withdraw.Amount
	if err := coreRepo.Wallet().Update(ctx, wallet); err != nil {
		return nil
	}
	if err := withdrawRepo.Create(ctx, withdraw); err != nil {
		return err
	}
	trx := &core.Transaction{
		Amount:        -withdraw.Amount,
		BlockedAmount: withdraw.Amount,
		WalletID:      withdraw.WalletID,
		Description:   "blocking for withdrawal",
		Reference:     withdraw.ID,
	}
	if err := coreRepo.Transaction().Create(ctx, trx); err != nil {
		return err
	}
	withdraw.BlockTransactionID = trx.ID
	withdraw.Status = enums.NEW
	if err := withdrawRepo.Update(ctx, withdraw); err != nil {
		return err
	}
	return withdrawRepo.Commit()
}

func (s *service) Reverse(ctx context.Context, withdraw *Withdrawal) error {
	if withdraw.Status == enums.FAILED || withdraw.Status == enums.SUCCESS {
		return ErrInvalidState
	}
	withdrawRepo := s.withdrawRepoFactory.New(nil)
	coreRepo := s.coreRepoFactory.New(withdrawRepo.GetDBTransaction())
	defer func() {
		_ = withdrawRepo.RollBack()
	}()
	wallet, err := coreRepo.Wallet().GetOrCreateForUpdate(ctx, withdraw.WalletID)
	if err != nil {
		return err
	}
	wallet.AvailableBalance += withdraw.Amount
	wallet.BlockedBalance -= withdraw.Amount
	if err := coreRepo.Wallet().Update(ctx, wallet); err != nil {
		return nil
	}
	trx := &core.Transaction{
		Amount:        withdraw.Amount,
		BlockedAmount: -withdraw.Amount,
		WalletID:      withdraw.WalletID,
		Description:   "withdraw cancellation",
		Reference:     withdraw.ID,
	}
	if err := coreRepo.Transaction().Create(ctx, trx); err != nil {
		return err
	}
	withdraw.ReverserTransactionID = trx.ID
	withdraw.Status = enums.FAILED
	if err := withdrawRepo.Update(ctx, withdraw); err != nil {
		return err
	}
	return withdrawRepo.Commit()
}

func (s *service) MarkAsSent(ctx context.Context, withdraw *Withdrawal) error {
	if withdraw.Status != enums.NEW {
		return ErrInvalidState
	}
	withdrawRepo := s.withdrawRepoFactory.New(nil)
	defer func() {
		_ = withdrawRepo.RollBack()
	}()
	withdraw.Status = enums.SENT
	if err := withdrawRepo.Update(ctx, withdraw); err != nil {
		return err
	}
	return withdrawRepo.Commit()
}

func (s *service) Complete(ctx context.Context, withdraw *Withdrawal) error {
	if withdraw.Status != enums.SENT {
		return ErrInvalidState
	}
	withdrawRepo := s.withdrawRepoFactory.New(nil)
	coreRepo := s.coreRepoFactory.New(withdrawRepo.GetDBTransaction())
	defer func() {
		_ = withdrawRepo.RollBack()
	}()
	wallet, err := coreRepo.Wallet().GetOrCreateForUpdate(ctx, withdraw.WalletID)
	if err != nil {
		return err
	}
	wallet.BlockedBalance -= withdraw.Amount
	if err := coreRepo.Wallet().Update(ctx, wallet); err != nil {
		return nil
	}
	trx := &core.Transaction{
		Amount:      withdraw.Amount,
		WalletID:    withdraw.WalletID,
		Description: "withdraw completion",
		Reference:   withdraw.ID,
	}
	if err := coreRepo.Transaction().Create(ctx, trx); err != nil {
		return err
	}
	withdraw.ReverserTransactionID = trx.ID
	withdraw.Status = enums.SUCCESS
	if err := withdrawRepo.Update(ctx, withdraw); err != nil {
		return err
	}
	return withdrawRepo.Commit()

}
