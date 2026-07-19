package api

import (
	"context"
	"fmt"
)

type BotCommandScope interface{ botCommandScope() }

type BotCommandScopeDefault struct {
	Type string `json:"type"`
}

func (BotCommandScopeDefault) botCommandScope()   {}
func DefaultCommandScope() BotCommandScopeDefault { return BotCommandScopeDefault{Type: "default"} }

type BotCommandScopeAllPrivateChats struct {
	Type string `json:"type"`
}

func (BotCommandScopeAllPrivateChats) botCommandScope() {}
func AllPrivateChatsCommandScope() BotCommandScopeAllPrivateChats {
	return BotCommandScopeAllPrivateChats{Type: "all_private_chats"}
}

type BotCommandScopeAllGroupChats struct {
	Type string `json:"type"`
}

func (BotCommandScopeAllGroupChats) botCommandScope() {}
func AllGroupChatsCommandScope() BotCommandScopeAllGroupChats {
	return BotCommandScopeAllGroupChats{Type: "all_group_chats"}
}

type BotCommandScopeAllChatAdministrators struct {
	Type string `json:"type"`
}

func (BotCommandScopeAllChatAdministrators) botCommandScope() {}
func AllChatAdministratorsCommandScope() BotCommandScopeAllChatAdministrators {
	return BotCommandScopeAllChatAdministrators{Type: "all_chat_administrators"}
}

type BotCommandScopeChat struct {
	Type   string `json:"type"`
	ChatID any    `json:"chat_id"`
}

func (BotCommandScopeChat) botCommandScope() {}
func ChatCommandScope(chatID any) BotCommandScopeChat {
	return BotCommandScopeChat{Type: "chat", ChatID: chatID}
}

type BotCommandScopeChatAdministrators struct {
	Type   string `json:"type"`
	ChatID any    `json:"chat_id"`
}

func (BotCommandScopeChatAdministrators) botCommandScope() {}
func ChatAdministratorsCommandScope(chatID any) BotCommandScopeChatAdministrators {
	return BotCommandScopeChatAdministrators{Type: "chat_administrators", ChatID: chatID}
}

type BotCommandScopeChatMember struct {
	Type   string `json:"type"`
	ChatID any    `json:"chat_id"`
	UserID int64  `json:"user_id"`
}

func (BotCommandScopeChatMember) botCommandScope() {}
func ChatMemberCommandScope(chatID any, userID int64) BotCommandScopeChatMember {
	return BotCommandScopeChatMember{Type: "chat_member", ChatID: chatID, UserID: userID}
}

type SetMyCommandsParams struct {
	Commands     []BotCommand    `json:"commands"`
	Scope        BotCommandScope `json:"scope,omitempty"`
	LanguageCode string          `json:"language_code,omitempty"`
}

func (b *Client) SetMyCommands(ctx context.Context, params SetMyCommandsParams) error {
	if len(params.Commands) == 0 {
		return fmt.Errorf("hermes: setMyCommands commands are required")
	}
	return b.callTrue(ctx, "setMyCommands", params)
}

type DeleteMyCommandsParams struct {
	Scope        BotCommandScope `json:"scope,omitempty"`
	LanguageCode string          `json:"language_code,omitempty"`
}

func (b *Client) DeleteMyCommands(ctx context.Context, params DeleteMyCommandsParams) error {
	return b.callTrue(ctx, "deleteMyCommands", params)
}

type GetMyCommandsParams struct {
	Scope        BotCommandScope `json:"scope,omitempty"`
	LanguageCode string          `json:"language_code,omitempty"`
}

func (b *Client) GetMyCommands(ctx context.Context, params GetMyCommandsParams) ([]BotCommand, error) {
	var commands []BotCommand
	if err := b.Call(ctx, "getMyCommands", params, &commands); err != nil {
		return nil, err
	}
	return commands, nil
}
