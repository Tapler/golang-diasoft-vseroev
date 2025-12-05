package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_calendar/internal/app"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_calendar/internal/storage/memory"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configFile)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logg := logger.New(config.Logger.Level)

	storage := memorystorage.New()
	calendar := app.New(logg, storage)

	server := internalhttp.NewServer(logg, calendar, config.HTTP.Host, config.HTTP.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer shutdownCancel()

		if err := server.Stop(shutdownCtx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
