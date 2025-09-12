package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wallet/lib/config"
	"wallet/lib/core"
	"wallet/lib/deposits"
	"wallet/lib/deposits/repository"
	"wallet/lib/utils"
	"wallet/lib/utils/db"
	"wallet/lib/utils/logger"
)

type Config struct {
	DbDsn         string
	Prefix        string
	SleepInterval time.Duration
}

func (c *Config) FillDefaults() {
	if c.SleepInterval == time.Duration(0) {
		c.SleepInterval = 10 * time.Second
	}
}

func main() {
	conf := Config{}
	config.Load(&conf)
	db, err := db.Connect(conf.DbDsn)
	if err != nil {
		logger.Get().Error("failed to connect DB:", "err", utils.Stringify(err))
		os.Exit(1)
	}
	coreFactory := core.NewFactory(db)
	depFactory := repository.NewFactory(db)
	service := deposits.New(coreFactory, depFactory)

	// Graceful shutdown setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		logger.Get().Info("shutting down")
		cancel()
	}()

	logger.Get().Info("deposit worker started", "prefix", conf.Prefix)

	// Infinite loop
	for {
		select {
		case <-ctx.Done():
			logger.Get().Info("exiting loop")
			return
		default:
			err := processDeposits(ctx, service, conf.Prefix)
			if err != nil {
				logger.Get().Error("error processing deposits", "err", utils.Stringify(err))
			}
			time.Sleep(conf.SleepInterval)
		}
	}
}

func processDeposits(ctx context.Context, service deposits.Service, prefix string) error {
	depositsList, err := service.GetApplicableDeposits(ctx, prefix)
	if err != nil {
		return err
	}

	for _, d := range depositsList {
		logger := logger.Get().With("deposit_id", d.ID)
		err := service.Apply(ctx, &d)
		if err != nil {
			logger.Error("failed to apply deposit", "err", utils.Stringify(err))
		} else {
			logger.Info("applied deposit")
		}
	}

	return nil
}
