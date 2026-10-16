package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/yukiisen/ohana/config"
	"github.com/yukiisen/ohana/handler"
	"github.com/yukiisen/ohana/logger"
	"github.com/yukiisen/ohana/worker"
)

func dataPath() string {
	return path.Join(os.Getenv("HOME"), ".local/share/osu/files")
}

func main() {
	mux := http.NewServeMux()

	config, err := config.Load("config.json")
	if err != nil { log.Fatal(err) }

	clientMan := handler.ClientHandler {}

	leaderboardWorker, err := worker.NewLeadboardWorker(config, &clientMan)
	if err != nil { log.Fatal(err) }
	defer leaderboardWorker.Wg.Wait()

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

	mux.Handle("/", http.FileServer(http.Dir("./public/")))

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
