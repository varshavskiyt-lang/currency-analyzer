package analyzer

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

const (
	MinRatesCount    = 5
	MaxRatesCount    = 100
	MaxSubjectLength = 100

	CategoryLow      = "Низький"
	CategoryBelowPar = "Нижче паритету"
	CategoryAbovePar = "Вище паритету"
	CategoryHigh     = "Високий"
)

// CurrencyPair представляє предметне поняття валютної пари (наприклад, "EUR/USD").
type CurrencyPair string

// CategoryStat зберігає назву предметної категорії та кількість котирувань у ній.
type CategoryStat struct {
	Name  string
	Count int
}

// RepeatedRate зберігає повторюване значення курсу та частоту його появи.
type RepeatedRate struct {
	Value     float64
	Frequency int
}

// QuoteSeries інкапсулює перевірений набір котирувань для заданої валютної пари.
type QuoteSeries struct {
	pair  CurrencyPair
	rates []float64
}

// AnalysisReport містить усі похідні результати аналізу набору котирувань.
type AnalysisReport struct {
	Pair            CurrencyPair
	Count           int
	Min             float64
	Max             float64
	Sum             float64
	Average         float64
	Categories      []CategoryStat
	Repeated        []RepeatedRate
	TotalDuplicates int
	Trends          []string
}

// NewQuoteSeries перевіряє інваріанти вхідних даних і створює коректний набір котирувань.
func NewQuoteSeries(rawPair string, rates []float64) (QuoteSeries, error) {
	trimmed := strings.TrimSpace(rawPair)
	if trimmed == "" {
		return QuoteSeries{}, errors.New("валютна пара не може бути порожньою")
	}
	if utf8.RuneCountInString(trimmed) > MaxSubjectLength {
		return QuoteSeries{}, fmt.Errorf("довжина валютної пари перевищує %d символів", MaxSubjectLength)
	}
	if len(rates) < MinRatesCount || len(rates) > MaxRatesCount {
		return QuoteSeries{}, fmt.Errorf("кількість значень має бути від %d до %d, отримано %d", MinRatesCount, MaxRatesCount, len(rates))
	}

	copied := make([]float64, len(rates))
	for i, v := range rates {
		if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
			return QuoteSeries{}, fmt.Errorf("некоректне значення курсу на позиції %d: %v (курс має бути додатним скінченним числом)", i, v)
		}
		copied[i] = v
	}

	return QuoteSeries{
		pair:  CurrencyPair(trimmed),
		rates: copied,
	}, nil
}

// Pair повертає валютну пару набору котирувань.
func (s QuoteSeries) Pair() CurrencyPair {
	return s.pair
}

// Count повертає кількість котирувань у наборі.
func (s QuoteSeries) Count() int {
	return len(s.rates)
}

// Rates повертає захисну копію зрізу котирувань, щоб зовнішній код не змінив внутрішній стан.
func (s QuoteSeries) Rates() []float64 {
	out := make([]float64, len(s.rates))
	copy(out, s.rates)
	return out
}

// Analyze виконує повний аналіз набору котирувань: статистику, категорії, повтори та динаміку.
func (s QuoteSeries) Analyze() (AnalysisReport, error) {
	if s.pair == "" || len(s.rates) < MinRatesCount {
		return AnalysisReport{}, errors.New("набір котирувань не ініціалізовано через конструктор")
	}

	minVal, maxVal, sumVal, avgVal := s.calcStats()
	categories := s.calcCategories()
	repeated, totalDuplicates := s.calcRepeats()
	trends := s.calcTrends()

	return AnalysisReport{
		Pair:            s.pair,
		Count:           len(s.rates),
		Min:             minVal,
		Max:             maxVal,
		Sum:             sumVal,
		Average:         avgVal,
		Categories:      categories,
		Repeated:        repeated,
		TotalDuplicates: totalDuplicates,
		Trends:          trends,
	}, nil
}

func (s QuoteSeries) calcStats() (minVal, maxVal, sumVal, avgVal float64) {
	minVal = s.rates[0]
	maxVal = s.rates[0]
	sumVal = 0.0

	for _, v := range s.rates {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
		sumVal += v
	}

	avgVal = sumVal / float64(len(s.rates))
	return minVal, maxVal, sumVal, avgVal
}

func (s QuoteSeries) calcCategories() []CategoryStat {
	catLow := 0
	catBelowPar := 0
	catAbovePar := 0
	catHigh := 0

	for _, v := range s.rates {
		if v < 0.900 {
			catLow++
		} else if v >= 0.900 && v < 1.000 {
			catBelowPar++
		} else if v >= 1.000 && v < 1.200 {
			catAbovePar++
		} else {
			catHigh++
		}
	}
	return []CategoryStat{
		{Name: CategoryLow, Count: catLow},
		{Name: CategoryBelowPar, Count: catBelowPar},
		{Name: CategoryAbovePar, Count: catAbovePar},
		{Name: CategoryHigh, Count: catHigh},
	}
}

func (s QuoteSeries) calcRepeats() ([]RepeatedRate, int) {
	freq := make(map[float64]int, len(s.rates))
	order := make([]float64, 0, len(s.rates))

	for _, v := range s.rates {
		if freq[v] == 0 {
			order = append(order, v)
		}
		freq[v]++
	}

	var repeated []RepeatedRate
	totalDuplicates := 0

	// Ітеруємо за зрізом order (порядок першої появи), а не за map, щоб гарантувати стабільний порядок.
	for _, v := range order {
		count := freq[v]
		if count > 1 {
			repeated = append(repeated, RepeatedRate{
				Value:     v,
				Frequency: count,
			})
			totalDuplicates += count - 1
		}
	}

	return repeated, totalDuplicates
}

func (s QuoteSeries) calcTrends() []string {
	trends := make([]string, 0, len(s.rates)-1)
	for i := 1; i < len(s.rates); i++ {
		if s.rates[i] < s.rates[i-1] {
			trends = append(trends, "Нижче")
		} else if s.rates[i] == s.rates[i-1] {
			trends = append(trends, "Таке ж")
		} else {
			trends = append(trends, "Вище")
		}
	}
	return trends
}
