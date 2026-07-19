package api

import (
	"context"
	"fmt"
	"unicode/utf8"
)

type ApproveSuggestedPostParams struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int   `json:"message_id"`
	SendDate  int64 `json:"send_date,omitempty"`
}

func (client *Client) ApproveSuggestedPost(ctx context.Context, params ApproveSuggestedPostParams) error {
	if params.ChatID == 0 || params.MessageID == 0 {
		return fmt.Errorf("hermes: approveSuggestedPost chat_id and message_id are required")
	}
	return client.callTrue(ctx, "approveSuggestedPost", params)
}

type DeclineSuggestedPostParams struct {
	ChatID    int64  `json:"chat_id"`
	MessageID int    `json:"message_id"`
	Comment   string `json:"comment,omitempty"`
}

func (client *Client) DeclineSuggestedPost(ctx context.Context, params DeclineSuggestedPostParams) error {
	if params.ChatID == 0 || params.MessageID == 0 {
		return fmt.Errorf("hermes: declineSuggestedPost chat_id and message_id are required")
	}
	if utf8.RuneCountInString(params.Comment) > 128 {
		return fmt.Errorf("hermes: declineSuggestedPost comment must not exceed 128 characters")
	}
	return client.callTrue(ctx, "declineSuggestedPost", params)
}
