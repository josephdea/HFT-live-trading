package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Params struct {
	paramName   string
	params      map[string][]string
	orderedArgs []string
}

func NewParams() Params {
	p := Params{}
	p.paramName = "DefaultName"
	p.params = make(map[string][]string)
	p.orderedArgs = make([]string, 0)
	return p
}

func GetSignalIdentifier(p Params) string {
	res := p.paramName
	for _, val := range p.orderedArgs {
		if val != "paired" {
			res += "_" + p.params[val][0]
		}
	}
	return res
}

func GetBaseIdentifier(p Params) string {
	res := p.paramName
	val1, ok1 := p.params["exchange"]
	val2, ok2 := p.params["coin"]
	if ok1 && ok2 {
		res = res + "_" + val1[0] + "_" + val2[0]
		return res
	}
	for _, val := range p.orderedArgs {
		if val != "paired" {
			res += "_" + p.params[val][0]
		}
	}
	return res
}

func ConvertPairedStruct(p Params) []Params {
	_, ok := p.params["paired"]
	res := []Params{}
	if !ok {
		res = append(res, p)
		return res
	}
	exchange := p.params["exchange"]
	coin := p.params["coin"]
	valid_keys := make([]string, 0)
	i := 0
	for k := range p.params {
		if k != "exchange" && k != "paired" && k != "coin" {
			valid_keys = append(valid_keys, k)
			i++
		}

	}
	i = 0
out:
	for {
		tmp_params := NewParams()
		tmp_params.paramName = p.paramName
		tmp_params.orderedArgs = p.orderedArgs
		tmp_params.params["exchange"] = exchange
		tmp_params.params["coin"] = coin
		for k := range valid_keys {
			key := valid_keys[k]
			if i >= len(p.params[key]) {
				break out
			}
			tmp_params.params[key] = []string{p.params[key][i]}
		}
		res = append(res, tmp_params)
		i++
	}

	//res.paramName = p.paramName
	//if(p.params.fi)
	return res
}

func ParseConfig(filename string) []Params {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("config file does not exist", filename)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	allParams := []Params{}
	//scanner.Scan()
	p := NewParams()
	for scanner.Scan() {
		line := strings.ReplaceAll(scanner.Text(), " ", "")
		if len(line) == 0 {
			continue
		}
		colon_idx := strings.Index(line, ":")
		field_id := line[:colon_idx]
		params := strings.Split(line[colon_idx+1:], ",")
		if field_id == "Name" {

			if len(p.params) != 0 {
				for _, tmpParam := range ConvertPairedStruct(p) {
					allParams = append(allParams, tmpParam)
				}
				p = NewParams()
			}
			p.paramName = params[0]
		} else {
			//if field_id != "paired" {
			p.params[field_id] = params
			p.orderedArgs = append(p.orderedArgs, field_id)
			//}

		}
		//fmt.Println(line[:colon_idx])
		// next := scanner.Scan()
		// fmt.Println(next)
	}
	for _, tmpParam := range ConvertPairedStruct(p) {
		allParams = append(allParams, tmpParam)
	}
	// for _, val := range allParams {
	// 	fmt.Println(val)
	// }
	return allParams
}
