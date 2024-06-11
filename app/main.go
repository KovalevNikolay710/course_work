package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"
)

func main() {
	// Чтение данных из файла
	dataFile, err := os.Open("./data/new_data.csv")
	if err != nil {
		log.Fatalf("Unable to read input file %s: %v", "./data/new_data.csv", err)
	}
	defer dataFile.Close()

	reader := csv.NewReader(dataFile)

	// Чтение оставшихся данных с обработкой неправильного количества полей
	var data [][]string
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		if len(record) > 0 {
			data = append(data, record)
		}
	}

	// Создание общего HTML файла
	createHTMLFile("Результаты")

	// Вывод таблицы в браузер
	appendTableToHTML("Исходная таблица данных", data)

	// Повторное чтение данных из файла
	dataFile.Close()
	dataFile, err = os.Open("./data/new_data.csv")
	if err != nil {
		log.Fatalf("Unable to read input file %s: %v", "./data/new_data.csv", err)
	}
	defer dataFile.Close()

	reader = csv.NewReader(dataFile)
	// Пропустить первые три строки (заголовки)
	for i := 0; i < 4; i++ {
		_, err = reader.Read()
		if err != nil {
			log.Fatalf("Unable to parse file as CSV: %v", err)
		}
	}

	// Чтение оставшихся данных с обработкой неправильного количества полей
	data = nil
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		if len(record) > 0 {
			data = append(data, record)
		}
	}

	// Получение списка граней и точек
	edges, points := extractEdgesAndPoints(data)

	// Вычисление результатов
	results1, results2 := calculateResults(data)
	distribution := calculateDistribution(results1, results2, edges)

	// Вывод результатов в браузер
	appendTableToHTML("Результаты 1", results1)
	appendTableToHTML("Результаты 2", results2)
	appendTableToHTML("Распределение", distribution)
	// Генерация случайного числа
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Float64()
	appendRandomNumberToHTML(randomNumber)
	randomNetwork := generateRandomNetwork(points, distribution, randomNumber)
	appendTableToHTML("Случайная сеть", randomNetwork)
	distMatrix := dijkstraAll(randomNetwork)
	appendTableToHTML("Матрица расстояний", distMatrix)
	extRad, intRad := calculateExternalDistances(distMatrix), calculateInternalDistances(distMatrix)
	extIntTable := calculateAndHighlightModelingResults(intRad, extRad, points)
	appendTableToHTML("Результаты модуляции", extIntTable)
}

// Функция для создания общего HTML файла
func createHTMLFile(title string) {
	htmlFile, err := os.Create("results.html")
	if err != nil {
		log.Fatalf("Unable to create HTML file: %v", err)
	}
	defer htmlFile.Close()

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
</head>
<body>
	<h1>%s</h1>
`, title, title)

	_, err = htmlFile.WriteString(htmlContent)
	if err != nil {
		log.Fatalf("Unable to write to HTML file: %v", err)
	}
}

// Функция для добавления таблицы в общий HTML файл
func appendTableToHTML(title string, data [][]string) {
	htmlFile, err := os.OpenFile("results.html", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to open HTML file: %v", err)
	}
	defer htmlFile.Close()

	htmlContent := fmt.Sprintf(`
	<h2>%s</h2>
	<table border="1">
		<tr>`, title)

	// Добавление заголовков таблицы
	for _, header := range data[0] {
		htmlContent += fmt.Sprintf("<th>%s</th>", header)
	}
	htmlContent += "</tr>"

	// Добавление данных таблицы
	for _, row := range data[1:] {
		htmlContent += "<tr>"
		for _, cell := range row {
			htmlContent += fmt.Sprintf("<td>%s</td>", cell)
		}
		htmlContent += "</tr>"
	}

	htmlContent += `
	</table>
`

	_, err = htmlFile.WriteString(htmlContent)
	if err != nil {
		log.Fatalf("Unable to write to HTML file: %v", err)
	}

	// Открытие HTML файла в браузере, если это первая запись
	if title == "Исходная таблица данных" {
		err = openBrowser("results.html")
		if err != nil {
			log.Fatalf("Unable to open HTML file in browser: %v", err)
		}
	}
}

// Функция для открытия файла в браузере
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch os := runtime.GOOS; os {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd"
		args = append(args, "/c", "start")
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// Функция для извлечения граней и точек
func extractEdgesAndPoints(data [][]string) ([]string, []string) {
	edges := make([]string, len(data))
	pointsSet := make(map[string]struct{})

	for i, row := range data {
		edges[i] = row[0]
		parts := strings.Split(row[0], ":")
		if len(parts) == 2 {
			pointsSet[parts[0]] = struct{}{}
			pointsSet[parts[1]] = struct{}{}
		}
	}

	points := make([]string, 0, len(pointsSet))
	for point := range pointsSet {
		points = append(points, point)
	}
	sort.Strings(points)
	return edges, points
}

func appendRandomNumberToHTML(randomNumber float64) {
	htmlFile, err := os.OpenFile("results.html", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to open HTML file: %v", err)
	}
	defer htmlFile.Close()

	htmlContent := fmt.Sprintf(`
	<h2>Случайное число</h2>
	<p>%f</p>
`, randomNumber)

	_, err = htmlFile.WriteString(htmlContent)
	if err != nil {
		log.Fatalf("Unable to write to HTML file: %v", err)
	}
}

func radMatrix(ext, iter []float64, points []string) [][]string {
	matrixSize := len(points)
	radMatrix := make([][]string, matrixSize+1)

	// Инициализация первой строки заголовков
	radMatrix[0] = []string{"Point", "External Radius", "Internal Radius"}

	// Заполнение матрицы данными
	for i := 0; i < matrixSize; i++ {
		radMatrix[i+1] = []string{
			points[i],
			fmt.Sprintf("%.2f", ext[i]),
			fmt.Sprintf("%.2f", iter[i]),
		}
	}

	return radMatrix
}

func calculateModelingResults(internalDistances, externalDistances []float64) [][]string {
	// Инициализация первой строки заголовков
	results := [][]string{
		{"Название", "Результат"},
	}

	// Нахождение минимальных внешних и внутренних радиусов
	minInternalRadius := math.MaxFloat64
	minExternalRadius := math.MaxFloat64

	for i := range internalDistances {
		if internalDistances[i] < minInternalRadius {
			minInternalRadius = internalDistances[i]
		}
		if externalDistances[i] < minExternalRadius {
			minExternalRadius = externalDistances[i]
		}
	}

	// Нахождение суммы минимальных радиусов
	sumRadius := minInternalRadius + minExternalRadius

	// Формирование таблицы результатов
	results = append(results, []string{"Минимальный внешний радиус", fmt.Sprintf("%.2f", minExternalRadius)})
	results = append(results, []string{"Минимальный внутренний радиус", fmt.Sprintf("%.2f", minInternalRadius)})
	results = append(results, []string{"Сумма радиусов (внешний + внутренний)", fmt.Sprintf("%.2f", sumRadius)})

	return results
}

func calculateAndHighlightModelingResults(internalDistances, externalDistances []float64, points []string) [][]string {
	matrixSize := len(points)
	results := make([][]string, matrixSize+1)

	// Инициализация первой строки заголовков
	results[0] = []string{"Вершина", "Внешний радиус", "Внутренний радиус", "Сумма радиусов"}

	// Переменные для нахождения минимальной суммы радиусов
	minSumRadius := math.MaxFloat64
	minIndex := -1

	// Заполнение таблицы данными
	for i := 0; i < matrixSize; i++ {
		sumRadius := internalDistances[i] + externalDistances[i]
		results[i+1] = []string{
			points[i],
			fmt.Sprintf("%.2f", externalDistances[i]),
			fmt.Sprintf("%.2f", internalDistances[i]),
			fmt.Sprintf("%.2f", sumRadius),
		}
		if sumRadius < minSumRadius {
			minSumRadius = sumRadius
			minIndex = i + 1 // +1 для учета заголовков
		}
	}

	// Подсветка строки с минимальной суммой радиусов
	for i := range results[minIndex] {
		results[minIndex][i] = fmt.Sprintf("<b>%s</b>", results[minIndex][i])
	}

	return results
}
