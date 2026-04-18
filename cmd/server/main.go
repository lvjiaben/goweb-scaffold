package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lvjiaben/goweb-core/app"
	"github.com/lvjiaben/goweb-scaffold/internal/bootstrap"
	"github.com/lvjiaben/goweb-scaffold/internal/gen"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	runtime, err := bootstrap.NewRuntime(*configPath)
	if err != nil {
		log.Fatalf("new runtime: %v", err)
	}

	if err := bootstrap.ApplyMigrations(runtime.DB, "migrations"); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	if err := gen.RegisterModules(runtime); err != nil {
		log.Fatalf("register modules: %v", err)
	}

	application := app.New(runtime.Config.App.Name, runtime.Config.App.Addr, runtime.Handler(), runtime.Logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- application.Run()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server stopped: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := application.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("shutdown server: %v", err)
		}
	}
}
