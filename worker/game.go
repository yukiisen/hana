package worker

import (
	"context"
	"encoding/json"
	"net/http"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/yukiisen/ohana/handler"
	"github.com/yukiisen/ohana/logger"
	"github.com/yukiisen/ohana/osumem"
)

type GameWorker struct {
	game osumem.Client
	clients 	*handler.ClientHandler
	beatmap 	*osumem.BeatmapInfo
	lbw 		*LeaderboardWorker
	cancelLb 	context.CancelFunc
	mu			sync.Mutex
}

func NewGameWorker(c *handler.ClientHandler, lbw *LeaderboardWorker) *GameWorker {
	return &GameWorker { clients: c, lbw: lbw, cancelLb: func() {} }
}

func (g *GameWorker) Run(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	defer func () { if g.game != nil { g.game.Close(); g.game = nil } }()
	defer g.cancelLb()

	for {
		if ctx.Err() != nil { 
			logger.Infof("Logger thread closing..")
			return 
		}

		if g.game == nil {
			g.loadGame(ctx)
			continue
		}

		select {
		case <-time.After(400 * time.Millisecond):
		case <-ctx.Done():
			logger.Infof("Logger thread closing..")
			return
		}

		g.mu.Lock()

		info, err := g.game.Beatmap()
		if err != nil {
			logger.Errorf(err.Error())
			if errors.Is(err, osumem.ErrProcessGone) { g.game.Close(); g.game = nil }
			g.mu.Unlock()
			continue
		} 

		if g.beatmap != nil && info.MapID == g.beatmap.MapID {
			g.mu.Unlock()
			continue
		}

		g.cancelLb()
		g.beatmap = info

		msg, err := json.Marshal(map[string]any{
			"type": "Beatmap",
			"data": info,
		})

		if err != nil {
			logger.Errorf(err.Error())
			g.mu.Unlock()
			continue
		}

		g.clients.Mu.Lock()

		for _, c := range g.clients.Clients {
			err = c.Write(ctx, websocket.MessageText, msg)

			if err != nil {
				logger.Errorf(err.Error())
			}
		}

		g.clients.Mu.Unlock()

		workerCtx, cancel := context.WithCancel(context.Background())
		g.cancelLb = cancel
		g.mu.Unlock()

		g.lbw.Wg.Add(1)
		go func () {
			g.lbw.Load(workerCtx, info.MapID)
			g.lbw.Wg.Done()
		}()

	}
}

func (g *GameWorker) loadGame(ctx context.Context) {
	game, err := osumem.Open()

	if err != nil {
		logger.Errorf("Could not find a running osu client, please make sure your game is open.")
		logger.Errorf(err.Error())

		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
		}

	} else {
		logger.Infof("Connected to %s client at pid %d", game.Kind(), game.PID())
		g.game = game
	}
}


func (g *GameWorker) ChangeLeaderboard(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ActiveLeaderboard string `json:"active_leaderboard"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if body.ActiveLeaderboard == "" {
		http.Error(w, "missing active_leaderboard", http.StatusBadRequest)
		return
	}

	g.lbw.mu.Lock()
	g.lbw.ActiveLeaderboard = body.ActiveLeaderboard
	g.lbw.mu.Unlock()

	g.mu.Lock()
	defer g.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	g.cancelLb()
	g.cancelLb = cancel

	g.lbw.Wg.Add(1)
	go func () {
		g.lbw.Load(ctx, g.beatmap.MapID)
		g.lbw.Wg.Done()
	}()

	w.WriteHeader(http.StatusNoContent)
}

func (g *GameWorker) GetLeaderboards(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(g.lbw.Cfg.Leaderboards()); err != nil {
		http.Error(w, "failed to encode leaderboards", http.StatusInternalServerError)
		return
	}
}

func (g *GameWorker) UpdateLeaderboards(w http.ResponseWriter, r *http.Request) {
	var boards map[string][]string
	if err := json.NewDecoder(r.Body).Decode(&boards); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if boards == nil {
		http.Error(w, "expected a JSON object", http.StatusBadRequest)
		return
	}

	g.lbw.Cfg.SetLeaderboards(boards)

	if err := g.lbw.Cfg.Save(); err != nil {
		http.Error(w, "failed to save leaderboards", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

	g.lbw.LoadUserIDs(context.Background())
}
