package main

// var registry map[string]string = map[string]string{"test":"sdfsdf",

func CreateEdges(SignalArr []string, SignalArgs map[string]map[string]string) map[string][]string {
	edges := make(map[string][]string)

	for _, val := range SignalArr {
		edges[val] = make([]string, 0)
	}
	for _, val := range SignalArr {
		for arg, _ := range SignalArgs[val] {
			_, ok := edges[arg]
			if ok {
				edges[arg] = append(edges[arg], val)
			}
		}
	}
	return edges
}

func TopoSort(Edges map[string][]string) []string {
	res := make([]string, 0)
	EdgeCtr := make(map[string]int)

	for key, _ := range Edges {
		EdgeCtr[key] = 0
	}
	for start, _ := range Edges {
		for _, end := range Edges[start] {
			EdgeCtr[end] += 1
		}
	}
	queue := make([]string, 0)
	for key, val := range EdgeCtr {
		if val == 0 {
			//res = append(res, key)
			queue = append(queue, key)
		}
	}
	for len(queue) != 0 {
		node := queue[0]
		res = append(res, node)
		queue = queue[1:]
		for _, nbr := range Edges[node] {
			EdgeCtr[nbr] -= 1
			if EdgeCtr[nbr] == 0 {
				queue = append(queue, nbr)
			}
		}
	}
	return res
}
