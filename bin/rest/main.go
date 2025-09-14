package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"wallet/lib/config"
	"wallet/lib/core"
	"wallet/lib/deposits"
	deposits_repository "wallet/lib/deposits/repository"
	"wallet/lib/rest"
	"wallet/lib/utils/db"
	"wallet/lib/utils/logger"
	"wallet/lib/withdraws"
	withdraws_repository "wallet/lib/withdraws/repository"
)

type Config struct {
	DbDsn     string
	AuthToken string
	BindAt    int           // port number (e.g. 8080)
	GraceTime time.Duration // graceful shutdown timeout
}

func (c *Config) FillDefaults() {
	if c.BindAt == 0 {
		c.BindAt = 8080
	}
	if c.GraceTime == time.Duration(0) {
		c.GraceTime = 5 * time.Second
	}
}

func main() {
	// Load configuration
	conf := Config{}
	config.Load(&conf)

	// Connect to database
	db, err := db.Connect(conf.DbDsn)
	if err != nil {
		logger.Get().Error("failed to connect DB", "err", err)
		os.Exit(1)
	}

	// Build repository factories
	coreRepoFactory := core.NewFactory(db)
	depositRepoFactory := deposits_repository.NewFactory(db)
	withdrawRepoFactory := withdraws_repository.NewFactory(db)

	// Build services
	depositService := deposits.New(coreRepoFactory, depositRepoFactory)
	withdrawService := withdraws.NewService(coreRepoFactory, withdrawRepoFactory)

	// Create HTTP server
	server := rest.New(conf.BindAt, conf.AuthToken, depositService, withdrawService, coreRepoFactory)

	// Handle signals for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Get().Info("REST server starting", "port", conf.BindAt)

	// Run server
	if err := server.Run(ctx); err != nil {
		log.Fatalf("server stopped: %v", err)
	}

	logger.Get().Info("REST server exited gracefully")
}
