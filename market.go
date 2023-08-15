package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Market struct {
	messageID   int
	updates     [][]float64
	updateTimes []time.Time
	record      bool
	writingctr  int
	prevUpdate  []float64
}

func NewMarket() *Market {
	m := Market{
		messageID:   -1,
		updates:     make([][]float64, 0),
		updateTimes: make([]time.Time, 0),
		record:      false,
		writingctr:  1,
	}
	return &m
}

func (m *Market) Update(timestamp time.Time, IndexPrice string, OraclePrice string, NextFundingRate string, OpenInterest string) {
	idxpx, _ := strconv.ParseFloat(IndexPrice, 64)
	if idxpx == 0 {
		idxpx = m.prevUpdate[0]
	}
	oraclepx, _ := strconv.ParseFloat(OraclePrice, 64)
	if oraclepx == 0 {
		oraclepx = m.prevUpdate[1]
	}
	nfr, _ := strconv.ParseFloat(NextFundingRate, 64)
	if nfr == 0 {
		nfr = m.prevUpdate[2]
	}
	oi, _ := strconv.ParseFloat(OpenInterest, 64)
	if oi == 0 {
		oi = m.prevUpdate[3]
	}
	newUpdate := []float64{idxpx, oraclepx, nfr, oi}
	//fmt.Println(idxpx, oraclepx, nfr, oi)
	// if reflect.DeepEqual(newUpdate, m.prevUpdate) {
	// 	return
	// }
	m.updates = append(m.updates, newUpdate)
	m.prevUpdate = newUpdate
	m.updateTimes = append(m.updateTimes, timestamp)
}
func (m *Market) WriteCSV(path string) {
	csvFile, _ := os.Create(path + "/" + "markets" + fmt.Sprint(m.writingctr) + ".csv")
	m.writingctr += 1
	csvwriter := csv.NewWriter(csvFile)
	curLength := len(m.updateTimes)
	//fmt.Println(len(t.updates), len(t.updateTimes))
	_ = csvwriter.Write([]string{"Time Received", "IndexPrice", "OraclePrice", "NextFundingRate", "OpenInterest"})
	loc, _ := time.LoadLocation("America/Chicago")
	for i := 0; i < curLength; i++ {
		//strRecord := FloatArrtoStringArr(record)
		line := append([]string{m.updateTimes[i].In(loc).String()}, FloatArrtoStringArr(m.updates[i][:])...)
		_ = csvwriter.Write(line)
		m.updates[i] = nil
	}
	m.updates = m.updates[curLength:]
	m.updateTimes = m.updateTimes[curLength:]
	csvwriter.Flush()
	csvFile.Close()

}
