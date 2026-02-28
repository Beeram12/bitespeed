package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/Beeram12/bitespeed-identify/db"

	httpapi "github.com/Beeram12/bitespeed-identify/handler"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	store, err := db.NewPostgresStore(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer store.Close()

	router := gin.Default()

	httpapi.RegisterRoutes(router, store)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
