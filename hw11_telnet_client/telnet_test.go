package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		t.Helper()
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("connection timeout", func(t *testing.T) {
		t.Helper()
		// Используем несуществующий адрес для проверки таймаута
		timeout := 1 * time.Second
		client := NewTelnetClient("192.0.2.1:9999", timeout, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})

		start := time.Now()
		err := client.Connect()
		elapsed := time.Since(start)

		require.Error(t, err)
		// Проверяем, что таймаут сработал примерно в указанное время
		require.Less(t, elapsed, 2*timeout)
	})

	t.Run("multiple messages exchange", func(t *testing.T) {
		t.Helper()
		// Создаем TCP listener на случайном порту для имитации сервера
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		// Горутина клиента - отправляет несколько строк
		go func() {
			defer wg.Done()

			in := bytes.NewBufferString("message1\nmessage2\nmessage3\n")
			out := &bytes.Buffer{}

			client := NewTelnetClient(l.Addr().String(), 10*time.Second, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			// Send() читает весь буфер до EOF и отправляет в сокет
			err := client.Send()
			require.NoError(t, err)
		}()

		// Горутина сервера - принимает и проверяет данные
		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			// Проверяем, что все три строки пришли целиком, в правильном порядке,
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			require.NoError(t, err)
			require.Equal(t, "message1\nmessage2\nmessage3\n", string(buf[:n]))
		}()

		wg.Wait()
	})
}
