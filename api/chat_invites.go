package api

import (
	"context"
	"fmt"
	"strings"
)

type ExportChatInviteLinkParams struct {
	ChatID any `json:"chat_id"`
}

func (b *Client) ExportChatInviteLink(ctx context.Context, chatID any) (string, error) {
	if err := validateChatID(chatID, "exportChatInviteLink"); err != nil {
		return "", err
	}
	var link string
	if err := b.Call(ctx, "exportChatInviteLink", ExportChatInviteLinkParams{ChatID: chatID}, &link); err != nil {
		return "", err
	}
	return link, nil
}

type CreateChatInviteLinkParams struct {
	ChatID             any    `json:"chat_id"`
	Name               string `json:"name,omitempty"`
	ExpireDate         int64  `json:"expire_date,omitempty"`
	MemberLimit        int    `json:"member_limit,omitempty"`
	CreatesJoinRequest bool   `json:"creates_join_request,omitempty"`
}

func (b *Client) CreateChatInviteLink(ctx context.Context, params CreateChatInviteLinkParams) (*ChatInviteLink, error) {
	if err := validateInviteCreate(params.ChatID, params.Name, params.MemberLimit, params.CreatesJoinRequest, "createChatInviteLink"); err != nil {
		return nil, err
	}
	var link ChatInviteLink
	if err := b.Call(ctx, "createChatInviteLink", params, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

type EditChatInviteLinkParams struct {
	ChatID             any     `json:"chat_id"`
	InviteLink         string  `json:"invite_link"`
	Name               *string `json:"name,omitempty"`
	ExpireDate         *int64  `json:"expire_date,omitempty"`
	MemberLimit        *int    `json:"member_limit,omitempty"`
	CreatesJoinRequest *bool   `json:"creates_join_request,omitempty"`
}

func (b *Client) EditChatInviteLink(ctx context.Context, params EditChatInviteLinkParams) (*ChatInviteLink, error) {
	if err := validateChatID(params.ChatID, "editChatInviteLink"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(params.InviteLink) == "" {
		return nil, fmt.Errorf("hermes: editChatInviteLink invite_link is required")
	}
	if params.Name != nil && len([]rune(*params.Name)) > 32 {
		return nil, fmt.Errorf("hermes: invite-link name must not exceed 32 characters")
	}
	if params.MemberLimit != nil && (*params.MemberLimit < 0 || *params.MemberLimit > 99999) {
		return nil, fmt.Errorf("hermes: invite-link member_limit must be 0-99999")
	}
	if params.CreatesJoinRequest != nil && *params.CreatesJoinRequest && params.MemberLimit != nil && *params.MemberLimit != 0 {
		return nil, fmt.Errorf("hermes: creates_join_request and member_limit are mutually exclusive")
	}
	var link ChatInviteLink
	if err := b.Call(ctx, "editChatInviteLink", params, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

type CreateChatSubscriptionInviteLinkParams struct {
	ChatID             any    `json:"chat_id"`
	Name               string `json:"name,omitempty"`
	SubscriptionPeriod int    `json:"subscription_period"`
	SubscriptionPrice  int    `json:"subscription_price"`
}

func (b *Client) CreateChatSubscriptionInviteLink(ctx context.Context, params CreateChatSubscriptionInviteLinkParams) (*ChatInviteLink, error) {
	if err := validateChatID(params.ChatID, "createChatSubscriptionInviteLink"); err != nil {
		return nil, err
	}
	if len([]rune(params.Name)) > 32 {
		return nil, fmt.Errorf("hermes: invite-link name must not exceed 32 characters")
	}
	if params.SubscriptionPeriod != 2592000 {
		return nil, fmt.Errorf("hermes: subscription_period must be 2592000 seconds")
	}
	if params.SubscriptionPrice < 1 || params.SubscriptionPrice > 10000 {
		return nil, fmt.Errorf("hermes: subscription_price must be 1-10000 Stars")
	}
	var link ChatInviteLink
	if err := b.Call(ctx, "createChatSubscriptionInviteLink", params, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

type EditChatSubscriptionInviteLinkParams struct {
	ChatID     any     `json:"chat_id"`
	InviteLink string  `json:"invite_link"`
	Name       *string `json:"name,omitempty"`
}

func (b *Client) EditChatSubscriptionInviteLink(ctx context.Context, params EditChatSubscriptionInviteLinkParams) (*ChatInviteLink, error) {
	if err := validateChatID(params.ChatID, "editChatSubscriptionInviteLink"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(params.InviteLink) == "" {
		return nil, fmt.Errorf("hermes: editChatSubscriptionInviteLink invite_link is required")
	}
	if params.Name != nil && len([]rune(*params.Name)) > 32 {
		return nil, fmt.Errorf("hermes: invite-link name must not exceed 32 characters")
	}
	var link ChatInviteLink
	if err := b.Call(ctx, "editChatSubscriptionInviteLink", params, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

type RevokeChatInviteLinkParams struct {
	ChatID     any    `json:"chat_id"`
	InviteLink string `json:"invite_link"`
}

func (b *Client) RevokeChatInviteLink(ctx context.Context, params RevokeChatInviteLinkParams) (*ChatInviteLink, error) {
	if err := validateChatID(params.ChatID, "revokeChatInviteLink"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(params.InviteLink) == "" {
		return nil, fmt.Errorf("hermes: revokeChatInviteLink invite_link is required")
	}
	var link ChatInviteLink
	if err := b.Call(ctx, "revokeChatInviteLink", params, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

func validateInviteCreate(chatID any, name string, memberLimit int, createsJoinRequest bool, method string) error {
	if err := validateChatID(chatID, method); err != nil {
		return err
	}
	if len([]rune(name)) > 32 {
		return fmt.Errorf("hermes: invite-link name must not exceed 32 characters")
	}
	if memberLimit < 0 || memberLimit > 99999 {
		return fmt.Errorf("hermes: invite-link member_limit must be 0-99999")
	}
	if createsJoinRequest && memberLimit != 0 {
		return fmt.Errorf("hermes: creates_join_request and member_limit are mutually exclusive")
	}
	return nil
}
