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

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("missing either name or url")
	}

	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("error add feed to database: %v", err)
	}

	fmt.Printf("%s feed successfully added to database for user %s\n", feed.Name, s.config.CurrentUserName)
	fmt.Printf("Full feed data: %v\n", feed)

	followFeedParams := database.CreateFeedFollowsParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	_, err = s.db.CreateFeedFollows(context.Background(), followFeedParams)
	if err != nil {
		return fmt.Errorf("error creating feed follow entry for user")
	}

	fmt.Printf("%s is now following %s feed", s.config.CurrentUserName, feed.Name)

	return nil
}

func handlerAgg(s *state, cmd command) error {
	const url string = "https://www.wagslane.dev/index.xml"

	rss, err := fetchFeed(context.Background(), url)
	if err != nil {
		return err
	}

	fmt.Printf("rss struct: %v", rss)
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching feeds from database: %v", err)
	}

	for _, feed := range feeds {
		fmt.Printf("* feed: %s, url: %s, user: %s\n", feed.Feed, feed.Url, feed.User)
	}

	return nil
}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("must provide feed url")
	}

	feed, err := s.db.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("feed not found, must add feed before following")
	}

	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Println("user is not yet following any feeds")
	}

	for _, userFeed := range feeds {
		if feed.ID == userFeed.ID {
			return fmt.Errorf("user is already following %s feed", feed.Name)
		}
	}

	followFeedParams := database.CreateFeedFollowsParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	feedFollow, err := s.db.CreateFeedFollows(context.Background(), followFeedParams)
	if err != nil {
		return fmt.Errorf("error following feed: %v", err)
	}

	fmt.Printf("User %s successfully followed feed %s!", feedFollow.User, feedFollow.Feed)
	return nil
}

func handlerFollowing(s *state, cmd command) error {
	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	feedsFollowing, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("user is not following any feeds")
	}

	fmt.Printf("User %s is following these feeds:\n", s.config.CurrentUserName)
	for _, feed := range feedsFollowing {
		fmt.Printf("* %s\n", feed.Feed)
	}

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

func handlerUsers(s *state, cmd command) error {
	currentUser := s.config.CurrentUserName

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching users: %v", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("no users to list")
	}

	for _, user := range users {
		if user == currentUser {
			user += " (current)"
		}
		fmt.Printf("* %s\n", user)
	}

	return nil
}
