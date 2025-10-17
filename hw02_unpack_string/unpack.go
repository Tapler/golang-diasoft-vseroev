package hw02unpackstring

import (
	"errors"
	"strings"
)

// ErrInvalidString возвращается при попытке распаковать некорректную строку.
// Некорректными считаются строки, начинающиеся с цифры или содержащие числа (несколько цифр подряд).
var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	// Пустая строка - валидный случай, возвращаем пустую строку
	if s == "" {
		return "", nil
	}

	// Преобразуем строку в массив рун для корректной работы с Unicode
	runes := []rune(s)

	var result strings.Builder

	// Проходим по всем рунам строки
	for i := 0; i < len(runes); i++ {
		if isDigit(runes[i]) {
			// Текущий символ - цифра, обрабатываем как счётчик повторений
			if err := processDigit(runes, i, &result); err != nil {
				return "", err
			}
		} else {
			// Текущий символ - не цифра, обрабатываем как обычный символ
			// processCharacter может вернуть i+1, если следующий символ - цифра
			i = processCharacter(runes, i, &result)
		}
	}

	// Возвращаем собранную строку
	return result.String(), nil
}

// isDigit проверяет, является ли руна цифрой.
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// processDigit обрабатывает цифру - повторяет предыдущий символ указанное количество раз.
// result - указатель на strings.Builder, чтобы изменять оригинальный объект, а не его копию.
// Возвращает ошибку, если цифра стоит в начале строки или после другой цифры.
func processDigit(runes []rune, i int, result *strings.Builder) error {
	// Проверяем, что цифра не в начале строки
	if i == 0 {
		return ErrInvalidString
	}

	// Получаем предыдущий символ
	prevRune := runes[i-1]

	// Проверяем, что предыдущий символ не цифра
	if isDigit(prevRune) {
		return ErrInvalidString
	}

	// Преобразуем руну-цифру в число
	count := int(runes[i] - '0')

	// Повторяем предыдущий символ count раз (минус 1, так как он уже добавлен)
	if count > 1 {
		// strings.Repeat создаёт строку из count-1 повторений символа prevRune
		result.WriteString(strings.Repeat(string(prevRune), count-1))
	}
	return nil
}

// processCharacter обрабатывает обычный символ (не цифру).
// result - указатель на strings.Builder, чтобы изменять оригинальный объект, а не его копию.
// Возвращает новую позицию для итератора (i или i+1).
func processCharacter(runes []rune, i int, result *strings.Builder) int {
	currentRune := runes[i]

	// Проверяем, есть ли следующий символ и является ли он цифрой
	// Если да, то текущий символ нужно повторить указанное количество раз
	if i+1 < len(runes) && isDigit(runes[i+1]) {
		// Преобразуем следующую руну-цифру в число
		count := int(runes[i+1] - '0')

		// Добавляем текущий символ count раз
		// strings.Repeat создаёт строку из count повторений символа currentRune
		if count > 0 {
			result.WriteString(strings.Repeat(string(currentRune), count))
		}

		// Возвращаем i+1, чтобы пропустить следующий символ (цифру)
		// Она уже обработана как счётчик повторений
		return i + 1
	}

	// Следующий символ не цифра или его нет
	// Добавляем текущий символ один раз как есть
	result.WriteRune(currentRune)
	return i
}
