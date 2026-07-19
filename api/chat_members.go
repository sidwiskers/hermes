package api

import (
	"context"
	"fmt"
	"strings"
)

type BanChatMemberParams struct {
	ChatID         any   `json:"chat_id"`
	UserID         int64 `json:"user_id"`
	UntilDate      int64 `json:"until_date,omitempty"`
	RevokeMessages bool  `json:"revoke_messages,omitempty"`
}

func (b *Client) BanChatMember(ctx context.Context, params BanChatMemberParams) error {
	if err := validateChatUser(params.ChatID, params.UserID, "banChatMember"); err != nil {
		return err
	}
	return b.callTrue(ctx, "banChatMember", params)
}

type UnbanChatMemberParams struct {
	ChatID       any   `json:"chat_id"`
	UserID       int64 `json:"user_id"`
	OnlyIfBanned bool  `json:"only_if_banned,omitempty"`
}

func (b *Client) UnbanChatMember(ctx context.Context, params UnbanChatMemberParams) error {
	if err := validateChatUser(params.ChatID, params.UserID, "unbanChatMember"); err != nil {
		return err
	}
	return b.callTrue(ctx, "unbanChatMember", params)
}

type RestrictChatMemberParams struct {
	ChatID                        any             `json:"chat_id"`
	UserID                        int64           `json:"user_id"`
	Permissions                   ChatPermissions `json:"permissions"`
	UseIndependentChatPermissions bool            `json:"use_independent_chat_permissions,omitempty"`
	UntilDate                     int64           `json:"until_date,omitempty"`
}

func (b *Client) RestrictChatMember(ctx context.Context, params RestrictChatMemberParams) error {
	if err := validateChatUser(params.ChatID, params.UserID, "restrictChatMember"); err != nil {
		return err
	}
	return b.callTrue(ctx, "restrictChatMember", params)
}

type PromoteChatMemberParams struct {
	ChatID                  any   `json:"chat_id"`
	UserID                  int64 `json:"user_id"`
	IsAnonymous             bool  `json:"is_anonymous,omitempty"`
	CanManageChat           bool  `json:"can_manage_chat,omitempty"`
	CanDeleteMessages       bool  `json:"can_delete_messages,omitempty"`
	CanManageVideoChats     bool  `json:"can_manage_video_chats,omitempty"`
	CanRestrictMembers      bool  `json:"can_restrict_members,omitempty"`
	CanPromoteMembers       bool  `json:"can_promote_members,omitempty"`
	CanChangeInfo           bool  `json:"can_change_info,omitempty"`
	CanInviteUsers          bool  `json:"can_invite_users,omitempty"`
	CanPostStories          bool  `json:"can_post_stories,omitempty"`
	CanEditStories          bool  `json:"can_edit_stories,omitempty"`
	CanDeleteStories        bool  `json:"can_delete_stories,omitempty"`
	CanPostMessages         bool  `json:"can_post_messages,omitempty"`
	CanEditMessages         bool  `json:"can_edit_messages,omitempty"`
	CanPinMessages          bool  `json:"can_pin_messages,omitempty"`
	CanManageTopics         bool  `json:"can_manage_topics,omitempty"`
	CanManageDirectMessages bool  `json:"can_manage_direct_messages,omitempty"`
	CanManageTags           bool  `json:"can_manage_tags,omitempty"`
}

func (b *Client) PromoteChatMember(ctx context.Context, params PromoteChatMemberParams) error {
	if err := validateChatUser(params.ChatID, params.UserID, "promoteChatMember"); err != nil {
		return err
	}
	return b.callTrue(ctx, "promoteChatMember", params)
}

type SetChatAdministratorCustomTitleParams struct {
	ChatID      any    `json:"chat_id"`
	UserID      int64  `json:"user_id"`
	CustomTitle string `json:"custom_title"`
}

func (b *Client) SetChatAdministratorCustomTitle(ctx context.Context, params SetChatAdministratorCustomTitleParams) error {
	if err := validateChatUser(params.ChatID, params.UserID, "setChatAdministratorCustomTitle"); err != nil {
		return err
	}
	if len([]rune(params.CustomTitle)) > 16 {
		return fmt.Errorf("hermes: administrator custom title must not exceed 16 characters")
	}
	return b.callTrue(ctx, "setChatAdministratorCustomTitle", params)
}

type SetChatMemberTagParams struct {
	ChatID any    `json:"chat_id"`
	UserID int64  `json:"user_id"`
	Tag    string `json:"tag,omitempty"`
}

func (b *Client) SetChatMemberTag(ctx context.Context, params SetChatMemberTagParams) error {
	if err := validateChatUser(params.ChatID, params.UserID, "setChatMemberTag"); err != nil {
		return err
	}
	if len([]rune(params.Tag)) > 16 {
		return fmt.Errorf("hermes: member tag must not exceed 16 characters")
	}
	return b.callTrue(ctx, "setChatMemberTag", params)
}

type ChatSenderParams struct {
	ChatID       any   `json:"chat_id"`
	SenderChatID int64 `json:"sender_chat_id"`
}

func (b *Client) BanChatSenderChat(ctx context.Context, params ChatSenderParams) error {
	if err := validateSenderChat(params, "banChatSenderChat"); err != nil {
		return err
	}
	return b.callTrue(ctx, "banChatSenderChat", params)
}

func (b *Client) UnbanChatSenderChat(ctx context.Context, params ChatSenderParams) error {
	if err := validateSenderChat(params, "unbanChatSenderChat"); err != nil {
		return err
	}
	return b.callTrue(ctx, "unbanChatSenderChat", params)
}

type SetChatPermissionsParams struct {
	ChatID                        any             `json:"chat_id"`
	Permissions                   ChatPermissions `json:"permissions"`
	UseIndependentChatPermissions bool            `json:"use_independent_chat_permissions,omitempty"`
}

func (b *Client) SetChatPermissions(ctx context.Context, params SetChatPermissionsParams) error {
	if err := validateChatID(params.ChatID, "setChatPermissions"); err != nil {
		return err
	}
	return b.callTrue(ctx, "setChatPermissions", params)
}

type GetChatAdministratorsParams struct {
	ChatID     any  `json:"chat_id"`
	ReturnBots bool `json:"return_bots,omitempty"`
}

func (b *Client) GetChatAdministrators(ctx context.Context, params GetChatAdministratorsParams) ([]ChatMember, error) {
	if err := validateChatID(params.ChatID, "getChatAdministrators"); err != nil {
		return nil, err
	}
	var members []ChatMember
	if err := b.Call(ctx, "getChatAdministrators", params, &members); err != nil {
		return nil, err
	}
	return members, nil
}

type GetChatMemberParams struct {
	ChatID any   `json:"chat_id"`
	UserID int64 `json:"user_id"`
}

func (b *Client) GetChatMember(ctx context.Context, params GetChatMemberParams) (*ChatMember, error) {
	if err := validateChatUser(params.ChatID, params.UserID, "getChatMember"); err != nil {
		return nil, err
	}
	var member ChatMember
	if err := b.Call(ctx, "getChatMember", params, &member); err != nil {
		return nil, err
	}
	return &member, nil
}

type ChatJoinRequestParams struct {
	ChatID any   `json:"chat_id"`
	UserID int64 `json:"user_id"`
}

func (b *Client) ApproveChatJoinRequest(ctx context.Context, params ChatJoinRequestParams) error {
	if err := validateChatUser(params.ChatID, params.UserID, "approveChatJoinRequest"); err != nil {
		return err
	}
	return b.callTrue(ctx, "approveChatJoinRequest", params)
}

func (b *Client) DeclineChatJoinRequest(ctx context.Context, params ChatJoinRequestParams) error {
	if err := validateChatUser(params.ChatID, params.UserID, "declineChatJoinRequest"); err != nil {
		return err
	}
	return b.callTrue(ctx, "declineChatJoinRequest", params)
}

const (
	JoinRequestApprove = "approve"
	JoinRequestDecline = "decline"
	JoinRequestQueue   = "queue"
)

type AnswerChatJoinRequestQueryParams struct {
	ChatJoinRequestQueryID string `json:"chat_join_request_query_id"`
	Result                 string `json:"result"`
}

func (b *Client) AnswerChatJoinRequestQuery(ctx context.Context, params AnswerChatJoinRequestQueryParams) error {
	if strings.TrimSpace(params.ChatJoinRequestQueryID) == "" {
		return fmt.Errorf("hermes: answerChatJoinRequestQuery query id is required")
	}
	switch params.Result {
	case JoinRequestApprove, JoinRequestDecline, JoinRequestQueue:
	default:
		return fmt.Errorf("hermes: invalid join request result %q", params.Result)
	}
	return b.callTrue(ctx, "answerChatJoinRequestQuery", params)
}

type SendChatJoinRequestWebAppParams struct {
	ChatJoinRequestQueryID string `json:"chat_join_request_query_id"`
	WebAppURL              string `json:"web_app_url"`
}

func (b *Client) SendChatJoinRequestWebApp(ctx context.Context, params SendChatJoinRequestWebAppParams) error {
	if strings.TrimSpace(params.ChatJoinRequestQueryID) == "" || strings.TrimSpace(params.WebAppURL) == "" {
		return fmt.Errorf("hermes: sendChatJoinRequestWebApp query id and URL are required")
	}
	return b.callTrue(ctx, "sendChatJoinRequestWebApp", params)
}

func validateChatUser(chatID any, userID int64, method string) error {
	if err := validateChatID(chatID, method); err != nil {
		return err
	}
	if userID == 0 {
		return fmt.Errorf("hermes: %s user_id is required", method)
	}
	return nil
}

func validateSenderChat(params ChatSenderParams, method string) error {
	if err := validateChatID(params.ChatID, method); err != nil {
		return err
	}
	if params.SenderChatID == 0 {
		return fmt.Errorf("hermes: %s sender_chat_id is required", method)
	}
	return nil
}
