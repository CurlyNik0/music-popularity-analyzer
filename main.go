package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/CurlyNik0/music-popularity-analyzer/analyzer"
	"github.com/compilyator/VNTU_GoLang/labs/Lab2Data/musicplays"
)

func main() {
	fmt.Println("==================================================================")
	fmt.Println("Лабораторна робота №3. Варіант 26: Музична популярність")
	fmt.Println("Показник: Кількість прослуховувань виконавця за період (int)")
	fmt.Println("Допустимий діапазон значень: 0 - 10,000,000 прослуховувань")
	fmt.Println("==================================================================")

	reader := bufio.NewReader(os.Stdin)

	var count int
	for {
		fmt.Print("Введіть кількість значень (від 5 до 100): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		parsedCount, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || parsedCount < analyzer.MinSampleCount || parsedCount > analyzer.MaxSampleCount {
			fmt.Println("Помилка: кількість має бути цілим числом від 5 до 100.")
			continue
		}
		count = parsedCount
		break
	}

	var artist string
	for {
		fmt.Print("Введіть назву виконавця (предметний параметр): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		artist = strings.TrimSpace(input)
		if artist == "" {
			fmt.Println("Помилка: назва виконавця не може бути порожньою.")
			continue
		}
		break
	}

	values, err := musicplays.Generate(count, artist)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ПОМИЛКА ГЕНЕРАТОРА:", err)
		os.Exit(1)
	}

	series, err := analyzer.NewArtistPlaySeries(artist, values)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ПОМИЛКА ДАНИХ:", err)
		os.Exit(1)
	}

	report := series.Analyze()
	fmt.Printf("\nВиконавець: %s | Кількість: %d | Сума: %d | Мін: %d | Макс: %d | Середнє: %.2f\n",
		report.Artist, report.Stats.Count, report.Stats.Sum, report.Stats.Min, report.Stats.Max, report.Stats.Average)
}
