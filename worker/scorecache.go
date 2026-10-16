package worker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/yukiisen/ohana/cache"
	"github.com/yukiisen/ohana/logger"
)

type ScoreCache struct {
	store *cache.Cache
	ttl   time.Duration
}

func NewScoreCache(store *cache.Cache, ttl time.Duration) *ScoreCache {
	return &ScoreCache{store: store, ttl: ttl}
}

func scoreKey(beatmapID, userID int) string {
	return fmt.Sprintf("%d:%d", beatmapID, userID)
}

func (c *ScoreCache) Get(beatmapID, userID int) (Score, bool) {
	raw, ok := c.store.Score(scoreKey(beatmapID, userID))
	if !ok {
		return Score{}, false
	}

	var score Score
	if err := json.Unmarshal(raw, &score); err != nil {
		logger.Errorf("Failed to decode cached score: %v", err)
		return Score{}, false
	}

	return score, true
}

func (c *ScoreCache) Set(score Score) {
	raw, err := json.Marshal(score)
	if err != nil {
		logger.Errorf("Failed to encode score for cache: %v", err)
		return
	}

	c.store.SetScore(scoreKey(score.BeatmapID, score.UserID), raw, c.ttl)
}
