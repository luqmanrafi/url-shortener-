package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/luqmanrafi/urlshortener/internal/handler"
	"github.com/luqmanrafi/urlshortener/internal/middleware"
	"github.com/luqmanrafi/urlshortener/internal/repository"
	"github.com/luqmanrafi/urlshortener/internal/service"
)

func main(){
	port :=  getenv("PORT", "3000")
	baseURL := getenv("BASE_URL", "http://localhost:"+port)
	dbPath := getenv("DB_PATH", "urls.db")


	repo, err := repository.NewSQLite(dbPath)
	if err != nil {
		log.Fatalf("init database: %v", err)
	}
	
	svc := service.New(repo)
	h := handler.New(svc, baseURL)

	limiter := middleware.NewRateLimiter(5, 10)
	handler := middleware.Logging(limiter.Middleware(h.Routes()))
	srv := &http.Server{
		Addr: ":" + port,
		Handler: handler,
		ReadHeaderTimeout: 5*time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func(){
		log.Printf("listening on port :%s\n", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed){
			log.Fatalf("server error: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	if err:= srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func getenv(key, fallback string)string{
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}