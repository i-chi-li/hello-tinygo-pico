package main

import (
	"log"
	"time"
)

func main() {
	for {
		log.Println("Hello, Go!")
		time.Sleep(time.Second)
	}
}
