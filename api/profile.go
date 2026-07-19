package api

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

type GetUserProfilePhotosParams struct {
	UserID int64 `json:"user_id"`
	Offset int   `json:"offset,omitempty"`
	Limit  int   `json:"limit,omitempty"`
}

type GetUserProfileAudiosParams = GetUserProfilePhotosParams

func validateUserProfilePage(userID int64, offset, limit int, method string) error {
	if userID == 0 {
		return fmt.Errorf("hermes: %s user_id is required", method)
	}
	if offset < 0 {
		return fmt.Errorf("hermes: %s offset must not be negative", method)
	}
	if limit < 0 || limit > 100 {
		return fmt.Errorf("hermes: %s limit must be between 1 and 100 when set", method)
	}
	return nil
}

func (client *Client) GetUserProfilePhotos(ctx context.Context, params GetUserProfilePhotosParams) (UserProfilePhotos, error) {
	if err := validateUserProfilePage(params.UserID, params.Offset, params.Limit, "getUserProfilePhotos"); err != nil {
		return UserProfilePhotos{}, err
	}
	return Call[UserProfilePhotos](ctx, client, "getUserProfilePhotos", params)
}

func (client *Client) GetUserProfileAudios(ctx context.Context, params GetUserProfileAudiosParams) (UserProfileAudios, error) {
	if err := validateUserProfilePage(params.UserID, params.Offset, params.Limit, "getUserProfileAudios"); err != nil {
		return UserProfileAudios{}, err
	}
	return Call[UserProfileAudios](ctx, client, "getUserProfileAudios", params)
}

type SetUserEmojiStatusParams struct {
	UserID                    int64  `json:"user_id"`
	EmojiStatusCustomEmojiID  string `json:"emoji_status_custom_emoji_id,omitempty"`
	EmojiStatusExpirationDate int64  `json:"emoji_status_expiration_date,omitempty"`
}

func (client *Client) SetUserEmojiStatus(ctx context.Context, params SetUserEmojiStatusParams) error {
	if params.UserID == 0 {
		return fmt.Errorf("hermes: setUserEmojiStatus user_id is required")
	}
	if params.EmojiStatusExpirationDate < 0 {
		return fmt.Errorf("hermes: setUserEmojiStatus expiration date must not be negative")
	}
	return client.callTrue(ctx, "setUserEmojiStatus", params)
}

type GetUserChatBoostsParams struct {
	ChatID any   `json:"chat_id"`
	UserID int64 `json:"user_id"`
}

func (client *Client) GetUserChatBoosts(ctx context.Context, params GetUserChatBoostsParams) (UserChatBoosts, error) {
	if err := validateChatID(params.ChatID, "getUserChatBoosts"); err != nil {
		return UserChatBoosts{}, err
	}
	if params.UserID == 0 {
		return UserChatBoosts{}, fmt.Errorf("hermes: getUserChatBoosts user_id is required")
	}
	return Call[UserChatBoosts](ctx, client, "getUserChatBoosts", params)
}

type GetBusinessConnectionParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
}

func (client *Client) GetBusinessConnection(ctx context.Context, params GetBusinessConnectionParams) (BusinessConnection, error) {
	if strings.TrimSpace(params.BusinessConnectionID) == "" {
		return BusinessConnection{}, fmt.Errorf("hermes: getBusinessConnection business_connection_id is required")
	}
	return Call[BusinessConnection](ctx, client, "getBusinessConnection", params)
}

type SetMyNameParams struct {
	Name         string `json:"name,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

func (client *Client) SetMyName(ctx context.Context, params SetMyNameParams) error {
	if utf8.RuneCountInString(params.Name) > 64 {
		return fmt.Errorf("hermes: setMyName name must not exceed 64 characters")
	}
	return client.callTrue(ctx, "setMyName", params)
}

type GetMyNameParams struct {
	LanguageCode string `json:"language_code,omitempty"`
}

func (client *Client) GetMyName(ctx context.Context, params GetMyNameParams) (BotName, error) {
	return Call[BotName](ctx, client, "getMyName", params)
}

type SetMyDescriptionParams struct {
	Description  string `json:"description,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

func (client *Client) SetMyDescription(ctx context.Context, params SetMyDescriptionParams) error {
	if utf8.RuneCountInString(params.Description) > 512 {
		return fmt.Errorf("hermes: setMyDescription description must not exceed 512 characters")
	}
	return client.callTrue(ctx, "setMyDescription", params)
}

type GetMyDescriptionParams struct {
	LanguageCode string `json:"language_code,omitempty"`
}

func (client *Client) GetMyDescription(ctx context.Context, params GetMyDescriptionParams) (BotDescription, error) {
	return Call[BotDescription](ctx, client, "getMyDescription", params)
}

type SetMyShortDescriptionParams struct {
	ShortDescription string `json:"short_description,omitempty"`
	LanguageCode     string `json:"language_code,omitempty"`
}

func (client *Client) SetMyShortDescription(ctx context.Context, params SetMyShortDescriptionParams) error {
	if utf8.RuneCountInString(params.ShortDescription) > 120 {
		return fmt.Errorf("hermes: setMyShortDescription short_description must not exceed 120 characters")
	}
	return client.callTrue(ctx, "setMyShortDescription", params)
}

type GetMyShortDescriptionParams struct {
	LanguageCode string `json:"language_code,omitempty"`
}

func (client *Client) GetMyShortDescription(ctx context.Context, params GetMyShortDescriptionParams) (BotShortDescription, error) {
	return Call[BotShortDescription](ctx, client, "getMyShortDescription", params)
}

type SetChatMenuButtonParams struct {
	ChatID     int64       `json:"chat_id,omitempty"`
	MenuButton *MenuButton `json:"menu_button,omitempty"`
}

func validateMenuButton(button *MenuButton) error {
	if button == nil {
		return nil
	}
	switch button.Type {
	case MenuButtonTypeCommands, MenuButtonTypeDefault:
		return nil
	case MenuButtonTypeWebApp:
		if strings.TrimSpace(button.Text) == "" || button.WebApp == nil || strings.TrimSpace(button.WebApp.URL) == "" {
			return fmt.Errorf("hermes: web_app menu button requires text and web_app URL")
		}
		return nil
	default:
		return fmt.Errorf("hermes: unsupported menu button type %q", button.Type)
	}
}

func (client *Client) SetChatMenuButton(ctx context.Context, params SetChatMenuButtonParams) error {
	if err := validateMenuButton(params.MenuButton); err != nil {
		return err
	}
	return client.callTrue(ctx, "setChatMenuButton", params)
}

type GetChatMenuButtonParams struct {
	ChatID int64 `json:"chat_id,omitempty"`
}

func (client *Client) GetChatMenuButton(ctx context.Context, params GetChatMenuButtonParams) (MenuButton, error) {
	return Call[MenuButton](ctx, client, "getChatMenuButton", params)
}

type SetMyDefaultAdministratorRightsParams struct {
	Rights      *ChatAdministratorRights `json:"rights,omitempty"`
	ForChannels bool                     `json:"for_channels,omitempty"`
}

func (client *Client) SetMyDefaultAdministratorRights(ctx context.Context, params SetMyDefaultAdministratorRightsParams) error {
	return client.callTrue(ctx, "setMyDefaultAdministratorRights", params)
}

type GetMyDefaultAdministratorRightsParams struct {
	ForChannels bool `json:"for_channels,omitempty"`
}

func (client *Client) GetMyDefaultAdministratorRights(ctx context.Context, params GetMyDefaultAdministratorRightsParams) (ChatAdministratorRights, error) {
	return Call[ChatAdministratorRights](ctx, client, "getMyDefaultAdministratorRights", params)
}

type VerifyUserParams struct {
	UserID            int64  `json:"user_id"`
	CustomDescription string `json:"custom_description,omitempty"`
}

func (client *Client) VerifyUser(ctx context.Context, params VerifyUserParams) error {
	if params.UserID == 0 {
		return fmt.Errorf("hermes: verifyUser user_id is required")
	}
	if utf8.RuneCountInString(params.CustomDescription) > 70 {
		return fmt.Errorf("hermes: verifyUser custom_description must not exceed 70 characters")
	}
	return client.callTrue(ctx, "verifyUser", params)
}

type VerifyChatParams struct {
	ChatID            any    `json:"chat_id"`
	CustomDescription string `json:"custom_description,omitempty"`
}

func (client *Client) VerifyChat(ctx context.Context, params VerifyChatParams) error {
	if err := validateChatID(params.ChatID, "verifyChat"); err != nil {
		return err
	}
	if utf8.RuneCountInString(params.CustomDescription) > 70 {
		return fmt.Errorf("hermes: verifyChat custom_description must not exceed 70 characters")
	}
	return client.callTrue(ctx, "verifyChat", params)
}

type RemoveUserVerificationParams struct {
	UserID int64 `json:"user_id"`
}

func (client *Client) RemoveUserVerification(ctx context.Context, params RemoveUserVerificationParams) error {
	if params.UserID == 0 {
		return fmt.Errorf("hermes: removeUserVerification user_id is required")
	}
	return client.callTrue(ctx, "removeUserVerification", params)
}

type RemoveChatVerificationParams struct {
	ChatID any `json:"chat_id"`
}

func (client *Client) RemoveChatVerification(ctx context.Context, params RemoveChatVerificationParams) error {
	if err := validateChatID(params.ChatID, "removeChatVerification"); err != nil {
		return err
	}
	return client.callTrue(ctx, "removeChatVerification", params)
}
