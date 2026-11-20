package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/compilyator/VNTU_GoLang/labs/Lab2Data/exchange"
	"github.com/varshavskiyt-lang/currency-analyzer/analyzer"
)

const (
	appName    = "currency-analyzer"
	appVersion = "1.0.0"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "Помилка: не вказано підкоманду. Доступні команди: analyze, help, version.")
		fmt.Fprintf(stderr, "Використайте '%s help' для перегляду довідки.\n", appName)
		return 2
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	case "version", "-v", "--version":
		fmt.Fprintf(stdout, "%s version %s\n", appName, appVersion)
		return 0
	case "analyze":
		return runAnalyze(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "Помилка: невідома підкоманда %q. Використайте '%s help' для довідки.\n", args[0], appName)
		return 2
	}
}

func runAnalyze(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(stderr)

	count := fs.Int("count", 0, "Кількість значень для генерації (від 5 до 100, обов'язковий)")
	subject := fs.String("subject", "", "Валютна пара для аналізу, наприклад EUR/USD (обов'язковий)")
	seedRaw := fs.String("seed", "", "Необов'язкове зерно генератора (ціле число int64)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printAnalyzeUsage(stdout)
			return 0
		}
		return 2
	}

	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "Помилка: неочікувані позиційні аргументи: %s\n", strings.Join(fs.Args(), " "))
		return 2
	}

	trimmedSubject := strings.TrimSpace(*subject)
	if trimmedSubject == "" {
		fmt.Fprintln(stderr, "Помилка: обов'язковий параметр --subject не задано або він порожній.")
		return 2
	}
	if utf8.RuneCountInString(trimmedSubject) > analyzer.MaxSubjectLength {
		fmt.Fprintf(stderr, "Помилка: довжина --subject не може перевищувати %d символів.\n", analyzer.MaxSubjectLength)
		return 2
	}

	if *count < analyzer.MinRatesCount {
		fmt.Fprintf(stderr, "Помилка: параметр --count має бути від %d до %d (отримано %d).\n", analyzer.MinRatesCount, analyzer.MaxRatesCount, *count)
		return 2
	}
	if *count > analyzer.MaxRatesCount {
		fmt.Fprintf(stderr, "Помилка: параметр --count має бути від %d до %d (отримано %d).\n", analyzer.MinRatesCount, analyzer.MaxRatesCount, *count)
		return 2
	}

	var (
		seed    int64
		hasSeed bool
	)
	trimmedSeed := strings.TrimSpace(*seedRaw)
	if trimmedSeed != "" {
		parsedSeed, err := strconv.ParseInt(trimmedSeed, 10, 64)
		if err != nil {
			fmt.Fprintf(stderr, "Помилка: параметр --seed має бути цілим числом int64 (отримано %q).\n", trimmedSeed)
			return 2
		}
		seed = parsedSeed
		hasSeed = true
	}

	var (
		rates []float64
		err   error
	)
	if hasSeed {
		rates, err = exchange.GenerateWithSeed(*count, trimmedSubject, seed)
	} else {
		rates, err = exchange.Generate(*count, trimmedSubject)
	}
	if err != nil {
		fmt.Fprintf(stderr, "Помилка генерації котирувань: %v\n", err)
		return 1
	}

	series, err := analyzer.NewQuoteSeries(trimmedSubject, rates)
	if err != nil {
		fmt.Fprintf(stderr, "Помилка валідації серії: %v\n", err)
		return 1
	}

	report, err := series.Analyze()
	if err != nil {
		fmt.Fprintf(stderr, "Помилка аналізу: %v\n", err)
		return 1
	}

	printReport(stdout, report, rates, hasSeed, seed)
	return 0
}

func printReport(w io.Writer, report analyzer.AnalysisReport, rates []float64, hasSeed bool, seed int64) {
	fmt.Fprintf(w, "=== Аналіз валютної пари %s ===\n", report.Pair)
	if hasSeed {
		fmt.Fprintf(w, "Кількість котирувань: %d (seed: %d)\n", report.Count, seed)
	} else {
		fmt.Fprintf(w, "Кількість котирувань: %d (випадкова генерація)\n", report.Count)
	}
	fmt.Fprintln(w, "Згенеровані значення:", formatRates(rates))
	fmt.Fprintf(w, "Мінімум: %.3f\n", report.Min)
	fmt.Fprintf(w, "Максимум: %.3f\n", report.Max)
	fmt.Fprintf(w, "Сума: %.3f\n", report.Sum)
	fmt.Fprintf(w, "Середнє: %.3f\n\n", report.Average)

	fmt.Fprintln(w, "Категорії:")
	for _, cat := range report.Categories {
		fmt.Fprintf(w, "  %s: %d\n", cat.Name, cat.Count)
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Кількість повторів: %d\n", report.TotalDuplicates)
	if len(report.Repeated) > 0 {
		for _, rep := range report.Repeated {
			fmt.Fprintf(w, "  Курс %.3f — %d раз(и)\n", rep.Value, rep.Frequency)
		}
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "Динаміка котирувань: [%s]\n", strings.Join(report.Trends, " "))
}

func formatRates(rates []float64) string {
	parts := make([]string, len(rates))
	for i, v := range rates {
		parts[i] = fmt.Sprintf("%.3f", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Використання:\n  %s <підкоманда> [прапорці]\n\n", appName)
	fmt.Fprintln(w, "Підкоманди:")
	fmt.Fprintln(w, "  analyze   Згенерувати та проаналізувати серію котирувань валютної пари")
	fmt.Fprintln(w, "  help      Показати цю довідку")
	fmt.Fprintln(w, "  version   Показати версію програми")
	fmt.Fprintln(w, "\nПриклади:")
	fmt.Fprintf(w, "  %s analyze --count 10 --subject \"EUR/USD\"\n", appName)
	fmt.Fprintf(w, "  %s analyze --count 5 --subject \"EUR/USD\" --seed 12345\n", appName)
}

func printAnalyzeUsage(w io.Writer) {
	fmt.Fprintf(w, "Використання:\n  %s analyze --count <5..100> --subject <пара> [--seed <int64>]\n\n", appName)
	fmt.Fprintln(w, "Прапорці:")
	fmt.Fprintln(w, "  --count     Кількість значень для генерації від 5 до 100 (обов'язковий)")
	fmt.Fprintln(w, "  --subject   Назва валютної пари, наприклад EUR/USD (обов'язковий)")
	fmt.Fprintln(w, "  --seed      Ціле число int64 для відтворюваної генерації (необов'язковий)")
}
