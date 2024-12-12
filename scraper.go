package main

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Cprakhar/rss-aggregator/internal/database"
	"github.com/google/uuid"
)

func startScraping(db *database.Queries, concurrency int, timeBetweenRequest time.Duration) {
	log.Printf("Scraping on %v goroutines every %s durations", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)

	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Printf("Couldn't scrap the rss feed: %v", err)
			continue
		}

		wg := &sync.WaitGroup{}

		for _, feed := range feeds {
			wg.Add(1)

			go scrapFeed(db, wg, feed)
		}
		wg.Wait()
	}
}

func scrapFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()
	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Error marking feed as fetched: %v", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Printf("Error fetching the feed: %v", err)
		return
	}
	for _, item := range rssFeed.Channel.Item {

		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}

		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("Error parsing date %v, error: %v", item.PubDate, err)
			continue
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			PublishedAt: pubDate,
			FeedID: feed.ID,
			Title: item.Title,
			Description: description,
			Url: item.Link,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Printf("Failed to create post: %v", err)
		}
	}
	log.Printf("Feed %s collected, %d posts found", feed.Name, len(rssFeed.Channel.Item))
}