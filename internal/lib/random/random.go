package random

import (
	"math/rand/v2"
)

func NewRandomString(length int) string {
	// 1. Инициализируем генератор случайных чисел

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	// 2. Создаем слайс нужной длины
	b := make([]rune, length)

	// 3. Заполняем его случайными символами
	for i := range b {
		b[i] = chars[rand.IntN(len(chars))]
	}

	// 4. Превращаем в строку и возвращаем
	return string(b)
}
