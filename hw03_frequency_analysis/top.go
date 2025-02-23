package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(inputS string) []string {
	// Разбитие входного текста на "слова"
	sliceText := strings.Fields(inputS)

	// Инициализация карты для хранения "слов" с их числом вхождений в текст
	mapWords := make(map[string]int)

	// Проход по тексту с подсчетом количества слов
	for _, v := range sliceText {
		mapWords[v]++
	}

	// Инициализация среза для хранения структуры -- слово и число вхождений
	sliceWords := make([]struct {
		word  string
		count int
	}, 0, len(mapWords))

	// Перенос данных из mapWords в sliceWords
	for key, cont := range mapWords {
		sliceWords = append(sliceWords, struct {
			word  string
			count int
		}{key, cont})
	}

	// Сортировка по структуре:
	// 1. обратная соритровка по числу вхождений
	// 2. если число вхождений равно, то сравнение слов
	sort.Slice(sliceWords, func(i, j int) bool {
		if sliceWords[i].count != sliceWords[j].count {
			return sliceWords[i].count > sliceWords[j].count
		}
		return sliceWords[i].word < sliceWords[j].word
	})

	// Объявление среза для формирование выхода
	var result []string

	// Переносим первые 10-ть слов в результирующий срез
	// ! Не выдаем sliceWords[:10], так как будет храниться весь sliceWords
	for i := 0; i < len(sliceWords) && i < 10; i++ {
		result = append(result, sliceWords[i].word)
	}

	return result
}
