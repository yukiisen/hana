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
	"github.com/yukiisen/ohana/config"
	"github.com/yukiisen/ohana/handler"
	"github.com/yukiisen/ohana/logger"
)

type LeaderboardWorker struct {
	Clients 			*handler.ClientHandler
	Api 				*Client // osu api
	Cfg 				*config.Config
	UserIDs 			map[string]int
	ActiveLeaderboard 	string
	mu 					sync.Mutex
}

func NewLeadboardWorker(cfg *config.Config, clients *handler.ClientHandler) (*LeaderboardWorker, error) {
	api, err := NewClient(context.Background(), cfg.API.ClientID, cfg.API.ClientSecret)
	if err != nil {
		return nil, err
	}

	l := &LeaderboardWorker{
		Clients:           clients,
		Api:               api,
		Cfg:			   cfg,
		UserIDs:           make(map[string]int),
		ActiveLeaderboard: "friends",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := l.loadUserIDs(ctx); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *LeaderboardWorker) loadUserIDs(ctx context.Context) error {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, usernames := range l.Cfg.Leaderboards() {
		for _, username := range usernames {
			if _, exists := l.UserIDs[username]; exists {
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

			l.UserIDs[username] = user.ID
			logger.Debugf("Cached user ID: %s = %d", username, user.ID)
		}
	}

	logger.Infof("Cached %d user IDs", len(l.UserIDs))
	return nil
}

type leaderboardMessage struct {
	Type string  `json:"type"`
	Data []Score `json:"data"`
}

func (l *LeaderboardWorker) Load(ctx context.Context, beatmapID int32) error {
	// wait for a while in case the user changes beatmap quickly
	select {
	case <-time.After(2 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	name := l.ActiveLeaderboard

	if name == "Global" {
		var result BeatmapScores
		path := fmt.Sprintf("/beatmaps/%d/scores", beatmapID)

		if err := l.Api.GetJSON(ctx, path, &result); err != nil {
			logger.Errorf("Failed to fetch global leaderboard: %v", err)
			return fmt.Errorf("fetch global leaderboard: %w", err)
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		logger.Infof("Loaded global leaderboard for beatmap %d (%d scores)", beatmapID, len(result.Scores))

		return l.broadcast(ctx, leaderboardMessage{
			Type: "Leaderboard",
			Data: result.Scores,
		})
	}

	usernames, ok := l.Cfg.Leaderboards()[name]
	if !ok {
		return fmt.Errorf("unknown leaderboard %q", name)
	}

	scores := 0

	for _, username := range usernames {
        userID, ok := l.UserIDs[username]
        if !ok {
        	logger.Errorf("No cached user ID for %q", username)
        	continue
        }
        
        var result BeatmapUserScore
        path := fmt.Sprintf("/beatmaps/%d/scores/users/%d", beatmapID, userID)
        
        if err := l.Api.GetJSON(ctx, path, &result); err != nil {
        	logger.Errorf("Failed to fetch score for %q on beatmap %d: %v",
        		username, beatmapID, err)
        	continue
        }
        
        if err := l.broadcast(ctx, leaderboardMessage{
        	Type: "Leaderboard",
        	Data: []Score{result.Score},
        }); err != nil {
        	return err
        }

		scores++
	}

	logger.Infof("Finished loading leaderboard %q for beatmap %d (%d scores)", name, beatmapID, scores)
	return nil
}

func (l *LeaderboardWorker) broadcast(ctx context.Context, msg leaderboardMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal leaderboard message: %w", err)
	}

	l.Clients.Mu.Lock()
	clients := l.Clients.Clients

	for _, client := range clients {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := client.Write(ctx, websocket.MessageText, data); err != nil {
			logger.Errorf("Failed to broadcast leaderboard: %v", err)
		}
	}

	l.Clients.Mu.Unlock()

	return ctx.Err()
}

func (l *LeaderboardWorker) ChangeLeaderboard(w http.ResponseWriter, r *http.Request) {
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

	l.mu.Lock()
	l.ActiveLeaderboard = body.ActiveLeaderboard
	l.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
