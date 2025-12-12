package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Парсинг аргументов командной строки
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	// Проверка количества аргументов
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--timeout=<duration>] <host> <port>\n", os.Args[0])
		os.Exit(1)
	}

	host := flag.Arg(0)
	port := flag.Arg(1)
	address := net.JoinHostPort(host, port)

	// Создание клиента
	client := NewTelnetClient(address, *timeout, os.Stdin, os.Stdout)

	// Подключение к серверу
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...Failed to connect to %s: %v\n", address, err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	// Создание контекста с обработкой сигналов
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Канал для синхронизации завершения горутин
	done := make(chan struct{})

	// Горутина для отправки данных (STDIN -> сокет)
	go func() {
		defer close(done)
		if err := client.Send(); err != nil {
			fmt.Fprintf(os.Stderr, "...Error sending data: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "...EOF")
		}
	}()

	// Горутина для приема данных (сокет -> STDOUT)
	go func() {
		if err := client.Receive(); err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n")
			} else {
				fmt.Fprintf(os.Stderr, "...Error receiving data: %v\n", err)
			}
		}
		cancel() // Отменяем контекст при закрытии соединения сервером
	}()

	// Ожидание завершения
	select {
	case <-ctx.Done():
		// Получен сигнал или сервер закрыл соединение
		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			// Принудительное завершение если горутина зависла
		}
	case <-done:
		// Завершение ввода (Ctrl+D)
	}
}
