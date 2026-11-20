package main

import (
	"fmt"

	"github.com/charmbracelet/huh" // Додали бібліотеку huh
	"github.com/compilyator/VNTU_GoLang/labs/Lab2Data/exchange"
)

func main() {
	// Створюємо змінні використати пакети
	var count int
	var pair string

	// Використовуємо пакет fmt, щоб він не зникав
	fmt.Print("Введіть кількість значень (від 5 до 100): ")
	fmt.Scan(&count)
	if count < 5 || count > 100 {
		fmt.Println("Помилка: кількість має бути від 5 до 100")
		return
	}

	// --- Інтерактивне меню huh ---
	errHuh := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Оберіть валютну пару (Стрілки Вгору/Вниз, Enter для вибору):").
				Options(
					huh.NewOption("Євро до Долара (EUR/USD)", "EUR/USD"),
					huh.NewOption("Фунт до Гривні (GBP/UAH)", "GBP/UAH"),
					huh.NewOption("Долар до Гривні (USD/UAH)", "USD/UAH"),
					huh.NewOption("Зламати програму порожньою назвою (для тесту)", ""),
				).
				Value(&pair), // Вибране значення запишеться у нашу змінну pair
		),
	).Run()

	// Якщо користувач раптом скасував меню (нажав Ctrl+C)
	if errHuh != nil {
		fmt.Println("Помилка меню:", errHuh)
		return
	}

	// Використовуємо пакет exchange, щоб він теж не зникав
	values, err := exchange.Generate(count, pair)
	if err != nil {
		fmt.Println("Помилка:", err)
		return
	}

	fmt.Println("Отримані дані.", values)

	// --- Етап 3: Обчислення ---
	// Беремо перше згенероване число і кажемо, що воно пои що найменше і найбільше
	min := values[0]
	max := values[0]
	sum := 0.0

	//Запускаємо цикл: перебираємо кожен "вагончик" у нашому "потязі" (зрізі)
	for _, v := range values {
		//Якщо поточне число менше за наш мінімум - оновлюємо мінімум
		if v < min {
			min = v
		}
		//Якщо більше за максимум - оновлюємо максимум
		if v > max {
			max = v
		}
		//Додаємо чило до загальної суми
		sum += v
	}

	//Рахуємо серднє(ділимо суму на кількість вагончиків)
	// len(values) - це довжина зрізу, Ми перетворюємо її на float64, бо сума теж дробова.
	avg := sum / float64(len(values))

	//Виводимо результати красиво (%.3f - означає 3 знаки після коми)
	fmt.Printf("Мінімум: %.3f\n", min)
	fmt.Printf("Максимум: %.3f\n", max)
	fmt.Printf("Сума: %.3f\n", sum)
	fmt.Printf("Середнє: %.3f\n\n", avg)

	// --- Етап 4: Категоризація ---
	// Створюємо 4 порожні "коробочки" (лічильники) для кожної категорії
	catLow := 0      // Низький до 0.899
	catBeLowpar := 0 // Нижче паритету 0.900 - 0.999
	catAbovePar := 0 // Вище паритету 1.000 - 1.199
	catHigh := 0     // Високий від 1.200

	// Знову пробігаємося по всіх "вагончиках" (зрізу)
	for _, v := range values {
		if v < 0.900 {
			catLow++ // "++" означає "додати 1" до лічильника
		} else if v >= 0.900 && v < 1.000 {
			catBeLowpar++
		} else if v >= 1.000 && v < 1.200 {
			catAbovePar++
		} else { // Якщо жодна з умов вище не підійшла, число точно >= 1.200
			catHigh++
		}
	}

	// Виводимо результати категорій на екран (%d - для цілих чисел)
	fmt.Println("Категорії:")
	fmt.Printf("  Низький: %d\n", catLow)
	fmt.Printf("  Нижче паритету: %d\n", catBeLowpar)
	fmt.Printf("  Вище паритету: %d\n", catAbovePar)
	fmt.Printf("  Високий: %d\n\n", catHigh)

	// --- Етап 5: Підрахунок повторів ---

	// Спосіб 1: Вкладеними циклами
	duplicatesLoop := 0
	for i := 0; i < len(values); i++ {
		isDuplicate := false
		// Дивимось усі попередні "вагончики" до поточного i
		for j := 0; j < i; j++ {
			if values[i] == values[j] {
				isDuplicate = true
				break // Знайшли повтор - далі перевіряти немає сенсу
			}
		}
		if isDuplicate {
			duplicatesLoop++
		}
	}

	// Спосіб 2: За допомогою map (словника
	freqMap := make(map[float64]int)
	for _, v := range values {
		freqMap[v]++ // Рахуємо, скільки разів випало кожне число
	}

	duplicatesMap := 0
	for _, countFreq := range freqMap {
		if countFreq > 1 {
			//Якщо число випало 3 рази, цу 2 повтори (3 - 1)
			duplicatesMap += (countFreq - 1)
		}
	}

	fmt.Printf("Кількість повторів (вкладені цикли): %d\n", duplicatesLoop)
	fmt.Printf("Кількість повторів (через map): %d\n\n", duplicatesMap)

	// --- Етап 6: Обов'язкові 0перевірки з умови завдання ---
	fmt.Println("--- ОБОВ'ЯЗКОВІ ПЕРЕВІРКИ ---")

	// Перевірка 1: Порожній рядок замість валюти має повернути помилку
	// (Використовуємо _ замість змінної values, бо нам цікава лише наявність помилки)
	_, errEmpty := exchange.Generate(10, "")
	if errEmpty != nil {
		fmt.Println("1. Порожня валюта повертає error: Успіх")
	} else {
		fmt.Println("1. Порожня валюта повертає error: Провал")
	}

	//Перевірка 2: GenerateWithSeed з однвковим аргументами
	seed1 := int64(12345)
	seedVals1, _ := exchange.GenerateWithSeed(5, "EUR/USD", seed1)
	seedVals2, _ := exchange.GenerateWithSeed(5, "EUR/USD", seed1)

	// Порівнюємо два зрізи
	isSame := true
	for i := 0; i < len(seedVals1); i++ {
		if seedVals1[i] != seedVals2[i] {
			isSame = false // Якщо хоча б одне число відрізняється
		}
	}
	if isSame {
		fmt.Println("2. Однакові seed дають однакові зрізи: УСПІХ")
	} else {
		fmt.Println("2. Однакові seed дають однакові зрізи: ПРОВАЛ")
	}

	// Перевірка 3: Зміна seed змінює послідовність
	seed2 := int64(99999)
	seedVals3, _ := exchange.GenerateWithSeed(5, "EUR/USD", seed2)

	isDifferent := false
	for i := 0; i < len(seedVals1); i++ {
		if seedVals1[i] != seedVals3[i] {
			isDifferent = true // Знайшли відмінність
		}
	}
	if isDifferent {
		fmt.Println("3. Зміна seed змінює зріз: УСПІХ")
	} else {
		fmt.Println("3. Зміна seed змінює зріз: ПРОВАЛ")
	}

	var trends []string
	for i := 1; i < len(values); i++ {
		if values[i] < values[i-1] {
			trends = append(trends, "Нижче")
		} else if values[i] == values[i-1] {
			trends = append(trends, "Таке ж")
		} else {
			trends = append(trends, "Вище")
		}
	}
	fmt.Println("Динаміка котирувань:", trends)
}
