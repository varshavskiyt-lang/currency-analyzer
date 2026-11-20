package analyzer

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestNewQuoteSeries(t *testing.T) {
	validRates := []float64{0.950, 1.020, 1.150, 0.880, 1.250}

	tests := []struct {
		name      string
		pair      string
		rates     []float64
		wantErr   bool
		wantPair  CurrencyPair
		wantCount int
	}{
		{
			name:      "коректні дані та обрізання пробілів",
			pair:      "  EUR/USD  ",
			rates:     validRates,
			wantErr:   false,
			wantPair:  CurrencyPair("EUR/USD"),
			wantCount: 5,
		},
		{
			name:    "порожня валютна пара",
			pair:    "   ",
			rates:   validRates,
			wantErr: true,
		},
		{
			name:    "занадто довга назва валютної пари",
			pair:    strings.Repeat("A", MaxSubjectLength+1),
			rates:   validRates,
			wantErr: true,
		},
		{
			name:    "менше мінімальної кількості значень",
			pair:    "EUR/USD",
			rates:   []float64{1.0, 1.1, 1.2, 1.3},
			wantErr: true,
		},
		{
			name:    "більше максимальної кількості значень",
			pair:    "EUR/USD",
			rates:   make([]float64, MaxRatesCount+1),
			wantErr: true,
		},
		{
			name:    "від'ємний курс у наборі",
			pair:    "EUR/USD",
			rates:   []float64{1.0, -0.5, 1.1, 1.2, 1.3},
			wantErr: true,
		},
		{
			name:    "нульовий курс у наборі",
			pair:    "EUR/USD",
			rates:   []float64{1.0, 0.0, 1.1, 1.2, 1.3},
			wantErr: true,
		},
		{
			name:    "NaN у наборі",
			pair:    "EUR/USD",
			rates:   []float64{1.0, math.NaN(), 1.1, 1.2, 1.3},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			series, err := NewQuoteSeries(tc.pair, tc.rates)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("очікувалась помилка, але отримано nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("неочікувана помилка: %v", err)
			}
			if series.Pair() != tc.wantPair {
				t.Errorf("Pair() = %q, очікувалось %q", series.Pair(), tc.wantPair)
			}
			if series.Count() != tc.wantCount {
				t.Errorf("Count() = %d, очікувалось %d", series.Count(), tc.wantCount)
			}
		})
	}
}

func TestQuoteSeriesDefensiveCopy(t *testing.T) {
	input := []float64{0.850, 0.950, 1.050, 1.150, 1.250}
	series, err := NewQuoteSeries("USD/UAH", input)
	if err != nil {
		t.Fatalf("неочікувана помилка: %v", err)
	}

	// Зміна вихідного зрізу не повинна впливати на внутрішній стан QuoteSeries
	input[0] = 999.999
	got := series.Rates()
	if got[0] != 0.850 {
		t.Errorf("внутрішній зріз змінився після модифікації вхідного: отримано %v", got[0])
	}

	// Зміна зрізу, повернутого методом Rates(), також не повинна змінювати внутрішній стан
	got[1] = 888.888
	if series.Rates()[1] != 0.950 {
		t.Errorf("внутрішній зріз змінився після модифікації копії Rates()")
	}
}

func TestQuoteSeriesAnalyze(t *testing.T) {
	rates := []float64{0.850, 0.950, 0.950, 1.100, 1.250, 1.250}
	series, err := NewQuoteSeries("EUR/USD", rates)
	if err != nil {
		t.Fatalf("неочікувана помилка створення серії: %v", err)
	}

	report, err := series.Analyze()
	if err != nil {
		t.Fatalf("неочікувана помилка аналізу: %v", err)
	}

	if report.Min != 0.850 {
		t.Errorf("Min = %.3f, очікувалось 0.850", report.Min)
	}
	if report.Max != 1.250 {
		t.Errorf("Max = %.3f, очікувалось 1.250", report.Max)
	}
	if math.Abs(report.Sum-6.350) > 1e-9 {
		t.Errorf("Sum = %.3f, очікувалось 6.350", report.Sum)
	}
	wantAvg := 6.350 / 6.0
	if math.Abs(report.Average-wantAvg) > 1e-9 {
		t.Errorf("Average = %.6f, очікувалось %.6f", report.Average, wantAvg)
	}

	wantCategories := []CategoryStat{
		{Name: CategoryLow, Count: 1},
		{Name: CategoryBelowPar, Count: 2},
		{Name: CategoryAbovePar, Count: 1},
		{Name: CategoryHigh, Count: 2},
	}
	if !reflect.DeepEqual(report.Categories, wantCategories) {
		t.Errorf("Categories = %+v, очікувалось %+v", report.Categories, wantCategories)
	}

	wantRepeated := []RepeatedRate{
		{Value: 0.950, Frequency: 2},
		{Value: 1.250, Frequency: 2},
	}
	if !reflect.DeepEqual(report.Repeated, wantRepeated) {
		t.Errorf("Repeated = %+v, очікувалось %+v", report.Repeated, wantRepeated)
	}
	if report.TotalDuplicates != 2 {
		t.Errorf("TotalDuplicates = %d, очікувалось 2", report.TotalDuplicates)
	}

	wantTrends := []string{"Вище", "Таке ж", "Вище", "Вище", "Таке ж"}
	if !reflect.DeepEqual(report.Trends, wantTrends) {
		t.Errorf("Trends = %v, очікувалось %v", report.Trends, wantTrends)
	}
}

func TestUninitializedQuoteSeriesAnalyze(t *testing.T) {
	var empty QuoteSeries
	if _, err := empty.Analyze(); err == nil {
		t.Fatal("очікувалась помилка для неініціалізованої структури QuoteSeries")
	}
}
