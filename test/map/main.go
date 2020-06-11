package main

import (
	"fmt"
	"sort"
)

func main() {
	m := make(map[string]string)

	m["3"] = "3"
	m["hello"] = "hello"
	m["4"] = "4"
	m["1"] = "1"
	m["2"] = "2"

	ar := make([]string, len(m))
	i := 0
	for _, n := range m {
		ar[i] = n
		i++
	}
	sort.Strings(ar)
	fmt.Println(ar)
}
