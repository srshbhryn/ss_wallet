package withdraws

import (
	"context"
	"time"
	"wallet/lib/core"
	"wallet/lib/withdraws/integrations"
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

type Worker interface {
	SendToBank(context.Context, *Withdrawal) error
	GetStatus(context.Context, *Withdrawal) error
	Run(ctx context.Context)
	Stop()
}

func NewWorker(
	updateStatus func(context.Context, *Withdrawal) error,
	concurrency int,
	backOff time.Duration,
	retryCount int,
	client integrations.BankClient,
) Worker {
	return &worker{
		updateStatus: updateStatus,
		concurrency:  concurrency,
		backOff:      backOff,
		retryCount:   retryCount,
		client:       client,
		jobs:         make(chan job),
		done:         make(chan struct{}),
	}
}
