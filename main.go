package main

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/yukiisen/hana/cache"
	"github.com/yukiisen/hana/config"
	"github.com/yukiisen/hana/handler"
	"github.com/yukiisen/hana/logger"
	"github.com/yukiisen/hana/paths"
	"github.com/yukiisen/hana/worker"
)

func dataPath() string {
	return path.Join(os.Getenv("HOME"), ".local/share/osu/files")
}

// openBrowser opens url in the user's default browser. It launches the
// platform's handler and returns immediately without waiting for it.
func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	serverLog.Infof("Opening browser at %s", url)

	if err := cmd.Start(); err != nil {
		serverLog.Warnf("Failed to open browser: %v", err)
		return
	}

	serverLog.Debugf("Launched %s", cmd.Path)
}

//go:embed public/*
var publicFS embed.FS

// serverLog scopes top-level lifecycle messages so they are tagged [server].
var serverLog = logger.New(logger.LevelInfo).WithScope("server")

func main() {
	mux := http.NewServeMux()

	configPath, err := paths.ConfigFile("config.json")
	if err != nil {
		serverLog.Errorf("%v", err)
		os.Exit(1)
	}

	config, err := config.Load(configPath)
	if err != nil {
		serverLog.Errorf("%v", err)
		os.Exit(1)
	}
	serverLog.Infof("Loaded config from %s", configPath)

	cachePath, err := paths.CacheFile("cache.json")
	if err != nil {
		serverLog.Errorf("%v", err)
		os.Exit(1)
	}

	store, err := cache.Load(cachePath)
	if err != nil {
		serverLog.Warnf("Failed to load cache.json, starting fresh: %v", err)
		store = cache.New(cachePath)
	}
	serverLog.Infof("Loaded cache from %s (%d user IDs)", cachePath, store.UserIDCount())
	defer func() {
		if err := store.Save(); err != nil {
			serverLog.Errorf("Failed to persist cache: %v", err)
		}
	}()

	clientMan := handler.ClientHandler{}

	leaderboardWorker, err := worker.NewLeadboardWorker(config, store, &clientMan)
	if err != nil {
		serverLog.Errorf("%v", err)
		os.Exit(1)
	}
	defer leaderboardWorker.Close()

	gameWorker := worker.NewGameWorker(&clientMan, leaderboardWorker)

	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go gameWorker.Run(&wg, ctx)

	mux.HandleFunc("/ws", clientMan.Connection)

	mux.HandleFunc("PUT /leaderboard", gameWorker.ChangeLeaderboard)
	mux.HandleFunc("PUT /leaderboards", gameWorker.UpdateLeaderboards)
	mux.HandleFunc("GET /leaderboards", gameWorker.GetLeaderboards)


	mux.Handle("/lazer/",
		http.StripPrefix("/lazer/", http.FileServer(http.Dir(dataPath()))),
	)

	subFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		serverLog.Errorf("%v", err)
		os.Exit(1)
	}

	mux.Handle("/", http.FileServer(http.FS(subFS)))

	server := &http.Server{Addr: "0.0.0.0:8080", Handler: mux}

	go func() {
		serverLog.Infof("Listening on port %s", server.Addr)

		openBrowser("http://localhost" + server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverLog.Errorf("HTTP server stopped: %v", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	serverLog.Infof("Received %s, shutting down...", sig)

	s_ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(s_ctx); err != nil {
		serverLog.Errorf("HTTP server shutdown: %v", err)
	}
	serverLog.Infof("Shutdown complete")
}
