// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package ai

import (
	"context"
	"fmt"

	orsdk "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/operations"
	"github.com/spf13/viper"
)

const (
	InputTypeSearchDocument = "search_document"
	InputTypeSearchQuery    = "search_query"
)

// EmbedOptions configures an embedding request.
type EmbedOptions struct {
	InputType  string
	Dimensions int64
}

// EmbedClient calls oroute embedding APIs.
type EmbedClient struct {
	api   *orsdk.OpenRouter
	model string
}

// NewEmbedClient returns an embedding client loaded from app.ai.embed config.
func NewEmbedClient() *EmbedClient {
	return &EmbedClient{
		api: orsdk.New(
			orsdk.WithSecurity(viper.GetString("app.ai.api_key")),
		),
		model: viper.GetString("app.ai.embed.model"),
	}
}

// Generate generates float embeddings for the given texts.
func (c *EmbedClient) Generate(ctx context.Context, texts []string, opts ...EmbedOptions) ([][]float64, Usage, error) {
	request := operations.CreateEmbeddingsRequest{
		Model:      c.model,
		Input:      operations.CreateInputUnionArrayOfStr(texts),
		InputType:  new(opts[0].InputType),
		Dimensions: &opts[0].Dimensions,
	}

	response, err := c.api.Embeddings.Generate(ctx, request)
	if err != nil {
		return nil, Usage{}, fmt.Errorf("ai embed: %w", err)
	}

	if response == nil || response.CreateEmbeddingsResponseBody == nil {
		return nil, Usage{}, ErrNoEmbeddings
	}

	body := response.CreateEmbeddingsResponseBody
	embeddings := make([][]float64, 0, len(body.Data))
	for _, item := range body.Data {
		if item.Embedding.Type != operations.EmbeddingTypeArrayOfNumber {
			return nil, Usage{}, fmt.Errorf("ai embed: unexpected embedding type %q", item.Embedding.Type)
		}

		embeddings = append(embeddings, item.Embedding.ArrayOfNumber)
	}

	if len(embeddings) == 0 {
		return nil, Usage{}, ErrNoEmbeddings
	}

	usage := Usage{}
	if body.Usage != nil {
		var cost int64

		amount := body.Usage.GetCost()
		if amount != nil {
			cost = CostFromUSD(*amount)
		}

		usage = Usage{
			PromptTokens: body.Usage.GetPromptTokens(),
			TotalTokens:  body.Usage.GetTotalTokens(),
			Cost:         cost,
		}
	}

	return embeddings, usage, nil
}

// EmbedDocuments generates embeddings optimized for indexing documents.
func (c *EmbedClient) EmbedDocuments(ctx context.Context, texts []string, opts EmbedOptions) ([][]float64, Usage, error) {
	opts.InputType = InputTypeSearchDocument

	return c.Generate(ctx, texts, opts)
}

// EmbedQuery generates an embedding optimized for search queries.
func (c *EmbedClient) EmbedQuery(ctx context.Context, text string, opts EmbedOptions) ([]float64, Usage, error) {
	opts.InputType = InputTypeSearchQuery
	res, u, err := c.Generate(ctx, []string{text}, opts)
	if err != nil {
		return nil, Usage{}, err
	}

	return res[0], u, err
}
