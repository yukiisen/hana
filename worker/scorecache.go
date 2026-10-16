package worker

import (
	"sync"
	"time"
)

type Key struct {
	BeatmapID int
	UserID    int
}

type entry struct {
	score     Score
	expiresAt time.Time
}

type ScoreCache struct {
	mu      sync.RWMutex
	scores  map[Key]entry
	ttl     time.Duration
}

func NewScoreCache(ttl time.Duration) *ScoreCache {
	return &ScoreCache{
		scores: make(map[Key]entry),
		ttl:    ttl,
	}
}

func (c *ScoreCache) Get(beatmapID, userID int) (Score, bool) {
	key := Key{beatmapID, userID}

	c.mu.RLock()
	item, ok := c.scores[key]
	if ok && time.Now().Before(item.expiresAt) {
		c.mu.RUnlock()
		return item.score, true
	}
	c.mu.RUnlock()

	// Remove expired entries.
	if ok {
		c.mu.Lock()
		if current, exists := c.scores[key]; exists &&
			!time.Now().Before(current.expiresAt) {
			delete(c.scores, key)
		}
		c.mu.Unlock()
	}

	return Score{}, false
}

func (c *ScoreCache) Set(score Score) {
	key := Key{score.BeatmapID, score.UserID}

	c.mu.Lock()
	c.scores[key] = entry{
		score:     score,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}
