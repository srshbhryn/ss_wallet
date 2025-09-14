package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wallet/lib/config"
	"wallet/lib/utils"
	"wallet/lib/utils/db"
	"wallet/lib/utils/logger"
	"wallet/lib/withdraws"
	"wallet/lib/withdraws/enums"
	"wallet/lib/withdraws/integrations"
	"wallet/lib/withdraws/repository"
)

// WorkerConfig defines worker pool tuning
type WorkerConfig struct {
	Concurrency int
	BackOff     time.Duration
	RetryCount  int
}

func (w *WorkerConfig) FillDefaults() {
	if w.Concurrency == 0 {
		w.Concurrency = 5
	}
	if w.BackOff == 0 {
		w.BackOff = 2 * time.Second
	}
	if w.RetryCount == 0 {
		w.RetryCount = 3
	}
}

// Config defines all settings for withdraw worker
type Config struct {
	DbDsn         string
	Prefix        string
	SleepInterval time.Duration
	Bank          string
	BankConfig    map[string]any
	Worker        WorkerConfig
}

func (c *Config) FillDefaults() {
	if c.SleepInterval == 0 {
		c.SleepInterval = 10 * time.Second
	}
	c.Worker.FillDefaults()
}

func main() {
	// load config
	conf := Config{}
	config.Load(&conf)
	conf.FillDefaults()

	// init DB
	db, err := db.Connect(conf.DbDsn)
	if err != nil {
		logger.Get().Error("failed to connect DB", "err", utils.Stringify(err))
		os.Exit(1)
	}

	// init repo
	withdrawRepoFactory := repository.NewFactory(db) //TODO shouldn't I use service instead of repo????

	// init bank client
	client, err := integrations.NewBankClient(enums.BankType(conf.Bank), conf.BankConfig)
	if err != nil {
		logger.Get().Error("failed to init bank client", "err", utils.Stringify(err))
		os.Exit(1)
	}

	// init worker with updateStatus callback
	wrk := withdraws.NewWorker(
		func(ctx context.Context, wd *withdraws.Withdrawal) error {
			return withdrawRepoFactory.New(nil).Update(ctx, wd)
		},
		conf.Worker.Concurrency,
		conf.Worker.BackOff,
		conf.Worker.RetryCount,
		client,
	)

	// graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Get().Info("shutting down withdraw worker")
		cancel()
	}()

	// start worker pool
	wrk.Run(ctx)
	defer wrk.Stop()

	logger.Get().Info("withdraw worker started",
		"prefix", conf.Prefix,
		"bank", conf.Bank,
		"concurrency", conf.Worker.Concurrency,
	)

	// loop
	for {
		select {
		case <-ctx.Done():
			logger.Get().Info("exiting loop")
			return
		default:
			err := processWithdraws(ctx, wrk, withdrawRepoFactory.New(nil), conf.Prefix, enums.BankType(conf.Bank))
			if err != nil {
				logger.Get().Error("error processing withdraws", "err", utils.Stringify(err))
			}
			time.Sleep(conf.SleepInterval)
		}
	}
}

func processWithdraws(
	ctx context.Context,
	wrk withdraws.Worker,
	repo repository.Repo,
	prefix string,
	bank enums.BankType,
) error {
	withdraws, err := repo.GetUnFinishedWithdraws(ctx, prefix, bank)
	if err != nil {
		return err
	}

	for _, wd := range withdraws {
		log := logger.Get().With("withdraw_id", wd.ID, "status", wd.Status)

		switch wd.Status {
		case enums.NEW:
			if err := wrk.SendToBank(ctx, &wd); err != nil {
				log.Error("failed to enqueue withdrawal send", "err", utils.Stringify(err))
			} else {
				log.Info("enqueued withdrawal send")
			}

		case enums.SENT:
			if err := wrk.GetStatus(ctx, &wd); err != nil {
				log.Error("failed to enqueue withdrawal status check", "err", utils.Stringify(err))
			} else {
				log.Info("enqueued withdrawal status check")
			}

		default:
			log.Info("skipping withdrawal with status", "status", wd.Status)
		}
	}

	return nil
}
