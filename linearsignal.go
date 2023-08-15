package main

// type LinearSignal struct {
// 	signals   []func() float64
// 	weights   []float64
// 	outputs   []float64
// 	intercept float64
// }

// // func NewLinearSignal() *Signal {
// // 	var a Signal = LinearSignal{}
// // 	return &a
// // }

// // initialize(SignalNexus map[string]*Signal,args map[string]string)
// func (s *LinearSignal) initialize(N *Nexus, Args map[string]string) {
// 	for key, _ := range Args {
// 		_, ok := N.Signals[key]
// 		if ok {
// 			s.signals = append(s.signals, N.Signals[key])
// 			f, _ := strconv.ParseFloat(Args[key], 64)
// 			s.weights = append(s.weights, f)
// 		}
// 	}
// 	s.outputs = make([]float64, len(s.weights))
// 	f, _ := strconv.ParseFloat(Args["Intercept"], 64)
// 	s.intercept = f
// }
// func (s LinearSignal) forward() float64 {
// 	var wg sync.WaitGroup
// 	wg.Add(len(s.signals))
// 	for i := range s.signals {
// 		go func(x int) {
// 			s.outputs[x] = s.weights[x] * s.signals[x]()
// 			wg.Done()
// 		}(i)

// 	}
// 	var res float64 = s.intercept
// 	wg.Wait()
// 	for i := range s.outputs {
// 		res += s.outputs[i]
// 	}
// 	return res

// }
