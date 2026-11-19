package presentation

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/CurlyNik0/music-popularity-analyzer/analyzer"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiCyan   = "\033[36m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
)

// Formatter відповідає за табличне та кольорове оформлення результатів аналізу у переданий io.Writer.
type Formatter struct {
	w        io.Writer
	useColor bool
}

// NewFormatter створює екземпляр форматувальника виводу.
func NewFormatter(w io.Writer, useColor bool) Formatter {
	return Formatter{
		w:        w,
		useColor: useColor,
	}
}

func (f Formatter) style(code, text string) string {
	if !f.useColor {
		return text
	}
	return code + text + ansiReset
}

// RenderReport виводить повний структурований звіт за результатами аналізу популярності виконавця.
func (f Formatter) RenderReport(report analyzer.PopularityReport) error {
	header := fmt.Sprintf("=== Аналіз музичної популярності: %s ===", report.Artist)
	if _, err := fmt.Fprintln(f.w, f.style(ansiBold+ansiCyan, header)); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(f.w, f.style(ansiBold, "\n1. Згенеровані значення прослуховувань:")); err != nil {
		return err
	}
	for i, val := range report.Plays {
		if _, err := fmt.Fprintf(f.w, "  [%2d] %d\n", i+1, val); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintln(f.w, f.style(ansiBold, "\n2. Статистичні показники:")); err != nil {
		return err
	}
	twStats := tabwriter.NewWriter(f.w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(twStats, "  Кількість значень:\t%d\n", report.Stats.Count)
	fmt.Fprintf(twStats, "  Сума прослуховувань:\t%d\n", report.Stats.Sum)
	fmt.Fprintf(twStats, "  Мінімум:\t%d\n", report.Stats.Min)
	fmt.Fprintf(twStats, "  Максимум:\t%d\n", report.Stats.Max)
	fmt.Fprintf(twStats, "  Середнє арифметичне:\t%.2f\n", report.Stats.Average)
	fmt.Fprintf(twStats, "  Кількість вище середнього:\t%d\n", report.Stats.AboveAvgCnt)
	fmt.Fprintf(twStats, "  Сума різниць вище середнього:\t%.2f\n", report.Stats.AboveAvgDiff)
	if err := twStats.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(f.w, f.style(ansiBold, "\n3. Розподіл за категоріями:")); err != nil {
		return err
	}
	twTiers := tabwriter.NewWriter(f.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(twTiers, "  Категорія\tКількість")
	for _, t := range report.Tiers {
		fmt.Fprintf(twTiers, "  %s\t%d\n", t.Tier, t.Count)
	}
	if err := twTiers.Flush(); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(f.w, f.style(ansiBold, "\n4. Повторювані значення:")); err != nil {
		return err
	}
	if report.AlgorithmsOK {
		fmt.Fprintln(f.w, f.style(ansiGreen, "  Перевірка O(n^2) та O(n): результати збігаються."))
	} else {
		fmt.Fprintln(f.w, f.style(ansiYellow, "  Увага: виявлено розбіжність між алгоритмами пошуку повторів!"))
	}

	if len(report.Repeats) == 0 {
		_, err := fmt.Fprintln(f.w, "  Повторюваних значень не виявлено.")
		return err
	}

	twRep := tabwriter.NewWriter(f.w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(twRep, "  Значення\tЧастота")
	for _, r := range report.Repeats {
		fmt.Fprintf(twRep, "  %d\t%d\n", r.Value, r.Frequency)
	}
	return twRep.Flush()
}

// RenderComparison виводить порівняльну таблицю для двох виконавців.
func (f Formatter) RenderComparison(first, second analyzer.PopularityReport) error {
	title := fmt.Sprintf("=== Порівняння популярності: %s vs %s ===", first.Artist, second.Artist)
	if _, err := fmt.Fprintln(f.w, f.style(ansiBold+ansiCyan, title)); err != nil {
		return err
	}

	tw := tabwriter.NewWriter(f.w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "  Показник\t%s\t%s\n", first.Artist, second.Artist)
	fmt.Fprintf(tw, "  Кількість вибірки\t%d\t%d\n", first.Stats.Count, second.Stats.Count)
	fmt.Fprintf(tw, "  Сума прослуховувань\t%d\t%d\n", first.Stats.Sum, second.Stats.Sum)
	fmt.Fprintf(tw, "  Мінімум\t%d\t%d\n", first.Stats.Min, second.Stats.Min)
	fmt.Fprintf(tw, "  Максимум\t%d\t%d\n", first.Stats.Max, second.Stats.Max)
	fmt.Fprintf(tw, "  Середнє значення\t%.2f\t%.2f\n", first.Stats.Average, second.Stats.Average)
	return tw.Flush()
}
