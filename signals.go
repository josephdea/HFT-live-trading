package main

type Signal interface {
	forward()
	preprocess()
	initParam(param map[string][]string)
}

type SignalStruct struct {
	signal Signal
}
