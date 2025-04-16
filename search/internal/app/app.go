package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/nzb3/closer"
	"golang.org/x/sync/errgroup"

	"github.com/nzb3/diploma/search/internal/configurator"
)

// App is a structure that configure and run application
type App struct {
	serviceProvider *ServiceProvider
	server          *http.Server
}

// NewApp creates exemplar of App
func NewApp(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

// Start starts the App
func (a *App) Start(ctx context.Context) error {
	const op = "app.Start"
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		slog.Info("Starting server")
		a.server.BaseContext = func(_ net.Listener) context.Context {
			return ctx
		}
		return a.server.ListenAndServe()
	})

	return fmt.Errorf("%s: %w", op, eg.Wait())
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

func (a *App) initConfig(_ context.Context) error {
	const op = "app.initConfig"
	err := configurator.LoadConfig("configs", ".env", "env")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
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
