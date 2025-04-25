package main

import (
	"BloomFilter/pkg"
	"fmt"
)

func main() {
	bf := pkg.NewBloomFilter(1000)
	bf.Add("hello")

	fmt.Println("Exists:", bf.Exists("hello")) // true
	fmt.Println("Exists:", bf.Exists("world")) // false
}
