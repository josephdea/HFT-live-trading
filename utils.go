package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ConcatMultipleSlices[T any](slices [][]T) []T {
	var totalLen int

	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]T, totalLen)

	var i int

	for _, s := range slices {
		i += copy(result[i:], s)
	}

	return result
}

func Insert[T any](a []T, index int, value T) []T {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

// remove value at index of splice
func Remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

func CarvePath(path string, coin string) string {
	curTime := time.Now()
	os.MkdirAll(path, os.ModePerm)
	curDate := curTime.Format("01-02-2006")
	coinPath := filepath.Join(path, coin)
	os.MkdirAll(coinPath, os.ModePerm)
	datePath := filepath.Join(coinPath, curDate)
	os.MkdirAll(datePath, os.ModePerm)
	return datePath
}

func CarvePathDate(path string, coin string, exchange string) string {
	curTime := time.Now()
	os.MkdirAll(path, os.ModePerm)
	curDate := curTime.Format("01-02-2006")
	datePath := filepath.Join(path, curDate)
	os.MkdirAll(datePath, os.ModePerm)
	coinPath := filepath.Join(datePath, coin)
	os.MkdirAll(coinPath, os.ModePerm)
	exchPath := filepath.Join(coinPath, exchange)
	os.MkdirAll(exchPath, os.ModePerm)

	return exchPath
}

func FloatArrtoStringArr(arr []float64) []string {
	res := make([]string, len(arr))
	for i, val := range arr {
		res[i] = fmt.Sprintf("%f", val)
	}
	return res
}

func Write_CSV(path string, fileName string, data [][]float64, timeStamps []time.Time, columns []string) {
	csvFile, _ := os.Create(path + "/" + fileName)
	csvwriter := csv.NewWriter(csvFile)
	_ = csvwriter.Write(columns)
	for i, record := range data {
		strRecord := FloatArrtoStringArr(record)
		line := append([]string{timeStamps[i].String()}, strRecord...)
		_ = csvwriter.Write(line)
	}
	csvwriter.Flush()
	csvFile.Close()
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tHeapAlloc = %v MiB", bToMb(m.HeapAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}
func removeDuplicate[T string | float64](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
