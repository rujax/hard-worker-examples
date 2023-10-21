package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	hw "github.com/rujax/hard-worker"
	"github.com/rujax/hard-worker-examples/gin/jobs"
)

const (
	maxWorkers = 100
	port       = 8080
)

func main() {
	workDispatcher := hw.NewDispatcher(maxWorkers)
	workers := workDispatcher.Run()

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		workDispatcher.JobQueue <- &jobs.GinJob{Message: "hard-worker is working!"}

		c.String(http.StatusOK, "Received a job.")
	})

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Listen: %+v", err)
		}
	}()

	quit := make(chan os.Signal, 2)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	workDispatcher.Exit()

	log.Println("Dispatcher exited.")

	for _, worker := range workers {
		worker.Stop()
	}

	log.Println("Workers stopped.")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown: %+v", err)
	}

	log.Println("Server exited.")
}
