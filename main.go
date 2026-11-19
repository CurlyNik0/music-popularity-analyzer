package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/CurlyNik0/music-popularity-analyzer/analyzer"
	"github.com/CurlyNik0/music-popularity-analyzer/presentation"
	"github.com/compilyator/VNTU_GoLang/labs/Lab2Data/musicplays"
)

const (
	appName     = "music-popularity-analyzer"
	appVersion  = "1.0.0"
	variantInfo = "Варіант 26: Музична популярність (прослуховування виконавця, int)"
)

const (
	exitOK       = 0
	exitDataErr  = 1
	exitUsageErr = 2
)

// PlaySource абстрагує джерело генерації даних для зручного тестування CLI.
type PlaySource interface {
	Generate(count int, artist string) ([]int, error)
	GenerateWithSeed(count int, artist string, seed int64) ([]int, error)
}

type musicPlaysSource struct{}

func (musicPlaysSource) Generate(count int, artist string) ([]int, error) {
	return musicplays.Generate(count, artist)
}

func (musicPlaysSource) GenerateWithSeed(count int, artist string, seed int64) ([]int, error) {
	return musicplays.GenerateWithSeed(count, artist, seed)
}

type int64OptionalFlag struct {
	set   bool
	value int64
}

func (f *int64OptionalFlag) String() string {
	if !f.set {
		return ""
	}
	return fmt.Sprintf("%d", f.value)
}

func (f *int64OptionalFlag) Set(raw string) error {
	var parsed int64
	if _, err := fmt.Sscanf(strings.TrimSpace(raw), "%d", &parsed); err != nil {
		return fmt.Errorf("некоректне ціле число для --seed: %q", raw)
	}
	// Перевіряємо, що рядок не містить зайвих символів після числа
	if fmt.Sprintf("%d", parsed) != strings.TrimSpace(raw) {
		return fmt.Errorf("некоректне ціле число для --seed: %q", raw)
	}
	f.value = parsed
	f.set = true
	return nil
}

func main() {
	code := runCLI(os.Args[1:], os.Stdout, os.Stderr, musicPlaysSource{}, isTerminal(os.Stdout))
	os.Exit(code)
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func runCLI(args []string, stdout, stderr io.Writer, src PlaySource, stdoutIsTTY bool) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Помилка: не вказано підкоманду. Використовуйте: analyze, compare, help або version.")
		printUsage(stderr)
		return exitUsageErr
	}

	switch args[0] {
	case "help", "-h", "--help":
		if len(args) > 1 {
			fmt.Fprintln(stderr, "Помилка: підкоманда help не приймає додаткових аргументів.")
			return exitUsageErr
		}
		printUsage(stdout)
		return exitOK

	case "version", "-v", "--version":
		if len(args) > 1 {
			fmt.Fprintln(stderr, "Помилка: підкоманда version не приймає додаткових аргументів.")
			return exitUsageErr
		}
		fmt.Fprintf(stdout, "%s v%s (%s)\n", appName, appVersion, variantInfo)
		return exitOK

	case "analyze":
		return runAnalyze(args[1:], stdout, stderr, src, stdoutIsTTY)

	case "compare":
		return runCompare(args[1:], stdout, stderr, src, stdoutIsTTY)

	default:
		fmt.Fprintf(stderr, "Помилка: невідома підкоманда %q.\n", args[0])
		printUsage(stderr)
		return exitUsageErr
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "%s — CLI-аналізатор музичної популярності\n", appName)
	fmt.Fprintf(w, "%s\n\n", variantInfo)
	fmt.Fprintln(w, "Використання:")
	fmt.Fprintln(w, "  music-popularity-analyzer analyze --count <5..100> --subject <виконавець> [--seed <число>] [--no-color]")
	fmt.Fprintln(w, "  music-popularity-analyzer compare --count <5..100> --first <виконавець1> --second <виконавець2> [--seed <число>] [--no-color]")
	fmt.Fprintln(w, "  music-popularity-analyzer help")
	fmt.Fprintln(w, "  music-popularity-analyzer version")
}

func runAnalyze(args []string, stdout, stderr io.Writer, src PlaySource, stdoutIsTTY bool) int {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var count int
	var subject string
	var seedFlag int64OptionalFlag
	var noColor bool

	fs.IntVar(&count, "count", 0, "кількість значень прослуховувань (від 5 до 100, обов'язковий)")
	fs.StringVar(&subject, "subject", "", "назва виконавця (непорожній рядок до 100 символів, обов'язковий)")
	fs.Var(&seedFlag, "seed", "фіксоване зерно генератора для відтворюваного запуску (необов'язковий)")
	fs.BoolVar(&noColor, "no-color", false, "вимкнути ANSI-кольори у виводі")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printUsage(stdout)
			return exitOK
		}
		return exitUsageErr
	}

	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "Помилка: виявлено зайві позиційні аргументи: %v\n", fs.Args())
		return exitUsageErr
	}

	if count < analyzer.MinSampleCount || count > analyzer.MaxSampleCount {
		fmt.Fprintf(stderr, "Помилка: параметр --count є обов'язковим і має бути в межах від %d до %d (отримано: %d).\n",
			analyzer.MinSampleCount, analyzer.MaxSampleCount, count)
		return exitUsageErr
	}

	cleanedSubject, err := validateSubject(subject, "--subject")
	if err != nil {
		fmt.Fprintf(stderr, "Помилка: %v\n", err)
		return exitUsageErr
	}

	var raw []int
	if seedFlag.set {
		raw, err = src.GenerateWithSeed(count, cleanedSubject, seedFlag.value)
	} else {
		raw, err = src.Generate(count, cleanedSubject)
	}
	if err != nil {
		fmt.Fprintf(stderr, "Помилка генератора даних: %v\n", err)
		return exitDataErr
	}

	series, err := analyzer.NewArtistPlaySeries(cleanedSubject, raw)
	if err != nil {
		fmt.Fprintf(stderr, "Помилка перевірки вхідних даних: %v\n", err)
		return exitDataErr
	}

	report := series.Analyze()
	useColor := stdoutIsTTY && !noColor
	formatter := presentation.NewFormatter(stdout, useColor)
	if err := formatter.RenderReport(report); err != nil {
		fmt.Fprintf(stderr, "Помилка форматування виводу: %v\n", err)
		return exitDataErr
	}

	return exitOK
}

func runCompare(args []string, stdout, stderr io.Writer, src PlaySource, stdoutIsTTY bool) int {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var count int
	var first, second string
	var seedFlag int64OptionalFlag
	var noColor bool

	fs.IntVar(&count, "count", 0, "кількість значень для кожного виконавця (від 5 до 100)")
	fs.StringVar(&first, "first", "", "перший виконавець")
	fs.StringVar(&second, "second", "", "другий виконавець")
	fs.Var(&seedFlag, "seed", "фіксоване зерно генератора")
	fs.BoolVar(&noColor, "no-color", false, "вимкнути ANSI-кольори")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsageErr
	}

	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "Помилка: зайві аргументи: %v\n", fs.Args())
		return exitUsageErr
	}

	if count < analyzer.MinSampleCount || count > analyzer.MaxSampleCount {
		fmt.Fprintf(stderr, "Помилка: параметр --count має бути від %d до %d.\n",
			analyzer.MinSampleCount, analyzer.MaxSampleCount)
		return exitUsageErr
	}

	cleanedFirst, err := validateSubject(first, "--first")
	if err != nil {
		fmt.Fprintf(stderr, "Помилка: %v\n", err)
		return exitUsageErr
	}
	cleanedSecond, err := validateSubject(second, "--second")
	if err != nil {
		fmt.Fprintf(stderr, "Помилка: %v\n", err)
		return exitUsageErr
	}

	raw1, err := fetchPlays(src, count, cleanedFirst, seedFlag)
	if err != nil {
		fmt.Fprintf(stderr, "Помилка генератора для першого виконавця: %v\n", err)
		return exitDataErr
	}
	raw2, err := fetchPlays(src, count, cleanedSecond, seedFlag)
	if err != nil {
		fmt.Fprintf(stderr, "Помилка генератора для другого виконавця: %v\n", err)
		return exitDataErr
	}

	s1, err := analyzer.NewArtistPlaySeries(cleanedFirst, raw1)
	if err != nil {
		fmt.Fprintf(stderr, "Помилка даних першого виконавця: %v\n", err)
		return exitDataErr
	}
	s2, err := analyzer.NewArtistPlaySeries(cleanedSecond, raw2)
	if err != nil {
		fmt.Fprintf(stderr, "Помилка даних другого виконавця: %v\n", err)
		return exitDataErr
	}

	useColor := stdoutIsTTY && !noColor
	formatter := presentation.NewFormatter(stdout, useColor)
	if err := formatter.RenderComparison(s1.Analyze(), s2.Analyze()); err != nil {
		fmt.Fprintf(stderr, "Помилка виводу: %v\n", err)
		return exitDataErr
	}
	return exitOK
}

func fetchPlays(src PlaySource, count int, artist string, seedFlag int64OptionalFlag) ([]int, error) {
	if seedFlag.set {
		return src.GenerateWithSeed(count, artist, seedFlag.value)
	}
	return src.Generate(count, artist)
}

func validateSubject(raw, flagName string) (string, error) {
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return "", fmt.Errorf("параметр %s є обов'язковим і не може бути порожнім", flagName)
	}
	if utf8.RuneCountInString(cleaned) > analyzer.MaxArtistLen {
		return "", fmt.Errorf("параметр %s перевищує %d символів", flagName, analyzer.MaxArtistLen)
	}
	return cleaned, nil
}
