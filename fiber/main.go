package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	hw "github.com/rujax/hard-worker"
	"github.com/rujax/hard-worker-examples/fiber/jobs"
)

const (
	maxWorkers = 100
	port       = 8080
)

func main() {
	workDispatcher := hw.NewDispatcher(maxWorkers)
	workers := workDispatcher.Run()

	app := fiber.New(fiber.Config{
		ReadTimeout: time.Second * 5,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		workDispatcher.JobQueue <- &jobs.GinJob{Message: "hard-worker is working!"}

		return c.SendString("Received a job.")
	})

	go func() {
		if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatalf("Server shutdown: %+v", err)
	}

	log.Println("Server exited.")
}
