package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type stubPlaySource struct {
	data []int
	err  error
}

func (s stubPlaySource) Generate(count int, artist string) ([]int, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.data[:count], nil
}

func (s stubPlaySource) GenerateWithSeed(count int, artist string, seed int64) ([]int, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.data[:count], nil
}

func TestRunCLI_TableDriven(t *testing.T) {
	validSource := stubPlaySource{
		data: []int{5000, 150000, 5000, 1000000, 2500000, 40000, 900, 300000, 4000000, 12000},
	}

	tests := []struct {
		name           string
		args           []string
		src            PlaySource
		wantCode       int
		wantStdoutPart string
		wantStderrPart string
	}{
		{
			name:           "help_command",
			args:           []string{"help"},
			src:            validSource,
			wantCode:       exitOK,
			wantStdoutPart: "Варіант 26",
		},
		{
			name:           "version_command",
			args:           []string{"version"},
			src:            validSource,
			wantCode:       exitOK,
			wantStdoutPart: "music-popularity-analyzer v1.0.0",
		},
		{
			name:           "analyze_valid_with_seed_and_no_color",
			args:           []string{"analyze", "--count", "5", "--subject", "Title Fight", "--seed", "42", "--no-color"},
			src:            validSource,
			wantCode:       exitOK,
			wantStdoutPart: "Сума прослуховувань:",
		},
		{
			name:           "analyze_subject_starting_with_dash",
			args:           []string{"analyze", "--count=5", "--subject=-band-name-"},
			src:            validSource,
			wantCode:       exitOK,
			wantStdoutPart: "-band-name-",
		},
		{
			name:           "compare_two_artists",
			args:           []string{"compare", "--count", "5", "--first", "Thursday", "--second", "Embrace", "--seed", "7"},
			src:            validSource,
			wantCode:       exitOK,
			wantStdoutPart: "Порівняння популярності: Thursday vs Embrace",
		},
		{
			name:           "missing_count_flag",
			args:           []string{"analyze", "--subject", "Thursday"},
			src:            validSource,
			wantCode:       exitUsageErr,
			wantStderrPart: "--count",
		},
		{
			name:           "count_out_of_bounds",
			args:           []string{"analyze", "--count", "4", "--subject", "Thursday"},
			src:            validSource,
			wantCode:       exitUsageErr,
			wantStderrPart: "від 5 до 100",
		},
		{
			name:           "empty_subject_rejected",
			args:           []string{"analyze", "--count", "5", "--subject", "   "},
			src:            validSource,
			wantCode:       exitUsageErr,
			wantStderrPart: "--subject",
		},
		{
			name:           "unknown_subcommand",
			args:           []string{"unknown"},
			src:            validSource,
			wantCode:       exitUsageErr,
			wantStderrPart: "невідома підкоманда",
		},
		{
			name:           "generator_failure_returns_exit_1",
			args:           []string{"analyze", "--count", "5", "--subject", "Thursday"},
			src:            stubPlaySource{err: errors.New("генератор недоступний")},
			wantCode:       exitDataErr,
			wantStderrPart: "Помилка генератора даних",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runCLI(tc.args, &stdout, &stderr, tc.src, false)

			if code != tc.wantCode {
				t.Fatalf("код завершення: очікувалося %d, отримано %d (stderr: %s)", tc.wantCode, code, stderr.String())
			}
			if tc.wantStdoutPart != "" && !strings.Contains(stdout.String(), tc.wantStdoutPart) {
				t.Errorf("stdout не містить %q; отримано:\n%s", tc.wantStdoutPart, stdout.String())
			}
			if tc.wantStderrPart != "" && !strings.Contains(stderr.String(), tc.wantStderrPart) {
				t.Errorf("stderr не містить %q; отримано:\n%s", tc.wantStderrPart, stderr.String())
			}
		})
	}
}
