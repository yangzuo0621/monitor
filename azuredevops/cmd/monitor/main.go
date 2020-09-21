package main

import (
	"fmt"
	"time"
)

func main() {
	for true {
		fmt.Println("Run...")
		time.Sleep(5 * time.Second)
	}
}
