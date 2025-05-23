package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/d-shames3/gatorcli/internal/config"
	"github.com/d-shames3/gatorcli/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)
	st := state{dbQueries, &cfg}
	cmds := commands{make(map[string]func(*state, command) error)}

	err = cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.register("agg", handlerAgg)
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.register("feeds", handlerFeeds)
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.register("follow", middlewareLoggedIn(handlerFollow))
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.register("following", middlewareLoggedIn(handlerFollowing))
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.register("login", handlerLogin)
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.register("register", handlerRegister)
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.register("reset", handlerReset)
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.register("users", handlerUsers)
	if err != nil {
		log.Fatal(err)
	}

	err = cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
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
}
