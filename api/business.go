package api

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

func validateBusinessConnectionID(connectionID, method string) error {
	if strings.TrimSpace(connectionID) == "" {
		return fmt.Errorf("hermes: %s business_connection_id is required", method)
	}
	return nil
}

type ReadBusinessMessageParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	ChatID               int64  `json:"chat_id"`
	MessageID            int    `json:"message_id"`
}

func (client *Client) ReadBusinessMessage(ctx context.Context, params ReadBusinessMessageParams) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "readBusinessMessage"); err != nil {
		return err
	}
	if params.ChatID == 0 || params.MessageID == 0 {
		return fmt.Errorf("hermes: readBusinessMessage chat_id and message_id are required")
	}
	return client.callTrue(ctx, "readBusinessMessage", params)
}

type DeleteBusinessMessagesParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	MessageIDs           []int  `json:"message_ids"`
}

func (client *Client) DeleteBusinessMessages(ctx context.Context, params DeleteBusinessMessagesParams) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "deleteBusinessMessages"); err != nil {
		return err
	}
	if len(params.MessageIDs) == 0 || len(params.MessageIDs) > 100 {
		return fmt.Errorf("hermes: deleteBusinessMessages requires 1-100 message_ids")
	}
	return client.callTrue(ctx, "deleteBusinessMessages", params)
}

type SetBusinessAccountNameParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	FirstName            string `json:"first_name"`
	LastName             string `json:"last_name,omitempty"`
}

func (client *Client) SetBusinessAccountName(ctx context.Context, params SetBusinessAccountNameParams) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "setBusinessAccountName"); err != nil {
		return err
	}
	firstLength := utf8.RuneCountInString(params.FirstName)
	if firstLength == 0 || firstLength > 64 || utf8.RuneCountInString(params.LastName) > 64 {
		return fmt.Errorf("hermes: business first_name must contain 1-64 characters and last_name at most 64")
	}
	return client.callTrue(ctx, "setBusinessAccountName", params)
}

type SetBusinessAccountUsernameParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	Username             string `json:"username,omitempty"`
}

func (client *Client) SetBusinessAccountUsername(ctx context.Context, params SetBusinessAccountUsernameParams) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "setBusinessAccountUsername"); err != nil {
		return err
	}
	if utf8.RuneCountInString(params.Username) > 32 {
		return fmt.Errorf("hermes: business username must not exceed 32 characters")
	}
	return client.callTrue(ctx, "setBusinessAccountUsername", params)
}

type SetBusinessAccountBioParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	Bio                  string `json:"bio,omitempty"`
}

func (client *Client) SetBusinessAccountBio(ctx context.Context, params SetBusinessAccountBioParams) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "setBusinessAccountBio"); err != nil {
		return err
	}
	if utf8.RuneCountInString(params.Bio) > 140 {
		return fmt.Errorf("hermes: business bio must not exceed 140 characters")
	}
	return client.callTrue(ctx, "setBusinessAccountBio", params)
}

type SetBusinessAccountProfilePhotoParams struct {
	BusinessConnectionID string            `json:"business_connection_id"`
	Photo                InputProfilePhoto `json:"photo"`
	IsPublic             bool              `json:"is_public,omitempty"`
}

func (client *Client) SetBusinessAccountProfilePhoto(ctx context.Context, params SetBusinessAccountProfilePhotoParams, uploads ...Upload) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "setBusinessAccountProfilePhoto"); err != nil {
		return err
	}
	if err := validateInputProfilePhoto(params.Photo, "setBusinessAccountProfilePhoto"); err != nil {
		return err
	}
	if err := validateAttachmentUploads(params.Photo, uploads, "setBusinessAccountProfilePhoto"); err != nil {
		return err
	}
	fields := make(formFields, 3)
	fields.String("business_connection_id", params.BusinessConnectionID)
	fields.Bool("is_public", params.IsPublic)
	if err := fields.JSON("photo", params.Photo); err != nil {
		return err
	}
	return callMultipartTrue(ctx, client, "setBusinessAccountProfilePhoto", fields, uploads)
}

type RemoveBusinessAccountProfilePhotoParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	IsPublic             bool   `json:"is_public,omitempty"`
}

func (client *Client) RemoveBusinessAccountProfilePhoto(ctx context.Context, params RemoveBusinessAccountProfilePhotoParams) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "removeBusinessAccountProfilePhoto"); err != nil {
		return err
	}
	return client.callTrue(ctx, "removeBusinessAccountProfilePhoto", params)
}

type SetBusinessAccountGiftSettingsParams struct {
	BusinessConnectionID string            `json:"business_connection_id"`
	ShowGiftButton       bool              `json:"show_gift_button"`
	AcceptedGiftTypes    AcceptedGiftTypes `json:"accepted_gift_types"`
}

func (client *Client) SetBusinessAccountGiftSettings(ctx context.Context, params SetBusinessAccountGiftSettingsParams) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "setBusinessAccountGiftSettings"); err != nil {
		return err
	}
	return client.callTrue(ctx, "setBusinessAccountGiftSettings", params)
}

type TransferBusinessAccountStarsParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
	StarCount            int    `json:"star_count"`
}

func (client *Client) TransferBusinessAccountStars(ctx context.Context, params TransferBusinessAccountStarsParams) error {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "transferBusinessAccountStars"); err != nil {
		return err
	}
	if params.StarCount < 1 || params.StarCount > 10000 {
		return fmt.Errorf("hermes: transferBusinessAccountStars star_count must be between 1 and 10000")
	}
	return client.callTrue(ctx, "transferBusinessAccountStars", params)
}

type GetBusinessAccountStarBalanceParams struct {
	BusinessConnectionID string `json:"business_connection_id"`
}

func (client *Client) GetBusinessAccountStarBalance(ctx context.Context, params GetBusinessAccountStarBalanceParams) (StarAmount, error) {
	if err := validateBusinessConnectionID(params.BusinessConnectionID, "getBusinessAccountStarBalance"); err != nil {
		return StarAmount{}, err
	}
	return Call[StarAmount](ctx, client, "getBusinessAccountStarBalance", params)
}
