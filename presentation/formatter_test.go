package presentation

import (
	"bytes"
	"strings"
	"testing"

	"github.com/CurlyNik0/music-popularity-analyzer/analyzer"
)

func TestFormatter_RenderReport_NoColorAndANSI(t *testing.T) {
	series, err := analyzer.NewArtistPlaySeries("Title Fight", []int{5000, 150000, 5000, 1000000, 2500000})
	if err != nil {
		t.Fatalf("неочікувана помилка: %v", err)
	}
	report := series.Analyze()

	t.Run("no_color_mode_has_no_ansi_sequences", func(t *testing.T) {
		var buf bytes.Buffer
		f := NewFormatter(&buf, false)
		if err := f.RenderReport(report); err != nil {
			t.Fatalf("помилка форматування: %v", err)
		}

		out := buf.String()
		if strings.Contains(out, "\033[") {
			t.Fatalf("у режимі без кольорів виявлено ANSI-послідовності: %q", out)
		}

		expectedFragments := []string{
			"Title Fight",
			"3660000",
			"732000.00",
			"Низька",
			"Середня",
			"Висока",
			"Дуже висока",
			"5000",
		}
		for _, fragment := range expectedFragments {
			if !strings.Contains(out, fragment) {
				t.Errorf("у виводі відсутній очікуваний фрагмент %q", fragment)
			}
		}
	})

	t.Run("color_mode_includes_ansi_sequences", func(t *testing.T) {
		var buf bytes.Buffer
		f := NewFormatter(&buf, true)
		if err := f.RenderReport(report); err != nil {
			t.Fatalf("помилка форматування: %v", err)
		}

		if !strings.Contains(buf.String(), "\033[") {
			t.Fatalf("у кольоровому режимі очікувались ANSI-послідовності")
		}
	})
}
