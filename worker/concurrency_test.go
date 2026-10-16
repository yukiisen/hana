package worker

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/yukiisen/ohana/cache"
	"github.com/yukiisen/ohana/config"
	"github.com/yukiisen/ohana/handler"
)

func newTestWorker(t *testing.T) *LeaderboardWorker {
	t.Helper()

	return &LeaderboardWorker{
		Clients:           &handler.ClientHandler{},
		Api:               &Client{httpClient: &http.Client{Timeout: time.Second}},
		Cfg:               &config.Config{},
		Store:             cache.New(filepath.Join(t.TempDir(), "cache.json")),
		ActiveLeaderboard: "Global",
	}
}

// broadcast must not hold the client mutex after returning, even when the
// context is already cancelled.
func TestBroadcastReleasesClientLock(t *testing.T) {
	lw := newTestWorker(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := lw.broadcast(ctx, leaderboardMessage{Type: "Leaderboard"}); err == nil {
		t.Fatal("expected error from cancelled context")
	}

	done := make(chan struct{})
	go func() {
		lw.Clients.Mu.Lock()
		lw.Clients.Mu.Unlock()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("broadcast leaked the client mutex")
	}
}

// StartLoad after Close must be a no-op, and concurrent StartLoad/Close must
// not trip the WaitGroup. Run with -race.
func TestStartLoadClose(t *testing.T) {
	lw := newTestWorker(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lw.Close()

	lw.StartLoad(ctx, 1)

	lw2 := newTestWorker(t)

	loadCtx, loadCancel := context.WithCancel(context.Background())
	loadCancel()

	for i := 0; i < 50; i++ {
		lw2.StartLoad(loadCtx, int32(i))
	}
	lw2.Close()
	lw2.StartLoad(loadCtx, 99)
}
