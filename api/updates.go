package api

import (
	"context"

	telegram "github.com/sidwiskers/hermes/types"
)

type GetUpdatesParams struct {
	Offset         int64    `json:"offset,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

func (b *Client) GetUpdates(ctx context.Context, params GetUpdatesParams) ([]Update, error) {
	raw, err := b.callJSON(ctx, "getUpdates", params)
	if err != nil {
		return nil, err
	}
	updates, err := telegram.DecodeUpdates(raw, b.preserveRawUpdates)
	if err != nil {
		return nil, err
	}
	return updates, nil
}
