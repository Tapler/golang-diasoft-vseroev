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
	"github.com/google/uuid"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/scheduler_config.toml", "Path to configuration file")
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
	producer, err := kafka.NewProducer(config.Kafka.Brokers)
	if err != nil {
		logg.Error(fmt.Sprintf("Failed to create Kafka producer: %v", err))
		panic("failed to create Kafka producer: " + err.Error())
	}
	logg.Info("Successfully connected to Kafka")

	scanInterval, err := time.ParseDuration(config.Scheduler.ScanInterval)
	if err != nil {
		logg.Error(fmt.Sprintf("Invalid scan interval: %v", err))
		panic("invalid scan interval: " + err.Error())
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer shutdownCancel()

		if err := producer.Close(); err != nil {
			logg.Error("failed to close producer: " + err.Error())
		}

		if err := stor.Close(shutdownCtx); err != nil {
			logg.Error("failed to close storage: " + err.Error())
		} else {
			logg.Info("Storage closed successfully")
		}
	}()

	logg.Info("Scheduler is running...")

	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logg.Info("Shutting down scheduler...")
			return
		case <-ticker.C:
			if err := processEvents(ctx, logg, stor, producer, config.Kafka.Topic); err != nil {
				logg.Error(fmt.Sprintf("Error processing events: %v", err))
			}

			if err := cleanupOldEvents(ctx, logg, stor); err != nil {
				logg.Error(fmt.Sprintf("Error cleaning up old events: %v", err))
			}
		}
	}
}

func processEvents(
	ctx context.Context,
	logg *logger.Logger,
	stor storage.Storage,
	producer queue.Producer,
	topic string,
) error {
	now := time.Now().UTC()
	logg.Debug(fmt.Sprintf("Checking for events to notify at %s", now.Format(time.RFC3339)))

	events, err := stor.GetEventsToNotify(ctx, now)
	if err != nil {
		return fmt.Errorf("failed to get events to notify: %w", err)
	}

	logg.Debug(fmt.Sprintf("Found %d events to check", len(events)))

	metrics.EventsScannedTotal.Inc()

	if len(events) == 0 {
		return nil
	}

	logg.Info(fmt.Sprintf("Found %d events to notify", len(events)))

	for _, event := range events {
		notification := storage.Notification{
			ID:        uuid.New().String(),
			EventID:   event.ID,
			Title:     event.Title,
			EventTime: event.StartTime,
			UserID:    event.UserID,
		}

		data, err := json.Marshal(notification)
		if err != nil {
			logg.Error(fmt.Sprintf("Failed to marshal notification: %v", err))
			metrics.NotificationsSendErrors.Inc()
			continue
		}

		if err := producer.SendMessage(ctx, topic, []byte(notification.EventID), data); err != nil {
			logg.Error(fmt.Sprintf("Failed to send notification: %v", err))
			metrics.NotificationsSendErrors.Inc()
			continue
		}

		metrics.NotificationsSentTotal.Inc()
		logg.Info(fmt.Sprintf("Sent notification for event %s", event.ID))
	}

	return nil
}

func cleanupOldEvents(ctx context.Context, logg *logger.Logger, stor storage.Storage) error {
	oneYearAgo := time.Now().UTC().AddDate(-1, 0, 0)

	if err := stor.DeleteOldEvents(ctx, oneYearAgo); err != nil {
		return fmt.Errorf("failed to delete old events: %w", err)
	}

	metrics.OldEventsDeletedTotal.Inc()
	logg.Info("Cleaned up old events")
	return nil
}
