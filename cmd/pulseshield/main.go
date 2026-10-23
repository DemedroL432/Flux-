package main

import (
	"log"

	appmain "pulseshield/app/main"
)

var version = "0.6.0"

func main() {
	if err := appmain.Run(version); err != nil {
		log.Fatal(err)
	}
}
