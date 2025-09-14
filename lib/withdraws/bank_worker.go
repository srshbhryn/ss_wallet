package withdraws

import (
	"context"
	"errors"
	"time"

	"wallet/lib/withdraws/enums"
	"wallet/lib/withdraws/integrations"
)

type jobType int

const (
	jobSend jobType = iota
	jobCheck
)

type job struct {
	ctx context.Context
	wd  *Withdrawal
	t   jobType
}

type worker struct {
	service     Service
	concurrency int
	backOff     time.Duration
	retryCount  int
	client      integrations.BankClient

	jobs chan job
	done chan struct{}
}

func (w *worker) Run(ctx context.Context) {
	for i := 0; i < w.concurrency; i++ {
		go func() {
			for {
				select {
				case j := <-w.jobs:
					switch j.t {
					case jobSend:
						_ = w.doSend(j.ctx, j.wd)
					case jobCheck:
						_ = w.doCheck(j.ctx, j.wd)
					}
				case <-ctx.Done():
					return
				case <-w.done:
					return
				}
			}
		}()
	}
}

func (w *worker) Stop() {
	close(w.done)
}

func (w *worker) doWithRetry(ctx context.Context, fn func() error) error {
	var err error
	for i := 0; i < w.retryCount; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return ctx.Err()
		}
		select {
		case <-time.After(w.backOff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return err
}

func (w *worker) doSend(ctx context.Context, wd *Withdrawal) error {
	var status enums.PayoutStatus
	err := w.doWithRetry(ctx, func() error {
		var e error
		status, e = w.client.Send(wd.Iban, wd.Amount, wd.ID.String())
		return e
	})
	if errors.Is(err, integrations.ErrDuplicatePayout) {
		w.service.MarkAsSent(ctx, wd)
		return nil
	}
	if err != nil {
		return err
	}
	if status == enums.SUCCESS {
		return w.service.Complete(ctx, wd)
	}
	if status == enums.FAILED {
		return w.service.Reverse(ctx, wd)
	}
	if status == enums.SENT {
		return w.service.MarkAsSent(ctx, wd)
	}
	return nil
}

func (w *worker) doCheck(ctx context.Context, wd *Withdrawal) error {
	var status enums.PayoutStatus
	err := w.doWithRetry(ctx, func() error {
		var e error
		status, e = w.client.GetStatus(wd.ID.String())
		return e
	})
	if err != nil {
		return err
	}
	if status == enums.SUCCESS {
		return w.service.Complete(ctx, wd)
	}
	if status == enums.FAILED {
		return w.service.Reverse(ctx, wd)
	}
	return nil
}

func (w *worker) SendToBank(ctx context.Context, wd *Withdrawal) error {
	select {
	case w.jobs <- job{ctx: ctx, wd: wd, t: jobSend}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *worker) GetStatus(ctx context.Context, wd *Withdrawal) error {
	select {
	case w.jobs <- job{ctx: ctx, wd: wd, t: jobCheck}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
