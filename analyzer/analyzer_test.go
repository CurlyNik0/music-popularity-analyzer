package analyzer

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/compilyator/VNTU_GoLang/labs/Lab2Data/musicplays"
)

func almostEqual(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}

func TestArtistPlaySeries_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		artist         string
		rawPlays       []int
		wantErr        error
		wantStats      SummaryStats
		wantTiers      []TierStat
		wantRepeats    []RepeatEntry
		wantNormalized string
	}{
		{
			name:           "valid_static_with_repeats_and_all_tiers",
			artist:         "  Title Fight  ",
			rawPlays:       []int{5000, 150000, 5000, 1000000, 2500000},
			wantNormalized: "Title Fight",
			wantStats: SummaryStats{
				Count:        5,
				Sum:          3660000,
				Min:          5000,
				Max:          2500000,
				Average:      732000.0,
				AboveAvgCnt:  2,
				AboveAvgDiff: 2036000.0, // (1000000 - 732000) + (2500000 - 732000) = 268000 + 1768000
			},
			wantTiers: []TierStat{
				{Tier: TierLow, Count: 2},
				{Tier: TierMedium, Count: 1},
				{Tier: TierHigh, Count: 1},
				{Tier: TierVeryHigh, Count: 1},
			},
			wantRepeats: []RepeatEntry{
				{Value: 5000, Frequency: 2},
			},
		},
		{
			name:           "tier_boundaries_exact_thresholds",
			artist:         "Thursday",
			rawPlays:       []int{0, 9999, 10000, 199999, 200000, 1999999, 2000000, 10000000},
			wantNormalized: "Thursday",
			wantStats: SummaryStats{
				Count:        8,
				Sum:          14419997,
				Min:          0,
				Max:          10000000,
				Average:      1802499.625,
				AboveAvgCnt:  3,
				AboveAvgDiff: 8592500.125, // (1999999 - avg) + (2000000 - avg) + (10000000 - avg)
			},
			wantTiers: []TierStat{
				{Tier: TierLow, Count: 2},      // 0, 9999
				{Tier: TierMedium, Count: 2},   // 10000, 199999
				{Tier: TierHigh, Count: 2},     // 200000, 1999999
				{Tier: TierVeryHigh, Count: 2}, // 2000000, 10000000
			},
			wantRepeats: []RepeatEntry{},
		},
		{
			name:           "multiple_repeats_preserve_first_appearance_order",
			artist:         "Embrace",
			rawPlays:       []int{30000, 12000, 30000, 500, 12000, 12000},
			wantNormalized: "Embrace",
			wantStats: SummaryStats{
				Count:        6,
				Sum:          96500,
				Min:          500,
				Max:          30000,
				Average:      16083.333333333334,
				AboveAvgCnt:  2,
				AboveAvgDiff: 27833.333333333332,
			},
			wantTiers: []TierStat{
				{Tier: TierLow, Count: 1},
				{Tier: TierMedium, Count: 5},
				{Tier: TierHigh, Count: 0},
				{Tier: TierVeryHigh, Count: 0},
			},
			wantRepeats: []RepeatEntry{
				{Value: 30000, Frequency: 2},
				{Value: 12000, Frequency: 3},
			},
		},
		{
			name:     "empty_artist_rejected",
			artist:   "   \t\n ",
			rawPlays: []int{100, 200, 300, 400, 500},
			wantErr:  ErrEmptyArtist,
		},
		{
			name:     "too_long_artist_rejected",
			artist:   strings.Repeat("a", MaxArtistLen+1),
			rawPlays: []int{100, 200, 300, 400, 500},
			wantErr:  ErrArtistTooLong,
		},
		{
			name:     "too_few_elements_rejected",
			artist:   "Nuvolascura",
			rawPlays: []int{100, 200, 300, 400},
			wantErr:  ErrInvalidSampleSize,
		},
		{
			name:     "empty_slice_rejected",
			artist:   "Nuvolascura",
			rawPlays: []int{},
			wantErr:  ErrInvalidSampleSize,
		},
		{
			name:     "negative_play_count_rejected",
			artist:   "Lifelover",
			rawPlays: []int{100, -1, 300, 400, 500},
			wantErr:  ErrPlayOutOfRange,
		},
		{
			name:     "above_max_play_count_rejected",
			artist:   "Lifelover",
			rawPlays: []int{100, 200, 10_000_001, 400, 500},
			wantErr:  ErrPlayOutOfRange,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			series, err := NewArtistPlaySeries(tc.artist, tc.rawPlays)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("очікувалась помилка %v, отримано %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("неочікувана помилка конструктора: %v", err)
			}

			if series.Artist() != tc.wantNormalized {
				t.Errorf("Artist(): очікувалося %q, отримано %q", tc.wantNormalized, series.Artist())
			}

			report := series.Analyze()
			if !report.AlgorithmsOK {
				t.Errorf("алгоритми O(n^2) та O(n) не збіглися")
			}

			if report.Stats.Count != tc.wantStats.Count ||
				report.Stats.Sum != tc.wantStats.Sum ||
				report.Stats.Min != tc.wantStats.Min ||
				report.Stats.Max != tc.wantStats.Max ||
				report.Stats.AboveAvgCnt != tc.wantStats.AboveAvgCnt ||
				!almostEqual(report.Stats.Average, tc.wantStats.Average, 1e-6) ||
				!almostEqual(report.Stats.AboveAvgDiff, tc.wantStats.AboveAvgDiff, 1e-3) {
				t.Errorf("Stats: очікувалося %+v, отримано %+v", tc.wantStats, report.Stats)
			}

			if !reflect.DeepEqual(report.Tiers, tc.wantTiers) {
				t.Errorf("Tiers: очікувалося %+v, отримано %+v", tc.wantTiers, report.Tiers)
			}

			if !reflect.DeepEqual(report.Repeats, tc.wantRepeats) {
				t.Errorf("Repeats: очікувалося %+v, отримано %+v", tc.wantRepeats, report.Repeats)
			}
		})
	}
}

func TestArtistPlaySeries_DefensiveCopy(t *testing.T) {
	input := []int{1000, 2000, 3000, 4000, 5000}
	series, err := NewArtistPlaySeries("False Effect", input)
	if err != nil {
		t.Fatalf("неочікувана помилка: %v", err)
	}

	// Зміна вихідного зрізу після створення серії не повинна порушувати внутрішній стан
	input[0] = 9_999_999
	plays := series.Plays()
	if plays[0] != 1000 {
		t.Fatalf("інваріант порушено через зміну вхідного зрізу: отримано %d, очікувалося 1000", plays[0])
	}

	// Зміна отриманого через Plays() зрізу також не повинна впливати на об'єкт
	plays[1] = 8_888_888
	if series.Plays()[1] != 2000 {
		t.Fatalf("інваріант порушено через зміну повернутого зрізу: отримано %d, очікувалося 2000", series.Plays()[1])
	}
}

func TestIntegrationWithFixedSeed(t *testing.T) {
	const (
		count  = 10
		artist = "Title Fight"
		seed   = int64(42)
	)

	raw1, err := musicplays.GenerateWithSeed(count, artist, seed)
	if err != nil {
		t.Fatalf("помилка GenerateWithSeed (1): %v", err)
	}
	raw2, err := musicplays.GenerateWithSeed(count, artist, seed)
	if err != nil {
		t.Fatalf("помилка GenerateWithSeed (2): %v", err)
	}

	s1, err := NewArtistPlaySeries(artist, raw1)
	if err != nil {
		t.Fatalf("помилка створення серії 1: %v", err)
	}
	s2, err := NewArtistPlaySeries(artist, raw2)
	if err != nil {
		t.Fatalf("помилка створення серії 2: %v", err)
	}

	r1 := s1.Analyze()
	r2 := s2.Analyze()

	if !reflect.DeepEqual(r1, r2) {
		t.Fatalf("результати з фіксованим seed=%d не детерміновані: %+v != %+v", seed, r1, r2)
	}
}
