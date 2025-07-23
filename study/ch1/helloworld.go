package main

import (
	"github.com/tobgu/qframe"
)

func main() {
	df := qframe.New(map[string]interface{}{
		"A": []int{1, 2, 3},
		"B": []int{4, 5, 6},
	})
	// 加法 (A + B)
	df.Select("A")+df.Select("B")
	

}
