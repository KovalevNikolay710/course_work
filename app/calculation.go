package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"

	"gonum.org/v1/gonum/stat/distuv"
)

// I
// 1.E, мат ожидание
// 2.Omega, среднее математическое отклонение
// 3-4.Диапозон отклонения
// 5-8.O1234, Кол-во поподаний в каждый промежуток функции в диапозоне
// 9-12.E1234, Коэф-т значимости промежутка (0.16, 0.34)
// 13.X^2
// 14.(X^2)>=[X^2(приведен)]

// II
// 1.E, мат ожидание
// 2.Omega, среднее математическое отклонение
// 3.a
// 4.b
// 5-8.O1234, Кол-во поподаний в каждый промежуток функции в диапозоне
// 9-12.E1234, Коэф-т значимости промежутка (0.01, 0.49)
// 13.X^2
// 14.(X^2)>=[X^2(приведен)]

// III
// 1.Тип ГСЧ
//   Для равномерного Для нормального
// 2. 0				  E
// 3. 0				  Omega
// 4. 0				  z
// 5. 0				  a
// 6. 0				  b

// Выводим случайное число

// Генерирование случайной взвешенной сети

// Матрица расстояний

// Считаем внешний и внутренний радиус

// Выводим результаты моделирования

// Выводим таблицу внешне-внутрненних радиусов

// Выводим гистограмму по этой таблице

type edge struct {
	origin      string
	destination string
}

func generateRandomNetwork(points []string, distribution [][]string, randomNumber float64) [][]string {
	matrixSize := len(points)
	randomNetwork := make([][]string, matrixSize)
	for i := range randomNetwork {
		randomNetwork[i] = make([]string, matrixSize)
		for j := range randomNetwork[i] {
			randomNetwork[i][j] = "0" // Инициализация всех значений нулями
		}
	}

	for i := 0; i < matrixSize; i++ {
		for j := 0; j < matrixSize; j++ {
			if i == j {
				randomNetwork[i][j] = "0"
				continue
			}

			key, r_key := fmt.Sprintf("%d:%d", i+1, j+1), fmt.Sprintf("%d:%d", j+1, i+1)
			var dist []string
			for _, d := range distribution {
				if d[0] == key || d[0] == r_key {
					dist = d
					break
				}
			}

			if len(dist) == 0 {
				continue
			}

			if dist[1] == "нормальное" {
				E, _ := strconv.ParseFloat(dist[2], 64)
				Omega, _ := strconv.ParseFloat(dist[3], 64)
				z := randomNumber
				randomNetwork[i][j] = fmt.Sprintf("%.2f", E+Omega*z)
			} else {
				a, _ := strconv.ParseFloat(dist[5], 64)
				b, _ := strconv.ParseFloat(dist[6], 64)
				randomNetwork[i][j] = fmt.Sprintf("%.2f", a+(b-a)*randomNumber)
			}
			randomNetwork[j][i] = randomNetwork[i][j]
		}
	}

	return randomNetwork
}

func dijkstraAll(graph [][]string) [][]string {
	matrixSize := len(graph)
	distanceMatrix := make([][]string, matrixSize)
	for i := 0; i < matrixSize; i++ {
		distanceMatrix[i] = dijkstra(graph, i)
	}
	return distanceMatrix
}

func dijkstra(graph [][]string, start int) []string {
	matrixSize := len(graph)
	distances := make([]float64, matrixSize)
	visited := make([]bool, matrixSize)

	for i := range distances {
		distances[i] = math.Inf(1)
	}
	distances[start] = 0

	for i := 0; i < matrixSize; i++ {
		minDist := math.Inf(1)
		minIndex := -1
		for j := 0; j < matrixSize; j++ {
			if !visited[j] && distances[j] < minDist {
				minDist = distances[j]
				minIndex = j
			}
		}

		if minIndex == -1 {
			break
		}

		visited[minIndex] = true

		for j := 0; j < matrixSize; j++ {
			if !visited[j] && graph[minIndex][j] != "0" {
				edgeDist, _ := strconv.ParseFloat(graph[minIndex][j], 64)
				newDist := distances[minIndex] + edgeDist
				if newDist < distances[j] {
					distances[j] = newDist
				}
			}
		}
	}

	result := make([]string, matrixSize)
	for i, dist := range distances {
		if dist == math.Inf(1) {
			result[i] = "∞"
		} else {
			result[i] = fmt.Sprintf("%.2f", dist)
		}
	}
	return result
}

func a(E float64, Omega float64) float64 {
	return E - math.Sqrt(3)*Omega
}

func b(E float64, a float64) float64 {
	return E + math.Sqrt(3)*a
}

func chisqDistRT(x, k float64) float64 {
	if x < 0 || k <= 0 {
		return 0
	}

	chiSquared := distuv.ChiSquared{K: k}
	return 1 - chiSquared.CDF(x)
}

func Omega(data []int) float64 {
	n := len(data)
	if n == 0 {
		return 0
	}
	mean := 0.0
	for _, x := range data {
		mean += float64(x)
	}
	mean /= float64(n)

	variance := 0.0
	for _, x := range data {
		variance += math.Pow(float64(x)-mean, 2)
	}
	variance /= float64(n - 1)
	return math.Sqrt(variance)
}

func O(data []int, E float64, omega float64) (O1, O2, O3, O4 float64) {
	for _, x := range data {
		if float64(x) <= (E - omega) {
			O1++
		}
		if float64(x) > (E-omega) && float64(x) <= E {
			O2++
		}
		if float64(x) > E && float64(x) <= (E+omega) {
			O3++
		}
		if float64(x) > (E + omega) {
			O4++
		}
	}
	return
}

func sO(data []int, E float64, a float64, b float64) (O1, O2, O3, O4 float64) {
	for _, x := range data {
		if float64(x) <= a {
			O1++
		}
		if float64(x) > a && float64(x) <= E {
			O2++
		}
		if float64(x) > E && float64(x) <= E+b {
			O3++
		}
		if float64(x) > b {
			O4++
		}
	}
	return
}

func E(data []int, x1, x2 float64) (E1, E2, E3, E4 float64) {
	n := float64(len(data))
	E1 = x1 * n
	E2 = x2 * n
	E3 = x2 * n
	E4 = x1 * n
	return
}

func AVG(data []int) float64 {
	var sum int
	n := len(data)
	if n == 0 {
		return 0
	}
	for _, value := range data {
		sum += value
	}
	return float64(sum) / float64(n)
}

func calculateResults(data [][]string) ([][]string, [][]string) {
	var results1 [][]string
	var results2 [][]string

	header1 := []string{"ребро", "µ", "σ", "µ+σ", "µ-σ", "O1", "O2", "O3", "O4", "Хи2 норм.", "вероятность"}
	header2 := []string{"ребро", "µ", "σ", "a", "b", "O1", "O2", "O3", "O4", "Хи2 равн.", "вероятность"}
	results1 = append(results1, header1)
	results2 = append(results2, header2)

	for _, record := range data {
		intData := make([]int, len(record)-1)
		for i := 1; i < len(record); i++ {
			intValue, err := strconv.Atoi(record[i])
			if err != nil {
				log.Fatalf("Unable to convert %s to int: %v", record[i], err)
				continue
			}
			intData[i-1] = intValue
		}

		edge := record[0]

		avg := AVG(intData)
		omega := Omega(intData)

		O1, O2, O3, O4 := O(intData, avg, omega)
		E1, E2, E3, E4 := E(intData, 0.16, 0.34)
		x2 := (math.Pow(O1-E1, 2))/E1 + (math.Pow(O2-E2, 2))/E2 + (math.Pow(O3-E3, 2))/E3 + (math.Pow(O4-E4, 2))/E4
		p := chisqDistRT(x2, 1)

		results1 = append(results1, []string{
			edge,
			fmt.Sprintf("%.2f", avg),
			fmt.Sprintf("%.2f", omega),
			fmt.Sprintf("%.2f", avg-omega),
			fmt.Sprintf("%.2f", avg+omega),
			fmt.Sprintf("%.2f", O1),
			fmt.Sprintf("%.2f", O2),
			fmt.Sprintf("%.2f", O3),
			fmt.Sprintf("%.2f", O4),
			fmt.Sprintf("%.2f", x2),
			fmt.Sprintf("%.2f", p),
		})

		a := a(avg, omega)
		b := b(avg, a)
		O1, O2, O3, O4 = sO(intData, avg, a, b)
		E1, E2, E3, E4 = E(intData, 0.01, 0.49)
		x2 = (math.Pow(O1-E1, 2))/E1 + (math.Pow(O2-E2, 2))/E2 + (math.Pow(O3-E3, 2))/E3 + (math.Pow(O4-E4, 2))/E4
		p = chisqDistRT(x2, 1)

		results2 = append(results2, []string{
			edge,
			fmt.Sprintf("%.2f", avg),
			fmt.Sprintf("%.2f", omega),
			fmt.Sprintf("%.2f", a),
			fmt.Sprintf("%.2f", b),
			fmt.Sprintf("%.2f", O1),
			fmt.Sprintf("%.2f", O2),
			fmt.Sprintf("%.2f", O3),
			fmt.Sprintf("%.2f", O4),
			fmt.Sprintf("%.2f", x2),
			fmt.Sprintf("%.2f", p),
		})
	}

	return results1, results2
}

func calculateDistribution(results1, results2 [][]string, edges []string) [][]string {

	var distribution [][]string
	header := []string{"Ребро", "Распределение", "E", "Omega", "Рандом", "a", "b"}
	distribution = append(distribution, header)

	for i, edge := range edges {
		if i+1 < len(results1) && i+1 < len(results2) && len(results1[i+1]) > 9 && len(results2[i+1]) > 9 {
			pN, _ := strconv.ParseFloat(results1[i+1][9], 64)
			pR, _ := strconv.ParseFloat(results2[i+1][9], 64)
			if pN > pR {
				randValue := math.Sqrt(-2*math.Log(rand.Float64())) * math.Cos(2*math.Pi*rand.Float64())
				distribution = append(distribution, []string{
					edge,
					"нормальное",
					results1[i+1][0],
					results1[i+1][1],
					fmt.Sprintf("%f", randValue),
					"0",
					"0",
				})
			} else {
				distribution = append(distribution, []string{
					edge,
					"равномерное",
					"0",
					"0",
					"0",
					results2[i+1][2],
					results2[i+1][3],
				})
			}
		}
	}

	return distribution
}

func calculateModelingResults(internalDistances, externalDistances []float64) ([][]string, int) {
	results := [][]string{}
	minRadius := math.MaxFloat64
	minIndex := -1

	for i := range internalDistances {
		sumRadius := internalDistances[i] + externalDistances[i]
		results = append(results, []string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("%.2f", sumRadius),
		})
		if sumRadius < minRadius {
			minRadius = sumRadius
			minIndex = i
		}
	}

	return results, minIndex
}

func calculateInternalDistances(distanceMatrix [][]string) []float64 {
	matrixSize := len(distanceMatrix)
	internalDistances := make([]float64, matrixSize)
	for i := 0; i < matrixSize; i++ {
		maxDist := 0.0
		for j := 0; j < matrixSize; j++ {
			dist, _ := strconv.ParseFloat(distanceMatrix[i][j], 64)
			if dist > maxDist {
				maxDist = dist
			}
		}
		internalDistances[i] = maxDist
	}
	return internalDistances
}

func calculateExternalDistances(distanceMatrix [][]string) []float64 {
	matrixSize := len(distanceMatrix)
	externalDistances := make([]float64, matrixSize)
	for j := 0; j < matrixSize; j++ {
		maxDist := 0.0
		for i := 0; i < matrixSize; i++ {
			dist, _ := strconv.ParseFloat(distanceMatrix[i][j], 64)
			if dist > maxDist {
				maxDist = dist
			}
		}
		externalDistances[j] = maxDist
	}
	return externalDistances
}
