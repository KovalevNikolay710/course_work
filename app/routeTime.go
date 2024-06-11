package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const apiKey = "Ahf1n8TklGr8yzG0t9nbehYvG0j3an4Ox2umkDNyDe5Unoqbb-9zSnRbFOUJHdKl"

var points = map[int]string{
	1: "55.816753,37.646121",
	2: "55.816895,37.660638",
	3: "55.810956,37.643093",
	4: "55.808123,37.652310",
	5: "55.804789,37.642483",
	6: "55.800838,37.639374",
	7: "55.799077,37.646130",
}

var edges = []string{
	"1:2",
	"1:3",
	"3:2",
	"3:5",
	"5:4",
	"4:2",
	"5:6",
	"6:7",
	"7:4",
	"3:4",
}

var times = []string{"09:00", "12:00", "15:00", "18:00", "20:00", "23:00"}
var dates = generateNewDates()

var sem = make(chan struct{}, 5) // Семафор с лимитом на 5 одновременных запросов

// GetRouteDurationForEachEdge
func smain() {

	matrix := make([][]string, len(edges))
	for i := range matrix {
		matrix[i] = make([]string, 31)
		matrix[i][0] = edges[i]
	}

	var wg sync.WaitGroup

	for i, edge := range edges {
		pointsSplit := strings.Split(edge, ":")
		originIndex, _ := strconv.Atoi(pointsSplit[0])
		destinationIndex, _ := strconv.Atoi(pointsSplit[1])
		origin := points[originIndex]
		destination := points[destinationIndex]
		for j, date := range dates {
			for k, time := range times {
				dateTime := fmt.Sprintf("%sT%s:00Z", date, time)
				wg.Add(1)
				sem <- struct{}{} // Захватываем слот
				go getRouteDuration(origin, destination, dateTime, i, j*len(times)+k+1, &wg, matrix)
			}
		}
	}

	wg.Wait()

	newFile, err := os.Create("./data/new_data.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()

	writer := csv.NewWriter(newFile)
	defer writer.Flush()

	headers := generateCSVHeaders(dates, times)
	for _, header := range headers {
		writer.Write(header)
	}

	for _, row := range matrix {
		writer.Write(row)
	}
}

type DistanceMatrixResponse struct {
	ResourceSets []struct {
		Resources []struct {
			Results []struct {
				TravelDuration float64 `json:"travelDuration"`
			} `json:"results"`
		} `json:"resources"`
	} `json:"resourceSets"`
}

// urls counter
var amount, counter int = len(edges) * len(dates) * len(times), 0

func getRouteDuration(origin, destination, dateTime string, edgeIndex, cellIndex int, wg *sync.WaitGroup, matrix [][]string) {
	defer wg.Done()
	defer func() { <-sem }() // Освобождаем слот

	travelMode := "driving"
	time.Sleep(500 * time.Millisecond)
	url := fmt.Sprintf("https://dev.virtualearth.net/REST/v1/Routes/DistanceMatrix?origins=%s&destinations=%s&traffic=true&startTime=%s&travelMode=%s&key=%s", origin, destination, dateTime, travelMode, apiKey)
	fmt.Printf("%d/%d Request URL: %s\n", counter, amount, url)
	counter++
	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error: received non-200 response code %d\n", resp.StatusCode)
		matrix[edgeIndex][cellIndex] = "Ошибка раз"
		return
	}

	var data DistanceMatrixResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Println("Error decoding response:", err)
		matrix[edgeIndex][cellIndex] = "Ошибка два"
		return
	}

	if len(data.ResourceSets) > 0 && len(data.ResourceSets[0].Resources) > 0 && len(data.ResourceSets[0].Resources[0].Results) > 0 {
		duration := data.ResourceSets[0].Resources[0].Results[0].TravelDuration
		matrix[edgeIndex][cellIndex] = fmt.Sprintf("%.0f", duration)
	} else {
		log.Printf("No data found for %s to %s at %s", origin, destination, dateTime)
		matrix[edgeIndex][cellIndex] = "Ошибка три"
	}
}

func generateNewDates() []string {
	now := time.Now()
	var dates []string
	for i := 0; i < 5; i++ {
		date := now.AddDate(0, 0, -3*i)
		formattedDate := date.Format("2006-01-02")
		dates = append(dates, formattedDate)
	}
	return dates
}

func generateCSVHeaders(dates []string, times []string) [][]string {
	var headers [][]string
	header1 := []string{"Множество рёбер"}
	header2 := []string{""}
	header3 := []string{""}
	for _, date := range dates {
		header1 = append(header1, "Дата")
		header2 = append(header2, date)
		header3 = append(header3, "Время")
		for i := 0; i <= len(times)-2; i++ {
			header1 = append(header1, "")
			header2 = append(header2, "")
			header3 = append(header3, "")
		}
	}
	headers = append(headers, header1)
	headers = append(headers, header2)
	headers = append(headers, header3)
	header4 := []string{""}
	for range dates {
		for _, time := range times {
			header4 = append(header4, time)
		}
	}
	headers = append(headers, header4)

	return headers
}
