package main

import (
	"fmt"
	"math"
	"strings"
)

type GeometricSignal struct {
	nexus           *Nexus
	exchange        string
	coin            string
	spreadIdx       int
	paramToIncoming map[int]int
	num_signals     int
	write_idx       int
}

func NewGeometricSignal(n *Nexus, w_idx int) *GeometricSignal {
	s := GeometricSignal{}
	s.nexus = n
	s.paramToIncoming = make(map[int]int, 0)
	s.write_idx = w_idx
	return &s
}

func (s *GeometricSignal) initParam(param map[string][]string) {
	//GeometricSignal_TradeSignal_BNB_BTC-USD_0_0_1e0_1_1_0.000000_3.000000
	signalList, ok := param["inc_signal"]
	s.num_signals = len(signalList)
	if !ok {
		fmt.Println("not inc_signal in geometric signal")
	}
	for idx, val := range signalList {
		s.paramToIncoming[idx] = s.nexus.UniqueSignals[val]
		//GeometricSignal_TradeSignal_BNB_BTC-USD_0_0_1e0_1_1_0.000000_3.000000
		s.nexus.UniqueSignals["GeometricSignal_"+val] = s.write_idx + idx
		s.nexus.OrderedSignalNames = append(s.nexus.OrderedSignalNames, "GeometricSignal_"+val)
		//n.UniqueSignals[GetSignalIdentifier(val)] = w_idx
	}
	for key, _ := range s.nexus.UniqueSignals {
		if strings.Contains(key, "EffectiveSpread") {
			s.spreadIdx = s.nexus.UniqueSignals[key]
		}
	}
}

func (s *GeometricSignal) preprocess() {
	return
}

func (s *GeometricSignal) forward() {
	effective_spread := s.nexus.SignalOut[s.spreadIdx]
	for key, val := range s.paramToIncoming {
		inc_signal_val := s.nexus.SignalOut[val]
		res := 0.0
		if inc_signal_val < 0 {
			res = -1 * math.Pow(math.Abs(inc_signal_val)*effective_spread, 0.5)
		} else {
			res = math.Pow(math.Abs(inc_signal_val)*effective_spread, 0.5)
		}
		s.nexus.SignalOut[s.write_idx+key] = res
	}
}
