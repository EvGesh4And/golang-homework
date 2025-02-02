// Пакет main демонстрирует использование пакета reverse для инвертирования строки
package main

import (
	"fmt" // Импортируем встроенный пакет fmt для форматированного ввода/вывода

	"golang.org/x/example/hello/reverse" // Импортируем пакет reverse для работы с инверсией строк
)

func main() {
	s := "Hello, OTUS!"            // Исходная строка, которую нужно перевернуть
	fmt.Println(reverse.String(s)) // Выводим результат реверса строки в консоль
}
