package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"wallet/lib/config"
	"wallet/lib/core"
	"wallet/lib/deposits"
	"wallet/lib/rest/internal/middlewares"
	"wallet/lib/utils/logger"
	"wallet/lib/withdraws"

	"github.com/gin-gonic/gin"
	slog_gin "github.com/samber/slog-gin"
)

type server struct {
	engine          *gin.Engine
	httpServer      *http.Server
	depositService  deposits.Service
	withdrawService withdraws.Service
	coreRepoFactory core.RepoFactory
}

func New(
	bindAt int,
	authToken string,
	depositService deposits.Service,
	withdrawService withdraws.Service,
	coreRepoFactory core.RepoFactory,

) *server {
	if config.Env != config.DEV {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(slog_gin.New(logger.Get().WithGroup("gin")))

	s := &server{
		engine:          engine,
		depositService:  depositService,
		withdrawService: withdrawService,
		coreRepoFactory: coreRepoFactory,
	}
	engine.Use(middlewares.Auth(authToken))

	s.registerHandlers()

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", bindAt),
		Handler: engine,
	}

	return s
}

func (s *server) registerHandlers() {
	api := s.engine.Group("/api/v1")
	api.GET("/balance", s.GetBalanceHandler)
	api.GET("/transactions", s.getTransactionsHistoryHandler)
	api.POST("/withdraw", s.createWithdrawHandler)
	api.POST("/deposit", s.createDepositHandler)
}

func (s *server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return s.Stop()
	case err := <-errCh:
		return err
	}
}

// Stop shuts down the server gracefully.
func (s *server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
