// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
)

const (
	// DefaultCardPath is where A2A agents serve their card.
	DefaultCardPath = "/.well-known/agent-card.json"
	// MaxCardSize caps the card body read from an instance.
	MaxCardSize = 1 << 20
	// CardFetchTimeout bounds each card fetch attempt.
	CardFetchTimeout = 5 * time.Second
)

var (
	ErrCardTooLarge     = errors.New("agent card too large")
	ErrInvalidAgentCard = errors.New("agent card is not a JSON object")
)

// Card fetches agent cards from agent instances.
type Card struct {
	http *req.Client
}

// NewCard returns a card fetcher with the fetch timeout applied.
func NewCard() *Card {
	return &Card{
		http: req.C().SetTimeout(CardFetchTimeout),
	}
}

// Fetch reads the agent card from an instance, trying https first and then http.
func (c *Card) Fetch(ctx context.Context, address string, port int, path string) (map[string]any, error) {
	var lastErr error
	for _, scheme := range []string{"https", "http"} {
		target := url.URL{
			Scheme: scheme,
			Host:   net.JoinHostPort(address, strconv.Itoa(port)),
			Path:   path,
		}

		resp, err := c.http.R().
			SetContext(ctx).
			SetHeader("Accept", "application/json").
			DisableAutoReadResponse().
			Get(target.String())
		if err != nil {
			log.Debug().
				Err(err).
				Str("url", target.String()).
				Msg("Failed to fetch agent card")
			lastErr = fmt.Errorf("fetch agent card: %w", err)
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, MaxCardSize+1))
		resp.Body.Close()
		if err != nil {
			log.Debug().
				Err(err).
				Str("url", target.String()).
				Msg("Failed to fetch agent card")
			lastErr = fmt.Errorf("fetch agent card: %w", err)
			continue
		}

		if !resp.IsSuccessState() {
			log.Debug().
				Str("url", target.String()).
				Int("status", resp.StatusCode).
				Msg("Agent card request failed")
			lastErr = fmt.Errorf("fetch agent card: %s: status %d", target.String(), resp.StatusCode)
			continue
		}

		if len(body) > MaxCardSize {
			log.Warn().
				Str("url", target.String()).
				Msg("Agent card too large")
			return nil, ErrCardTooLarge
		}

		var card map[string]any
		if json.Unmarshal(body, &card) != nil || card == nil {
			log.Warn().
				Str("url", target.String()).
				Msg("Agent card is not a JSON object")
			return nil, ErrInvalidAgentCard
		}

		log.Debug().
			Str("url", target.String()).
			Msg("Agent card fetched")

		return card, nil
	}

	log.Warn().
		Err(lastErr).
		Str("address", address).
		Int("port", port).
		Str("path", path).
		Msg("Failed to fetch agent card")

	return nil, lastErr
}
