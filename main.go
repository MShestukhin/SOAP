package main

import (
	"math/rand"
	"time"
)

func main() {
	// fmt.Println("ok")
	rand.Seed(time.Now().UnixNano())
	start()
}
