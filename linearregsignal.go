package main

import (
	"fmt"
	"os"
	"strconv"
)

type LinearSignal struct {
	nexus        *Nexus
	intercept    float64
	idxes        []int
	coefficients []float64
	write_idx    int
}

func NewLinearSignal(n *Nexus, w_idx int) *LinearSignal {
	s := LinearSignal{}
	s.nexus = n
	s.intercept = 0.0
	s.idxes = make([]int, 0)
	s.coefficients = make([]float64, 0)
	s.write_idx = w_idx
	return &s
}
func (s *LinearSignal) initParam(param map[string][]string) {
	signalList, ok := param["Signals"]
	if !ok {
		fmt.Println("no Signals in linear signal")
		os.Exit(1)
	}
	for _, val := range signalList {
		_, ok = s.nexus.UniqueSignals[val]
		if !ok {
			fmt.Println("no", val, "detected in linear regression lookup")
		}
		s.idxes = append(s.idxes, s.nexus.UniqueSignals[val])
	}
	coeffList, ok := param["Coefficients"]
	for _, val := range coeffList {
		val_f, _ := strconv.ParseFloat(val, 64)
		s.coefficients = append(s.coefficients, val_f)
	}
	val, ok := param["Intercept"]
	val_f, _ := strconv.ParseFloat(val[0], 64)
	if !ok {
		fmt.Println("no Intercept for LinearSignal")
		os.Exit(1)
	}
	s.intercept = val_f
}

func (s *LinearSignal) preprocess() {
	return
}

func (s *LinearSignal) forward() {
	res := s.intercept
	for idx, val := range s.coefficients {
		res += s.nexus.SignalOut[s.idxes[idx]] * val
	}
	s.nexus.SignalOut[s.write_idx] = res
}
