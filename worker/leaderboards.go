package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/yukiisen/ohana/cache"
	"github.com/yukiisen/ohana/config"
	"github.com/yukiisen/ohana/handler"
	"github.com/yukiisen/ohana/logger"
)

type LeaderboardWorker struct {
	Clients *handler.ClientHandler
	Api     *Client
	Cfg     *config.Config
	Store   *cache.Cache

	ActiveLeaderboard string
	scoreCache        *ScoreCache
	mu                sync.Mutex

	// loadMu ensures only one leaderboard load runs at a time so that a
	// cancelled load cannot interleave broadcasts with its replacement.
	loadMu sync.Mutex

	// wgMu guards closed and the Add side of loading; Close sets closed and
	// then waits, so no Add can race with Wait.
	wgMu    sync.Mutex
	loading sync.WaitGroup
	closed  bool
}

func NewLeadboardWorker(
	cfg *config.Config,
	store *cache.Cache,
	clients *handler.ClientHandler,
) (*LeaderboardWorker, error) {
	api, err := NewClient(context.Background(), cfg.API.ClientID, cfg.API.ClientSecret)
	if err != nil {
		return nil, err
	}

	l := &LeaderboardWorker{
		Clients:           clients,
		Api:               api,
		Cfg:               cfg,
		Store:             store,
		ActiveLeaderboard: "friends",
		scoreCache:        NewScoreCache(store, 5*time.Minute),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := l.LoadUserIDs(ctx); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *LeaderboardWorker) UserID(username string) (int, bool) {
	return l.Store.UserID(username)
}

// StartLoad runs a leaderboard load in the background. It is a no-op after
// Close, and it is safe to call concurrently with Close.
func (l *LeaderboardWorker) StartLoad(ctx context.Context, beatmapID int32) {
	l.wgMu.Lock()
	if l.closed {
		l.wgMu.Unlock()
		return
	}
	l.loading.Add(1)
	l.wgMu.Unlock()

	go func() {
		defer l.loading.Done()

		if err := l.Load(ctx, beatmapID); err != nil {
			logger.Errorf("%v", err)
		}
	}()
}

// Close prevents new loads from starting and waits for in-flight loads.
func (l *LeaderboardWorker) Close() {
	l.wgMu.Lock()
	l.closed = true
	l.wgMu.Unlock()

	l.loading.Wait()
}

func (l *LeaderboardWorker) LoadUserIDs(ctx context.Context) error {
	defer func() {
		if err := l.Store.Save(); err != nil {
			logger.Errorf("Failed to persist cache: %v", err)
		}
	}()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, usernames := range l.Cfg.Leaderboards() {
		for _, username := range usernames {
			if _, exists := l.Store.UserID(username); exists {
				continue
			}

			endpoint := "https://mirror.hinamizawa.ai/api/v1/hinai/player/" +
				url.PathEscape(username)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
			if err != nil {
				return fmt.Errorf("create lookup request for %q: %w", username, err)
			}

			req.Header.Set("User-Agent", "osu hana/1.0")

			resp, err := httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("resolve user %q: %w", username, err)
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return fmt.Errorf("resolve user %q: Hinai returned HTTP %d",
					username, resp.StatusCode)
			}

			var user User
			err = json.NewDecoder(resp.Body).Decode(&user)
			resp.Body.Close()

			if err != nil {
				return fmt.Errorf("decode user %q: %w", username, err)
			}

			if user.ID <= 0 {
				return fmt.Errorf("resolve user %q: invalid user ID %d",
					username, user.ID)
			}

			l.Store.SetUserID(username, user.ID)

			logger.Debugf("Cached user ID: %s = %d", username, user.ID)
		}
	}

	logger.Infof("Cached %d user IDs", l.Store.UserIDCount())
	return nil
}

type leaderboardMessage struct {
	Type string  `json:"type"`
	Leaderboard string `json:"leaderboard"`
	Data []Score `json:"data"`
}

func (l *LeaderboardWorker) Load(ctx context.Context, beatmapID int32) error {
	l.loadMu.Lock()
	defer l.loadMu.Unlock()

	defer func() {
		if err := l.Store.Save(); err != nil {
			logger.Errorf("Failed to persist cache: %v", err)
		}
	}()

	l.mu.Lock()
	name := l.ActiveLeaderboard
	l.mu.Unlock()

	if name == "Global" {
		var result BeatmapScores
		path := fmt.Sprintf("/beatmaps/%d/scores", beatmapID)

		if err := l.Api.GetJSON(ctx, path, &result); err != nil {
			return fmt.Errorf("fetch global leaderboard: %w", err)
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		logger.Infof(
			"Loaded global leaderboard for beatmap %d (%d scores)",
			beatmapID, len(result.Scores),
		)

		return l.broadcast(ctx, leaderboardMessage{
			Type: "Leaderboard",
			Leaderboard: name,
			Data: result.Scores,
		})
	}

    usernames, ok := l.Cfg.Leaderboards()[name]
    if !ok {
    	return fmt.Errorf("unknown leaderboard %q", name)
    }
    
    scores := 0
    missing := make([]string, 0)
    
    // Serve cached scores immediately; collect cache misses.
    for _, username := range usernames {
    	if err := ctx.Err(); err != nil {
    		return err
    	}
    
    	userID, ok := l.UserID(username)
    	if !ok {
    		logger.Errorf("No cached user ID for %q", username)
    		continue
    	}
    
    	score, ok := l.scoreCache.Get(int(beatmapID), userID)
    	if !ok {
    		missing = append(missing, username)
    		continue
    	}
    
    	if err := l.broadcast(ctx, leaderboardMessage{
    		Type:        "Leaderboard",
    		Leaderboard: name,
    		Data:        []Score{score},
    	}); err != nil {
    		return err
    	}
    
    	scores++
    }
    
    // No API requests if all scores were cached.
    if len(missing) > 0 {
    	select {
    	case <-time.After(2 * time.Second):
    	case <-ctx.Done():
    		return ctx.Err()
    	}
    }
    
    // Fetch only the scores that weren't cached.
    for _, username := range missing {
    	if err := ctx.Err(); err != nil {
    		return err
    	}
    
    	userID, ok := l.UserID(username)
    	if !ok {
    		continue
    	}
    
    	var result BeatmapUserScore
    	path := fmt.Sprintf(
    		"/beatmaps/%d/scores/users/%d",
    		beatmapID,
    		userID,
    	)
    
    	if err := l.Api.GetJSON(ctx, path, &result); err != nil {
    		if ctx.Err() != nil {
    			return ctx.Err()
    		}
    
    		// logger.Errorf(
    		// 	"Failed to fetch score for %q on beatmap %d: %v",
    		// 	username, beatmapID, err,
    		// )
    		continue
    	}
    
    	if err := ctx.Err(); err != nil {
    		return err
    	}
    
    	score := result.Score
    	l.scoreCache.Set(score)
    
    	if err := l.broadcast(ctx, leaderboardMessage{
    		Type:        "Leaderboard",
    		Leaderboard: name,
    		Data:        []Score{score},
    	}); err != nil {
    		return err
    	}
    
    	scores++
    }
    
    logger.Infof(
    	"Finished loading leaderboard %q for beatmap %d (%d scores)",
    	name, beatmapID, scores,
    )

	return nil
}

func (l *LeaderboardWorker) broadcast(ctx context.Context, msg leaderboardMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal leaderboard message: %w", err)
	}

	// Snapshot the client list so we don't hold the lock while writing.
	l.Clients.Mu.Lock()
	clients := append([]*websocket.Conn(nil), l.Clients.Clients...)
	l.Clients.Mu.Unlock()

	for _, client := range clients {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := client.Write(ctx, websocket.MessageText, data); err != nil {
			logger.Errorf("Failed to broadcast leaderboard: %v", err)
		}
	}

	return ctx.Err()
}
