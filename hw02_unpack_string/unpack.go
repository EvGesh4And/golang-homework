package hw02unpackstring

import (
	"errors"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(s string) (string, error) {
	// strings.Builder позволяет эффектино хранить и дополнять строку
	// так как s = s + newS приводит к аллокации строки
	var builder strings.Builder

	// Символ для записи
	var symbolRune rune
	// Текущий статус
	var status int

	for _, r := range s {
		switch status {
		// Режим записи.
		case 2:
			// Если текущий символ цифра -- повторяем символ для записи
			if '0' <= r && r <= '9' {
				builder.WriteString(strings.Repeat(string(symbolRune), int(r-'0')))
				// После записи становимся готовыми к идентификации
				status = 0
				continue
			} else {
				// Если символ не цифра, то записываем символ для записи
				builder.WriteRune(symbolRune)
			}
			// И идем в идентификацию для определения нового символа для записи
			fallthrough
		// Режим идентификации.
		case 0:
			switch {
			// Если цифра, то ошибка
			case '0' <= r && r <= '9':
				return "", ErrInvalidString
			// Если имеем дело с возможным экранированием, то идем к следующему символу с проверкой
			case r == '\\':
				status = 1
			// Если не / и не цифра, то символ для записи
			default:
				symbolRune = r
				status = 2
			}

		case 1:
			// Проверка претендента на экранирование
			// Он должен быть либо \, либо цифра
			if r != '\\' && ('0' > r || r > '9') {
				return "", ErrInvalidString
			}
			// Экранированный символ для записи
			symbolRune = r
			status = 2
		}
	}

	// После выхода из цикла
	// Если символ для записи -- записывем его
	if status == 2 {
		builder.WriteRune(symbolRune)
	} else if status == 1 {
		// Если открыто экранирование -- ошибка
		return "", ErrInvalidString
	}
	// Возращаем распакованную строку
	return builder.String(), nil
}
