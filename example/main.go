package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	starter "github.com/anatollupacescu/curly-spin"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Start server
	e.Logger.Fatal(e.Start(":1323"))

	// Handler

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)

	db := starter.NewFn("database", dbFn)
	httpc := starter.NewFn("http", httpFn)

	httpc.DependsOn(db)

	errC := starter.Run(ctx, db, httpc)

	// Routes
	e.GET("/stop/:id", func(c echo.Context) error {
		id := c.Param("id")
		switch id {
		case "database":
		case "http":
		default:
			panic(id)
		}
		return c.String(http.StatusOK, id)
	})

	interruptChannel := make(chan os.Signal, 1)

	go func() {
		signal.Notify(interruptChannel, syscall.SIGINT, syscall.SIGTERM)
	}()

	select {
	case <-ctx.Done():
	case <-interruptChannel:
		cancel()
	}

	select {
	case <-errC:
		cancel()
		log.Println("done")
	case <-time.After(6 * time.Second):
		log.Println("timeout, exiting with running services")
	case <-interruptChannel:
		log.Println("killed")
	}
}

func dbFn(ctx context.Context) error {
	return nil
}

func httpFn(ctx context.Context) error {
	return nil
}
