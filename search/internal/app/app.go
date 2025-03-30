package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/nzb3/closer"
	"golang.org/x/sync/errgroup"

	"github.com/nzb3/diploma/search/internal/configurator"
)

type App struct {
	serviceProvider *serviceProvider
	server          *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) Start(ctx context.Context) error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		slog.Info("Starting server")
		return a.server.ListenAndServe()
	})

	return eg.Wait()
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initConfig,
		a.initServiceProvider,
		a.initLogger,
		a.initServer,
	}

	for _, init := range inits {
		if err := init(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initConfig(ctx context.Context) error {
	err := configurator.LoadConfig("configs", ".env", "env")
	if err != nil {
		return err
	}
	return nil
}

func (a *App) initServiceProvider(ctx context.Context) error {
	a.serviceProvider = NewServiceProvider()
	return nil
}

func (a *App) initLogger(ctx context.Context) error {
	a.serviceProvider.Logger(ctx)
	return nil
}

func (a *App) initServer(ctx context.Context) error {
	a.server = a.serviceProvider.Server(ctx)
	return nil
}
