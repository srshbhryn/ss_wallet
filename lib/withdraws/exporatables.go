package withdraws

import (
	"context"
	"wallet/lib/core"
	"wallet/lib/withdraws/repository"
)

type Withdrawal = repository.Withdrawal

type Service interface {
	Create(context.Context, *Withdrawal) error
	Reverse(context.Context, *Withdrawal) error
	MarkAsSent(context.Context, *Withdrawal) error
	Complete(context.Context, *Withdrawal) error
}

func NewService(
	coreRepoFactory core.RepoFactory,
	withdrawRepoFactory repository.RepoFactory,
) Service {
	return &service{
		coreRepoFactory:     coreRepoFactory,
		withdrawRepoFactory: withdrawRepoFactory,
	}
}
