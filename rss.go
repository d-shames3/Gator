package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/d-shames3/gator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func scrapeFeeds(db *database.Queries) error {
	feed, err := db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error getting next feed to fetch: %v", err)
	}

	markedFeed, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %v", err)
	}

	fmt.Printf("Successfully marked feed %s as last fetched %v!\n", markedFeed.Name, markedFeed.LastFetchedAt.Time)

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("error fetching RSS feed %s: %v", markedFeed.Name, err)
	}

	fmt.Printf("Successfully fetched RSS feed %s!\n", rssFeed.Channel.Title)

	for _, post := range rssFeed.Channel.Item {
		validDesc := true
		if post.Description == "" {
			validDesc = false
		}

		timeFormats := []string{time.RFC1123, time.RFC1123Z, time.RFC822, time.RFC822Z, time.RFC850, time.RFC3339, time.RFC3339Nano, time.ANSIC, time.UnixDate, time.RubyDate}
		validTime := false
		var publishedAt time.Time
		for _, timeFormat := range timeFormats {
			publishedAt, err = time.Parse(timeFormat, post.PubDate)
			if err == nil {
				validTime = true
				break
			}
		}

		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       post.Title,
			Url:         post.Link,
			Description: sql.NullString{String: post.Description, Valid: validDesc},
			PublishedAt: sql.NullTime{Time: publishedAt, Valid: validTime},
			FeedID:      feed.ID,
		}

		savedPost, err := db.CreatePost(context.Background(), postParams)
		if err != nil {
			pqErr, ok := err.(*pq.Error)
			if !ok {
				return fmt.Errorf("error parsing sql error: %v", err)
			}
			if pqErr.Code == "23505" && pqErr.Table == "posts" && pqErr.Constraint == "posts_url_key" {
				continue
			} else {
				return fmt.Errorf("error code %s from operation on %s table and %s constraint", pqErr.Code, pqErr.Table, pqErr.Constraint)
			}
		}
		fmt.Printf("Successfully saved post %v in db (link: %v)!\n", savedPost.Title, savedPost.Url)
	}

	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error creating fetch RSS feed request: %v", err)
	}

	req.Header.Set("User-Agent", "gator")
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error executing RSS feed request: %v", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error reading RSS feed response body: %v", err)
	}

	var rss *RSSFeed
	if err = xml.Unmarshal(body, &rss); err != nil {
		return &RSSFeed{}, fmt.Errorf("error decoding RSS feed xml: %v", err)
	}

	rss.Channel.Title = html.UnescapeString(rss.Channel.Title)
	rss.Channel.Description = html.UnescapeString(rss.Channel.Description)

	for i, item := range rss.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		rss.Channel.Item[i] = item
	}

	return rss, nil
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}
