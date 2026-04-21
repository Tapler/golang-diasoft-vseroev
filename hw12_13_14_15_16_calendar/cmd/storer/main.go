package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/metrics"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/queue"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/queue/kafka"
	"github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/storage"
	sqlstorage "github.com/Tapler/golang-diasoft-vseroev/hw12_13_14_15_16_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/storer_config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	config, err := NewConfig(configFile)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logg := logger.New(config.Logger.Level)

	logg.Info(fmt.Sprintf("Initializing SQL storage with DSN: %s", config.Database.DSN))
	stor, err := sqlstorage.New(config.Database.DSN)
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to create SQL storage: %v", err))
		panic("failed to create SQL storage: " + err.Error())
	}

	connCtx := context.Background()
	if err := stor.Connect(connCtx); err != nil {
		logg.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		panic("failed to connect to database: " + err.Error())
	}
	logg.Info("Successfully connected to SQL database")

	logg.Info(fmt.Sprintf("Connecting to Kafka brokers: %v", config.Kafka.Brokers))
	consumer := kafka.NewConsumer(config.Kafka.Brokers, config.Kafka.GroupID)
	logg.Info("Kafka consumer created")

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer shutdownCancel()

		if err := consumer.Close(); err != nil {
			logg.Error("failed to close consumer: " + err.Error())
		}

		if err := stor.Close(shutdownCtx); err != nil {
			logg.Error("failed to close storage: " + err.Error())
		} else {
			logg.Info("Storage closed successfully")
		}
	}()

	logg.Info("Storer is running...")

	handler := createMessageHandler(logg, stor)

	if err := consumer.Consume(ctx, []string{config.Kafka.Topic}, handler); err != nil {
		logg.Error(fmt.Sprintf("Consumer error: %v", err))
	}
}

func createMessageHandler(logg *logger.Logger, stor storage.Storage) queue.MessageHandler {
	return func(ctx context.Context, message []byte) error {
		var notification storage.Notification
		if err := json.Unmarshal(message, &notification); err != nil {
			metrics.NotificationsSaveErrors.Inc()
			return fmt.Errorf("failed to unmarshal notification: %w", err)
		}

		metrics.NotificationsReceivedTotal.Inc()
		logg.Info(fmt.Sprintf("Received notification for event %s", notification.EventID))

		if err := stor.CreateNotification(ctx, notification); err != nil {
			metrics.NotificationsSaveErrors.Inc()
			return fmt.Errorf("failed to save notification: %w", err)
		}

		metrics.NotificationsSavedTotal.Inc()
		logg.Info(fmt.Sprintf("Saved notification %s to database", notification.ID))
		return nil
	}
}
