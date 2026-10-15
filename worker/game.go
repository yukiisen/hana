package worker

import (
	"context"
	"encoding/json"
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
	clients *handler.ClientHandler
	beatmap *osumem.BeatmapInfo
	lbw *LeaderboardWorker
	cancelLb context.CancelFunc
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

		info, err := g.game.Beatmap()
		if err != nil {
			logger.Errorf(err.Error())
			if errors.Is(err, osumem.ErrProcessGone) { g.game.Close(); g.game = nil }
			continue
		} 

		if g.beatmap != nil && info.MapID == g.beatmap.MapID {
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

		wg.Add(1)
		go func () {
			g.lbw.Load(workerCtx, info.MapID)
			wg.Done()
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
