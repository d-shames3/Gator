package main

import (
	"fmt"
	"log"

	"github.com/d-shames3/gatorcli/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Start config: %v\n", cfg)

	err = cfg.SetUser("david")
	if err != nil {
		log.Fatal(err)
	}

	newConfig, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Current config: %v\n", newConfig)

}
