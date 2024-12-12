package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Cprakhar/rss-aggregator/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}


func main() {
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT is not found in the environment.")
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL is not found in the environment.")
	}

	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Error connecting to database", err)
	}

	db := database.New(conn)
	apiCfg := apiConfig{
		DB: db,
	}

	go startScraping(db, 10, time.Minute)

	router := chi.NewRouter()
	

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: false,
		MaxAge: 300,

	}))

	v1Router := chi.NewRouter()
	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerError)
	v1Router.Post("/users", apiCfg.handlerCreateUser)
	v1Router.Get("/users", apiCfg.middlewareHandler(apiCfg.handleGetUser))

	v1Router.Post("/feeds", apiCfg.middlewareHandler(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetFeeds)

	v1Router.Post("/feed_follows", apiCfg.middlewareHandler(apiCfg.handlerCreateFeedFollow))
	v1Router.Get("/feed_follows", apiCfg.middlewareHandler(apiCfg.handlerGetFeedFollows))
	v1Router.Delete("/feed_follows/{feed_id}", apiCfg.middlewareHandler(apiCfg.handleDeleteFeedFollow))

	v1Router.Get("/posts", apiCfg.middlewareHandler(apiCfg.handleGetPosts))

	router.Mount("/v1", v1Router)


	server := &http.Server {
		Handler: router,
		Addr: ":"+port,
	}

	log.Printf("Server start running on port: %v", port)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Port: ", port)
}