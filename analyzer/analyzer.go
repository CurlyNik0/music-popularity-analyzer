package analyzer

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	MinSampleCount = 5
	MaxSampleCount = 100
	MaxArtistLen   = 100
	MinPlayCount   = 0
	MaxPlayCount   = 10_000_000
	LowTierMax     = 9_999
	MediumTierMax  = 199_999
	HighTierMax    = 1_999_999
)

var (
	ErrEmptyArtist       = errors.New("назва виконавця не може бути порожньою")
	ErrArtistTooLong     = errors.New("назва виконавця перевищує допустиму довжину (100 символів)")
	ErrInvalidSampleSize = errors.New("кількість значень має бути в межах від 5 до 100")
	ErrPlayOutOfRange    = errors.New("кількість прослуховувань виходить за межі [0, 10000000]")
)

type PlayCount int

type PopularityTier string

const (
	TierLow      PopularityTier = "Низька"
	TierMedium   PopularityTier = "Середня"
	TierHigh     PopularityTier = "Висока"
	TierVeryHigh PopularityTier = "Дуже висока"
)

func OrderedTiers() []PopularityTier {
	return []PopularityTier{TierLow, TierMedium, TierHigh, TierVeryHigh}
}

func (p PlayCount) Classify() PopularityTier {
	switch {
	case p <= LowTierMax:
		return TierLow
	case p <= MediumTierMax:
		return TierMedium
	case p <= HighTierMax:
		return TierHigh
	default:
		return TierVeryHigh
	}
}

type ArtistPlaySeries struct {
	artist string
	plays  []PlayCount
}

func NewArtistPlaySeries(artist string, rawPlays []int) (ArtistPlaySeries, error) {
	cleanedArtist := strings.TrimSpace(artist)
	if cleanedArtist == "" {
		return ArtistPlaySeries{}, ErrEmptyArtist
	}
	if utf8.RuneCountInString(cleanedArtist) > MaxArtistLen {
		return ArtistPlaySeries{}, fmt.Errorf("%w: отримано %d", ErrArtistTooLong, utf8.RuneCountInString(cleanedArtist))
	}
	if len(rawPlays) < MinSampleCount || len(rawPlays) > MaxSampleCount {
		return ArtistPlaySeries{}, fmt.Errorf("%w: отримано %d", ErrInvalidSampleSize, len(rawPlays))
	}

	copied := make([]PlayCount, len(rawPlays))
	for i, val := range rawPlays {
		if val < MinPlayCount || val > MaxPlayCount {
			return ArtistPlaySeries{}, fmt.Errorf("%w: індекс %d, значення %d", ErrPlayOutOfRange, i, val)
		}
		copied[i] = PlayCount(val)
	}

	return ArtistPlaySeries{
		artist: cleanedArtist,
		plays:  copied,
	}, nil
}

func (s ArtistPlaySeries) Artist() string {
	return s.artist
}

func (s ArtistPlaySeries) Plays() []PlayCount {
	out := make([]PlayCount, len(s.plays))
	copy(out, s.plays)
	return out
}

type SummaryStats struct {
	Count        int
	Sum          int64
	Min          PlayCount
	Max          PlayCount
	Average      float64
	AboveAvgCnt  int
	AboveAvgDiff float64
}

type TierStat struct {
	Tier  PopularityTier
	Count int
}

type RepeatEntry struct {
	Value     PlayCount
	Frequency int
}

type PopularityReport struct {
	Artist       string
	Plays        []PlayCount
	Stats        SummaryStats
	Tiers        []TierStat
	Repeats      []RepeatEntry
	AlgorithmsOK bool
}

func (s ArtistPlaySeries) SummarizeStats() SummaryStats {
	var sum int64
	minVal := s.plays[0]
	maxVal := s.plays[0]

	for _, p := range s.plays {
		sum += int64(p)
		if p < minVal {
			minVal = p
		}
		if p > maxVal {
			maxVal = p
		}
	}

	avg := float64(sum) / float64(len(s.plays))
	aboveCnt := 0
	diffSum := 0.0
	for _, p := range s.plays {
		if float64(p) > avg {
			aboveCnt++
			diffSum += float64(p) - avg
		}
	}

	return SummaryStats{
		Count:        len(s.plays),
		Sum:          sum,
		Min:          minVal,
		Max:          maxVal,
		Average:      avg,
		AboveAvgCnt:  aboveCnt,
		AboveAvgDiff: diffSum,
	}
}

func (s ArtistPlaySeries) CountByTier() []TierStat {
	counts := make(map[PopularityTier]int, 4)
	for _, p := range s.plays {
		counts[p.Classify()]++
	}

	order := OrderedTiers()
	result := make([]TierStat, len(order))
	for i, tier := range order {
		result[i] = TierStat{
			Tier:  tier,
			Count: counts[tier],
		}
	}
	return result
}

func (s ArtistPlaySeries) FindRepeats() ([]RepeatEntry, bool) {
	var repeatsQuadratic []PlayCount
	for i := 0; i < len(s.plays); i++ {
		already := false
		for _, added := range repeatsQuadratic {
			if added == s.plays[i] {
				already = true
				break
			}
		}
		if already {
			continue
		}
		for j := i + 1; j < len(s.plays); j++ {
			if s.plays[i] == s.plays[j] {
				repeatsQuadratic = append(repeatsQuadratic, s.plays[i])
				break
			}
		}
	}

	freq := make(map[PlayCount]int, len(s.plays))
	for _, p := range s.plays {
		freq[p]++
	}

	var repeatsLinear []PlayCount
	seen := make(map[PlayCount]bool, len(freq))
	for _, p := range s.plays {
		if freq[p] > 1 && !seen[p] {
			seen[p] = true
			repeatsLinear = append(repeatsLinear, p)
		}
	}

	consistent := len(repeatsQuadratic) == len(repeatsLinear)
	if consistent {
		for i := range repeatsQuadratic {
			if repeatsQuadratic[i] != repeatsLinear[i] {
				consistent = false
				break
			}
		}
	}

	entries := make([]RepeatEntry, len(repeatsLinear))
	for i, val := range repeatsLinear {
		entries[i] = RepeatEntry{
			Value:     val,
			Frequency: freq[val],
		}
	}
	return entries, consistent
}

func (s ArtistPlaySeries) Analyze() PopularityReport {
	repeats, ok := s.FindRepeats()
	return PopularityReport{
		Artist:       s.artist,
		Plays:        s.Plays(),
		Stats:        s.SummarizeStats(),
		Tiers:        s.CountByTier(),
		Repeats:      repeats,
		AlgorithmsOK: ok,
	}
}
