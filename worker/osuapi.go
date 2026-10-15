package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yukiisen/ohana/logger"
)

const tokenURL = "https://osu.ppy.sh/oauth/token"
const baseURL = "https://osu.ppy.sh/api/v2"

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func GenerateToken(
	ctx context.Context,
	clientID int,
	clientSecret string,
) (*Token, error) {
	data := url.Values{
		"client_id":     {strconv.Itoa(clientID)},
		"client_secret": {clientSecret},
		"grant_type":    {"client_credentials"},
		"scope":         {"public"},
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		tokenURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("osu! OAuth: %s", resp.Status)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

type Client struct {
	clientID     int
	clientSecret string
	httpClient   *http.Client

	mu        sync.Mutex
	token     Token
	expiresAt time.Time

	// rate limiter
	rateMu       sync.Mutex
	nextRequest  time.Time
	requestDelay time.Duration
}

func NewClient(
	ctx context.Context,
	clientID int,
	clientSecret string,
) (*Client, error) {
	c := &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		requestDelay: 100 * time.Millisecond,
	}

	if err := c.renewToken(ctx); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) renewToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Another caller may already have renewed it.
	if c.token.AccessToken != "" &&
		time.Until(c.expiresAt) > time.Minute {
		return nil
	}

	token, err := GenerateToken(ctx, c.clientID, c.clientSecret)
	if err != nil {
		logger.Errorf("renew osu! token: %v", err)
		return err
	}

	c.token = *token
	c.expiresAt = time.Now().Add(
		time.Duration(token.ExpiresIn) * time.Second,
	)

	return nil
}

// Do sends an authenticated request to an API v2 endpoint.
// Pass paths such as "/users/2".
func (c *Client) Do(
	ctx context.Context,
	method string,
	path string,
) (*http.Response, error) {
	if err := c.renewToken(ctx); err != nil {
		return nil, err
	}

	if err := c.waitRateLimit(ctx); err != nil {
		return nil, err
	}

	c.mu.Lock()
	accessToken := c.token.AccessToken
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		baseURL+path,
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-version", "20220705")

	return c.httpClient.Do(req)
}

func (c *Client) GetJSON(
    ctx context.Context,
    path string,
    dst any,
) error {
    resp, err := c.Do(ctx, http.MethodGet, path)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        logger.Errorf("osu! API returned %s", resp.Status)
		return errors.New("HTTP Status " + resp.Status)
    }

    if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
        logger.Errorf("decode osu! API response: %w", err)
		return err
    }

    return nil
}

func (c *Client) waitRateLimit(ctx context.Context) error {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()

	wait := time.Until(c.nextRequest)
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Reserve the next slot.
	c.nextRequest = time.Now().Add(c.requestDelay)
	return nil
}

// TokenExpiresAt returns the current token's expiration time.
func (c *Client) TokenExpiresAt() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.expiresAt
}

// ClientID returns the configured OAuth client ID.
func (c *Client) ClientID() string {
	return strconv.Itoa(c.clientID)
}


type Score struct {
	ID                  int64             `json:"id"`
	UserID              int               `json:"user_id"`
	BeatmapID           int               `json:"beatmap_id"`
	Accuracy            float64           `json:"accuracy"`
	Mods                []Mod             `json:"mods"`
	MaxCombo            int               `json:"max_combo"`
	IsPerfectCombo      bool              `json:"is_perfect_combo"`
	LegacyPerfect       bool              `json:"legacy_perfect"`
	Passed              bool              `json:"passed"`
	PP                  *float64          `json:"pp"`
	Rank                string            `json:"rank"`
	EndedAt             time.Time         `json:"ended_at"`
	HasReplay           bool              `json:"has_replay"`
	LegacyScoreID       *int64            `json:"legacy_score_id"`
	LegacyTotalScore    int64             `json:"legacy_total_score"`
	ClassicTotalScore   *int64            `json:"classic_total_score"`
	TotalScore          int64             `json:"total_score"`
	TotalScoreWithoutMods *int64          `json:"total_score_without_mods"`
	Type                string            `json:"type"`
	Statistics          map[string]*int   `json:"statistics"`
	User                *User             `json:"user"`
}

type Mod struct {
	Acronym  string         `json:"acronym"`
	Settings map[string]any `json:"settings,omitempty"`
}

type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	CountryCode string `json:"country_code"`
}

type BeatmapUserScore struct {
	Position int   `json:"position"`
	Score    Score `json:"score"`
}

type BeatmapScores struct {
	Scores    []Score `json:"scores"`
	UserScore *Score  `json:"user_score"`
}
