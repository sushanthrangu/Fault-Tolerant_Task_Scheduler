package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"task-scheduler/internal/api"
	"task-scheduler/internal/config"
	mysqlrepo "task-scheduler/internal/repo/mysql"
)

func main() {
	cfg := config.Load()
	if cfg.DBDSN == "" {
		log.Fatal("DB_DSN is required")
	}

	db, err := mysqlrepo.Open(cfg.DBDSN)
	if err != nil {
		log.Fatalf("db open failed: %v", err)
	}
	defer db.Close()

	jobRepo := mysqlrepo.NewJobRepo(db)
	server := api.NewServer(jobRepo)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on http://localhost:%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down api...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}
