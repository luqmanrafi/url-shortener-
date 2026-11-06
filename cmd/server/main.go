package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/luqmanrafi/urlshortener/internal/handler"
	"github.com/luqmanrafi/urlshortener/internal/repository"
	"github.com/luqmanrafi/urlshortener/internal/service"
)

func main(){
	baseURL := getenv("BASE_URL", "http://localhost:3000")
	dbPath := getenv("DB_PATH", "urls.db")

	repo, err := repository.NewSQLite(dbPath)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	
	svc := service.New(repo)
	h := handler.New(svc, baseURL)

	srv := &http.Server{
		Addr: ":3000",
		Handler: h.Routes(),
		ReadHeaderTimeout: 5*time.Second,
	}

	log.Println("listening on port :3000")
	log.Fatal(srv.ListenAndServe())
}

func getenv(key, fallback string)string{
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}