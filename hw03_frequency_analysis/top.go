package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

const topN = 10

// хранит слово и его частоту использования.
type wordFreq struct {
	word  string // слово из текста
	count int    // количество вхождений слова
}

// Top10 Возвращает:
//   - слайс из максимум 10 самых частых слов (или меньше, если уникальных слов меньше 10)
//   - пустой слайс, если входной текст пустой
func Top10(sourceText string) []string {
	// Обработка пустого текста
	if sourceText == "" {
		return []string{}
	}

	// Разбиваем текст на слова по пробельным символам (пробел, табуляция, перенос строки)
	// strings.Fields автоматически удаляет пустые элементы
	words := strings.Fields(sourceText)

	// Подсчитываем частоту каждого слова в тексте
	frequency := buildFrequencyMap(words)

	// Преобразуем map в слайс для возможности сортировки
	// (map в Go не поддерживает сортировку напрямую)
	wordFreqs := mapToSlice(frequency)

	sortByFrequencyAndLexicographically(wordFreqs)

	// Извлекаем первые N слов из отсортированного списка
	return extractTopWords(wordFreqs, topN)
}

// Подсчитывает частоту встречаемости каждого слова.
// Возвращает:
//   - map, где ключ - слово, значение - количество его вхождений
func buildFrequencyMap(words []string) map[string]int {
	frequency := make(map[string]int)
	// Проходим по всем словам и увеличиваем счётчик для каждого
	for _, word := range words {
		frequency[word]++
	}
	return frequency
}

// Преобразует map частот в слайс структур wordFreq.
func mapToSlice(frequency map[string]int) []wordFreq {
	// Создаём слайс с предварительно выделенной ёмкостью для оптимизации
	wordFreqs := make([]wordFreq, 0, len(frequency))
	// Переносим данные из map в слайс
	for word, count := range frequency {
		wordFreqs = append(wordFreqs, wordFreq{word: word, count: count})
	}
	return wordFreqs
}

// Сортирует слова по частоте (убывание),
// при равной частоте - лексикографически (возрастание).
func sortByFrequencyAndLexicographically(wordFreqs []wordFreq) {
	sort.Slice(wordFreqs, func(i, j int) bool {
		// Если частота одинаковая, сравниваем слова лексикографически
		if wordFreqs[i].count == wordFreqs[j].count {
			return wordFreqs[i].word < wordFreqs[j].word // алфавитный порядок (A-Z)
		}
		// Иначе сортируем по частоте в порядке убывания (от большего к меньшему)
		return wordFreqs[i].count > wordFreqs[j].count
	})
}

// Извлекает первые n слов из отсортированного слайса.
func extractTopWords(wordFreqs []wordFreq, n int) []string {
	// Определяем размер результата: минимум из n и количества уникальных слов
	resultSize := n
	if len(wordFreqs) < n {
		resultSize = len(wordFreqs)
	}

	// Создаём итоговый слайс и заполняем его словами
	result := make([]string, resultSize)
	for i := 0; i < resultSize; i++ {
		result[i] = wordFreqs[i].word
	}

	return result
}
