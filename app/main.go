package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var peaks []string

func main() {
	createHTMLFile("Результаты")

	// Проверка наличия папки и создание её, если нет
	path := "./data"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			log.Fatalf("Ошибка при создании директории: %v", err)
		}
		fmt.Println("Директория 'data' создана. Начинаем процесс создания новых данных.")
		generateNewTable()
		processData("./data/new_data.csv")
	} else {
		files, err := listFilesInDirectory(path)
		if err != nil {
			log.Fatalf("Ошибка при получении списка файлов: %v", err)
		}

		if len(files) == 0 {
			fmt.Println("Директория 'data' пуста. Начинаем процесс создания новых данных.")
			generateNewTable()
			processData("./data/new_data.csv")
		} else {
			var choice string
			fmt.Println("Директория 'data' не пуста.")
			fmt.Println("Введите '1' для выбора файла или '2' для создания новой таблицы:")
			fmt.Scan(&choice)

			if choice == "1" {
				fmt.Println("Доступные файлы:")
				for i, file := range files {
					fmt.Printf("%d: %s\n", i+1, file)
				}
				var fileChoice int
				fmt.Println("Введите номер файла, который вы хотите загрузить:")
				fmt.Scan(&fileChoice)
				if fileChoice < 1 || fileChoice > len(files) {
					log.Fatalf("Некорректный выбор файла")
				}
				selectedFile := filepath.Join(path, files[fileChoice-1])
				_, err := loadDataFromFile(selectedFile)
				if err != nil {
					log.Fatalf("Ошибка при загрузке данных из файла: %v", err)
				}
				processData(selectedFile)
			} else if choice == "2" {
				generateNewTable()
				processData("./data/new_data.csv")
			} else {
				log.Fatalf("Некорректный выбор: %s", choice)
			}
		}
	}
}

func processData(filePath string) {
	// Вывод таблицы в браузер
	data, err := loadDataFromFile(filePath)
	if err != nil {
		log.Fatalf("Ошибка при загрузке данных из файла: %v", err)
	}
	appendTableToHTML("Исходная таблица данных", data)

	// Повторное чтение данных из файла
	dataFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Unable to read input file %s: %v", filePath, err)
	}
	defer dataFile.Close()

	reader := csv.NewReader(dataFile)
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
	edges, peaks = extractEdgesAndPoints(data)
	log.Printf("Edges: %v, Peaks: %v", edges, peaks)

	// Вычисление результатов
	results1, results2 := calculateResults(data)
	distribution := calculateDistribution(results1, results2, edges)
	appendImageToHTML("Граф", "graph.png")
	// Вывод результатов в браузер
	appendTableToHTML("Результаты 1", results1)
	appendTableToHTML("Результаты 2", results2)
	appendTableToHTML("Распределение", distribution)

	// Генерация случайного числа
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Float64()
	appendRandomNumberToHTML(randomNumber)

	randomNetwork := generateRandomNetwork(peaks, distribution, randomNumber)
	appendTableToHTML("Случайная сеть", randomNetwork)
	distMatrix := dijkstraAll(randomNetwork)
	appendTableToHTML("Матрица расстояний", distMatrix)

	extRad, intRad := calculateExternalDistances(distMatrix), calculateInternalDistances(distMatrix)
	extIntTable := calculateAndHighlightModelingResults(intRad, extRad, peaks)
	appendTableToHTML("Результаты модуляции", extIntTable)

	histogramFilename := "histogram.png"
	createHistogram(extIntTable, "Гистограмма суммы радиусов", histogramFilename)
	appendImageToHTML("Гистограмма суммы радиусов", histogramFilename)

	log.Printf("Generating full histogram with %d peaks", len(peaks))
	generateFullHist(results1, results2, "Гистограмма рамещения", "full_histogram.png")
	appendImageToHTML("Гистограмма рамещения", "full_histogram.png")

	err = openBrowser("results.html")
	if err != nil {
		log.Fatalf("Unable to open HTML file in browser: %v", err)
	}
}

// Функция для загрузки данных из файла
func loadDataFromFile(filename string) ([][]string, error) {
	dataFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer dataFile.Close()

	reader := csv.NewReader(dataFile)
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
	return data, nil
}

// Функция для получения списка файлов в директории
func listFilesInDirectory(directory string) ([]string, error) {
	var files []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, info.Name())
		}
		return nil
	})
	return files, err
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
}

// Функция для открытия файла в браузере
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
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

// Функция для создания гистограммы
func createHistogram(data [][]string, title string, filename string) {
	// Пропустить заголовок и первую строку с названиями столбцов
	values := make([]float64, len(data)-1)
	for i, row := range data[1:] {
		// Очистка значения от HTML-тегов перед парсингом
		cleanValue := stripHTMLTags(row[3]) // Используем столбец "Сумма радиусов"
		value, err := strconv.ParseFloat(cleanValue, 64)
		if err != nil {
			log.Fatalf("Unable to parse value from data: %v", err)
		}
		values[i] = value
	}

	// Создание данных для гистограммы
	barValues := make(plotter.Values, len(values))
	for i, v := range values {
		barValues[i] = v
	}

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Номер вершины"
	p.Y.Label.Text = "Сумма радиусов"
	p.Y.Min = 0 // Установить минимум оси Y на 0

	bars, err := plotter.NewBarChart(barValues, vg.Points(20))
	if err != nil {
		log.Fatalf("Unable to create bar chart: %v", err)
	}

	// Настройка меток оси X
	labels := make([]string, len(values))
	for i := range labels {
		labels[i] = strconv.Itoa(i + 1)
	}
	p.NominalX(labels...)

	p.Add(bars)

	// Найти индекс минимального значения
	minIndex := 0
	for i, v := range barValues {
		if v < barValues[minIndex] {
			minIndex = i
		}
	}

	highlight, err := plotter.NewBarChart(plotter.Values{barValues[minIndex]}, vg.Points(20))
	if err != nil {
		log.Fatalf("Unable to create highlight bar: %v", err)
	}

	p.Add(highlight)

	if err := p.Save(8*vg.Inch, 4*vg.Inch, filename); err != nil {
		log.Fatalf("Unable to save bar chart: %v", err)
	}
}

// Функция для добавления изображения в HTML
func appendImageToHTML(title, filename string) {
	htmlFile, err := os.OpenFile("results.html", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to open HTML file: %v", err)
	}
	defer htmlFile.Close()

	htmlContent := fmt.Sprintf(`
	<h2>%s</h2>
	<img src="%s" alt="%s">
`, title, filename, title)

	_, err = htmlFile.WriteString(htmlContent)
	if err != nil {
		log.Fatalf("Unable to write to HTML file: %v", err)
	}

	log.Printf("Added image %s to HTML file with title %s", filename, title)
}

// Функция для удаления HTML-тегов
func stripHTMLTags(input string) string {
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(input, "")
}

func generateFullHist(results1, results2 [][]string, title, filename string) {
	pointSet := make(map[string]int, len(points))
	for _, val := range points {
		pointSet[val] = 0
	}

	// Simulation to populate pointSet
	for i := 0; i < 10001; i++ {
		distribution := calculateDistribution(results1, results2, edges)
		randomNumber := rand.Float64()
		randomNetwork := generateRandomNetwork(peaks, distribution, randomNumber)
		distMatrix := dijkstraAll(randomNetwork)
		extRad, intRad := calculateExternalDistances(distMatrix), calculateInternalDistances(distMatrix)
		extIntTable := calculateAndHighlightModelingResults(intRad, extRad, peaks)
		ind, err := extractValueFromFirstColumn(extIntTable)
		if err != nil {
			log.Fatalf("Unable to extract value from table: %v", err)
		}
		pointSet[ind]++
	}

	// Find the maximum index
	maxIndex := 0
	for k := range pointSet {
		i, err := strconv.Atoi(k)
		if err != nil {
			log.Printf("Invalid index %s: %v", k, err)
			continue
		}
		if i > maxIndex {
			maxIndex = i
		}
	}

	// Create data for the histogram
	barValues := make(plotter.Values, maxIndex)
	for k, v := range pointSet {
		i, err := strconv.Atoi(k)
		if err != nil {
			log.Printf("Invalid index %s: %v", k, err)
			continue
		}
		if i-1 >= len(barValues) {
			log.Printf("Index %d is out of bounds for barValues of length %d", i, len(barValues))
			continue
		}
		barValues[i-1] = float64(v)
	}

	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = "Номер вершины"
	p.Y.Label.Text = "Эффективность расположения"
	p.Y.Min = 0 // Set the minimum Y axis value to 0

	bars, err := plotter.NewBarChart(barValues, vg.Points(20))
	if err != nil {
		log.Fatalf("Unable to create bar chart: %v", err)
	}

	// Set X-axis labels
	labels := make([]string, len(barValues))
	for i := range labels {
		labels[i] = strconv.Itoa(i + 1) // Correctly match labels with indices
	}
	p.NominalX(labels...)

	p.Add(bars)

	if err := p.Save(8*vg.Inch, 4*vg.Inch, filename); err != nil {
		log.Fatalf("Unable to save bar chart: %v", err)
	}
}

func extractValueFromFirstColumn(matrix [][]string) (string, error) {
	var value string

	// Регулярное выражение для поиска конструкции <b>...</b>
	re := regexp.MustCompile(`<b>([^<]+)</b>`)

	for _, row := range matrix {
		if len(row) == 0 || len(row) == 8 {
			continue
		}

		// Проверяем первый столбец на наличие конструкции <b>...</b>
		match := re.FindStringSubmatch(row[0])
		if match != nil {
			value = match[1]
			break
		}
	}

	return value, nil
}
