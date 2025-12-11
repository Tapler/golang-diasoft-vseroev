package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/app"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/logger"
	internalhttp "github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/server/http"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/storage"
	memorystorage "github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/storage/sql"
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

	// Создание storage на основе конфигурации
	var stor storage.Storage
	switch config.Database.Type {
	case "sql":
		logg.Info(fmt.Sprintf("Initializing SQL storage with DSN: %s", config.Database.DSN))
		sqlStor, err := sqlstorage.New(config.Database.DSN)
		if err != nil {
			logg.Error(fmt.Sprintf("Failed to create SQL storage: %v", err))
			panic("failed to create SQL storage: " + err.Error())
		}

		// Подключение к БД
		connCtx := context.Background()
		if err := sqlStor.Connect(connCtx); err != nil {
			logg.Error(fmt.Sprintf("Failed to connect to database: %v", err))
			panic("failed to connect to database: " + err.Error())
		}
		logg.Info("Successfully connected to SQL database")
		stor = sqlStor
	case "memory":
		logg.Info("Initializing in-memory storage")
		stor = memorystorage.New()
	default:
		panic(fmt.Sprintf("unknown storage type: %s", config.Database.Type))
	}

	calendar := app.New(logg, stor)

	server := internalhttp.NewServer(logg, calendar, config.HTTP.Host, config.HTTP.Port)

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer shutdownCancel()

		// Останавливаем HTTP сервер
		if err := server.Stop(shutdownCtx); err != nil {
			logg.Error("failed to stop http server: " + err.Error())
		}

		// Закрываем соединение с БД (если это SQL storage)
		if sqlStor, ok := stor.(*sqlstorage.Storage); ok {
			if err := sqlStor.Close(shutdownCtx); err != nil {
				logg.Error("failed to close storage: " + err.Error())
			} else {
				logg.Info("Storage closed successfully")
			}
		}
	}()

	logg.Info("calendar is running...")

	if err := server.Start(ctx); err != nil {
		logg.Error("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
