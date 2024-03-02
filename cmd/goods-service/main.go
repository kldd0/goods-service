package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/kldd0/goods-service/internal/clients/redis"
	"github.com/kldd0/goods-service/internal/config"
	"github.com/kldd0/goods-service/internal/http-server/handlers/good/get"
	"github.com/kldd0/goods-service/internal/http-server/handlers/good/patch"
	"github.com/kldd0/goods-service/internal/http-server/handlers/good/post"
	"github.com/kldd0/goods-service/internal/logger"
	"github.com/kldd0/goods-service/internal/storage/postgres"
	"github.com/nats-io/nats.go"
)

func main() {
	config := config.MustLoad()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.InitLogger(config.Env)

	log.Info(
		"starting goods-service",
		slog.String("env", config.Env),
		slog.String("ver", "1.0"),
	)
	log.Debug("debug messages are enabled")

	db, err := postgres.New(config.DBUri)
	if err != nil {
		log.Error("failed connecting to database", logger.Err(err))
	}
	defer db.Close()

	// init nats connection
	nc, err := nats.Connect(fmt.Sprintf("nats://%s", config.NATSAddr))
	if err != nil {
		log.Error("failed connecting to NATS", logger.Err(err))
	}
	defer nc.Flush()
	defer nc.Close()

	/*
		cm, err := sub.New(nc)

		if err != nil {
			log.Error("failed creating consumer", logger.Err(err))
		}

		// subscribe for messages
		ch, err := cm.Subscribe()
		if err != nil {
			log.Error("failed to subscribe to cluster", logger.Err(err))
		}

		// main service init
		svc := service.New(ctx, db)

		// start business logic
		go svc.Run(ch)
	*/

	// creating cache
	cache, err := redis.New(ctx, config)
	if err != nil {
		log.Info("failed connecting to redis", logger.Err(err))
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	// router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// healthcheck route
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(map[string]bool{
			"pong": true,
		})
	})

	router.Route("/good", func(r chi.Router) {
		r.Get("/{id}", get.New(log, db, cache))

		r.Post("/create", post.New(log, db))
		r.Patch("/update", patch.New(log, db, cache))
	})

	// router.Get("/goods/list")

	log.Info("starting http server", slog.String("address", config.HTTPServer.Address))

	// server configuration
	srv := &http.Server{
		Addr:         config.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  config.HTTPServer.Timeout,
		WriteTimeout: config.HTTPServer.Timeout,
		IdleTimeout:  config.HTTPServer.IdleTimeout,
	}

	// listen to OS signals and gracefully shutdown HTTP server
	done := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		log.Info("stopping server")

		if err := srv.Shutdown(ctx); err != nil {
			log.Info("http server shutdown error", logger.Err(err))
		}

		close(done)
	}()

	// start http server
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error("http server ListenAndServe error:", logger.Err(err))
	}

	<-done

	log.Info("http server stopped")
}
