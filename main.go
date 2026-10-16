package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/yukiisen/ohana/cache"
	"github.com/yukiisen/ohana/config"
	"github.com/yukiisen/ohana/handler"
	"github.com/yukiisen/ohana/logger"
	"github.com/yukiisen/ohana/paths"
	"github.com/yukiisen/ohana/worker"
)

func dataPath() string {
	return path.Join(os.Getenv("HOME"), ".local/share/osu/files")
}

//go:embed public/*
var publicFS embed.FS

func main() {
	mux := http.NewServeMux()

	configPath, err := paths.ConfigFile("config.json")
	if err != nil { log.Fatal(err) }

	config, err := config.Load(configPath)
	if err != nil { log.Fatal(err) }

	cachePath, err := paths.CacheFile("cache.json")
	if err != nil { log.Fatal(err) }

	store, err := cache.Load(cachePath)
	if err != nil {
		logger.Warnf("Failed to load cache.json, starting fresh: %v", err)
		store = cache.New(cachePath)
	}
	defer func() {
		if err := store.Save(); err != nil {
			logger.Errorf("Failed to persist cache: %v", err)
		}
	}()

	clientMan := handler.ClientHandler {}

	leaderboardWorker, err := worker.NewLeadboardWorker(config, store, &clientMan)
	if err != nil { log.Fatal(err) }
	defer leaderboardWorker.Close()

	gameWorker := worker.NewGameWorker(&clientMan, leaderboardWorker)

	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel();
	go gameWorker.Run(&wg, ctx);

	mux.HandleFunc("/ws", clientMan.Connection)

	mux.HandleFunc("PUT /leaderboard", gameWorker.ChangeLeaderboard)
	mux.HandleFunc("PUT /leaderboards", gameWorker.UpdateLeaderboards)
	mux.HandleFunc("GET /leaderboards", gameWorker.GetLeaderboards)

	mux.Handle("/lazer/",
		http.StripPrefix("/lazer/", http.FileServer(http.Dir(dataPath()))),
	)

	mux.Handle("/", http.FileServer(http.FS(publicFS)))

	server := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		logger.Infof("Listening on port %s", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s_ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(s_ctx)
}
