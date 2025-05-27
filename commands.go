package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
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
	_, exists := c.commands[cmd.name]
	if !exists {
		return fmt.Errorf("invalid command, try again")
	}
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

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("missing either name or url")
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
	if len(cmd.args) == 0 {
		return fmt.Errorf("must provide a time between requests duration, formatted like 1s, 1m, 1h")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("error parsing time between requests duration - ensure formatting is similar to 1s, 1m, 1h, etc")
	}

	ticker := time.NewTicker(timeBetweenReqs)
	fmt.Printf("Collecting feeds every %v\n", timeBetweenReqs)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s.db)
		if err != nil {
			return err
		}
	}
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var err error
	limit := 2
	if len(cmd.args) > 0 {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("error parsing limit")
		}
	}

	userPostParams := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	userPosts, err := s.db.GetPostsForUser(context.Background(), userPostParams)
	if err != nil {
		return fmt.Errorf("error fetching posts for user %s: %v", s.config.CurrentUserName, err)
	}

	fmt.Printf("Successfully fetched posts for user %s!\n", s.config.CurrentUserName)

	for _, post := range userPosts {
		fmt.Printf("Feed: %s, Post Title: %s, Description: %s, URL: %s, Published At: %v\n", post.FeedName, post.PostTitle, post.Description.String, post.Url, post.PublishedAt.Time)
	}

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

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("must provide feed url")
	}

	feed, err := s.db.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("feed not found, must add feed before following")
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

	fmt.Printf("User %s successfully followed feed %s!\n", feedFollow.User, feedFollow.Feed)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feedsFollowing, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	if len(feedsFollowing) == 0 {
		fmt.Printf("User %s is not following any feeds\n", s.config.CurrentUserName)
		return nil
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

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("no feed url provided to unfollow")
	}

	feed, err := s.db.GetFeed(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("error fetching feed: %v", err)
	}

	unfollowParams := database.DeleteFeedFollowForUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.DeleteFeedFollowForUser(context.Background(), unfollowParams)
	if err != nil {
		return fmt.Errorf("error deleting feed %s for user %s", feed.Name, user.Name)
	}

	fmt.Printf("User %s is no longer following feed %s\n", user.Name, feed.Name)
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
