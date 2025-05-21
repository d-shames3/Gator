package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/d-shames3/gatorcli/internal/config"
	"github.com/d-shames3/gatorcli/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	err := c.commands[cmd.name](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) error {
	_, exists := c.commands[name]
	if exists {
		return fmt.Errorf("command already registered")
	}

	c.commands[name] = f
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no username arg provided for login")
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		fmt.Println("cannot login as an unregistered user - please register first")
		os.Exit(1)
		return nil
	}

	err = s.config.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Successfully logged in as user %s!\n", cmd.args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no username provided for registration")
	}

	registerUserParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}

	_, err := s.db.GetUser(context.Background(), registerUserParams.Name)
	if err == nil {
		fmt.Println("User already registered")
		os.Exit(1)
		return nil
	}

	user, err := s.db.CreateUser(context.Background(), registerUserParams)
	if err != nil {
		return fmt.Errorf("error registering user: %v", err)
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User %s was created!\n", user.Name)
	fmt.Printf("%s user data: %v\n", user.Name, user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		fmt.Printf("error resetting users: %v", err)
		os.Exit(1)
		return nil
	}

	fmt.Println("Successfully reset all users!")
	return nil
}
