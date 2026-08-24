package main

import (
	"log"
	"os"

	datasnoop "github.com/datasnoop/datasnoop/sdk/go"
)

func main() {
	if err := datasnoop.Demonstration(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
