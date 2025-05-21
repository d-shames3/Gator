package main

import (
	"fmt"
	"log"
	"os"

	"github.com/d-shames3/gatorcli/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Start config: %v\n", cfg)

	st := state{&cfg}
	cmds := commands{make(map[string]func(*state, command) error)}

	err = cmds.register("login", handlerLogin)
	if err != nil {
		log.Fatal(err)
	}

	argsRaw := os.Args
	if len(argsRaw) < 2 {
		log.Fatalln("no command line args provided")
	}

	args := make([]string, 0)
	cmdName := argsRaw[1]
	if len(argsRaw) > 2 {
		args = append(args, argsRaw[2:]...)
	}

	command := command{cmdName, args}
	err = cmds.run(&st, command)
	if err != nil {
		log.Fatal(err)
	}

	cfgNew, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("End config: %v\n", cfgNew)

}
