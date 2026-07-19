package api

import "context"

type GetChatParams struct {
	ChatID any `json:"chat_id"`
}

func (b *Client) GetChat(ctx context.Context, chatID any) (*ChatFullInfo, error) {
	if err := validateChatID(chatID, "getChat"); err != nil {
		return nil, err
	}
	var chat ChatFullInfo
	if err := b.Call(ctx, "getChat", GetChatParams{ChatID: chatID}, &chat); err != nil {
		return nil, err
	}
	return &chat, nil
}

func (b *Client) GetChatMemberCount(ctx context.Context, chatID any) (int, error) {
	if err := validateChatID(chatID, "getChatMemberCount"); err != nil {
		return 0, err
	}
	var count int
	err := b.Call(ctx, "getChatMemberCount", GetChatParams{ChatID: chatID}, &count)
	return count, err
}

func (b *Client) LeaveChat(ctx context.Context, chatID any) error {
	if err := validateChatID(chatID, "leaveChat"); err != nil {
		return err
	}
	return b.callTrue(ctx, "leaveChat", GetChatParams{ChatID: chatID})
}
