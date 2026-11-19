package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/compilyator/VNTU_GoLang/labs/Lab2Data/musicplays"
)

func main() {
	// 1. Інформаційне повідомлення про варіант
	fmt.Println("==================================================================")
	fmt.Println("Лабораторна робота №2. Варіант 26: Музична популярність")
	fmt.Println("Показник: Кількість прослуховувань виконавця за період (int)")
	fmt.Println("Допустимий діапазон значень: 0 - 10,000,000 прослуховувань")
	fmt.Println("Категорії:")
	fmt.Println("  1. Низька:      0 - 9,999")
	fmt.Println("  2. Середня:     10,000 - 199,999")
	fmt.Println("  3. Висока:      200,000 - 1,999,999")
	fmt.Println("  4. Дуже висока: від 2,000,000")
	fmt.Println("==================================================================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// 2. Зчитування та валідація кількості (count: 5..100)
	var count int
	for {
		fmt.Print("Введіть кількість значень (від 5 до 100): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Помилка читання введення. Спробуйте ще раз.")
			continue
		}
		input = strings.TrimSpace(input)

		parsedCount, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Помилка: Введено нечислове значення або неправильний формат!")
			continue
		}

		if parsedCount < 5 || parsedCount > 100 {
			fmt.Printf("Помилка: Кількість має бути в межах від 5 до 100! (Ви ввели: %d)\n", parsedCount)
			continue
		}

		count = parsedCount
		break
	}

	// 3. Зчитування та валідація предметного параметра (artist)
	var artist string
	for {
		fmt.Print("Введіть назву виконавця (предметний параметр): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Помилка читання введення. Спробуйте ще раз.")
			continue
		}
		artist = strings.TrimSpace(input)

		if artist == "" {
			fmt.Println("Помилка: Назва виконавця не може бути порожньою!")
			continue
		}
		break
	}

	fmt.Printf("\nЗапит даних для виконавця: \"%s\", кількість елементів: %d...\n\n", artist, count)

	// 4. Виклик генератора пакета
	values, err := musicplays.Generate(count, artist)

	// 5. Явна обробка помилки генератора
	if err != nil {
		fmt.Println("ПОМИЛКА ГЕНЕРАТОРА:", err)
		fmt.Println("Аналіз зупинено через помилку отримання даних.")
		return
	}

	// 6. Перевірка отриманого зрізу на цілісність та діапазон
	if len(values) != count {
		fmt.Printf("ПОМИЛКА ЦІЛІСНОСТІ: Очікувалося %d значень, отримано %d\n", count, len(values))
		return
	}

	for i, val := range values {
		if val < 0 || val > 10000000 {
			fmt.Printf("ПОМИЛКА ДІАПАЗОНУ: Значення за індексом %d (%d) виходить за межі [0, 10000000]\n", i, val)
			return
		}
	}

	// 7. Виведення отриманого зрізу
	fmt.Println("--- Отримані дані ---")
	fmt.Printf("Виконавець: %s\n", artist)
	fmt.Println("Згенеровані прослуховування:")
	for i, val := range values {
		fmt.Printf("  [%2d] %d прослуховувань\n", i+1, val)
	}
	fmt.Println()

	// 8. Обчислення суми, мінімуму, максимуму та середнього
	sum := 0
	minVal := values[0]
	maxVal := values[0]

	for _, val := range values {
		sum += val
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	avg := float64(sum) / float64(count)

	fmt.Println("--- Статистичний аналіз ---")
	fmt.Printf("Загальна кількість: %d\n", count)
	fmt.Printf("Сума прослуховувань: %d\n", sum)
	fmt.Printf("Мінімум: %d прослуховувань\n", minVal)
	fmt.Printf("Максимум: %d прослуховувань\n", maxVal)
	fmt.Printf("Середнє арифметичне: %.2f прослуховувань\n\n", avg)

	aboveAvgCount := 0.0
	diffSum := 0.0

	for _, val := range values {
		if float64(val) > avg {
			aboveAvgCount++
			diffSum += float64(val) - avg
		}
	}

	fmt.Printf("Кількість значень, більших за середнє: %f\n", aboveAvgCount)
	fmt.Printf("Сума різниць середніх значень: %.2f\n\n", diffSum)

	// 9. Категоризація та накопичення частот у map
	categoryCounts := map[string]int{
		"Низька":      0,
		"Середня":     0,
		"Висока":      0,
		"Дуже висока": 0,
	}

	for _, val := range values {
		switch {
		case val <= 9999:
			categoryCounts["Низька"]++
		case val <= 199999:
			categoryCounts["Середня"]++
		case val <= 1999999:
			categoryCounts["Висока"]++
		default:
			categoryCounts["Дуже висока"]++
		}
	}

	// 10. Виведення категорій у СУВОРОМУ порядку умови
	categoriesOrder := []string{"Низька", "Середня", "Висока", "Дуже висока"}
	fmt.Println("--- Розподіл за категоріями ---")
	for _, cat := range categoriesOrder {
		fmt.Printf("  %-12s: %d значень\n", cat, categoryCounts[cat])
	}
	fmt.Println()

	// 11. Пошук повторів СПОСОБОМ 1 (Вкладені цикли - O(n^2))
	var repeatsMethod1 []int
	for i := 0; i < count; i++ {
		// Перевіряємо, чи вже додавали цей елемент до списку повторів
		alreadyAdded := false
		for _, added := range repeatsMethod1 {
			if added == values[i] {
				alreadyAdded = true
				break
			}
		}
		if alreadyAdded {
			continue
		}

		// Шукаємо збіг далі по масиву
		isRepeat := false
		for j := i + 1; j < count; j++ {
			if values[i] == values[j] {
				isRepeat = true
				break
			}
		}

		if isRepeat {
			repeatsMethod1 = append(repeatsMethod1, values[i])
		}
	}

	// 12. Пошук повторів СПОСОБОМ 2 (Частотна map - O(n))
	freqMap := make(map[int]int)
	for _, val := range values {
		freqMap[val]++
	}

	var repeatsMethod2 []int
	// Зберігаємо порядок першої появи
	for _, val := range values {
		if freqMap[val] > 1 {
			alreadyAdded := false
			for _, added := range repeatsMethod2 {
				if added == val {
					alreadyAdded = true
					break
				}
			}
			if !alreadyAdded {
				repeatsMethod2 = append(repeatsMethod2, val)
			}
		}
	}

	// 13. Зіставлення результатів обох способів та виведення
	fmt.Println("--- Аналіз повторів ---")

	// Перевірка ідентичності двох способів
	sameResults := len(repeatsMethod1) == len(repeatsMethod2)
	if sameResults {
		for i := range repeatsMethod1 {
			if repeatsMethod1[i] != repeatsMethod2[i] {
				sameResults = false
				break
			}
		}
	}

	if !sameResults {
		fmt.Println("УВАГА: Виявлено розбіжність між результатами O(n^2) та O(n)!")
	} else {
		fmt.Println("Підтвердження: Висновки алгоритмів O(n^2) та O(n) повністю збігаються.")
	}

	if len(repeatsMethod1) == 0 {
		fmt.Println("Повторюваних значень у вибірці НЕ виявлено.")
	} else {
		fmt.Println("Знайдені повтори (у порядку першої появи):")
		for _, val := range repeatsMethod1 {
			fmt.Printf("  Значення %d зустрічається %d раз(и)\n", val, freqMap[val])
		}
	}
	fmt.Println("==================================================================")
}
