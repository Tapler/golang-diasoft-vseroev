package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	// Потоковое чтение для минимального использования памяти
	scanner := bufio.NewScanner(r)
	// Предварительное вычисление суффикса домена
	domainSuffix := "." + domain

	for scanner.Scan() {
		// Получаем байты напрямую из буфера сканера (без копирования)
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Unmarshaling только нужного поля Email
		var user struct {
			Email string
		}

		if err := json.Unmarshal(line, &user); err != nil {
			return nil, fmt.Errorf("unmarshal error: %w", err)
		}

		// Завершаем для пустых email
		if user.Email == "" {
			continue
		}

		// Быстрая проверка наличия домена (strings.Contains быстрее regexp)
		if !strings.Contains(user.Email, domainSuffix) {
			continue
		}

		// Извлечение домена через IndexByte (быстрее чем SplitN)
		atIndex := strings.IndexByte(user.Email, '@')
		if atIndex == -1 {
			continue
		}

		emailDomain := strings.ToLower(user.Email[atIndex+1:])
		result[emailDomain]++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return result, nil
}
